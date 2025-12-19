package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"hash/crc32"
	"io"
	"os"
)

// SSTableFooter contains metadata about an SSTable file, stored at the end of the file.
type SSTableFooter struct {
	MinTimestamp       string          `json:"min_timestamp"`
	MaxTimestamp       string          `json:"max_timestamp"`
	EventCount         int64           `json:"event_count"`
	Version            int             `json:"version"`
	CompressionType    CompressionType `json:"compression_type"`
	CompressionVersion int             `json:"compression_version"`
	BlockSize          int             `json:"block_size"`
	NumBlocks          int             `json:"num_blocks"`
	RawSize            int64           `json:"raw_size"`
	CompressedSize     int64           `json:"compressed_size"`
	BlockOffsets       []int64         `json:"block_offsets"`
}

// SSTableWriter writes sorted events to an immutable file.
type SSTableWriter struct {
	file         *os.File
	minTimestamp string
	maxTimestamp string
	count        int64
	blockBuffer  *bytes.Buffer
	compression        CompressionProvider
	maxBlockSize       int
	blockOffsets       []int64
	totalRawSize       int64
	totalCompressedSize int64
}

func NewSSTableWriter(path string) (*SSTableWriter, error) {
	return NewSSTableWriterWithCompression(path, &ZlibProvider{})
}

func NewSSTableWriterWithCompression(path string, provider CompressionProvider) (*SSTableWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &SSTableWriter{
		file:          f,
		blockBuffer:   new(bytes.Buffer),
		compression:   provider,
		maxBlockSize:  64 * 1024, // 64KB
		blockOffsets:  make([]int64, 0),
	}, nil
}

func (w *SSTableWriter) WriteBatch(events []buffer.Event) error {
	for _, event := range events {
		if w.minTimestamp == "" || event.Timestamp < w.minTimestamp {
			w.minTimestamp = event.Timestamp
		}
		if w.maxTimestamp == "" || event.Timestamp > w.maxTimestamp {
			w.maxTimestamp = event.Timestamp
		}
		w.count++

		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		w.totalRawSize += int64(len(data) + 1)
		w.blockBuffer.Write(data)
		w.blockBuffer.WriteByte('\n')

		if w.blockBuffer.Len() >= w.maxBlockSize {
			if err := w.flushBlock(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *SSTableWriter) flushBlock() error {
	if w.blockBuffer.Len() == 0 {
		return nil
	}

	// Record start offset of this block
	offset, _ := w.file.Seek(0, io.SeekCurrent)
	w.blockOffsets = append(w.blockOffsets, offset)

	compressed, err := w.compression.Compress(w.blockBuffer.Bytes())
	if err != nil {
		return err
	}
	w.totalCompressedSize += int64(len(compressed) + 8) // +4 for size, +4 for crc

	// Write block size
	if err := binary.Write(w.file, binary.LittleEndian, int32(len(compressed))); err != nil {
		return err
	}
	
	// Write CRC32 checksum
	checksum := crc32.ChecksumIEEE(compressed)
	if err := binary.Write(w.file, binary.LittleEndian, checksum); err != nil {
		return err
	}

	// Write compressed data
	if _, err := w.file.Write(compressed); err != nil {
		return err
	}

	w.blockBuffer.Reset()
	return nil
}

func (w *SSTableWriter) Close() error {
	if err := w.flushBlock(); err != nil {
		return err
	}

	// Write footer before closing
	footer := SSTableFooter{
		MinTimestamp:       w.minTimestamp,
		MaxTimestamp:       w.maxTimestamp,
		EventCount:         w.count,
		Version:            1,
		CompressionType:    w.compression.Type(),
		CompressionVersion: 1,
		BlockSize:          w.maxBlockSize,
		NumBlocks:          len(w.blockOffsets),
		RawSize:            w.totalRawSize,
		CompressedSize:     w.totalCompressedSize,
		BlockOffsets:       w.blockOffsets,
	}
	
	footerData, _ := json.Marshal(footer)
	if _, err := w.file.Write(append(footerData, '\n')); err != nil {
		return err
	}
	
	footerSize := int64(len(footerData) + 1)
	binary.Write(w.file, binary.LittleEndian, footerSize)

	if err := w.file.Sync(); err != nil {
		return err
	}
	return w.file.Close()
}
