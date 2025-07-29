// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package ast

// 🧱 AST Design Principles
//
// * All nodes carry **position metadata** (line and column) from the originating token.
// * AST is kept **simple and orthogonal**: statements, expressions, and lvalues are distinct types.
// * Comments and whitespace are ignored by the AST.
// * Statements and expressions are **fully typed** for extensibility.

// ---

// 🧩 Common Position Struct

type Pos struct {
	Line   int
	Column int
}

// ---

// 🔨 AST Nodes

// Top-level Program

type Program struct {
	Statements []Stmt
}

// ---

// 🟦 Statement Interface and Types

type Stmt interface {
	Pos() Pos
	stmtNode()
}

// Assignment

type AssignStmt struct {
	Target LValue
	Value  Expr
	At     Pos
}

func (s *AssignStmt) Pos() Pos  { return s.At }
func (s *AssignStmt) stmtNode() {}

// Function Call (as statement)

type CallStmt struct {
	Call *CallExpr
	At   Pos
}

func (s *CallStmt) Pos() Pos  { return s.At }
func (s *CallStmt) stmtNode() {}

// If Statement

type IfStmt struct {
	Condition Expr
	Then      []Stmt
	Else      []Stmt
	At        Pos
}

func (s *IfStmt) Pos() Pos  { return s.At }
func (s *IfStmt) stmtNode() {}

// For Loop

type ForStmt struct {
	VarName  string
	Iterator Expr
	Body     []Stmt
	At       Pos
}

func (s *ForStmt) Pos() Pos  { return s.At }
func (s *ForStmt) stmtNode() {}

// ---

// 🟩 Expression Interface and Types

type Expr interface {
	Pos() Pos
	exprNode()
}

// Literal (Number, String, Boolean)

type LiteralExpr struct {
	Value interface{} // float64, string, bool
	At    Pos
}

func (e *LiteralExpr) Pos() Pos  { return e.At }
func (e *LiteralExpr) exprNode() {}

// Variable or Property Access

type IdentExpr struct {
	Name string
	At   Pos
}

func (e *IdentExpr) Pos() Pos  { return e.At }
func (e *IdentExpr) exprNode() {}

// Binary Operation

type BinaryExpr struct {
	Left     Expr
	Operator string // "+", "=", etc.
	Right    Expr
	At       Pos
}

func (e *BinaryExpr) Pos() Pos  { return e.At }
func (e *BinaryExpr) exprNode() {}

// Function Call (as expression)

type CallExpr struct {
	FuncName string
	Args     []Expr
	At       Pos
}

func (e *CallExpr) Pos() Pos  { return e.At }
func (e *CallExpr) exprNode() {}

// ---

// 🟨 LValue Path
//
// LValues represent assignable locations: `map.hexes[2].terrain`

type LValue struct {
	Root  string
	Steps []LValueStep
	At    Pos
}

func (lv *LValue) Pos() Pos { return lv.At }

type LValueStep interface {
	lvalueStepNode()
	Pos() Pos
}

type PropAccess struct {
	Name string
	At   Pos
}

func (p *PropAccess) Pos() Pos        { return p.At }
func (p *PropAccess) lvalueStepNode() {}

type IndexAccess struct {
	Index Expr
	At    Pos
}

func (i *IndexAccess) Pos() Pos        { return i.At }
func (i *IndexAccess) lvalueStepNode() {}

// ---

// 🧪 Example AST
//
// The DSL:
//
// pascal
// map.hexes[0].terrain := "swamp";
//
//
// Produces:
//
//
// &AssignStmt{
//   At: Pos{Line: 1, Column: 1},
//   Target: LValue{
//     Root: "map",
//     Steps: []LValueStep{
//       &PropAccess{"hexes", Pos{1, 5}},
//       &IndexAccess{&LiteralExpr{0, Pos{1, 11}}, Pos{1, 10}},
//       &PropAccess{"terrain", Pos{1, 14}},
//     },
//     At: Pos{1, 1},
//   },
//   Value: &LiteralExpr{"swamp", Pos{1, 24}},
// }
