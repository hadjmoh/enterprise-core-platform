package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// SSTableIterator reads events from a single SSTable file.
type SSTableIterator struct {
	file        *os.File
	footer      SSTableFooter
	compression CompressionProvider
	
	blockReader  *bufio.Scanner
	currentBlock int
	currentPos   int64
	dataEnd      int64
}

func NewSSTableIterator(path string) (*SSTableIterator, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// 1. Read footer size (last 8 bytes)
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if fi.Size() < 8 {
		f.Close()
		return nil, io.ErrUnexpectedEOF
	}

	f.Seek(-8, io.SeekEnd)
	var footerSize int64
	if err := binary.Read(f, binary.LittleEndian, &footerSize); err != nil {
		f.Close()
		return nil, err
	}

	// 2. Read footer
	f.Seek(-(footerSize + 8), io.SeekEnd)
	footerData := make([]byte, footerSize)
	if _, err := io.ReadFull(f, footerData); err != nil {
		f.Close()
		return nil, err
	}
	var footer SSTableFooter
	if err := json.Unmarshal(footerData, &footer); err != nil {
		f.Close()
		return nil, err
	}

	// 3. Setup for block reading
	var provider CompressionProvider
	switch footer.CompressionType {
	case CompressionZlib:
		provider = &ZlibProvider{}
	default:
		provider = &NoneProvider{}
	}

	return &SSTableIterator{
		file:         f,
		footer:       footer,
		compression:  provider,
		currentBlock: 0,
		currentPos:   0,
		dataEnd:      fi.Size() - (footerSize + 8),
	}, nil
}

// SeekToBlock positions the iterator at the start of a specific block.
func (it *SSTableIterator) SeekToBlock(blockIdx int) error {
	if blockIdx < 0 || blockIdx >= len(it.footer.BlockOffsets) {
		return io.ErrUnexpectedEOF
	}
	it.currentBlock = blockIdx
	it.currentPos = it.footer.BlockOffsets[blockIdx]
	it.blockReader = nil
	return nil
}

func (it *SSTableIterator) Next(ctx context.Context) (buffer.Event, error) {
	for {
		if it.blockReader != nil && it.blockReader.Scan() {
			line := it.blockReader.Bytes()
			var event buffer.Event
			if err := json.Unmarshal(line, &event); err != nil {
				continue // Skip malformed
			}
			return event, nil
		}

		if it.blockReader != nil && it.blockReader.Err() != nil {
			return buffer.Event{}, it.blockReader.Err()
		}

		// Need to read next block
		if it.currentBlock >= len(it.footer.BlockOffsets) {
			return buffer.Event{}, io.EOF
		}

		// Optimization: if SeekToBlock wasn't called, currentPos is already correct.
		// If it was, SeekToBlock updated currentPos.
		it.file.Seek(it.currentPos, io.SeekStart)
		var blockSize int32
		if err := binary.Read(it.file, binary.LittleEndian, &blockSize); err != nil {
			return buffer.Event{}, err
		}

		// Integrity check (CRC32)
		var checksum uint32
		if err := binary.Read(it.file, binary.LittleEndian, &checksum); err != nil {
			return buffer.Event{}, err
		}
		
		compressed := make([]byte, blockSize)
		if _, err := io.ReadFull(it.file, compressed); err != nil {
			return buffer.Event{}, err
		}

		if crc32.ChecksumIEEE(compressed) != checksum {
			return buffer.Event{}, fmt.Errorf("block checksum mismatch")
		}
		
		decompressed, err := it.compression.Decompress(compressed)
		if err != nil {
			return buffer.Event{}, err
		}

		it.blockReader = bufio.NewScanner(bytes.NewReader(decompressed))
		it.currentPos += 4 + int64(blockSize)
		it.currentBlock++
	}
}

func (it *SSTableIterator) Close() error {
	return it.file.Close()
}

type MergeIterator struct {
	iters []StorageReader
	nexts []buffer.Event
	valid []bool
}

func NewMergeIterator(iters []StorageReader) *MergeIterator {
	return &MergeIterator{
		iters: iters,
		nexts: make([]buffer.Event, len(iters)),
		valid: make([]bool, len(iters)),
	}
}

func (m *MergeIterator) Next(ctx context.Context) (buffer.Event, error) {
	var earliest buffer.Event
	var selectedIdx = -1

	for i, it := range m.iters {
		if !m.valid[i] {
			ev, err := it.Next(ctx)
			if err == nil {
				m.nexts[i] = ev
				m.valid[i] = true
			} else {
				continue
			}
		}

		if selectedIdx == -1 || m.nexts[i].Timestamp < earliest.Timestamp {
			earliest = m.nexts[i]
			selectedIdx = i
		}
	}
	
	if selectedIdx == -1 {
		return buffer.Event{}, io.EOF
	}
	
	m.valid[selectedIdx] = false
	return earliest, nil
}

func (m *MergeIterator) Close() error {
	for _, it := range m.iters {
		it.Close()
	}
	return nil
}
