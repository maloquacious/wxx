// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package dsl

import (
	"fmt"
	"strings"
	"unicode"
)

// TODO
// Add AST node types (if desired) or interpret directly.
//
// Extend parseExpr() to support operators, precedence, and nested expressions.
//
// Add support for ., [expr], and complex lvalue parsing.
//
// Add error reporting with position info and recovery if needed.

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, line: 1}
}

func (lx *Lexer) peek() rune {
	if lx.pos >= len(lx.input) {
		return 0
	}
	return rune(lx.input[lx.pos])
}

func (lx *Lexer) advance() rune {
	if lx.pos >= len(lx.input) {
		return 0
	}
	r := rune(lx.input[lx.pos])
	lx.pos++
	if r == '\n' {
		lx.line++
		lx.column = 0
	} else {
		lx.column++
	}
	return r
}

func (lx *Lexer) skipWhitespace() {
	for unicode.IsSpace(lx.peek()) {
		lx.advance()
	}
}

func (lx *Lexer) NextToken() Token {
	lx.skipWhitespace()

	startLine := lx.line
	startCol := lx.column
	ch := lx.peek()

	if ch == 0 {
		return Token{Type: TokenEOF, Line: startLine, Column: startCol}
	}

	switch ch {
	case ';':
		lx.advance()
		return Token{Type: TokenSemicolon, Value: ";", Line: startLine, Column: startCol}
	case '.':
		lx.advance()
		return Token{Type: TokenDot, Value: ".", Line: startLine, Column: startCol}
	case ',':
		lx.advance()
		return Token{Type: TokenComma, Value: ",", Line: startLine, Column: startCol}
	case '(':
		lx.advance()
		return Token{Type: TokenLParen, Value: "(", Line: startLine, Column: startCol}
	case ')':
		lx.advance()
		return Token{Type: TokenRParen, Value: ")", Line: startLine, Column: startCol}
	case '[':
		lx.advance()
		return Token{Type: TokenLBracket, Value: "[", Line: startLine, Column: startCol}
	case ']':
		lx.advance()
		return Token{Type: TokenRBracket, Value: "]", Line: startLine, Column: startCol}
	case ':':
		lx.advance()
		if lx.peek() == '=' {
			lx.advance()
			return Token{Type: TokenAssign, Value: ":=", Line: startLine, Column: startCol}
		}
	case '"', '\'':
		return lx.lexString()
	}

	if isIdentStart(ch) {
		return lx.lexIdentifierOrKeyword()
	}

	if unicode.IsDigit(ch) {
		return lx.lexNumber()
	}

	if strings.ContainsRune("+-*/=<>", ch) {
		return lx.lexOperator()
	}

	panic(fmt.Sprintf("unexpected character '%c' at %d:%d", ch, lx.line, lx.column))
}

func (lx *Lexer) lexIdentifierOrKeyword() Token {
	start := lx.pos
	startLine := lx.line
	startCol := lx.column

	for isIdentPart(lx.peek()) {
		lx.advance()
	}

	word := lx.input[start:lx.pos]
	kw := strings.ToLower(word)
	switch kw {
	case "if":
		return Token{Type: TokenIf, Value: word, Line: startLine, Column: startCol}
	case "then":
		return Token{Type: TokenThen, Value: word, Line: startLine, Column: startCol}
	case "else":
		return Token{Type: TokenElse, Value: word, Line: startLine, Column: startCol}
	case "end":
		return Token{Type: TokenEnd, Value: word, Line: startLine, Column: startCol}
	case "for":
		return Token{Type: TokenFor, Value: word, Line: startLine, Column: startCol}
	case "in":
		return Token{Type: TokenIn, Value: word, Line: startLine, Column: startCol}
	case "do":
		return Token{Type: TokenDo, Value: word, Line: startLine, Column: startCol}
	case "true":
		return Token{Type: TokenTrue, Value: word, Line: startLine, Column: startCol}
	case "false":
		return Token{Type: TokenFalse, Value: word, Line: startLine, Column: startCol}
	default:
		return Token{Type: TokenIdentifier, Value: word, Line: startLine, Column: startCol}
	}
}

func (lx *Lexer) lexNumber() Token {
	start := lx.pos
	startLine := lx.line
	startCol := lx.column

	for unicode.IsDigit(lx.peek()) {
		lx.advance()
	}
	if lx.peek() == '.' {
		lx.advance()
		for unicode.IsDigit(lx.peek()) {
			lx.advance()
		}
	}
	return Token{Type: TokenNumber, Value: lx.input[start:lx.pos], Line: startLine, Column: startCol}
}

func (lx *Lexer) lexString() Token {
	quote := lx.advance() // consume opening quote
	start := lx.pos
	startLine := lx.line
	startCol := lx.column

	for lx.peek() != rune(quote) && lx.peek() != 0 {
		lx.advance()
	}
	val := lx.input[start:lx.pos]
	lx.advance() // consume closing quote
	return Token{Type: TokenString, Value: val, Line: startLine, Column: startCol}
}

func (lx *Lexer) lexOperator() Token {
	start := lx.pos
	_ = start // unused?
	startLine := lx.line
	startCol := lx.column

	ch := lx.advance()
	op := string(ch)

	if (ch == '<' || ch == '>') && lx.peek() == '=' {
		op += string(lx.advance())
	} else if ch == '<' && lx.peek() == '>' {
		op += string(lx.advance())
	}

	return Token{Type: TokenBinOp, Value: op, Line: startLine, Column: startCol}
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}
