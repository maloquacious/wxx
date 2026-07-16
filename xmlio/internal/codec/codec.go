// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package codec holds the parse/emit pairs and the schema -> codec selector
// (ADR 0004 Decision 4).
//
// It is internal to xmlio because of issue #41 requirement 5: the DISPATCHER
// picks the encoder, and a caller may name only an application version. A
// selector that takes a schema and hands back an encoder is exactly the reach
// requirement 5 denies -- #41 documents what it buys, a classic map emitted
// through the W2025 codec as 18,006 bytes of W2025 XML declaring release=""
// version="1.77" schema="", which then re-decodes silently as classic.
//
// Unexporting the selector inside xmlio would not do: xmlio's tests are an
// EXTERNAL test package (package xmlio_test) and could not see it. Go's internal
// rule is directory-based rather than package-name-based, so those tests -- which
// sit physically inside xmlio/ -- may still import this package, while nothing
// outside the xmlio subtree can. That is requirement 5's "test units may choose
// the encoder" exception, preserved by construction and with no escape hatch.
package codec

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/v0_77"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// DecodeFunc parses one schema's XML into the Map_t superset.
type DecodeFunc func(input []byte) (*wxx.Map_t, error)

// EncodeFunc emits a Map_t as one schema's XML, as the application version app.
//
// app is verbatim map/@version. The codec verifies it against the set it declares
// and rejects one it does not accept, so an encoder can never be talked into
// writing a release that never existed (issue #41 requirement 3). It has to be
// passed rather than derived because the schema cannot tell the codec which of
// the application versions sharing it the caller meant (ADR 0004 Decision 4).
type EncodeFunc func(m *wxx.Map_t, app string) ([]byte, error)

// Codec_t is the parse/emit pair a schema selects (ADR 0004 Decision 4).
//
// It is deliberately keyed off the schema and not off the release: two
// application versions sharing one schema share this pair and differ only in the
// string written to map/@version, which is caller-chosen data the codec cannot
// infer -- but not data it accepts unexamined, since each codec declares the
// application versions it accepts and Encode verifies against that set.
type Codec_t struct {
	Decode DecodeFunc
	Encode EncodeFunc
}

// Entry_t binds one schema to the codec that parses and emits it.
//
// Schema is verbatim map/@schema ("1.06"); "" is the implicit legacy (classic)
// schema, which files state by stating no @schema at all (ADR 0004 Decision 2).
// Nothing on disk can collide with "": ParseDotted rejects an empty string, so no
// file can state one.
type Entry_t struct {
	Schema string
	Codec  Codec_t
}

// VerifyTable returns an error unless every schema in entries selects exactly
// one codec.
//
// The check lives here, taking a table, rather than inline in init, so that it
// stays testable: a test can hand it a deliberately ambiguous table and inspect
// the error, which it could not do with a panic. This mirrors NewRegistry and
// appver.VerifyDisjoint, for the same reason.
//
// Entries MAY repeat a schema -- they simply must agree on the codec it selects.
// Disagreement means the schema no longer answers "which code path emits this",
// which is the one question ADR 0004 Decision 4 gives it.
func VerifyTable(entries ...Entry_t) error {
	seen := make(map[string]Codec_t, len(entries))
	for _, e := range entries {
		if e.Codec.Decode == nil || e.Codec.Encode == nil {
			return errors.Join(wxx.ErrMissingCodec, fmt.Errorf("schema %s: incomplete codec pair (decode nil: %v, encode nil: %v)", Label(e.Schema), e.Codec.Decode == nil, e.Codec.Encode == nil))
		}
		if prev, ok := seen[e.Schema]; ok && !SameCodec(prev, e.Codec) {
			return errors.Join(wxx.ErrAmbiguousSchemaCodec, fmt.Errorf("schema %s: named twice, selecting a different codec each time", Label(e.Schema)))
		}
		seen[e.Schema] = e.Codec
	}
	return nil
}

// SameCodec reports whether two codec pairs name the same functions.
//
// Func values are not comparable with ==, so this compares code pointers. Every
// codec is a package-level function rather than a closure, so a code pointer
// identifies it uniquely.
func SameCodec(a, b Codec_t) bool {
	return reflect.ValueOf(a.Decode).Pointer() == reflect.ValueOf(b.Decode).Pointer() &&
		reflect.ValueOf(a.Encode).Pointer() == reflect.ValueOf(b.Encode).Pointer()
}

// Label renders a schema for an error message.
func Label(schema string) string {
	if schema == "" {
		return "implicit (classic)"
	}
	return fmt.Sprintf("%q", schema)
}

// table returns the schema -> codec table: every schema this package parses and
// emits, and the codec that does it.
//
// Adding a schema is an entry here rather than a new switch arm. The package
// paths name the CODEC version, which by convention matches the schema the file
// states -- v1_06 implements the schema a file states as schema="1.06" -- and
// v0_77 implements the schema classic files state by stating none. "0.77" is a
// codec version that appears on no disk and is deliberately absent from the
// Schema column here.
func table() []Entry_t {
	return []Entry_t{
		{Schema: "", Codec: Codec_t{Decode: v0_77.Decode, Encode: v0_77.Encode}},
		{Schema: "1.06", Codec: Codec_t{Decode: v1_06.Decode, Encode: v1_06.Encode}},
	}
}

// bySchema indexes table() and is built by init.
var bySchema map[string]Codec_t

// init builds the index, panicking if the compiled-in table is ambiguous.
//
// The table is a constant of the program, so an ambiguity in it is a programming
// error rather than a runtime condition a caller could handle -- and the failure
// it would otherwise produce is a silently wrong codec at encode time. Failing at
// load makes it unmissable.
func init() {
	entries := table()
	if err := VerifyTable(entries...); err != nil {
		panic(fmt.Sprintf("xmlio/internal/codec: schema table: %v", err))
	}
	bySchema = make(map[string]Codec_t, len(entries))
	for _, e := range entries {
		bySchema[e.Schema] = e.Codec
	}
}

// ForSchema resolves the parse/emit pair a schema selects (ADR 0004 Decision 4).
// schema is verbatim map/@schema; "" asks for the implicit legacy (classic)
// schema.
//
// The schema answers "which code path parses/emits this"; the application
// version does not, and is caller-chosen data. Two application versions sharing a
// schema therefore resolve to one Codec_t here.
//
// The match is verbatim, for the same reason the registry keys application
// versions on Raw: "1.06" and "1.6" carry identical dotted components but name
// different files, so a component comparison would conflate them.
func ForSchema(schema string) (Codec_t, error) {
	if c, ok := bySchema[schema]; ok {
		return c, nil
	}
	return Codec_t{}, errors.Join(wxx.ErrUnsupportedMapSchema, fmt.Errorf("schema %s: not a supported schema", Label(schema)))
}
