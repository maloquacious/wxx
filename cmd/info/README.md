# info
`info` is a command line tool to print out information on a `.wxx` file.

## Usage

```bash
go run ./cmd/info testdata/input/blank-2017-1.77-1.0.wxx testdata/input/blank-2025-1.10-1.01.wxx 

testdata/input/blank-2017-1.77-1.0.wxx
	   10017 bytes on disk
	   10017 bytes compressed
	   98910 bytes uncompressed
	   98910 bytes utf-16/be encoded
	   49454 bytes utf-8 encoded
	xml header "<?xml version='1.0' encoding='utf-16'?>\n"
	   49414 bytes xml data
	H2017: version 1.77
testdata/input/blank-2025-1.10-1.01.wxx
	    5324 bytes on disk
	    5324 bytes compressed
	   52358 bytes uncompressed
	   52358 bytes utf-16/be encoded
	   26178 bytes utf-8 encoded
	xml header "<?xml version='1.1' encoding='utf-16'?>\n"
	   26138 bytes xml data
	W2025: version 1.10: schema 1.01
```