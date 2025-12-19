package query

import (
	"testing"
)

func TestParser_Basic(t *testing.T) {
	input := `error | limit 10`
	l := NewLexer(input)
	p := NewParser(l)

	pipeline, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(pipeline.Commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(pipeline.Commands))
	}

	if pipeline.Commands[0].Name != "search" {
		t.Errorf("Expected first command to be 'search', got %s", pipeline.Commands[0].Name)
	}
	
	if pipeline.Commands[0].Meta.Column != 1 {
		t.Errorf("Expected col 1, got %d", pipeline.Commands[0].Meta.Column)
	}

	if pipeline.Commands[1].Name != "limit" {
		t.Errorf("Expected second command to be 'limit', got %s", pipeline.Commands[1].Name)
	}
}

func TestParser_CommentsAndEscapes(t *testing.T) {
	input := `# this is a comment
	search "quoted \"with\" escape" | limit 5`
	l := NewLexer(input)
	p := NewParser(l)

	pipeline, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(pipeline.Commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(pipeline.Commands))
	}
	
	msg := pipeline.Commands[0].Args[0].Value
	expected := `quoted "with" escape`
	if msg != expected {
		t.Errorf("Expected: %s, Got: %s", expected, msg)
	}
	
	if pipeline.Commands[0].Meta.Line != 2 {
		t.Errorf("Expected line 2 for search command, got %d", pipeline.Commands[0].Meta.Line)
	}
}

func TestParser_Validation(t *testing.T) {
	// 1. Max length
	longInput := ""
	for i := 0; i < 5000; i++ {
		longInput += "a"
	}
	l1 := NewLexer(longInput)
	p1 := NewParser(l1)
	_, err := p1.Parse()
	if err == nil {
		t.Error("Expected error for too long query, got nil")
	}

	// 2. Syntax Error with Position
	input := `search service=`
	l2 := NewLexer(input)
	p2 := NewParser(l2)
	_, err = p2.Parse()
	if err == nil {
		t.Error("Expected syntax error, got nil")
	} else {
		t.Logf("Caught expected error: %v", err)
	}
}
