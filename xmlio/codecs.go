// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package xmlio implements the XML decoding and encoding pipeline for
// Worldographer .wxx files and hosts the registry that maps an application
// version to the encoder that writes it.
package xmlio

import (
	"errors"
	"fmt"

	"github.com/maloquacious/wxx"
	"github.com/maloquacious/wxx/xmlio/internal/appver"
	"github.com/maloquacious/wxx/xmlio/internal/codec"
	"github.com/maloquacious/wxx/xmlio/internal/v0_77"
	"github.com/maloquacious/wxx/xmlio/internal/v1_06"
)

// codecs returns every codec, in codec-version order.
//
// The dispatcher is the only place that sees all of the codecs at once, so it is
// the only place that can check a property ACROSS them and the only place that
// can build an index over them. What each codec accepts is NOT restated here: it
// is read from the codec that owns it, so this table cannot drift from what the
// codecs actually enforce and write.
//
// Adding a codec means adding it here, and that is the whole of adding one. A
// codec missing from this list is not checked against the others and is not
// reachable, which is the one way the guarantees below can be lost.
func codecs() []codec.Codec {
	return []codec.Codec{
		v0_77.Codec_t{},
		v1_06.Codec_t{},
	}
}

// byApp is the registry: verbatim application version -> the codec that writes
// it. It is built and validated once, by init.
//
// This is the whole registry, and issue #45 Decision 8 is why there is nothing
// else in it. It used to be a table of Release_t descriptors carrying the
// map/@release, the map/@schema and the XML declaration version each release
// writes, with the schema selecting the codec. Every one of those is a byte the
// ENCODER writes, so the encoder owns it and declares it (see appver.Set_t), and
// carrying a second copy here would be the same class of bug as issue #45 itself:
// one value with two sources that can disagree. Stripped of them, a Release_t held
// nothing but its own key -- so the map IS the registry, and at heart it is a
// switch statement.
//
// The key is the exact map/@version string, never the parsed components.
// Dotted{Raw: "2.06"} and Dotted{Raw: "2.6"} carry identical components ({2, 6})
// but name different files, so keying on components would conflate them (ADR 0004
// Decision 1).
var byApp map[string]codec.Codec

// init builds the registry and refuses to load a program whose codec table is
// invalid.
//
// The table is a constant of the program, so anything wrong with it is a
// programming error rather than a runtime condition a caller could handle -- and
// what it would otherwise produce is a silently wrong file at encode time.
// Failing at load makes it unmissable.
//
// Three properties are checked. Every declaration must be valid on its own terms
// and no application version may be accepted by two codecs, which is
// appver.VerifyDisjoint; and every declared XML version must have a header, which
// is verifyXMLVersions. Both live outside init so that they stay testable, since a
// panic cannot be inspected.
//
// VerifyDisjoint is doing the work of TWO checks that issue #41 was careful to
// distinguish: the registry's duplicate-application-version guard (one
// application version must not name two RELEASES) and the codec disjointness
// guard (one application version must not be accepted by two CODECS). They were
// different statements only because the registry had releases in it to be
// ambiguous about. Now that the registry IS application version -> codec, they
// are the same statement, and this is the survivor -- reporting
// wxx.ErrAmbiguousAppCodec, the disjointness guard's own error. The duplicate
// guard's error constant went with the registry, having lost its producer.
func init() {
	all := codecs()
	sets := make([]appver.Set_t, 0, len(all))
	for _, c := range all {
		sets = append(sets, c.AcceptedApps())
	}
	if err := appver.VerifyDisjoint(sets...); err != nil {
		panic(fmt.Sprintf("xmlio: codec application versions: %v", err))
	}
	if err := verifyXMLVersions(sets...); err != nil {
		panic(fmt.Sprintf("xmlio: codec xml versions: %v", err))
	}
	byApp = make(map[string]codec.Codec, len(all))
	for i, c := range all {
		for _, a := range sets[i].Apps {
			byApp[a.Version] = c
		}
	}
}

// verifyXMLVersions returns an error unless every declaration names an XML
// version some header exists for.
//
// This is the guard NewRegistry ran over Release_t.XMLVersion, rescued rather
// than lapsed. It cannot live in appver with the rest of the declaration's checks
// -- appver would have to import xmlio to see the header table, and xmlio imports
// appver -- so it lives with the table it is about. init is where every codec is
// visible at once, so init is where it runs.
//
// It is caught at load rather than at encode: a codec that cannot say how its
// files open cannot write one, and finding that out mid-write is finding it out
// too late. It is a real check and not a formality -- "1.0" and "1.1" are the only
// XML declarations any Worldographer build writes, so a codec naming "1.2" is
// describing a file that does not exist.
func verifyXMLVersions(sets ...appver.Set_t) error {
	for _, s := range sets {
		if _, ok := utf16XMLHeader(s.XMLVersion); !ok {
			return errors.Join(wxx.ErrUnknownXMLHeader, fmt.Errorf("codec %s: xml version %q: no header", s.Codec, s.XMLVersion))
		}
	}
	return nil
}

// codecFor resolves the codec that writes the application version app.
//
// This is the dispatch, and it is the ONLY way to a codec: a caller names an
// application version and gets bytes, never an encoder (issue #41 requirement 5).
// Naming the identity IS naming the codec here, because one codec both accepts
// the version and writes the identity that goes with it -- which is what makes
// #41's chimera, W2025 content declaring the classic identity, unaskable.
//
// An unregistered version is an error, never a best-effort nearest match (ADR
// 0004 Decision 5). "" is not a sentinel: it names no application version and is
// the same error as any other, because a caller who passes one has named a target
// badly rather than declined to name one. The match is verbatim; see byApp.
func codecFor(app string) (codec.Codec, error) {
	if c, ok := byApp[app]; ok {
		return c, nil
	}
	return nil, errors.Join(wxx.ErrUnsupportedMapVersion, fmt.Errorf("version %q: not a supported application version", app))
}
