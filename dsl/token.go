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

func (t Token) String() string {
	return fmt.Sprintf("%v(%q)", t.Type, t.Value)
}
