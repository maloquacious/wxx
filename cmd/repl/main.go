// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package main

import (
	"bufio"
	"fmt"
	"github.com/maloquacious/wxx/dsl"
	"github.com/maloquacious/wxx/dsl/ast"
	"os"
	"strings"
)

func main() {
	fmt.Println("WXX DSL REPL - type `$exit` to quit")

	vm := dsl.NewVM(dsl.NewMockMap())
	scanner := bufio.NewScanner(os.Stdin)

	var lines []string

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		} else if strings.HasPrefix(strings.TrimSpace(line), "$") {
			handleReplCommand(vm, strings.TrimSpace(line))
			continue
		}

		lines = append(lines, line)

		// Simple block detection (improve later)
		if blockComplete(lines) {
			input := strings.Join(lines, "\n")
			lines = nil

			tokens := []dsl.Token{}
			lexer := dsl.NewLexer(input)
			for {
				tok := lexer.NextToken()
				tokens = append(tokens, tok)
				if tok.Type == dsl.TokenEOF {
					break
				}
			}

			parser := dsl.NewParser(tokens)
			var prog *ast.Program
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Parse error:", r)
						prog = nil
					}
				}()
				prog = parser.ParseProgram()
			}()

			if prog != nil {
				if errs := dsl.Check(prog); len(errs) > 0 {
					for _, err := range errs {
						fmt.Println("Check error:", err)
					}
				} else if err := vm.Execute(prog); err != nil {
					fmt.Println("Runtime error:", err)
				}
			}
		}
	}
}

// A simple heuristic to know when the user is done typing a block:
// 📌 Note: This is crude, but good enough for early usage. Eventually you can:
// * Track open control blocks more reliably
// * Use the parser to detect incomplete inputs (e.g., recoverable errors)
func blockComplete(lines []string) bool {
	text := strings.Join(lines, "\n")
	open := strings.Count(text, "if") + strings.Count(text, "for")
	close := strings.Count(text, "end")
	return close >= open
}

func handleReplCommand(vm *dsl.VM, line string) {
	// drop any leading spaces and the '$' that signifies repl commands
	line = strings.TrimPrefix(strings.TrimSpace(line), "$")
	args := strings.Fields(line)
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "exit":
		os.Exit(0)
	//case "vars":
	//	for k := range vm.Vars() {
	//		fmt.Println(k)
	//	}
	//case "hexes":
	//	for i, h := range vm.Root().Hexes {
	//		fmt.Printf("hexes[%d] = %s\n", i, h.Terrain)
	//	}
	default:
		fmt.Printf("Unknown REPL command: %s\n", args[0])
	}
}
