GO_MOD_FILES := go.mod go.sum go.work go.work.sum

# populate the source file lists by package
GO_ROOT_SOURCES := $(shell find . -maxdepth 1 -name "*.go")
GO_DSL_SOURCES := $(shell find dsl -name "*.go")
GO_READERS_SOURCES := $(shell find readers -name "*.go")
GO_MODELS_SOURCES := $(shell find models -name "*.go")
GO_ADAPTERS_SOURCES := $(shell find adapters -name "*.go")
GO_WRITERS_SOURCES := $(shell find writers -name "*.go")
GO_XMLIO_SOURCES := $(shell find xmlio -name "*.go")
GO_GZUTF16_SOURCES := $(shell find gzutf16 -name "*.go")
GO_XML100_SOURCES := $(shell find xml_1_0_0 -name "*.go")

# common sources used by both commands
GO_COMMON_SOURCES := $(GO_ROOT_SOURCES) $(GO_DSL_SOURCES) $(GO_READERS_SOURCES) $(GO_MODELS_SOURCES) $(GO_ADAPTERS_SOURCES) $(GO_WRITERS_SOURCES) $(GO_XMLIO_SOURCES) $(GO_GZUTF16_SOURCES) $(GO_XML100_SOURCES)

# commands
REPL_BIN := dist/local/repl
REPL_SOURCES := $(GO_COMMON_SOURCES) cmd/repl/main.go
WXX_BIN := dist/local/wxx
WXX_SOURCES := $(GO_COMMON_SOURCES) cmd/wxx/main.go

#

.PHONY: all clean tools

all: tools

tools: $(REPL_BIN) $(WXX_BIN)
	@echo "all tools built"

# build steps

$(REPL_BIN): $(REPL_SOURCES) $(GO_MOD_FILES)
	@mkdir -p dist/local
	go build -o $@ ./cmd/repl

$(WXX_BIN): $(WXX_SOURCES) $(GO_MOD_FILES)
	@mkdir -p dist/local
	go build -o $@ ./cmd/wxx

repl: $(REPL_BIN)

wxx: $(WXX_BIN)

clean:
	rm -rf dist/*
