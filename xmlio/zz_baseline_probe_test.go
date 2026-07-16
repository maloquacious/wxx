// Copyright (c) 2026 Michael D Henderson. All rights reserved.

//go:build baselineprobe

// TEMPORARY probe for issue #45. Not for commit.
//
// It fingerprints the encoder's observable behavior -- output bytes and error
// text -- for every (fixture, application version) pair, so that the refactor
// can be PROVED byte-identical rather than assumed so. A green suite is not
// proof (see the #32 workflow notes).
package xmlio_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio"
)

// decodeFresh re-decodes per pair so that no encode can observe a mutation left
// by the previous one.
func decodeFresh(t *testing.T, path string) (*wxx.Map_t, error) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return xmlio.NewDecoder().Decode(bytes.NewReader(raw))
}

// probeMarshal records MarshalXML's result: a hash of the bytes, or the error.
// A panic is recorded too -- the classic ROWS path is known to assert.
func probeMarshal(m *wxx.Map_t, app string) (res string) {
	defer func() {
		if r := recover(); r != nil {
			res = fmt.Sprintf("PANIC(%v)", r)
		}
	}()
	b, err := xmlio.MarshalXML(m, app)
	if err != nil {
		return "ERR(" + err.Error() + ")"
	}
	sum := sha256.Sum256(b)
	return fmt.Sprintf("OK len=%d sha=%s", len(b), hex.EncodeToString(sum[:8]))
}

// probeEncode records the full transport pipeline's result. Gzip output embeds
// no timestamp here (compress/gzip writes a zero MTIME unless set), so the hash
// is stable across runs.
func probeEncode(m *wxx.Map_t, app string) (res string) {
	defer func() {
		if r := recover(); r != nil {
			res = fmt.Sprintf("PANIC(%v)", r)
		}
	}()
	var buf bytes.Buffer
	err := xmlio.NewEncoder(xmlio.WithTargetVersion(app)).Encode(&buf, m)
	if err != nil {
		return "ERR(" + err.Error() + ")"
	}
	sum := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("OK len=%d sha=%s", buf.Len(), hex.EncodeToString(sum[:8]))
}

func TestZZBaselineProbe(t *testing.T) {
	fixtures, err := filepath.Glob("../testdata/*.wxx")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(fixtures)
	if len(fixtures) == 0 {
		t.Fatal("no fixtures matched: probe would vacuously pass")
	}

	// Every supported version, plus the two rejects: "" (not a sentinel) and an
	// unregistered version.
	apps := []string{"1.73", "1.74", "1.77", "2.06", "", "9.99"}

	var out []string
	for _, f := range fixtures {
		base := filepath.Base(f)
		for _, app := range apps {
			m, derr := decodeFresh(t, f)
			if derr != nil {
				out = append(out, fmt.Sprintf("%s\t%-5s\tDECODE-ERR(%v)", base, app, derr))
				continue
			}
			out = append(out, fmt.Sprintf("%s\t%-5s\tmarshal=%s", base, app, probeMarshal(m, app)))

			m2, _ := decodeFresh(t, f)
			out = append(out, fmt.Sprintf("%s\t%-5s\tencode =%s", base, app, probeEncode(m2, app)))
		}
	}
	sort.Strings(out)
	for _, line := range out {
		fmt.Println(line)
	}
}
