// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Let’s build a **simple, position-aware VM** to execute your AST and stop on the first runtime error. This will provide the foundation for scripting over WXX files, with clean error reporting.

// ---

// ✅ VM Design Goals
//
// * Interpret the AST directly (no bytecode yet).
// * Track errors with `ast.Pos` for context.
// * Stop immediately on first error.
// * Operate on a simplified in-memory map structure (e.g., a fake `map.hexes`).

// ---

// 🧩 Types and Interfaces

package dsl

import (
	"fmt"
	"github.com/maloquacious/wxx/dsl/ast"
)

// ---

// 🧱 Runtime Error with Position

type RuntimeError struct {
	Msg      string
	Pos      ast.Pos
	Filename string
}

func (e RuntimeError) Error() string {
	if e.Filename != "" {
		return fmt.Sprintf("%s:%d:%d: Runtime error: %s", e.Filename, e.Pos.Line, e.Pos.Column, e.Msg)
	}
	return fmt.Sprintf("Runtime error at line %d: %s", e.Pos.Line, e.Msg)
}

func (vm *VM) newRuntimeError(msg string, pos ast.Pos) RuntimeError {
	return RuntimeError{
		Msg:      msg,
		Pos:      pos,
		Filename: vm.filename,
	}
}

// ---

// 🧠 Execution Context

type VM struct {
	vars     map[string]Value           // for loop vars, etc.
	funcs    map[string]BuiltinFunction // built-in function handlers
	root     *MapRoot                   // root object (e.g., WXX DOM)
	filename string                     // current script filename
}

// ---

// 👇 Map-Like Data Model (Stub for now)

type MapRoot struct {
	Hexes []Hex
}

type Hex struct {
	Terrain string
}

func NewMockMap() *MapRoot {
	return &MapRoot{
		Hexes: []Hex{
			{Terrain: "forest"},
			{Terrain: "plains"},
		},
	}
}

// ---

// 🔧 Value Type

type Value interface{}

// ---

// 🧪 VM Entry Point

func NewVM(root *MapRoot) *VM {
	vm := &VM{
		vars: make(map[string]Value),
		root: root,
	}

	vm.funcs = map[string]BuiltinFunction{}
	for k, v := range builtins {
		vm.funcs[k] = v.fn
	}

	return vm
}

func NewVMWithFilename(root *MapRoot, filename string) *VM {
	vm := NewVM(root)
	vm.filename = filename
	return vm
}

