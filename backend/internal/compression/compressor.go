package compression

import (
	"bytes"
	"compress/gzip"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"io"

	"github.com/klauspost/compress/snappy"
	"github.com/pierrec/lz4/v4"
)

type Algorithm string

const (
	AlgorithmGzip   Algorithm = "gzip"
	AlgorithmSnappy Algorithm = "snappy"
	AlgorithmLZ4    Algorithm = "lz4"
	AlgorithmNone   Algorithm = "none"
)

// Compressor handles data compression
type Compressor struct {
	algorithm Algorithm
	logger    *logger.Logger
}

func NewCompressor(algorithm Algorithm, logger *logger.Logger) *Compressor {
	return &Compressor{
		algorithm: algorithm,
		logger:    logger,
	}
}

func (c *Compressor) Compress(data []byte) ([]byte, error) {
	switch c.algorithm {
	case AlgorithmGzip:
		return c.compressGzip(data)
	case AlgorithmSnappy:
		return c.compressSnappy(data)
	case AlgorithmLZ4:
		return c.compressLZ4(data)
	case AlgorithmNone:
		return data, nil
	default:
		return nil, fmt.Errorf("unknown compression algorithm: %s", c.algorithm)
	}
}

func (c *Compressor) Decompress(data []byte) ([]byte, error) {
	switch c.algorithm {
	case AlgorithmGzip:
		return c.decompressGzip(data)
	case AlgorithmSnappy:
		return c.decompressSnappy(data)
	case AlgorithmLZ4:
		return c.decompressLZ4(data)
	case AlgorithmNone:
		return data, nil
	default:
		return nil, fmt.Errorf("unknown compression algorithm: %s", c.algorithm)
	}
}

func (c *Compressor) compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	
	if err := writer.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (c *Compressor) decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	return io.ReadAll(reader)
}

func (c *Compressor) compressSnappy(data []byte) ([]byte, error) {
	return snappy.Encode(nil, data), nil
}

func (c *Compressor) decompressSnappy(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}

func (c *Compressor) compressLZ4(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := lz4.NewWriter(&buf)
	
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	
	if err := writer.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (c *Compressor) decompressLZ4(data []byte) ([]byte, error) {
	reader := lz4.NewReader(bytes.NewReader(data))
	return io.ReadAll(reader)
}

// GetCompressionRatio returns the compression ratio
func (c *Compressor) GetCompressionRatio(original, compressed []byte) float64 {
	if len(original) == 0 {
		return 0
	}
	return float64(len(compressed)) / float64(len(original))
}

// SelectBestAlgorithm chooses the best compression algorithm for the data
func SelectBestAlgorithm(data []byte) Algorithm {
	// For small data, compression overhead isn't worth it
	if len(data) < 1024 {
		return AlgorithmNone
	}

	// For medium data, use Snappy (fast)
	if len(data) < 100*1024 {
		return AlgorithmSnappy
	}

	// For large data, use Gzip (better compression)
	return AlgorithmGzip
}
