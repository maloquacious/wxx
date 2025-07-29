// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package dsl

import (
	"fmt"
	"strings"
)

// BuiltinFunction returns a value or an error.
// The error is set if there are issues with the function (syntax or run-time).
// If the function wants to return an error to the user (for example we failed
// to open a file), then that error is in the value returned.
type BuiltinFunction func(args []Value) (any, error)

var (
	builtins = map[string]struct {
		numberOfInputArgs    int // use -1 for variable number of arguments
		numberOfReturnValues int // must be 0 or more
		fn                   BuiltinFunction
	}{
		"load":  {1, 1, biLoad},
		"print": {1, 0, biPrint},
		"save":  {1, 1, biSave},
	}
)

// biLoad tries to load a .wxx file, returning any errors.
func biLoad(args []Value) (any, error) {
	// arguments
	var fileName string
	if len(args) != 1 {
		return nil, fmt.Errorf("load: requires exactly one argument")
	} else if arg, ok := args[0].(string); !ok {
		//println(fmt.Sprintf("load: args %T %+v\n", args[0], args[0]))
		return nil, fmt.Errorf("load: argument must be a string")
	} else {
		fileName = arg
	}
	// return an error to the caller if filename is not a .wxx file
	if !strings.HasSuffix(fileName, ".wxx") {
		return fmt.Errorf("file name must end with .wxx"), nil
	}
	// For now, just return NewMockMap()
	return NewMockMap(), nil
}

// biPrint tries to display all of its arguments.
func biPrint(args []Value) (any, error) {
	for _, arg := range args {
		fmt.Println(arg)
	}
	return nil, nil
}

// biSave tries to write a .wxx file.
func biSave(args []Value) (any, error) {
	// arguments
	var fileName string
	if len(args) != 1 {
		return nil, fmt.Errorf("save: requires exactly one argument")
	} else if arg, ok := args[0].(string); !ok {
		//println(fmt.Sprintf("load: args %T %+v\n", args[0], args[0]))
		return nil, fmt.Errorf("save: argument must be a string")
	} else {
		fileName = arg
	}
	// return an error to the caller if filename is not a .wxx file
	if !strings.HasSuffix(fileName, ".wxx") {
		return fmt.Errorf("file name must end with .wxx"), nil
	}
	fmt.Printf("save: converting... mocking writing to: %q\n", fileName)
	return nil, nil
}
