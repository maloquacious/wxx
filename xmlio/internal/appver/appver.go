// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package appver models a codec's declaration of the application versions it
// accepts, and verifies that no two codecs claim the same one.
//
// The declaration belongs to the codec, not to the registry: a codec knows which
// application builds wrote the schema it implements, and that knowledge is what
// lets Encode reject a version it cannot honestly write (issue #41 requirements 3
// and 4). This package holds only the shape of the declaration and the checks
// over it, so that every codec states its set the same way and one check can look
// across all of them.
//
// It is internal to xmlio for the same reason the codecs are: only the dispatcher
// picks a codec, and only the dispatcher needs to see what each one accepts.
package appver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/maloquacious/wxx"
)

// Set_t is one codec's declaration: the application versions it accepts and the
// single schema it writes.
//
// The set is closed. An application version this set does not name is one the
// codec rejects, never one it writes on a best-effort basis (ADR 0004 Decision
// 5).
type Set_t struct {
	// Codec identifies the declaring codec by its package path element, e.g.
	// "v0_77". It is our identifier, not a file value, and it exists here only to
	// name the codec in an error message.
	Codec string

	// Schema is the single schema version this codec writes, verbatim as a file
	// states it ("1.06"); "" when the codec writes no map/@schema attribute at
	// all, which is classic's implicit legacy schema (ADR 0004 Decision 2).
	//
	// It is a declaration, not an output source. Every schema string that reaches
	// a file comes from the map's own identity, never from here, because verbatim
	// output is inviolable (ADR 0004 Decision 1).
	Schema string

	// Apps are the application versions the codec accepts, each verbatim as a
	// file states map/@version ("1.73"). Raw strings are authoritative: "2.06"
	// and "2.6" name different files even though their dotted components agree,
	// so membership is a string comparison and never a component comparison.
	Apps []string
}

// Clone returns a copy of s whose Apps slice the caller owns, so that handing out
// a declaration cannot hand out the ability to edit it.
func (s Set_t) Clone() Set_t {
	out := s
	out.Apps = append([]string(nil), s.Apps...)
	return out
}

// Accepts reports whether the codec accepts the application version app.
//
// The comparison is verbatim, for the reason given on Apps.
func (s Set_t) Accepts(app string) bool {
	for _, a := range s.Apps {
		if a == app {
			return true
		}
	}
	return false
}

// VerifyApp returns nil if the codec accepts app, and an error naming what it
// does accept otherwise.
//
// This is requirement 3 of issue #41 at its narrowest: an encoder writes exactly
// one schema, so the application versions that schema is valid for are a fixed
// set, and a version outside it describes a release that never existed. Writing
// it would produce a file claiming a build that did not write that format.
func (s Set_t) VerifyApp(app string) error {
	if s.Accepts(app) {
		return nil
	}
	return errors.Join(wxx.ErrUnacceptedAppVersion, fmt.Errorf("version %q: codec %s writes schema %s and accepts only %s", app, s.Codec, s.SchemaLabel(), s.appList()))
}

// SchemaLabel renders the declared schema for an error message.
func (s Set_t) SchemaLabel() string {
	if s.Schema == "" {
		return "implicit (classic)"
	}
	return fmt.Sprintf("%q", s.Schema)
}

// appList renders the accepted set for an error message.
func (s Set_t) appList() string {
	if len(s.Apps) == 0 {
		return "no application version"
	}
	quoted := make([]string, 0, len(s.Apps))
	for _, a := range s.Apps {
		quoted = append(quoted, fmt.Sprintf("%q", a))
	}
	return strings.Join(quoted, ", ")
}

// VerifyDisjoint returns an error unless every application version named by sets
// is accepted by no more than one codec.
//
// This is requirement 4 of issue #41 as a property the codecs state rather than
// one the registry happens to produce. The registry's own duplicate-application-
// version guard is a different check with a different subject: it stops one
// version naming two RELEASES, and it would pass a table in which two CODECS both
// claimed to write "2.06" for one release, because that table never repeats the
// version. Here the sets themselves must not overlap, so "which codec writes this
// application version" has exactly one answer before any table is built.
//
// An application version named by no set is fine: a codec may be written for a
// build we do not support, and a build may be supported by none.
func VerifyDisjoint(sets ...Set_t) error {
	claimed := make(map[string]string) // app version -> the codec that claimed it
	for _, s := range sets {
		for _, app := range s.Apps {
			if app == "" {
				return errors.Join(wxx.ErrMissingVersion, fmt.Errorf("codec %s: empty application version", s.Codec))
			}
			if prev, ok := claimed[app]; ok {
				if prev == s.Codec {
					return errors.Join(wxx.ErrAmbiguousAppCodec, fmt.Errorf("version %q: codec %s names it twice", app, s.Codec))
				}
				return errors.Join(wxx.ErrAmbiguousAppCodec, fmt.Errorf("version %q: accepted by codec %s and codec %s", app, prev, s.Codec))
			}
			claimed[app] = s.Codec
		}
	}
	return nil
}
