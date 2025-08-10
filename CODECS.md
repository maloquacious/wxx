# Guiding principles

* **Don’t return an `io.Writer`.** In Go, you *write to* a provided `io.Writer`.
* **Decode returns a value (or fills one). Encode writes to `io.Writer`.**
* **Pipelines, not mega-interfaces.** Keep small, testable transforms you can compose.
* **Have high‑level `Encoder`/`Decoder` types with options; keep low‑level transforms as funcs.**

# Suggested public API

We define MapDecoder and MapEncoder interfaces to read and write Worldographer data.

```go
package wxx

type MapDecoder interface {
	Decode(io.Reader) (*Map_t, error)
}

type MapEncoder interface {
	Encode(io.Writer, *Map_t) error
}
```

# Suggested implementation of interfaces

We encourage the version specific decoders and encoders to implement the functions listed in this section.

## High‑level: what users actually call

```go
package h2017v

type Decoder struct {
    opts decoderOpts
}
type Encoder struct {
    opts encoderOpts
}

// New with functional options (version, strictness, compression, etc.)
func NewDecoder(opts ...DecoderOption) *Decoder
func NewEncoder(opts ...EncoderOption) *Encoder

// Decode reads a WXX stream and returns a Map_t.
func (d *Decoder) Decode(r io.Reader) (*Map_t, error)

// Variant that reuses a Map_t (helps pooling).
func (d *Decoder) DecodeInto(r io.Reader, m *Map_t) error

// Encode writes a WXX stream to w.
func (e *Encoder) Encode(w io.Writer, m *Map_t) error

// Convenience helpers for one‑shot use.
func ReadMap(r io.Reader, opts ...DecoderOption) (*Map_t, error)
func WriteMap(w io.Writer, m *Map_t, opts ...EncoderOption) error
```

### Options (examples)

```go
type DecoderOption func(*decoderOpts)
type EncoderOption func(*encoderOpts)

func WithCompression(enabled bool) DecoderOption       // expect gzip on input
func WithUTF16BEInput(enabled bool) DecoderOption      // expect UTF-16/BE
func WithFixXMLHeader(enabled bool) DecoderOption
func WithWorldographerVersion(v string) DecoderOption  // parse/validate schema

func WithGzipOutput(enabled bool) EncoderOption
func WithUTF16BEOutput(enabled bool) EncoderOption
func WithXMLHeader(version, encoding string) EncoderOption
func WithWorldographerSchema(v string) EncoderOption   // emit specific schema
```

That gives you a clean, user‑facing surface and keeps the mechanics behind options.

---

## Low‑level building blocks (composable transforms)

Expose tiny helpers that can be used independently and make unit tests dead simple.

```go
// Transform a stream step-by-step
func Gunzip(r io.Reader) (io.Reader, error)
func GzipWriter(w io.Writer) (io.WriteCloser, error)

func UTF16BEToUTF8(r io.Reader) (io.Reader, error)
func UTF8ToUTF16BE(w io.Writer) (io.WriteCloser, error)

// Ensure/patch XML header encoding value.
func FixXMLHeaderToUTF8(r io.Reader) (io.Reader, error)

// Parse/serialize the XML form of Map_t without transport concerns.
func UnmarshalXML(r io.Reader) (*Map_t, error)
func MarshalXML(w io.Writer, m *Map_t) error
func MarshalXMLBytes(m *Map_t) ([]byte, error) // handy for tests
```

Internally, `Decoder.Decode` becomes a small pipeline:

```go
func (d *Decoder) Decode(r io.Reader) (*Map_t, error) {
    var err error
    if d.opts.compressed { r, err = Gunzip(r); if err != nil { return nil, err } }
    if d.opts.utf16be    { r, err = UTF16BEToUTF8(r); if err != nil { return nil, err } }
    if d.opts.fixHeader  { r, err = FixXMLHeaderToUTF8(r); if err != nil { return nil, err } }
    // now r is clean UTF‑8 XML
    return UnmarshalXML(r)
}
```

And `Encoder.Encode` is the mirror image:

```go
func (e *Encoder) Encode(w io.Writer, m *Map_t) error {
    // start with UTF‑8 XML
    var buf bytes.Buffer
    if err := MarshalXML(&buf, m); err != nil { return err }

    var out io.Writer = w
    var closers []io.Closer

    if e.opts.utf16be {
        conv, err := UTF8ToUTF16BE(out)
        if err != nil { return err }
        out = conv; closers = append(closers, conv)
    }
    if e.opts.compressed {
        gz, err := GzipWriter(out)
        if err != nil { return err }
        out = gz; closers = append(closers, gz)
    }

    if _, err := io.Copy(out, &buf); err != nil { return err }
    for i := len(closers)-1; i >= 0; i-- { if cerr := closers[i].Close(); cerr != nil && err == nil { err = cerr } }
    return err
}
```

---

# Why your current interfaces feel off

* **Returning `io.Writer`/`io.Reader` from interfaces** (e.g., `Write(*Map_t) (io.Writer, error)`) flips control and complicates lifetime/cleanup. Idiomatic Go passes *you* a writer/reader; you do work and return an error.
* **Debug “BytesReader/BytesWriter” that return a stream plus an error** makes error handling ambiguous (“is the stream safe to read if error != nil?”). Prefer either:

    * return fully‑materialized `[]byte` for debug helpers, or
    * return a result + a **diagnostics object** (logs/steps) with `error` being authoritative.
* **Too many similar interfaces** (`Utf8Reader`, `Utf8XmlReader`, `CompressedReader`, etc.) fragment the API. Users really want: “Decode this” with knobs; you compose the steps.

---

# Minimal, pragmatic interface set

If you like things ultra‑lean:

```go
type MapDecoder interface {
    Decode(io.Reader) (*Map_t, error)
}
type MapEncoder interface {
    Encode(io.Writer, *Map_t) error
}
```

…and provide concrete `wxx.Decoder`/`wxx.Encoder` types that satisfy them. That keeps room for mocks in tests without forcing users through a maze of interfaces.

---

# Tiny usage examples

```go
dec := wxx.NewDecoder(
    wxx.WithCompression(true),
    wxx.WithUTF16BEInput(true),
    wxx.WithFixXMLHeader(true),
)

m, err := dec.Decode(src) // src: io.Reader
if err != nil { /* handle */ }

// …

enc := wxx.NewEncoder(
    wxx.WithGzipOutput(true),
    wxx.WithUTF16BEOutput(true),
    wxx.WithXMLHeader("1.0", "UTF-16"),
)
if err := enc.Encode(dst, m); err != nil { /* handle */ } // dst: io.Writer
```

---

# Optional: diagnostics without bending errors

If you want the “return intermediate data even on error” for debugging, add:

```go
type Diagnostics struct {
    Raw           []byte
    Uncompressed  []byte
    UTF8          []byte
    FixedHeader   []byte
}
func WithDiagnostics(buf *Diagnostics) DecoderOption
```

Then fill it as you go. You still return a proper `error`, but the caller can inspect what got produced before the failure.

---

**Bottom line:** collapse to `Decoder.Decode(io.Reader) (*Map_t, error)` and `Encoder.Encode(io.Writer, *Map_t) error`, hang the knobs off options, and expose the individual transforms as simple functions for unit tests. That’ll feel natural in Go and cleanly separates concerns.
