// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"github.com/maloquacious/semver"
	"github.com/maloquacious/wxx/dsl"
	"os"
	"runtime/debug"
	"strings"
)

var (
	debugMode = false
	version   = semver.Version{Minor: 2}
)

func main() {
	var showVersion bool
	flag.BoolVar(&debugMode, "debug", debugMode, "enable debugging mode")
	flag.BoolVar(&showVersion, "version", false, "show version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("wxx %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	
	if len(args) == 0 {
		fmt.Println("Usage: wxx [--debug] [--version] <script.wxxsh>")
		fmt.Println("   or: wxx [--debug] <DSL statement>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  wxx 'for h in map.hexes do h.terrain := \"swamp\"; end'")
		fmt.Println("  wxx myscript.wxxsh")
		fmt.Println("  wxx --version")
		os.Exit(1)
	}

	input := args[0]
	
	// Check if it looks like a filename (starts with a path or contains a file extension)
	if strings.Contains(input, "/") || strings.Contains(input, "\\") || 
	   (strings.Contains(input, ".") && len(strings.Fields(input)) == 1) {
		// If it looks like a filename, it must be a .wxxsh file
		if !strings.HasSuffix(input, ".wxxsh") {
			fmt.Printf("Error: Script files must have .wxxsh extension (got: %s)\n", input)
			fmt.Println("This is a safety measure to distinguish WXX scripts from Worldographer data files (.wxx)")
			os.Exit(1)
		}
		// Try to read the file
		data, err := os.ReadFile(input)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", input, err)
			os.Exit(1)
		}
		input = string(data)
	} else {
		// Treat as direct statement - join all args
		input = strings.Join(args, " ")
	}

	// Determine the filename to use for error reporting
	var filename string
	if (strings.Contains(args[0], "/") || strings.Contains(args[0], "\\") || 
	    (strings.Contains(args[0], ".") && len(strings.Fields(args[0])) == 1)) && 
	   strings.HasSuffix(args[0], ".wxxsh") {
		filename = args[0]
	}

	if debugMode {
		fmt.Printf("Executing: %s\n", input)
		fmt.Println("---")
	}

	executeCode(input, filename)
}



func executeCode(input, filename string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
			if debugMode {
				fmt.Println("--- Stack Trace ---")
				debug.PrintStack()
			}
			os.Exit(1)
		}
	}()

	// Tokenize
	tokens := []dsl.Token{}
	var lexer *dsl.Lexer
	if filename != "" {
		lexer = dsl.NewLexerWithFilename(input, filename)
	} else {
		lexer = dsl.NewLexer(input)
	}
	
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == dsl.TokenEOF {
			break
		}
	}

	if debugMode {
		fmt.Println("Tokens:")
		for _, tok := range tokens {
			if tok.Type != dsl.TokenEOF {
				fmt.Printf("  %s\n", tok)
			}
		}
		fmt.Println("---")
	}

	// Parse
	var parser *dsl.Parser
	if filename != "" {
		parser = dsl.NewParserWithFilename(tokens, filename)
	} else {
		parser = dsl.NewParser(tokens)
	}
	prog := parser.ParseProgram()

	if debugMode {
		fmt.Printf("AST: %d statements\n", len(prog.Statements))
		fmt.Println("---")
	}

	// Check semantics
	if errs := dsl.Check(prog); len(errs) > 0 {
		for _, err := range errs {
			fmt.Println("Error:", err)
		}
		os.Exit(1)
	}

	// Execute
	var vm *dsl.VM
	if filename != "" {
		vm = dsl.NewVMWithFilename(dsl.NewMockMap(), filename)
	} else {
		vm = dsl.NewVM(dsl.NewMockMap())
	}
	if err := vm.Execute(prog); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if debugMode {
		fmt.Println("Execution completed successfully")
	}
}
