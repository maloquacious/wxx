# Agent

This project implements a Go package (github.com/maloquacious/wxx) to read and write Worldographer data files (also called "WXX files").

Worldographer is a map-generator written in Java. It stores data as XML in compressed (with GZip) and UTF-16 encoded (big-endian, I think) files.

There are two programs that create and update the files, both named Worldographer.

1. The original program is called "Worldographer." This program uses XML version 1.0 and doesn't contain an XML schema version in the file.

2. The newer program is called "Worldographer 2025." This version uses XML version 1.1 and stores the XML schema version as an attribute of the "map" entity.

We are working to implement a Sqlite3 data store. To do that, we need

1. Routines to detect the XML schema version in the WXX data file
2. Go structs that we can use to read and write the different versions of WXX files
3. A database schema
4. Routines to read and write the database data using Go structs

There is not much documentation available for the WXX files. We should create it as we go, taking care to track differences between the XML schema versions.

## Building
* REPL: `go build -o dist/local/repl ./cmd/repl`
* WXX runner: `go build -o dist/local/wxx ./cmd/wxx`
