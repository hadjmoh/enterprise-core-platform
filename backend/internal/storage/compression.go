package storage

import (
	"bytes"
	"compress/zlib"
	"io"
)

// CompressionType defines the algorithm used for block compression
type CompressionType string

const (
	CompressionNone CompressionType = "none"
	CompressionZlib CompressionType = "zlib"
)

// CompressionProvider handles data compression and decompression for blocks
type CompressionProvider interface {
	Compress(data []byte) ([]byte, error)
	Decompress(compressed []byte) ([]byte, error)
	Type() CompressionType
}

// ZlibProvider implements CompressionProvider using zlib
type ZlibProvider struct{}

func (p *ZlibProvider) Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

func (p *ZlibProvider) Decompress(compressed []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (p *ZlibProvider) Type() CompressionType {
	return CompressionZlib
}

// NoneProvider implements CompressionProvider with no compression
type NoneProvider struct{}

func (p *NoneProvider) Compress(data []byte) ([]byte, error)   { return data, nil }
func (p *NoneProvider) Decompress(compressed []byte) ([]byte, error) { return compressed, nil }
func (p *NoneProvider) Type() CompressionType                 { return CompressionNone }
