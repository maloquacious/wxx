# info
`info` is a command line tool to print out information on a `.wxx` file.

## Usage

```bash
go run ./cmd/info testdata/2025-2.06-13x11-941577-blank.wxx

info:	testdata/2025-2.06-13x11-941577-blank.wxx
	 h2025v1 codec: app 2.06, schema 1.06
	      11 tiles high
	      13 tiles wide
	       1 terrain tiles defined
```

Accepts multiple files. The first column is the codec that decoded the file,
followed by the two version axes it states (`MetaData.Version`): the application
version from `map/@version` and the schema version from `map/@schema`. A classic
file states no schema at all, and reports `schema implicit (classic)`:

```bash
go run ./cmd/info testdata/blank-2017-1.77-1.0.wxx

info:	testdata/blank-2017-1.77-1.0.wxx
	 h2017v1 codec: app 1.77, schema implicit (classic)
	       5 tiles high
	       3 tiles wide
	       1 terrain tiles defined
```
