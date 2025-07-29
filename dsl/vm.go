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
	Msg string
	Pos ast.Pos
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error at line %d: %s", e.Pos.Line, e.Msg)
}

// ---

// 🧠 Execution Context

type VM struct {
	vars  map[string]Value           // for loop vars, etc.
	funcs map[string]BuiltinFunction // built-in function handlers
	root  *MapRoot                   // root object (e.g., WXX DOM)
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

type BuiltinFunction func(args []Value) error

// ---

// 🧪 VM Entry Point

func NewVM(root *MapRoot) *VM {
	return &VM{
		vars: make(map[string]Value),
		funcs: map[string]BuiltinFunction{
			"print": func(args []Value) error {
				for _, arg := range args {
					fmt.Println(arg)
				}
				return nil
			},
			"save": func(args []Value) error {
				fmt.Printf("Saving to: %v (mock)\n", args[0])
				return nil
			},
		},
		root: root,
	}
}

func (vm *VM) Execute(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if err := vm.execStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ---

// 🔁 Statement Execution

func (vm *VM) execStmt(s ast.Stmt) error {
	switch s := s.(type) {
	case *ast.AssignStmt:
		val, err := vm.evalExpr(s.Value)
		if err != nil {
			return err
		}
		return vm.assign(&s.Target, val)
	case *ast.CallStmt:
		return vm.evalCall(s.Call)
	case *ast.IfStmt:
		cond, err := vm.evalExpr(s.Condition)
		if err != nil {
			return err
		}
		truthy, ok := cond.(bool)
		if !ok {
			return RuntimeError{"if condition must be true or false", s.At}
		}
		if truthy {
			for _, stmt := range s.Then {
				if err := vm.execStmt(stmt); err != nil {
					return err
				}
			}
		} else {
			for _, stmt := range s.Else {
				if err := vm.execStmt(stmt); err != nil {
					return err
				}
			}
		}
	case *ast.ForStmt:
		iter, err := vm.evalExpr(s.Iterator)
		if err != nil {
			return err
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
				return RuntimeError{"cannot iterate over this value - use something like 'map.hexes'", s.At}
			}
		}
		for _, item := range slice {
			vm.vars[s.VarName] = item
			for _, stmt := range s.Body {
				if err := vm.execStmt(stmt); err != nil {
					return err
				}
			}
		}
	default:
		return RuntimeError{"unknown statement type", s.Pos()}
	}
	return nil
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
			return nil, RuntimeError{"variable '" + e.Name + "' is not defined", e.At}
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
		return nil, RuntimeError{"unsupported binary operator: " + e.Operator, e.At}
	case *ast.CallExpr:
		return nil, vm.evalCall(e)
	default:
		return nil, RuntimeError{"unknown expression", e.Pos()}
	}
}

// ---

// 🧾 Call Evaluation

func (vm *VM) evalCall(c *ast.CallExpr) error {
	fn, ok := vm.funcs[c.FuncName]
	if !ok {
		return RuntimeError{"function '" + c.FuncName + "' does not exist", c.At}
	}
	args := []Value{}
	for _, arg := range c.Args {
		val, err := vm.evalExpr(arg)
		if err != nil {
			return err
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
						return RuntimeError{"terrain must be a text value", prop.Pos()}
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
					return RuntimeError{"index must be a number", indexStep.Pos()}
				}
				i := int(idx)
				if i < 0 || i >= len(vm.root.Hexes) {
					return RuntimeError{fmt.Sprintf("index %d is out of range (valid: 0 to %d)", i, len(vm.root.Hexes)-1), indexStep.Pos()}
				}
				if len(lv.Steps) == 3 {
					if prop2, ok := lv.Steps[2].(*ast.PropAccess); ok && prop2.Name == "terrain" {
						s, ok := val.(string)
						if !ok {
							return RuntimeError{"terrain must be a text value", prop2.Pos()}
						}
						vm.root.Hexes[i].Terrain = s
						return nil
					}
				}
			}
		}
	}
	return RuntimeError{"cannot assign to this - try 'map.hexes[index].terrain := \"value\"'", lv.At}
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
