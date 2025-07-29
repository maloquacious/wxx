// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Excellent. Let’s now **refactor the parser scaffold** to build and return the typed AST nodes we just defined, with token positions included for error reporting.

// ---

// 🧩 Assumptions
//
// * We’ll use the token and AST types from our previous responses.
// * We’re building a recursive descent parser.
// * We’ll build an in-memory `*ast.Program` and return it from `ParseProgram`.

// ---

// ✅ Updated Parser Scaffold

// Here’s a refactored and simplified parser that returns a full AST:

package dsl

import (
	"fmt"
	"github.com/maloquacious/wxx/dsl/ast"
	"strconv"
)

type Parser struct {
	tokens   []Token
	pos      int
	filename string
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func NewParserWithFilename(tokens []Token, filename string) *Parser {
	return &Parser{tokens: tokens, filename: filename}
}

func (p *Parser) formatError(tok Token, msg string) string {
	if p.filename != "" {
		return fmt.Sprintf("%s:%d:%d: %s", p.filename, tok.Line, tok.Column, msg)
	}
	return fmt.Sprintf("%d:%d: %s", tok.Line, tok.Column, msg)
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
		panic(p.formatError(tok, fmt.Sprintf("Syntax error: expected %s but found %s", expected, tok.Type)))
	}
	return tok
}

// ---

// 🎯 ParseProgram

func (p *Parser) ParseProgram() *ast.Program {
	stmts := []ast.Stmt{}
	for p.peek().Type != TokenEOF {
		stmt := p.parseStatement()
		stmts = append(stmts, stmt)
	}
	return &ast.Program{Statements: stmts}
}

// ---

// 🎯 parseStatement

func (p *Parser) parseStatement() ast.Stmt {
	switch tok := p.peek(); tok.Type {
	case TokenIf:
		return p.parseIf()
	case TokenFor:
		return p.parseFor()
	case TokenIdentifier:
		return p.parseAssignmentOrCall()
	default:
		panic(p.formatError(tok, fmt.Sprintf("Syntax error: unexpected %s '%s'", tok.Type, tok.Value)))
	}
}

// ---

// 🟨 Assignment or Function Call

func (p *Parser) parseAssignmentOrCall() ast.Stmt {
	start := p.peek()
	lv := p.parseLValue()

	if p.peek().Type == TokenAssign {
		p.match(TokenAssign)
		val := p.parseExpr()
		p.match(TokenSemicolon)
		return &ast.AssignStmt{
			Target: *lv,
			Value:  val,
			At:     toPos(start),
		}
	} else if p.peek().Type == TokenLParen {
		// Function call treated as a statement
		call := p.parseCallExpr(lv.Root)
		p.match(TokenSemicolon)
		return &ast.CallStmt{
			Call: call,
			At:   toPos(start),
		}
	}

	tok := p.peek()
	panic(p.formatError(tok, fmt.Sprintf("Syntax error: unexpected %s '%s' after identifier", tok.Type, tok.Value)))
}

// ---

// 🟨 LValue Parsing

func (p *Parser) parseLValue() *ast.LValue {
	start := p.match(TokenIdentifier)

	lv := &ast.LValue{
		Root:  start.Value,
		Steps: []ast.LValueStep{},
		At:    toPos(start),
	}

	for {
		switch p.peek().Type {
		case TokenDot:
			p.match(TokenDot)
			prop := p.match(TokenIdentifier)
			lv.Steps = append(lv.Steps, &ast.PropAccess{
				Name: prop.Value,
				At:   toPos(prop),
			})
		case TokenLBracket:
			p.match(TokenLBracket)
			idx := p.parseExpr()
			close := p.match(TokenRBracket)
			lv.Steps = append(lv.Steps, &ast.IndexAccess{
				Index: idx,
				At:    toPos(close), // position of `]`
			})
		default:
			return lv
		}
	}
}

// ---

// 🟦 Control Flow

// `if expr then ... [else ...] end`

