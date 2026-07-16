// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package v1_06

import "github.com/maloquacious/wxx/xmlio/internal/appver"

// acceptedApps declares what this codec accepts and what it writes.
//
// 2.06 is the supported W2025 baseline, the first post-beta build, and today it
// is the only application version on schema 1.06 that this codec accepts. The set
// is the codec's own knowledge and must be read here rather than inferred from
// the package path: the path is the CODEC version, which by convention matches
// the schema the file states, and it says nothing about which builds wrote that
// schema. A later build on the same schema is added here, not by adding a
// package.
//
// 2.06 writes release="2025". That is stated per application version rather than
// as a constant of the codec because map/@release is DERIVED from the application
// version (issue #45 Decision 5): a later build on schema 1.06 shipped under a
// different label writes a different map/@release and is added as another App_t
// here. Today there is exactly one W2025 build, which makes the mapping look
// constant. It is not one, and collapsing it into one would hard-code away the
// relabel scenario ADR 0004 exists to keep expressible.
//
// XMLVersion is "1.1", the declaration 2.06 opens its files with.
var acceptedApps = appver.Set_t{
	Codec:      "v1_06",
	Schema:     "1.06",
	XMLVersion: "1.1",
	Apps: []appver.App_t{
		{Version: "2.06", Release: "2025"},
	},
}

// AcceptedApps returns this codec's declaration of the application versions it
// accepts and the schema it writes. The returned set is the caller's own copy.
func AcceptedApps() appver.Set_t {
	return acceptedApps.Clone()
}
