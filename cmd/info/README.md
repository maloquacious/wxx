# info
`info` is a command line tool to print out information on a `.wxx` file.

## Usage

```bash
go run ./cmd/info testdata/2025-2.06-13x11-941577-blank.wxx

info:	testdata/2025-2.06-13x11-941577-blank.wxx
	 h2025v1 schema version "2025.1.6"
	      11 tiles high
	      13 tiles wide
	       1 terrain tiles defined
```

Accepts multiple files. The reported schema version is the decoded
`MetaData.DataVersion`, not the map's raw `version` attribute.