func (vm *VM) Execute(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if _, err := vm.execStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ---

// 🔁 Statement Execution

// execStmt returns the results of executing the statement.
// An error is returned if there is a syntax or run-time error.
func (vm *VM) execStmt(s ast.Stmt) (any, error) {
	switch s := s.(type) {
	case *ast.AssignStmt:
		val, err := vm.evalExpr(s.Value)
		if err != nil {
			return nil, err
		}
		return val, vm.assign(&s.Target, val)
	case *ast.CallStmt:
		return vm.evalCall(s.Call)
	case *ast.IfStmt:
		cond, err := vm.evalExpr(s.Condition)
		if err != nil {
			return nil, err
		}
		truthy, ok := cond.(bool)
		if !ok {
			return nil, vm.newRuntimeError("if condition must be true or false", s.At)
		}
		if truthy {
			for _, stmt := range s.Then {
				_, err = vm.execStmt(stmt)
				if err != nil {
					return nil, err
				}
			}
			return nil, nil
		}
		// not truthy, so execute the else
		for _, stmt := range s.Else {
			_, err = vm.execStmt(stmt)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *ast.ForStmt:
		iter, err := vm.evalExpr(s.Iterator)
		if err != nil {
			return nil, err
		}
		slice, ok := iter.([]Value)
		if !ok {
			// special case: hardcoded map.hexes
			if s.Iterator.(*ast.IdentExpr).Name == "map.hexes" {
				slice = make([]Value, len(vm.root.Hexes))
				for i := range vm.root.Hexes {
					slice[i] = &vm.root.Hexes[i]
				}
				//} else if s.Iterator.(*ast.LValue).Root == "map" {
				//	// impossible type assertion: s.Iterator.(*ast.LValue)
				//	//	*ast.LValue does not implement ast.Expr (missing method exprNode)
				//	slice = make([]Value, len(vm.root.Hexes))
				//	for i := range vm.root.Hexes {
				//		slice[i] = &vm.root.Hexes[i]
				//	}
			} else {
				return nil, vm.newRuntimeError("cannot iterate over this value - use something like 'map.hexes'", s.At)
			}
		}
		for _, item := range slice {
			vm.vars[s.VarName] = item
			for _, stmt := range s.Body {
				if _, err := vm.execStmt(stmt); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	default:
		return nil, vm.newRuntimeError("unknown statement type", s.Pos())
	}
}

// ---

// 🧮 Expression Evaluation

func (vm *VM) evalExpr(e ast.Expr) (Value, error) {
	switch e := e.(type) {
	case *ast.LiteralExpr:
		return e.Value, nil
	case *ast.IdentExpr:
		if e.Name == "map.hexes" {
			// Special case: return the hexes slice
			slice := make([]Value, len(vm.root.Hexes))
			for i := range vm.root.Hexes {
				slice[i] = &vm.root.Hexes[i]
			}
			return slice, nil
		}
		val, ok := vm.vars[e.Name]
		if !ok {
			return nil, vm.newRuntimeError("variable '"+e.Name+"' is not defined", e.At)
		}
		return val, nil
	case *ast.BinaryExpr:
		left, err := vm.evalExpr(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := vm.evalExpr(e.Right)
		if err != nil {
			return nil, err
		}
		switch e.Operator {
		case "=":
			return left == right, nil
		case "+":
			return fmt.Sprintf("%v%v", left, right), nil
		}
		return nil, vm.newRuntimeError("unsupported binary operator: "+e.Operator, e.At)
	case *ast.CallExpr:
		return vm.evalCall(e)
	default:
		return nil, vm.newRuntimeError("unknown expression", e.Pos())
	}
}

// ---

// 🧾 Call Evaluation

// evalCall returns the results of evaluating the call or an error.
// An error is returned for syntax or run-time errors.
func (vm *VM) evalCall(c *ast.CallExpr) (any, error) {
	fn, ok := vm.funcs[c.FuncName]
	if !ok {
		return nil, vm.newRuntimeError("function '"+c.FuncName+"' does not exist", c.At)
	}
	args := []Value{}
	for _, arg := range c.Args {
		val, err := vm.evalExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	return fn(args)
}

// ---

// 🖊 Assignment to LValue

func (vm *VM) assign(lv *ast.LValue, val Value) error {
	// Handle assignment to loop variable properties like h.terrain
	if len(lv.Steps) == 1 {
		if prop, ok := lv.Steps[0].(*ast.PropAccess); ok && prop.Name == "terrain" {
			if hexVar, ok := vm.vars[lv.Root]; ok {
				if hex, ok := hexVar.(*Hex); ok {
					s, ok := val.(string)
					if !ok {
						return vm.newRuntimeError("terrain must be a text value", prop.Pos())
					}
					hex.Terrain = s
					return nil
				}
			}
		}
	}

	if lv.Root == "map" && len(lv.Steps) >= 2 {
		// Handle map.hexes[i].terrain = "swamp"
		if prop1, ok := lv.Steps[0].(*ast.PropAccess); ok && prop1.Name == "hexes" {
			if indexStep, ok := lv.Steps[1].(*ast.IndexAccess); ok {
				idxVal, err := vm.evalExpr(indexStep.Index)
				if err != nil {
					return err
				}
				idx, ok := idxVal.(float64) // numbers are float64 in literal
				if !ok {
					return vm.newRuntimeError("index must be a number", indexStep.Pos())
				}
				i := int(idx)
				if i < 0 || i >= len(vm.root.Hexes) {
					return vm.newRuntimeError(fmt.Sprintf("index %d is out of range (valid: 0 to %d)", i, len(vm.root.Hexes)-1), indexStep.Pos())
				}
				if len(lv.Steps) == 3 {
					if prop2, ok := lv.Steps[2].(*ast.PropAccess); ok && prop2.Name == "terrain" {
						s, ok := val.(string)
						if !ok {
							return vm.newRuntimeError("terrain must be a text value", prop2.Pos())
						}
						vm.root.Hexes[i].Terrain = s
						return nil
					}
				}
			}
		}
	}
	return vm.newRuntimeError("cannot assign to this - try 'map.hexes[index].terrain := \"value\"'", lv.At)
}

// ---

// ✅ Example Integration
//
// lexer := dsl.NewLexer(`map.hexes[0].terrain := "swamp"; print("Done");`)
// var tokens []dsl.Token
// for {
//    tok := lexer.NextToken()
//    tokens = append(tokens, tok)
//    if tok.Type == dsl.TokenEOF {
//        break
//    }
// }
// parser := dsl.NewParser(tokens)
// prog := parser.ParseProgram()
//
// vm := dsl.NewVM(dsl.NewMockMap())
// err := vm.Execute(prog)
// if err != nil {
//    fmt.Println("Runtime error:", err)
// }

// ---

// ✅ Next Steps
//
// * Add support for dynamic maps, slice literals, string concat, math, etc.
// * Write helpers to pretty-print the modified map (e.g., dump `map.hexes`).
// * Optionally introduce a scope stack if you add nested blocks or functions.
