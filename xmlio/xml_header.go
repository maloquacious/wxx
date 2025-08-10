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
