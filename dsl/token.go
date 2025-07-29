// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package dsl

import "fmt"

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString
	TokenAssign    // :=
	TokenSemicolon // ;
	TokenDot       // .
	TokenComma     // ,
	TokenLParen    // (
	TokenRParen    // )
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenIf
	TokenThen
	TokenElse
	TokenEnd
	TokenFor
	TokenIn
	TokenDo
	TokenTrue
	TokenFalse
	TokenBinOp // +, -, *, /, =, <>, etc.
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "end of file"
	case TokenIdentifier:
		return "identifier"
	case TokenNumber:
		return "number"
	case TokenString:
		return "string"
	case TokenAssign:
		return ":="
	case TokenSemicolon:
		return ";"
	case TokenDot:
		return "."
	case TokenComma:
		return ","
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenIf:
		return "if"
	case TokenThen:
		return "then"
	case TokenElse:
		return "else"
	case TokenEnd:
		return "end"
	case TokenFor:
		return "for"
	case TokenIn:
		return "in"
	case TokenDo:
		return "do"
	case TokenTrue:
		return "true"
	case TokenFalse:
		return "false"
	case TokenBinOp:
		return "operator"
	default:
		return fmt.Sprintf("unknown token %d", int(t))
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)", t.Type, t.Value)
}
