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
	input    string
	pos      int
	line     int
	column   int
	filename string
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, line: 1}
}

func NewLexerWithFilename(input, filename string) *Lexer {
	return &Lexer{input: input, line: 1, filename: filename}
}

func (lx *Lexer) formatError(msg string) string {
	if lx.filename != "" {
		return fmt.Sprintf("%s:%d:%d: %s", lx.filename, lx.line, lx.column, msg)
	}
	return fmt.Sprintf("%d:%d: %s", lx.line, lx.column, msg)
}

func (lx *Lexer) peek() rune {
	if lx.pos >= len(lx.input) {
		return 0
	}
	return rune(lx.input[lx.pos])
}

func (lx *Lexer) peekNext() rune {
	if lx.pos+1 >= len(lx.input) {
		return 0
	}
	return rune(lx.input[lx.pos+1])
}

func (lx *Lexer) peekAhead(n int) rune {
	if lx.pos+n >= len(lx.input) {
		return 0
	}
	return rune(lx.input[lx.pos+n])
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

func (lx *Lexer) skipLineComment() {
	// Skip until end of line (but not the newline itself)
	for lx.peek() != '\n' && lx.peek() != 0 {
		lx.advance()
	}
}

func (lx *Lexer) skipBlockComment() {
	// Skip /* ... */
	lx.advance() // consume '/'
	lx.advance() // consume '*'
	
	for lx.peek() != 0 {
		if lx.peek() == '*' {
			lx.advance()
			if lx.peek() == '/' {
				lx.advance()
				return
			}
		} else {
			lx.advance()
		}
	}
	panic(lx.formatError("unterminated block comment"))
}

func (lx *Lexer) skipLuaBlockComment() {
	// Skip /*- ... -*/ with balanced dashes
	lx.advance() // consume '/'
	lx.advance() // consume '*'
	
	// Count opening dashes
	dashCount := 0
	for lx.peek() == '-' {
		dashCount++
		lx.advance()
	}
	
	for lx.peek() != 0 {
		if lx.peek() == '-' {
			// Check for closing pattern: dashCount dashes followed by */
			saved_pos := lx.pos
			saved_line := lx.line  
			saved_col := lx.column
			
			closeDashCount := 0
			for lx.peek() == '-' {
				closeDashCount++
				lx.advance()
			}
			
			if closeDashCount == dashCount && lx.peek() == '*' {
				lx.advance()
				if lx.peek() == '/' {
					lx.advance()
					return
				}
			}
			
			// Not a match, restore position
			lx.pos = saved_pos
			lx.line = saved_line
			lx.column = saved_col
			lx.advance()
		} else {
			lx.advance()
		}
	}
	panic(lx.formatError("unterminated Lua-style block comment"))
}

func (lx *Lexer) NextToken() Token {
	var startLine, startCol int
	var ch rune
	
	for {
		lx.skipWhitespace()

		startLine = lx.line
		startCol = lx.column
		ch = lx.peek()

		if ch == 0 {
			return Token{Type: TokenEOF, Line: startLine, Column: startCol}
		}

		// Handle comments
		if ch == '#' {
			lx.skipLineComment()
			continue // Skip to next token
		}
		
		if ch == '/' {
			next := lx.peekNext()
			if next == '/' {
				lx.advance() // consume '/'
				lx.skipLineComment()
				continue // Skip to next token
			} else if next == '*' {
				// Check for Lua-style block comment
				if lx.peekAhead(2) == '-' {
					lx.skipLuaBlockComment()
				} else {
					lx.skipBlockComment()
				}
				continue // Skip to next token
			}
		}

		// Regular token processing
		break
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

	panic(lx.formatError(fmt.Sprintf("unexpected character '%c'", ch)))
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
