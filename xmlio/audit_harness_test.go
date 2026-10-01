// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package xmlio_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// This file holds the element/attribute-set audit harness shared by the
// round-trip and downgrade tests: xmlAggregate summarises a UTF-8 XML document
// by element path, and computeLoss diffs two such summaries.

// pathAgg aggregates everything seen at one element path across a document.
type pathAgg struct {
	count   int                        // number of occurrences of this element path
	attrs   map[string]map[string]bool // attr local-name -> set of normalized values
	hasText bool                       // any non-whitespace chardata under this path
}

// xmlAggregate tokenizes a UTF-8 XML document and returns an aggregate keyed by
// element path (local-names joined with '/', e.g. "map/features/feature"). For
// each path it records the occurrence count, the union of attribute local-names
// (with their normalized value sets), and whether the element carries text.
// Processing instructions (the <?xml ...?> header), comments, and directives
// are ignored, as is pure-whitespace chardata.
func xmlAggregate(data []byte) (map[string]*pathAgg, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity
	// The input's <?xml?> header declares encoding="utf-16", but diagnostics
	// have already converted the bytes to UTF-8. Pass the reader through
	// unchanged so the declared (stale) charset does not error the tokenizer.
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	agg := map[string]*pathAgg{}
	var stack []string
	get := func(path string) *pathAgg {
		a := agg[path]
		if a == nil {
			a = &pathAgg{attrs: map[string]map[string]bool{}}
			agg[path] = a
		}
		return a
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			path := strings.Join(stack, "/")
			a := get(path)
			a.count++
			for _, attr := range t.Attr {
				if attr.Name.Local == "xmlns" || attr.Name.Space == "xmlns" {
					continue
				}
				vs := a.attrs[attr.Name.Local]
				if vs == nil {
					vs = map[string]bool{}
					a.attrs[attr.Name.Local] = vs
				}
				vs[normVal(attr.Value)] = true
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			if strings.TrimSpace(string(t)) == "" {
				continue
			}
			get(strings.Join(stack, "/")).hasText = true
		}
	}
	return agg, nil
}

// normVal canonicalizes an attribute value so non-semantic numeric formatting
// (e.g. "0" vs "0.0", "50" vs "50.0") does not register as an alteration. Any
// value that parses as a float is reformatted canonically; everything else
// (RGBA tuples, enums, names) is compared verbatim after trimming.
func normVal(s string) string {
	s = strings.TrimSpace(s)
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return strconv.FormatFloat(f, 'g', -1, 64)
	}
	return s
}

// computeLoss returns the sorted set of things present in the INPUT aggregate
// but missing or reduced in the OUTPUT aggregate. Each entry is a canonical
// tab-delimited string "<kind>\t<path>[\t<detail>]" so loss sets compare and
// print cleanly:
//   - element-dropped   path            (output count is 0)
//   - element-reduced   path   in=X out=Y
//   - attr-dropped      path   attr     (attr on this path in input, never in output)
//   - attr-altered      path   attr     (attr present in both, value set differs)
//   - text-dropped      path            (text under this path in input, not output)
//
// A fully dropped element implies its attrs/text are gone too, so those are not
// separately listed for it.
func computeLoss(in, out map[string]*pathAgg) []string {
	var loss []string
	for path, inA := range in {
		outA := out[path]
		if outA == nil || outA.count == 0 {
			loss = append(loss, fmt.Sprintf("element-dropped\t%s", path))
			continue
		}
		if outA.count < inA.count {
			loss = append(loss, fmt.Sprintf("element-reduced\t%s\tin=%d out=%d", path, inA.count, outA.count))
		}
		for attr, inVals := range inA.attrs {
			outVals, ok := outA.attrs[attr]
			if !ok {
				loss = append(loss, fmt.Sprintf("attr-dropped\t%s\t%s", path, attr))
				continue
			}
			if !equalStringSet(inVals, outVals) {
				loss = append(loss, fmt.Sprintf("attr-altered\t%s\t%s", path, attr))
			}
		}
		if inA.hasText && !outA.hasText {
			loss = append(loss, fmt.Sprintf("text-dropped\t%s", path))
		}
	}
	sort.Strings(loss)
	return loss
}

func equalStringSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
