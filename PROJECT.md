# Project Structure

```
wxx/
├── schema/                   # reference grammars, one per schema version (1.06.rnc); see schema/README.md
├── xmlio/                    # packages for decoding and encoding XML data
│   └── internal/            # codec packages; unimportable outside xmlio/
│       ├── appver/          # a codec's accepted-application-version declaration
│       ├── codec/           # the Codec interface the dispatcher holds
│       └── v1_06/           # W2025 (schema 1.06) decoder and encoder
└── cmd/
    ├── copy/                 # tool to copy a Worldographer file
    │   └── main.go
    ├── info/                 # tool to show information on XML data
    │   └── main.go
    ├── schema/               # tool to extract schema from XML data
    │   └── main.go
    ├── version/              # tool to show package version
    │   └── main.go
    └── ...                   # commands that operate on Map
```