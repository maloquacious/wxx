# wxx

`wxx` is a Go package and command-line toolkit for reading, writing, inspecting, modifying, and scripting [Worldographer](https://worldographer.com/) files.

The name comes from `.wxx`, Worldographer’s default file extension.

The project provides:

* A Go API for working with Worldographer data
* Reading and writing `.wxx` files
* Inspecting and validating maps
* Modifying and generating map content
* Automating map-processing workflows
* Running embedded Lua scripts with [GopherLua](https://github.com/yuin/gopher-lua)
* Interactively scripting against loaded maps

## Go package

The `wxx` package can be imported into other Go applications that need to work with Worldographer files.

```go
package main

import (
	"fmt"
	"log"

	"github.com/maloquacious/wxx"
)

func main() {
	world, err := wxx.ReadFile("world.wxx")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(world)
}
```

The public API is intended to support parsing, inspecting, modifying, validating, and writing Worldographer data without requiring the command-line tool.

The exact package path and API are still evolving.

## Command-line tool

The `wxx` command exposes the package’s functionality through a subcommand-based interface.

```console
wxx info world.wxx
wxx validate world.wxx
wxx run script.lua world.wxx
wxx shell world.wxx
```

The exact commands and syntax are still evolving.

## Lua scripting

`wxx` embeds Lua using GopherLua, allowing scripts to inspect and modify Worldographer data.

```console
wxx run generate.lua world.wxx
```

Scripts can use the Lua API exposed by `wxx` to read map data, make changes, and write the resulting file.

An interactive shell can also be opened for a map:

```console
wxx shell world.wxx
```

This keeps `wxx` as the name of the overall project while treating scripting and the interactive shell as parts of a broader toolkit.

## Why `wxx`?

The name directly reflects the file format the project works with. It is short, distinctive, and suitable for both a Go package and a command-line executable.

Other names considered included `wsh`, short for “Worldographer shell.” However, `wsh` is already strongly associated with Microsoft Windows Script Host, and it describes only the scripting portion of the project.

Using `wxx` leaves room for the project to cover the full Worldographer workflow:

* Use the Go package in other applications
* Inspect or convert files from the command line
* Automate changes with Lua scripts
* Explore map data through an interactive shell

