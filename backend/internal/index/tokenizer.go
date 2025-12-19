package index

import (
	"regexp"
	"strings"
)

// Tokenizer breaks down text into searchable tokens
type Tokenizer struct {
	Version   int
	stopWords map[string]struct{}
	reg       *regexp.Regexp
}

func NewTokenizer() *Tokenizer {
	return NewTokenizerWithVersion(1)
}

func NewTokenizerWithVersion(version int) *Tokenizer {
	// Common stop words to ignore
	words := []string{"a", "an", "and", "are", "as", "at", "be", "but", "by", "for", "if", "in", "into", "is", "it", "no", "not", "of", "on", "or", "such", "that", "the", "their", "then", "there", "these", "they", "this", "to", "was", "will", "with"}
	stopWords := make(map[string]struct{})
	for _, w := range words {
		stopWords[w] = struct{}{}
	}

	// Delimiters (non-alphanumeric characters)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)

	return &Tokenizer{
		Version:   version,
		stopWords: stopWords,
		reg:       reg,
	}
}

func (t *Tokenizer) Tokenize(text string) []string {
	// Convert to lowercase and split by delimiters
	text = strings.ToLower(text)
	parts := t.reg.Split(text, -1)

	var tokens []string
	seen := make(map[string]struct{})

	for _, p := range parts {
		if len(p) < 2 {
			continue
		}
		if _, ok := t.stopWords[p]; ok {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		tokens = append(tokens, p)
		seen[p] = struct{}{}
	}

	return tokens
}
