package query

import (
	"fmt"
)

type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string
	input     string
	depth     int
}

const (
	MaxPipelineDepth = 20
	MaxQueryLength   = 4096
	CurrentSPLVersion = 1
)

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l, input: l.input}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Parse() (*Pipeline, error) {
	if len(p.input) > MaxQueryLength {
		return nil, fmt.Errorf("query too long: %d > %d", len(p.input), MaxQueryLength)
	}

	pipeline := &Pipeline{
		Source: p.input,
		Meta: NodeMetadata{
			Line:    1,
			Column:  1,
			Version: CurrentSPLVersion,
		},
	}

	for p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenPipe {
			p.nextToken()
			continue
		}

		p.depth++
		if p.depth > MaxPipelineDepth {
			return nil, fmt.Errorf("pipeline depth limit exceeded: %d", MaxPipelineDepth)
		}

		var cmdName string
		if p.curToken.Type == TokenIdentifier {
			if isCommand(p.curToken.Literal) {
				cmdName = p.curToken.Literal
			} else {
				cmdName = "search"
				cmd, err := p.parseCommand(cmdName, true) // Pass true to not consume the token as name
				if err != nil {
					return nil, err
				}
				pipeline.Commands = append(pipeline.Commands, *cmd)
				continue
			}
		} else {
			cmdName = "search"
			cmd, err := p.parseCommand(cmdName, true)
			if err != nil {
				return nil, err
			}
			pipeline.Commands = append(pipeline.Commands, *cmd)
			continue
		}

		cmd, err := p.parseCommand(cmdName, false)
		if err != nil {
			return nil, err
		}
		pipeline.Commands = append(pipeline.Commands, *cmd)
	}

	return pipeline, nil
}

func isCommand(name string) bool {
	commands := map[string]bool{
		"search":    true,
		"where":     true,
		"limit":     true,
		"stats":     true,
		"fields":    true,
		"eval":      true,
		"sort":      true,
		"bucket":    true,
		"timechart": true,
	}
	return commands[name]
}

func (p *Parser) parseCommand(name string, implicit bool) (*CommandNode, error) {
	cmd := &CommandNode{
		Name: name,
		Meta: NodeMetadata{
			Line:    p.curToken.Line,
			Column:  p.curToken.Column,
			Version: CurrentSPLVersion,
		},
	}
	
	if !implicit {
		p.nextToken() // consume command name
	}

	for p.curToken.Type != TokenPipe && p.curToken.Type != TokenEOF {
		arg := Argument{}
		
		if p.curToken.Type == TokenIdentifier {
			val := p.curToken.Literal
			if p.peekToken.Type == TokenEqual {
				// key=value pair
				arg.Key = val
				p.nextToken() // move to =
				p.nextToken() // move to value
				if p.curToken.Type == TokenIdentifier || p.curToken.Type == TokenString {
					arg.Value = p.curToken.Literal
					arg.IsQuoted = (p.curToken.Type == TokenString)
				} else {
					return nil, fmt.Errorf("line %d, col %d: expected value after '=', got %s", p.curToken.Line, p.curToken.Column, p.curToken.Type)
				}
			} else {
				// standalone identifier (keyword)
				arg.Value = val
			}
		} else if p.curToken.Type == TokenString {
			arg.Value = p.curToken.Literal
			arg.IsQuoted = true
		} else if p.curToken.Type == TokenError {
			return nil, fmt.Errorf("line %d, col %d: %s", p.curToken.Line, p.curToken.Column, p.curToken.Literal)
		} else {
			return nil, fmt.Errorf("line %d, col %d: unexpected token in command arguments: %s", p.curToken.Line, p.curToken.Column, p.curToken.Literal)
		}

		// Look ahead to merge operators and parentheses into the current argument value
		for {
			switch p.peekToken.Type {
			case TokenPlus, TokenMinus, TokenStar, TokenSlash, TokenLPAREN, TokenRPAREN:
				p.nextToken() // consume operator
				arg.Value += p.curToken.Literal
				if p.peekToken.Type == TokenIdentifier || p.peekToken.Type == TokenString {
					p.nextToken() // consume next value
					arg.Value += p.curToken.Literal
				}
				continue
			}
			break
		}
		
		cmd.Args = append(cmd.Args, arg)
		p.nextToken()
	}

	return cmd, nil
}
