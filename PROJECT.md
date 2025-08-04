# Project Structure

wxx/
├── xmlio/                    # packages for reading and writing XML data
│   ├── maploader/            # translates XML into Map
│   ├── v1_0/                 # schema-specific readers/writers for version 1.0.*
│   │   ├── reader.go
│   │   └── writer.go
│   ├── v2_0/                 # schema-specific readers/writers for version 2.0.*
│   │   ├── reader.go
│   │   └── writer.go
│   └── dispatcher.go         # delegates actions to correct version of reader/writer
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

