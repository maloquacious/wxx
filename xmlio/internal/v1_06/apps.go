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
var acceptedApps = appver.Set_t{
	Codec:  "v1_06",
	Schema: "1.06",
	Apps:   []string{"2.06"},
}

// AcceptedApps returns this codec's declaration of the application versions it
// accepts and the schema it writes. The returned set is the caller's own copy.
func AcceptedApps() appver.Set_t {
	return acceptedApps.Clone()
}
