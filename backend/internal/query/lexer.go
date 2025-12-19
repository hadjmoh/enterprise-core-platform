package query

import (
	"fmt"
	"unicode"
)

type TokenType string

const (
	TokenIdentifier TokenType = "IDENTIFIER"
	TokenString     TokenType = "STRING"
	TokenEqual      TokenType = "EQUAL"
	TokenPipe       TokenType = "PIPE"
	TokenPlus       TokenType = "PLUS"
	TokenMinus      TokenType = "MINUS"
	TokenStar       TokenType = "STAR"
	TokenSlash      TokenType = "SLASH"
	TokenLPAREN     TokenType = "LPAREN"
	TokenRPAREN     TokenType = "RPAREN"
	TokenEOF        TokenType = "EOF"
	TokenError      TokenType = "ERROR"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '|':
		tok.Type = TokenPipe
		tok.Literal = "|"
	case '=':
		tok.Type = TokenEqual
		tok.Literal = "="
	case '#':
		l.skipComment()
		return l.NextToken()
	case '+':
		tok.Type = TokenPlus
		tok.Literal = "+"
	case '-':
		tok.Type = TokenMinus
		tok.Literal = "-"
	case '*':
		tok.Type = TokenStar
		tok.Literal = "*"
	case '/':
		tok.Type = TokenSlash
		tok.Literal = "/"
	case '(':
		tok.Type = TokenLPAREN
		tok.Literal = "("
	case ')':
		tok.Type = TokenRPAREN
		tok.Literal = ")"
	case '"':
		tok.Type = TokenString
		tok.Literal = l.readString()
		return tok // readString already calls readChar as needed
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
	default:
		if isLetter(l.ch) || unicode.IsDigit(rune(l.ch)) {
			tok.Literal = l.readIdentifier()
			tok.Type = TokenIdentifier
			return tok
		} else {
			tok.Type = TokenError
			tok.Literal = fmt.Sprintf("illegal character: %c", l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.skipWhitespace()
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || unicode.IsDigit(rune(l.ch)) || l.ch == '_' || l.ch == '.' {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() string {
	l.readChar() // skip "
	var out []byte
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // skip \
			switch l.ch {
			case 'n':
				out = append(out, '\n')
			case 't':
				out = append(out, '\t')
			case '"':
				out = append(out, '"')
			case '\\':
				out = append(out, '\\')
			default:
				out = append(out, l.ch)
			}
		} else {
			out = append(out, l.ch)
		}
		l.readChar()
	}
	l.readChar() // skip closing "
	return string(out)
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}
