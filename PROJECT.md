# Project Structure

wxx/
├── mapio/                    # packages for reading and writing Map data
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
    ├── wxxdemo/              # command showing package use
    │   └── main.go
    ├── wxxinfo/              # command showing information on XML data
    │   └── main.go
    ├── wxxschema/            # command to extract schema from XML data
    │   └── main.go
    └── ...                   # commands that operate on Map

