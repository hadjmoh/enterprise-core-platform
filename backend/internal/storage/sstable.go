package storage

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/encryption"
	"hash/crc32"
	"io"
	"os"
)

// SSTableFooter contains metadata about an SSTable file, stored at the end of the file.
type SSTableFooter struct {
	MinTimestamp       string          `json:"min_timestamp"`
	MaxTimestamp       string          `max_timestamp"`
	EventCount         int64           `json:"event_count"`
	Version            int             `json:"version"`
	CompressionType    CompressionType `json:"compression_type"`
	CompressionVersion int             `json:"compression_version"`
	BlockSize          int             `json:"block_size"`
	NumBlocks          int             `json:"num_blocks"`
	RawSize            int64           `json:"raw_size"`
	CompressedSize     int64           `json:"compressed_size"`
	BlockOffsets       []int64         `json:"block_offsets"`
	EncryptedKey       []byte          `json:"encrypted_key,omitempty"` // Data key protected by Master Key
}

// SSTableWriter writes sorted events to an immutable file.
type SSTableWriter struct {
	file         *os.File
	minTimestamp string
	maxTimestamp string
	count        int64
	blockBuffer  *bytes.Buffer
	compression         CompressionProvider
	kms                 *encryption.KeyManager
	cipher              cipher.AEAD
	encryptedDataKey    []byte
	maxBlockSize        int
	blockOffsets        []int64
	totalRawSize        int64
	totalCompressedSize int64
}

func NewSSTableWriterWithEncryption(path string, kms *encryption.KeyManager) (*SSTableWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}

	dataKey, encryptedKey, err := kms.GenerateDataKey()
	if err != nil {
		return nil, err
	}

	c, err := kms.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}

	return &SSTableWriter{
		file:             f,
		blockBuffer:      new(bytes.Buffer),
		compression:      &ZlibProvider{},
		kms:              kms,
		cipher:           c,
		encryptedDataKey: encryptedKey,
		maxBlockSize:     64 * 1024,
		blockOffsets:     make([]int64, 0),
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

	// Record start offset
	offset, _ := w.file.Seek(0, io.SeekCurrent)
	w.blockOffsets = append(w.blockOffsets, offset)

	// 1. Compress
	compressed, err := w.compression.Compress(w.blockBuffer.Bytes())
	if err != nil {
		return err
	}

	// 2. Encrypt if cipher is available
	var finalBlocks []byte
	if w.cipher != nil {
		nonce := make([]byte, w.cipher.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		encrypted := w.cipher.Seal(nil, nonce, compressed, nil)
		finalBlocks = append(nonce, encrypted...)
	} else {
		finalBlocks = compressed
	}
	
	w.totalCompressedSize += int64(len(finalBlocks) + 8)

	// Write block size
	if err := binary.Write(w.file, binary.LittleEndian, int32(len(finalBlocks))); err != nil {
		return err
	}
	
	// Write CRC
	checksum := crc32.ChecksumIEEE(finalBlocks)
	if err := binary.Write(w.file, binary.LittleEndian, checksum); err != nil {
		return err
	}

	// Write data
	if _, err := w.file.Write(finalBlocks); err != nil {
		return err
	}

	w.blockBuffer.Reset()
	return nil
}

func (w *SSTableWriter) Close() error {
	if err := w.flushBlock(); err != nil {
		return err
	}

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
		EncryptedKey:       w.encryptedDataKey,
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
