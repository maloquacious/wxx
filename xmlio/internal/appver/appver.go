// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package appver models a codec's declaration of the application versions it
// accepts, verifies that a declaration is valid on its own terms, and verifies
// that no two codecs claim the same application version.
//
// The declaration belongs to the codec, not to the registry: a codec knows which
// application builds wrote the schema it implements, and that knowledge is what
// lets Encode reject a version it cannot honestly write (issue #41 requirements 3
// and 4). It is also what lets Encode DERIVE the identity it writes -- the
// map/@release and XML declaration bytes that go with an application version --
// rather than echoing the identity of whatever file happened to be decoded into
// the map it was handed (issue #45). This package holds only the shape of the
// declaration and the checks over it, so that every codec states its set the same
// way and one check can look across all of them.
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

// App_t is one application version a codec accepts, paired with the map/@release
// string that application version writes.
//
// The pairing is here because map/@release is DERIVED from the application
// version and the codec owns the derivation (issue #45 Decision 5). It is not a
// constant of the codec: every application version a codec accepts today happens
// to agree on map/@release, which makes the mapping LOOK constant, but the two
// only diverge when a future build on the same schema is relabelled -- exactly
// the scenario ADR 0004 was written to keep expressible. Do not collapse this
// back into a per-codec field because the current data looks like one.
type App_t struct {
	// Version is map/@version verbatim as a file states it ("1.73", "2.06"). Raw
	// strings are authoritative: "2.06" and "2.6" name different files even though
	// their dotted components agree, so membership is a string comparison and
	// never a component comparison (ADR 0004 Decision 1).
	Version string

	// Release is the map/@release this application version writes, verbatim
	// ("2025"). "" means the file states no map/@release attribute AT ALL, which
	// is classic's identity (ADR 0004 Decision 2) and not "unknown" -- the two
	// must not be conflated, because an unknown release is a bug and an absent one
	// is a correct classic file.
	//
	// This is an output value. It is never read from the map being encoded.
	Release string
}

// Set_t is one codec's declaration: the application versions it accepts, the
// map/@release each of them writes, the single schema it writes, and the XML
// declaration version its files open with.
//
// The set is closed. An application version this set does not name is one the
// codec rejects, never one it writes on a best-effort basis (ADR 0004 Decision
// 5).
type Set_t struct {
	// Codec identifies the declaring codec by its package path element, e.g.
	// "v0_77". It is our identifier, not a file value, and it exists here only to
	// name the codec in an error message.
	//
	// A codec version is not a schema version: "0.77" appears on no disk and must
	// never reach Schema or a file.
	Codec string

	// Schema is the single schema version this codec writes, verbatim as a file
	// states it ("1.06"); "" when the codec writes no map/@schema attribute at
	// all, which is classic's implicit legacy schema (ADR 0004 Decision 2).
	//
	// Every schema string that reaches a file is copied from here verbatim and is
	// never re-rendered from a parsed version's components, because verbatim
	// output is inviolable (ADR 0004 Decision 1).
	Schema string

	// XMLVersion is the version in the XML declaration this codec's files open
	// with: "1.0" for classic, "1.1" for W2025.
	//
	// It is a per-CODEC constant TODAY, and only because that is what Inkwell
	// currently produces: every build that writes a given schema happens to open
	// its files with the same declaration. That is an observation about the
	// current builds, not a law. Should a later build on this codec's schema open
	// its files differently, this moves onto App_t and becomes per-application-
	// version, exactly as App_t.Release already is.
	//
	// What does not change either way is ownership: the XML declaration is bytes
	// the encoder writes, so the codec owns the application-version -> XML-version
	// mapping. One value per codec is that mapping's current range, not a claim
	// that its range can only ever hold one value.
	XMLVersion string

	// Apps are the application versions the codec accepts, each paired with the
	// map/@release it writes. See App_t.
	Apps []App_t
}

// Clone returns a copy of s whose Apps slice the caller owns, so that handing out
// a declaration cannot hand out the ability to edit it.
//
// App_t is two strings and nothing else, so copying the slice copies the whole
// declaration; there is no shared state left to reach.
func (s Set_t) Clone() Set_t {
	out := s
	out.Apps = append([]App_t(nil), s.Apps...)
	return out
}

// App returns the declaration of the application version app, and whether the
// codec accepts it.
//
// The comparison is verbatim, for the reason given on App_t.Version. The App_t
// returned when ok is false is the zero value and must not be read: its empty
// Release would be indistinguishable from classic's meaningful "write no
// map/@release attribute".
func (s Set_t) App(app string) (App_t, bool) {
	for _, a := range s.Apps {
		if a.Version == app {
			return a, true
		}
	}
	return App_t{}, false
}