func (p *Parser) parseIf() ast.Stmt {
	start := p.match(TokenIf)
	cond := p.parseExpr()
	p.match(TokenThen)

	thenBranch := []ast.Stmt{}
	for p.peek().Type != TokenElse && p.peek().Type != TokenEnd {
		thenBranch = append(thenBranch, p.parseStatement())
	}

	var elseBranch []ast.Stmt
	if p.peek().Type == TokenElse {
		p.match(TokenElse)
		for p.peek().Type != TokenEnd {
			elseBranch = append(elseBranch, p.parseStatement())
		}
	}

	p.match(TokenEnd)
	return &ast.IfStmt{
		Condition: cond,
		Then:      thenBranch,
		Else:      elseBranch,
		At:        toPos(start),
	}
}

// `for ident in expr do ... end`

func (p *Parser) parseFor() ast.Stmt {
	start := p.match(TokenFor)
	ident := p.match(TokenIdentifier)
	p.match(TokenIn)
	iter := p.parseExpr()
	p.match(TokenDo)

	body := []ast.Stmt{}
	for p.peek().Type != TokenEnd {
		body = append(body, p.parseStatement())
	}
	p.match(TokenEnd)

	return &ast.ForStmt{
		VarName:  ident.Value,
		Iterator: iter,
		Body:     body,
		At:       toPos(start),
	}
}

// ---

// 🟩 Expressions

// parseExpr (temporary version)

// We’ll replace this with a full precedence parser soon:

func (p *Parser) parseExpr() ast.Expr {
	tok := p.advance()
	switch tok.Type {
	case TokenNumber:
		return &ast.LiteralExpr{Value: parseNumber(tok.Value), At: toPos(tok)}
	case TokenString:
		return &ast.LiteralExpr{Value: tok.Value, At: toPos(tok)}
	case TokenTrue:
		return &ast.LiteralExpr{Value: true, At: toPos(tok)}
	case TokenFalse:
		return &ast.LiteralExpr{Value: false, At: toPos(tok)}
	case TokenIdentifier:
		if p.peek().Type == TokenLParen {
			return p.parseCallExpr(tok.Value)
		}
		// Check for property access like map.hexes
		if p.peek().Type == TokenDot {
			return p.parsePropertyAccess(tok)
		}
		return &ast.IdentExpr{Name: tok.Value, At: toPos(tok)}
	default:
		panic(p.formatError(tok, fmt.Sprintf("Syntax error: unexpected %s '%s' in expression", tok.Type, tok.Value)))
	}
}

// Function Call as Expr

func (p *Parser) parseCallExpr(name string) *ast.CallExpr {
	start := p.match(TokenLParen)
	args := []ast.Expr{}

	if p.peek().Type != TokenRParen {
		args = append(args, p.parseExpr())
		for p.peek().Type == TokenComma {
			p.match(TokenComma)
			args = append(args, p.parseExpr())
		}
	}
	p.match(TokenRParen)
	return &ast.CallExpr{
		FuncName: name,
		Args:     args,
		At:       toPos(start),
	}
}

// ---

// 🧩 Utilities

func toPos(tok Token) ast.Pos {
	return ast.Pos{Line: tok.Line, Column: tok.Column}
}

func (p *Parser) parsePropertyAccess(startTok Token) ast.Expr {
	// For expressions like "map.hexes", we'll create a special identifier
	// that includes the full path, since the VM knows how to handle "map.hexes"
	name := startTok.Value

	for p.peek().Type == TokenDot {
		p.match(TokenDot)
		prop := p.match(TokenIdentifier)
		name += "." + prop.Value
	}

	return &ast.IdentExpr{Name: name, At: toPos(startTok)}
}

func parseNumber(val string) interface{} {
	// Try parsing as int64 first
	if i, err := strconv.ParseInt(val, 10, 64); err == nil {
		return i
	}
	// Then try parsing as float64
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}
	// Fallback to string if parsing fails
	return val
}

// ---

// ✅ What You Now Have
//
// * A **recursive descent parser** that builds a well-structured AST with positions.
// * Full support for:
//
// * Assignments
// * Function calls
// * `if/then/else/end` and `for/in/do/end`
// * Nested paths like `a.b[0].c`
// * Literals and identifiers

// ---

// 🔜 What’s Next?
//
// 1. ✅ Implementing **expression parsing with operator precedence**?
// 2. ✅ Adding **AST validation and semantic checks** (e.g., undefined vars)?
// 3. 🔄 Building a **VM or interpreter** for this language?
