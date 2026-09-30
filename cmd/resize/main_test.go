// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maloquacious/wxx/xmlio"
)

// runAsResize is set in the environment of a child process to make the test
// binary run main() instead of the tests, so a test can run the command end to
// end -- flags, exit status and stderr -- without building a separate binary.
const runAsResize = "WXX_TEST_RUN_RESIZE"

func TestMain(m *testing.M) {
	if os.Getenv(runAsResize) == "1" {
		// os.Args[0] stays the test binary; the flags after "--" are resize's.
		for i, a := range os.Args {
			if a == "--" {
				os.Args = append([]string{"resize"}, os.Args[i+1:]...)
				break
			}
		}
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// resize runs the command with args and returns its exit code and stderr.
func resize(t *testing.T, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), runAsResize+"=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), stderr.String()
	} else if err != nil {
		t.Fatalf("run resize: %v", err)
	}
	return 0, stderr.String()
}

// TestOverCropFailsCleanly: cropping more than the map has must be a one-line
// error and a non-zero exit, not a panic (issue #61). cmd/crop, deleted by
// #61, panicked this way; resize, the command that replaces it for cropping,
// did too, because its size check ran after the allocation it was guarding.
func TestOverCropFailsCleanly(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	// The fixture is 5 x 3; removing 6 columns leaves -1.
	code, stderr := resize(t,
		"-input", filepath.Join("..", "..", "testdata", "2017-1.77-1.0-columns-blank.wxx"),
		"-output", out,
		"-left", "-2", "-right", "-4")
	if code == 0 {
		t.Fatalf("exit 0, want a failure; stderr:\n%s", stderr)
	}
	if strings.Contains(stderr, "panic:") {
		t.Fatalf("resize panicked:\n%s", stderr)
	}
	if !strings.Contains(stderr, "smaller than 2 x 2 (this resize gives -1 x 3)") {
		t.Errorf("stderr does not explain the failure:\n%s", stderr)
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("output file exists after a failed resize (stat err = %v)", err)
	}
}

// TestCrop: a crop within the map's bounds writes the smaller map. It pins the
// replacement for cmd/crop's job, which README now sends users to resize for.
func TestCrop(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.wxx")
	code, stderr := resize(t,
		"-input", filepath.Join("..", "..", "testdata", "2017-1.77-1.0-columns-blank.wxx"),
		"-output", out,
		"-left", "-2", "-bottom", "-1")
	if code != 0 {
		t.Fatalf("exit %d, want 0; stderr:\n%s", code, stderr)
	}
	m, err := xmlio.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if m.Tiles.TilesWide != 3 || m.Tiles.TilesHigh != 2 {
		t.Errorf("output is %d x %d, want 3 x 2 (5 x 3 less 2 columns and 1 row)", m.Tiles.TilesWide, m.Tiles.TilesHigh)
	}
}
