# Agent

This project implements a Go package (github.com/maloquacious/wxx) to read and write Worldographer data files (also called "WXX files").

## Worldographer
Worldographer is a map-generator written in Java.
It stores data as XML in compressed (with GZip) and UTF-16 encoded (big-endian, I think) files.

There are two programs that create and update the files, both named Worldographer.

1. The original program is called "Worldographer."
This program uses XML version 1.0 and doesn't contain an XML schema version in the file.

2. The newer program is called "Worldographer 2025."
This version uses XML version 1.1 and stores the XML schema version as an attribute of the "map" entity.

We will create decoders and encoders under the `xmlio/` path to read and write the different versions of Worldographer data files.

### Documentation
There is not much documentation available for the WXX files.
We should create it as we go, taking care to track differences between the XML schema versions.

## Go data structures
Map_t is the Go struct that we store the Worldographer data in.
There are multiple versions of Worldographer data, so our Map_t is a superset of that data.

The xmlio decoders target the Map_t structure; the encoders use it as their source. 

## Sqlite3
Implementing a Sqlite3 data store is on the roadmap and will be scheduled after we complete the xmlio decoders and encoders.

For the data store, we will need to create a database schema and routines to load and store Map_t into the data store.

## Building
We have tools in the `cmd/` path that we use for testing this package.

* Build a tool: `go build -o dist/local/version ./cmd/version`
* Run a tool: `dist/local/version`
