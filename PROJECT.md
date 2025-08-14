# Project Structure

wxx/
├── xmlio/                    # packages for decoding and encoding XML data
│   ├── h2017v1/              # schema-specific decoders and encoders
│   ├── h2025v1/              # schema-specific decoders and encoders
│   └── h2025v1/              # schema-specific decoders and encoders
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

