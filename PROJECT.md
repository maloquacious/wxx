# Project Structure

```
wxx/
├── schema/                   # RelaxNG schema (reference only, v1.73/classic; not enforced)
├── xmlio/                    # packages for decoding and encoding XML data
│   └── internal/            # codec packages; unimportable outside xmlio/
│       ├── v0_77/           # schema-specific decoders and encoders
│       └── v1_06/           # schema-specific decoders and encoders
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