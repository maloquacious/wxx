// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package dsl

import "fmt"

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) match(expected TokenType) Token {
	tok := p.advance()
	if tok.Type != expected {
		panic(fmt.Sprintf("expected %v, got %v at line %d", expected, tok.Type, tok.Line))
	}
	return tok
}

// Entry point
func (p *Parser) ParseProgram() {
	for p.peek().Type != TokenEOF {
		p.parseStatement()
	}
}

func (p *Parser) parseStatement() {
	switch tok := p.peek(); tok.Type {
	case TokenIf:
		p.parseIf()
	case TokenFor:
		p.parseFor()
	case TokenIdentifier:
		p.parseAssignmentOrCall()
	default:
		panic(fmt.Sprintf("unexpected token %v at line %d", tok.Type, tok.Line))
	}
}

func (p *Parser) parseAssignmentOrCall() {
	tok := p.match(TokenIdentifier)
	if p.peek().Type == TokenAssign {
		p.advance() // :=
		p.parseExpr()
		p.match(TokenSemicolon)
	} else if p.peek().Type == TokenLParen {
		p.parseCall(tok)
		p.match(TokenSemicolon)
	} else {
		panic(fmt.Sprintf("unexpected token after identifier: %v", p.peek().Type))
	}
}

func (p *Parser) parseCall(fn Token) {
	p.match(TokenLParen)
	if p.peek().Type != TokenRParen {
		p.parseExpr()
		for p.peek().Type == TokenComma {
			p.advance()
			p.parseExpr()
		}
	}
	p.match(TokenRParen)
}

func (p *Parser) parseIf() {
	p.match(TokenIf)
	p.parseExpr()
	p.match(TokenThen)
	for p.peek().Type != TokenElse && p.peek().Type != TokenEnd {
		p.parseStatement()
	}
	if p.peek().Type == TokenElse {
		p.match(TokenElse)
		for p.peek().Type != TokenEnd {
			p.parseStatement()
		}
	}
	p.match(TokenEnd)
}

func (p *Parser) parseFor() {
	p.match(TokenFor)
	p.match(TokenIdentifier)
	p.match(TokenIn)
	p.parseExpr()
	p.match(TokenDo)
	for p.peek().Type != TokenEnd {
		p.parseStatement()
	}
	p.match(TokenEnd)
}

func (p *Parser) parseExpr() {
	// Placeholder for now — you’ll expand this to handle precedence and nesting.
	tok := p.advance()
	if tok.Type == TokenNumber || tok.Type == TokenString || tok.Type == TokenTrue || tok.Type == TokenFalse || tok.Type == TokenIdentifier {
		return
	}
	panic(fmt.Sprintf("unexpected token in expression: %v", tok.Type))
}
