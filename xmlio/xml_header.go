// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package xmlio

var (
	// table of XML headers that we can accept
	xmlHeaders = []struct {
		heading  string
		version  string
		encoding string
	}{
		{heading: "<?xml version='1.0' encoding='utf-8'?>\n", version: "1.0", encoding: "utf-8"},
		{heading: "<?xml version='1.0' encoding='utf-16'?>\n", version: "1.0", encoding: "utf-16"},
		{heading: "<?xml version='1.1' encoding='utf-8'?>\n", version: "1.1", encoding: "utf-8"},
		{heading: "<?xml version='1.1' encoding='utf-16'?>\n", version: "1.1", encoding: "utf-16"},
	}
)

// utf16XMLHeader returns the UTF-16 XML declaration for an XML version ("1.0",
// "1.1"), reporting whether one exists.
//
// The version comes from the target release (Release_t.XMLVersion): the XML
// declaration is on-disk identity data bound to a release, not something to
// infer from the map. Worldographer writes UTF-16, so that is the only encoding
// this offers -- the utf-8 rows of the table exist for reading files a tool has
// already transcoded.
func utf16XMLHeader(xmlVersion string) (string, bool) {
	for _, h := range xmlHeaders {
		if h.version == xmlVersion && h.encoding == "utf-16" {
			return h.heading, true
		}
	}
	return "", false
}
