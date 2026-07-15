// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package wxx

// Version_t is a file's on-disk version identity: the two independent axes a
// Worldographer file states (ADR 0003) and nothing else.
//
// There is deliberately no family year. "2017" is a project coinage that appears
// in no classic file, and "2025" is a marketing label that may change while the
// data format does not; neither is a fact about the format, so neither belongs
// in a version identity (ADR 0004). The marketing label survives verbatim in
// Map_t.Release and MetaData.Worldographer.Release, where it is fidelity data
// rather than identity.
//
// Both members are Dotted, never semver: the components exist to compare and Raw
// is what goes back to disk.
type Version_t struct {
	// App is map/@version, the application build that wrote the file: "1.73",
	// "1.74" or "1.77" for classic, "2.06" for the W2025 baseline. Every
	// supported file states it, so it is a value rather than a pointer.
	App Dotted

	// Schema is map/@schema, the on-disk data format the file conforms to
	// ("1.06" for the W2025 baseline).
	//
	// nil is meaningful, and it does not mean "unknown": it identifies the one
	// implicit legacy (classic) schema, which states no @schema attribute at
	// all. Classic 1.73, 1.74 and 1.77 share an identical element vocabulary,
	// so the absence names a single schema rather than leaving a question open.
	Schema *Dotted
}
