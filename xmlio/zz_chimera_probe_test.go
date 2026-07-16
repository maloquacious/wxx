// Copyright (c) 2026 Michael D Henderson. All rights reserved.

//go:build baselineprobe

// TEMPORARY probe for issue #45. Not for commit.
//
// It answers exactly one question: does v1_06's application-version gate stop
// v1_06.Encode(classicMap, "2.06")?
package xmlio_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

func TestZZChimeraProbe(t *testing.T) {
	raw, err := os.ReadFile("../testdata/blank-2017-1.77-1.0.wxx")
	if err != nil {
		t.Fatal(err)
	}
	classic, err := xmlio.NewDecoder().Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("map as decoded: Release=%q Version=%q Schema=%q\n",
		classic.Release, classic.Version, classic.Schema)

	// The call the issue says a test unit is entitled to make: the W2025 codec,
	// a classic map, and an application version v1_06 DOES accept.
	out, err := v1_06.Encode(classic, "2.06")
	if err != nil {
		fmt.Printf("v1_06.Encode(classicMap, \"2.06\") -> ERROR: %v\n", err)
		fmt.Println(">>> GATE STOPPED IT")
		return
	}
	fmt.Printf("v1_06.Encode(classicMap, \"2.06\") -> OK, %d bytes\n", len(out))
	if i := strings.Index(string(out), ">"); i >= 0 {
		fmt.Printf("emitted <map ...>: %s>\n", string(out)[:i])
	}
	fmt.Println(">>> GATE DID NOT STOP IT")

	// Control: an application version v1_06 does NOT accept.
	if _, err := v1_06.Encode(classic, "1.77"); err != nil {
		fmt.Printf("\ncontrol -- v1_06.Encode(classicMap, \"1.77\") -> ERROR: %v\n", err)
	} else {
		fmt.Println("\ncontrol -- v1_06.Encode(classicMap, \"1.77\") -> OK (gate is broken outright)")
	}
}
