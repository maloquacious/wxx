// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package main implements the `wxx` command line tool. It hosts
// subcommands that operate on Worldographer (WXX) data files.
//
// Usage:
//
//	wxx <subcommand> [flags] [args...]
//
// Subcommands:
//
//	export   export content from a Worldographer WXX file
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	rootCmd := newRootCommand()

	err := rootCmd.ParseAndRun(context.Background(), reorderArgs(rootCmd, os.Args[1:]))
	if errors.Is(err, ff.ErrHelp) {
		fmt.Fprintln(os.Stderr, ffhelp.Command(rootCmd))
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, ffhelp.Command(rootCmd))
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCommand() *ff.Command {
	rootFlags := ff.NewFlagSet("wxx")
	rootCmd := &ff.Command{
		Name:      "wxx",
		Usage:     "wxx <subcommand> [flags] [args...]",
		ShortHelp: "tools for working with Worldographer WXX files",
		Flags:     rootFlags,
	}
	rootCmd.Subcommands = append(rootCmd.Subcommands, newExportCommand(rootFlags))
	return rootCmd
}

// reorderArgs rewrites args so that flags can appear before or after positional
// arguments. ff/v4 stops parsing flags at the first non-flag token, which
// surprises users who write e.g. `wxx export file.wxx --utf-8 out`. We walk
// the command tree to learn which flag names take a value, then at each
// subcommand level we split into (this-level args, sub-level args) and emit
// `<this-level flags> <this-level positionals> <subcommand> <sub-level...>`.
func reorderArgs(cmd *ff.Command, args []string) []string {
	return reorderAtLevel(cmd, args, collectValueFlags(cmd))
}

func reorderAtLevel(cmd *ff.Command, args []string, valueFlags map[string]bool) []string {
	// Walk args, skipping flag tokens (and their value, when known), until we
	// find a positional that matches a subcommand of cmd.
	splitIdx := -1
	var subCmd *ff.Command
	for i := 0; i < len(args); {
		a := args[i]
		if a == "--" {
			break
		}
		if n, isFlag := flagAdvance(args, i, valueFlags); isFlag {
			i = n
			continue
		}
		for _, sc := range cmd.Subcommands {
			if strings.EqualFold(a, sc.Name) {
				subCmd = sc
				splitIdx = i
				break
			}
		}
		if subCmd != nil {
			break
		}
		i++
	}

	if subCmd == nil {
		return splitFlagsFirst(args, valueFlags)
	}

	head := splitFlagsFirst(args[:splitIdx], valueFlags)
	tail := reorderAtLevel(subCmd, args[splitIdx+1:], valueFlags)
	out := make([]string, 0, len(head)+1+len(tail))
	out = append(out, head...)
	out = append(out, subCmd.Name)
	out = append(out, tail...)
	return out
}

// splitFlagsFirst returns args with all flag tokens (and their values) moved
// ahead of positional tokens, preserving relative order within each group.
func splitFlagsFirst(args []string, valueFlags map[string]bool) []string {
	var flags, rest []string
	for i := 0; i < len(args); {
		a := args[i]
		if a == "--" {
			rest = append(rest, args[i:]...)
			break
		}
		if n, isFlag := flagAdvance(args, i, valueFlags); isFlag {
			flags = append(flags, args[i:n]...)
			i = n
			continue
		}
		rest = append(rest, a)
		i++
	}
	return append(flags, rest...)
}

// flagAdvance reports whether args[i] is a flag token and, if so, returns the
// index past the flag (and any consumed value).
func flagAdvance(args []string, i int, valueFlags map[string]bool) (int, bool) {
	a := args[i]
	switch {
	case strings.HasPrefix(a, "--") && len(a) > 2:
		name := strings.TrimPrefix(a, "--")
		if eq := strings.IndexByte(name, '='); eq >= 0 {
			return i + 1, true // --name=value is self-contained
		}
		if valueFlags[name] && i+1 < len(args) {
			return i + 2, true
		}
		return i + 1, true
	case strings.HasPrefix(a, "-") && len(a) > 1 && a != "-":
		// Treat short-flag clusters as self-contained; only a bare "-x" with a
		// known value-taking short flag consumes the next arg.
		if len(a) == 2 && valueFlags[string(a[1])] && i+1 < len(args) {
			return i + 2, true
		}
		return i + 1, true
	}
	return i, false
}

// collectValueFlags walks the command tree and collects the long and short
// names of every flag that takes a value.
func collectValueFlags(cmd *ff.Command) map[string]bool {
	out := map[string]bool{}
	var walk func(*ff.Command)
	walk = func(c *ff.Command) {
		if c.Flags != nil {
			_ = c.Flags.WalkFlags(func(f ff.Flag) error {
				if flagIsBool(f) {
					return nil
				}
				if s, ok := f.GetShortName(); ok {
					out[string(s)] = true
				}
				if l, ok := f.GetLongName(); ok {
					out[l] = true
				}
				return nil
			})
		}
		for _, sc := range c.Subcommands {
			walk(sc)
		}
	}
	walk(cmd)
	return out
}

// flagIsBool reports whether f is a boolean flag that doesn't consume the next
// arg. ff/v4 doesn't expose IsBoolFlag on the public Flag interface, so we
// infer it from its placeholder and default-value strings.
func flagIsBool(f ff.Flag) bool {
	if f.GetPlaceholder() != "" {
		return false
	}
	d := f.GetDefault()
	return d == "" || d == "true" || d == "false"
}
