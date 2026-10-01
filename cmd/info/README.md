# info
`info` is a command line tool to print out information on a `.wxx` file.

## Usage

```bash
go run ./cmd/info testdata/2025-2.06-13x11-941577-blank.wxx

info:	testdata/2025-2.06-13x11-941577-blank.wxx
	   v1_06 codec: app 2.06, schema 1.06
	      11 tiles high
	      13 tiles wide
	       1 terrain tiles defined
```

Accepts multiple files. The first column is the codec package that decoded the
file (`v1_06`, the only codec), followed by the two version axes the file states
(`MetaData.Version`): the application version from `map/@version` and the schema
version from `map/@schema`.

wxx no longer reads classic (Worldographer 1.x) files (issue #103). `info`
prints the decoder's refusal for one and moves on to the next file:

```bash
go run ./cmd/info testdata/2017-1.77-1.0-columns-blank.wxx

info:	testdata/2017-1.77-1.0-columns-blank.wxx
	classic (Worldographer 1.x) map: convert it in Worldographer 2025 first
map: version "1.77": no release
```
