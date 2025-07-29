// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/maloquacious/wxx/dsl"
	"os"
	"runtime/debug"
	"strings"
)

var (
	// global flag for debugging, set on command line or `$debug` command.
	debugMode = false
)

func main() {
	flag.BoolVar(&debugMode, "debug", debugMode, "enable debugging mode")

	fmt.Println("WXX DSL REPL - type `$exit` to quit")

	vm := dsl.NewVM(dsl.NewMockMap())
	scanner := bufio.NewScanner(os.Stdin)

	var lines []string

	for {
		if len(lines) == 0 {
			fmt.Print("> ")
		} else {
			fmt.Print(". ") // continuation line
		}
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

			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Internal error:", r)
						if debugMode {
							fmt.Println("--- Stack Trace ---")
							printStack()
						}
					}
				}()

				// LEXING
				tokens := []dsl.Token{}
				lexer := dsl.NewLexer(input)
				for {
					tok := lexer.NextToken()
					tokens = append(tokens, tok)
					if tok.Type == dsl.TokenEOF {
						break
					}
				}

				// PARSING
				parser := dsl.NewParser(tokens)
				prog := parser.ParseProgram()

				// CHECKING
				if errs := dsl.Check(prog); len(errs) > 0 {
					for _, err := range errs {
						fmt.Println("Check error:", err)
					}
					return
				}

				// EXECUTION
				if err := vm.Execute(prog); err != nil {
					fmt.Println("Runtime error:", err)
				}
			}()
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
	case "debug":
		if len(args) > 1 && args[1] == "on" {
			debugMode = true
			fmt.Println("Debug mode now enabled")
		} else if len(args) > 1 && args[1] == "off" {
			debugMode = false
			fmt.Println("Debug mode now disabled")
		} else if debugMode {
			fmt.Println("Debug mode is enabled")
		} else {
			fmt.Println("Debug mode is disabled")
		}

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

func printStack() {
	debug.PrintStack()
}
