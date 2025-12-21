package storage

import (
	"bufio"
	"bytes"
	"context"
	"crypto/cipher"
	"encoding/binary"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/encryption"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// SSTableIterator reads events from a single SSTable file.
type SSTableIterator struct {
	file         *os.File
	footer       SSTableFooter
	compression  CompressionProvider
	kms          *encryption.KeyManager
	cipher       cipher.AEAD
	
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

func NewSSTableIteratorWithEncryption(path string, kms *encryption.KeyManager) (*SSTableIterator, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	f.Seek(-8, io.SeekEnd)
	var footerSize int64
	binary.Read(f, binary.LittleEndian, &footerSize)

	f.Seek(-(footerSize + 8), io.SeekEnd)
	footerData := make([]byte, footerSize)
	io.ReadFull(f, footerData)

	var footer SSTableFooter
	json.Unmarshal(footerData, &footer)

	// Decrypt Data Key if present
	var c cipher.AEAD
	if len(footer.EncryptedKey) > 0 {
		dataKey, err := kms.DecryptDataKey(footer.EncryptedKey)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to decrypt sstable data key: %w", err)
		}
		c, err = kms.NewCipher(dataKey)
		if err != nil {
			f.Close()
			return nil, err
		}
	}

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
		kms:          kms,
		cipher:       c,
		currentBlock: 0,
		currentPos:   0,
		dataEnd:      fi.Size() - (footerSize + 8),
	}, nil
}

func (it *SSTableIterator) Next(ctx context.Context) (buffer.Event, error) {
	for {
		if it.blockReader != nil && it.blockReader.Scan() {
			line := it.blockReader.Bytes()
			var event buffer.Event
			if err := json.Unmarshal(line, &event); err != nil {
				continue
			}
			return event, nil
		}

		if it.currentBlock >= len(it.footer.BlockOffsets) {
			return buffer.Event{}, io.EOF
		}

		it.file.Seek(it.currentPos, io.SeekStart)
		var blockSize int32
		binary.Read(it.file, binary.LittleEndian, &blockSize)

		var checksum uint32
		binary.Read(it.file, binary.LittleEndian, &checksum)
		
		data := make([]byte, blockSize)
		io.ReadFull(it.file, data)

		if crc32.ChecksumIEEE(data) != checksum {
			return buffer.Event{}, fmt.Errorf("block checksum mismatch")
		}
		
		var decompressed []byte
		var err error

		// Decrypt if needed
		if it.cipher != nil {
			nonceSize := it.cipher.NonceSize()
			if len(data) < nonceSize {
				return buffer.Event{}, fmt.Errorf("block too short for nonce")
			}
			nonce, ciphertext := data[:nonceSize], data[nonceSize:]
			decrypted, err := it.cipher.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				return buffer.Event{}, fmt.Errorf("block decryption failed: %w", err)
			}
			decompressed, err = it.compression.Decompress(decrypted)
		} else {
			decompressed, err = it.compression.Decompress(data)
		}

		if err != nil {
			return buffer.Event{}, err
		}

		it.blockReader = bufio.NewScanner(bytes.NewReader(decompressed))
		it.currentPos += 4 + 4 + int64(blockSize)
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
