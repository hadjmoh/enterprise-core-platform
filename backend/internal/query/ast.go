package query

import (
	"time"
)

// Node represents a component in the SPL AST
type Node interface {
	NodeType() string
	GetVersion() int
}

// NodeMetadata stores position, versioning and analytics info
type NodeMetadata struct {
	Line    int
	Column  int
	Version int
	Type    string        // timechart, stats, etc.
	Span    time.Duration // for time-series
}

// Pipeline represents a sequence of commands connected by pipes
type Pipeline struct {
	Commands []CommandNode
	Source   string // Original query string
	Meta     NodeMetadata
}

func (p *Pipeline) NodeType() string { return "Pipeline" }
func (p *Pipeline) GetVersion() int  { return p.Meta.Version }

// CommandNode represents a single SPL command (e.g., search, where, limit)
type CommandNode struct {
	Name string
	Args []Argument
	Meta NodeMetadata
}

func (c *CommandNode) NodeType() string { return "Command" }
func (c *CommandNode) GetVersion() int  { return c.Meta.Version }

// Argument represents a value or key-value pair passed to a command
type Argument struct {
	Key   string // Optional: for key=value
	Value string
	IsQuoted bool
}
