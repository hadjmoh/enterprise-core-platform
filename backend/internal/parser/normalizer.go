package parser

import (
	"time"
)

// Normalizer standardizes ingested data
type Normalizer struct {
	TimestampFormats []string
	FieldMappings    map[string]string // source field -> normalized field
}

func NewNormalizer() *Normalizer {
	return &Normalizer{
		TimestampFormats: []string{
			time.RFC3339,
			time.RFC1123,
			"2006-01-02 15:04:05",
			"2006/01/02 15:04:05",
			"Jan 02 15:04:05",
		},
		FieldMappings: make(map[string]string),
	}
}

func (n *Normalizer) Normalize(data map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})

	// Extract and normalize timestamp
	if ts := n.extractTimestamp(data); ts != nil {
		normalized["@timestamp"] = ts.Format(time.RFC3339)
	} else {
		normalized["@timestamp"] = time.Now().Format(time.RFC3339)
	}

	// Apply field mappings
	for srcField, dstField := range n.FieldMappings {
		if val, ok := data[srcField]; ok {
			normalized[dstField] = val
		}
	}

	// Copy remaining fields
	for k, v := range data {
		if _, exists := normalized[k]; !exists {
			normalized[k] = v
		}
	}

	return normalized
}

func (n *Normalizer) extractTimestamp(data map[string]interface{}) *time.Time {
	// Try common timestamp fields
	timestampFields := []string{"timestamp", "time", "@timestamp", "ts", "date"}

	for _, field := range timestampFields {
		if val, ok := data[field]; ok {
			if strVal, ok := val.(string); ok {
				for _, format := range n.TimestampFormats {
					if t, err := time.Parse(format, strVal); err == nil {
						return &t
					}
				}
			}
		}
	}

	return nil
}
