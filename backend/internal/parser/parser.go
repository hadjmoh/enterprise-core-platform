package parser

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// Parser interface for all data parsers
type Parser interface {
	Parse(data []byte) (map[string]interface{}, error)
	Type() string
}

// JSONParser handles JSON data
type JSONParser struct{}

func (j *JSONParser) Parse(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (j *JSONParser) Type() string {
	return "JSON"
}

// XMLParser handles XML data
type XMLParser struct{}

func (x *XMLParser) Parse(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (x *XMLParser) Type() string {
	return "XML"
}

// CSVParser handles CSV data
type CSVParser struct {
	Headers []string
}

func (c *CSVParser) Parse(data []byte) (map[string]interface{}, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV")
	}

	// Use first row as headers if not provided
	if c.Headers == nil && len(records) > 0 {
		c.Headers = records[0]
		records = records[1:]
	}

	result := make(map[string]interface{})
	result["rows"] = records
	result["headers"] = c.Headers
	return result, nil
}

func (c *CSVParser) Type() string {
	return "CSV"
}

// RegexParser extracts fields using regex patterns
type RegexParser struct {
	Patterns map[string]*regexp.Regexp
}

func NewRegexParser(patterns map[string]string) (*RegexParser, error) {
	compiled := make(map[string]*regexp.Regexp)
	for name, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex for %s: %w", name, err)
		}
		compiled[name] = re
	}
	return &RegexParser{Patterns: compiled}, nil
}

func (r *RegexParser) Parse(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	text := string(data)

	for name, pattern := range r.Patterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			result[name] = matches[1] // First capture group
		} else if len(matches) == 1 {
			result[name] = matches[0] // Full match
		}
	}

	return result, nil
}

func (r *RegexParser) Type() string {
	return "Regex"
}
