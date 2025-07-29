// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Now that the parser returns a well-structured AST, we can walk it and implement **semantic validation**. This will allow us to:
//
// * Detect use of undefined identifiers or bad lvalue chains.
// * Validate function calls (e.g. `save(...)` with wrong arity).
// * Warn about unreachable code or bad control flow (e.g., `for foo in 42`).
// * Produce clear, position-aware error messages.

// ---

// ✅ Step 1: Semantic Checker Structure
//
// We'll define a `Checker` type in `github.com/maloquacious/wxx/dsl`, which traverses the AST and collects errors.

package dsl

import (
	"fmt"
	"github.com/maloquacious/wxx/dsl/ast"
)

type CheckError struct {
	Msg  string
	Line int
	Col  int
}

func (e CheckError) Error() string {
	return fmt.Sprintf("Line %d:%d: %s", e.Line, e.Col, e.Msg)
}

type Checker struct {
	errors []CheckError
	funcs  map[string]int  // func name → arity
	vars   map[string]bool // local loop vars, simple scoping for now
}

// ---

// ✅ Step 2: Entry Point

func Check(prog *ast.Program) []CheckError {
	c := &Checker{
		funcs: map[string]int{
			"save":  1,
			"print": 1, // or -1 for variadic
		},
		vars: map[string]bool{},
	}
	c.checkProgram(prog)
	return c.errors
}

// ---

// ✅ Step 3: Walk the AST

func (c *Checker) checkProgram(prog *ast.Program) {
	for _, stmt := range prog.Statements {
		c.checkStmt(stmt)
	}
}

func (c *Checker) checkStmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.AssignStmt:
		c.checkLValue(&s.Target)
		c.checkExpr(s.Value)
	case *ast.CallStmt:
		c.checkCallExpr(s.Call)
	case *ast.IfStmt:
		c.checkExpr(s.Condition)
		for _, stmt := range s.Then {
			c.checkStmt(stmt)
		}
		for _, stmt := range s.Else {
			c.checkStmt(stmt)
		}
	case *ast.ForStmt:
		c.checkExpr(s.Iterator)
		c.vars[s.VarName] = true
		for _, stmt := range s.Body {
			c.checkStmt(stmt)
		}
		delete(c.vars, s.VarName)
	}
}

// ---

// ✅ Step 4: Expression and LValue Checks

func (c *Checker) checkExpr(e ast.Expr) {
	switch e := e.(type) {
	case *ast.LiteralExpr:
		// OK
	case *ast.IdentExpr:
		if !c.vars[e.Name] {
			c.error(e.Pos(), fmt.Sprintf("undefined variable: %s", e.Name))
		}
	case *ast.CallExpr:
		c.checkCallExpr(e)
	case *ast.BinaryExpr:
		c.checkExpr(e.Left)
		c.checkExpr(e.Right)
	}
}

func (c *Checker) checkLValue(lv *ast.LValue) {
	if lv.Root != "map" {
		if !c.vars[lv.Root] {
			c.error(lv.At, fmt.Sprintf("undefined variable: %s", lv.Root))
		}
	}
	for _, step := range lv.Steps {
		switch s := step.(type) {
		case *ast.PropAccess:
			// could validate known property names in the future
		case *ast.IndexAccess:
			c.checkExpr(s.Index)
		}
	}
}

// ---

// ✅ Step 5: Call Validation

func (c *Checker) checkCallExpr(call *ast.CallExpr) {
	arity, ok := c.funcs[call.FuncName]
	if !ok {
		c.error(call.At, fmt.Sprintf("unknown function: %s", call.FuncName))
	}
	if ok && arity >= 0 && len(call.Args) != arity {
		c.error(call.At, fmt.Sprintf("function %s expects %d argument(s), got %d", call.FuncName, arity, len(call.Args)))
	}
	for _, arg := range call.Args {
		c.checkExpr(arg)
	}
}

// ---

// ✅ Step 6: Error Helper

func (c *Checker) error(pos ast.Pos, msg string) {
	c.errors = append(c.errors, CheckError{
		Msg:  msg,
		Line: pos.Line,
		Col:  pos.Column,
	})
}

// ---

// ✅ Usage Example
//
// src := `for h in map.hexes do h.terrain := "swamp"; end`
// lexer := dsl.NewLexer(src)
//
// var tokens []dsl.Token
// for {
//    tok := lexer.NextToken()
//    tokens = append(tokens, tok)
//    if tok.Type == dsl.TokenEOF {
//        break
//    }
// }
//
// parser := dsl.NewParser(tokens)
// prog := parser.ParseProgram()
//
// errors := dsl.Check(prog)
// for _, err := range errors {
//    fmt.Println(err)
// }

// ---

// ✅ Next Steps?
// * Add **scoping** or validation for known map fields (`hexes`, `terrain`, etc)?
// * Hook into `parser.go` so errors from `Check` halt execution early?
// * Extend function registry to include optional/variadic arity and type hints?