// Accepts reports whether the codec accepts the application version app.
func (s Set_t) Accepts(app string) bool {
	_, ok := s.App(app)
	return ok
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

// Verify returns an error unless s is a valid declaration on its own terms,
// independent of any other codec's.
//
// A codec states a schema IF AND ONLY IF its application versions state a
// release: classic files carry neither map/@schema nor map/@release, W2025 files
// carry both (ADR 0003 Decision 2). This is the guard NewRegistry used to enforce
// over registry entries, which issue #41 called load-bearing and which has
// nowhere else to live once the registry stops carrying schema and release
// (issue #45 Decision 8). It moves here rather than lapsing.
//
// It belongs on the declaration because the declaration is what it is about. A
// codec pairing Schema "" with an app carrying Release "2025" is not describing a
// file that could exist -- it is an invalid input, and the encoder rejects
// invalid inputs rather than writing a file that is half classic and half W2025.
//
// The check is per-app against the codec's one schema, so it also catches a set
// whose apps disagree with each other about whether a release is written.
//
// It does NOT check that Release is any particular string. "2025" today and a
// relabelled "2026" tomorrow are both valid on the same schema (ADR 0004); only
// the presence or absence of a release is tied to the presence or absence of a
// schema.
func (s Set_t) Verify() error {
	for i, a := range s.Apps {
		if a.Version == "" {
			// An empty version is not a version: it would match a map that states
			// no map/@version at all and let it through the gate.
			return errors.Join(wxx.ErrInvalidCodecDeclaration, wxx.ErrMissingVersion, fmt.Errorf("codec %s: app %d: empty application version", s.Codec, i))
		}
		if (s.Schema == "") != (a.Release == "") {
			return errors.Join(wxx.ErrInvalidCodecDeclaration, fmt.Errorf("codec %s: version %q: release %s with schema %s: a codec states a schema if and only if its apps state a release", s.Codec, a.Version, releaseLabel(a.Release), s.SchemaLabel()))
		}
	}
	return nil
}

// SchemaLabel renders the declared schema for an error message.
func (s Set_t) SchemaLabel() string {
	if s.Schema == "" {
		return "implicit (classic)"
	}
	return fmt.Sprintf("%q", s.Schema)
}

// releaseLabel renders a map/@release for an error message, distinguishing the
// meaningful absence from a quoted value.
func releaseLabel(release string) string {
	if release == "" {
		return "absent (classic)"
	}
	return fmt.Sprintf("%q", release)
}

// appList renders the accepted set for an error message.
func (s Set_t) appList() string {
	if len(s.Apps) == 0 {
		return "no application version"
	}
	quoted := make([]string, 0, len(s.Apps))
	for _, a := range s.Apps {
		quoted = append(quoted, fmt.Sprintf("%q", a.Version))
	}
	return strings.Join(quoted, ", ")
}

// VerifyDisjoint returns an error unless every declaration in sets is valid on
// its own terms AND every application version named by sets is accepted by no
// more than one codec.
//
// The per-declaration check comes first because disjointness over a table of
// invalid declarations is not worth knowing: Set_t.Verify is what it runs, and it
// is the guard that keeps the schema/release convention honest now that the
// registry no longer carries either (issue #45 Decision 8).
//
// Disjointness itself is requirement 4 of issue #41 as a property the codecs
// state rather than one the registry happens to produce. The registry's own
// duplicate-application-version guard is a different check with a different
// subject: it stops one version naming two RELEASES, and it would pass a table in
// which two CODECS both claimed to write "2.06" for one release, because that
// table never repeats the version. Here the sets themselves must not overlap, so
// "which codec writes this application version" has exactly one answer before any
// table is built.
//
// An application version named by no set is fine: a codec may be written for a
// build we do not support, and a build may be supported by none.
func VerifyDisjoint(sets ...Set_t) error {
	for _, s := range sets {
		if err := s.Verify(); err != nil {
			return err
		}
	}
	claimed := make(map[string]string) // app version -> the codec that claimed it
	for _, s := range sets {
		for _, a := range s.Apps {
			if prev, ok := claimed[a.Version]; ok {
				if prev == s.Codec {
					return errors.Join(wxx.ErrAmbiguousAppCodec, fmt.Errorf("version %q: codec %s names it twice", a.Version, s.Codec))
				}
				return errors.Join(wxx.ErrAmbiguousAppCodec, fmt.Errorf("version %q: accepted by codec %s and codec %s", a.Version, prev, s.Codec))
			}
			claimed[a.Version] = s.Codec
		}
	}
	return nil
}
