# The codec packages: what the number in the path means

```
xmlio/internal/
├── appver/   # the shape of a codec's accepted-application-version declaration,
│             #   plus the check that no two codecs claim the same version
├── codec/    # the Codec interface: what the dispatcher holds for each codec
└── v1_06/    # codec version 1.06 -- W2025 (its DECODER is a work in progress)
```

`v1_06` is a **codec version expressed as a package path**. It is not a family
year, and it is not the set of application versions the codec accepts. That last
point is the one this document exists for, so it comes first.

There is one codec. The classic codec, `v0_77`, was removed by issue #103: wxx
reads and writes W2025 only, and the decoder refuses a classic (Worldographer
1.x) file with `wxx.ErrClassicMap`, telling the user to convert it in
Worldographer 2025 first.

---

## CAUTION: the path does not tell you which application versions a codec accepts

**`v1_06` is not "the 1.06 build". `1.06` is the schema it writes; the
application versions it accepts are a separate list.**

The set is the **codec's own knowledge**, and it lives inside the codec:

| codec | its declaration | what `Encode` checks |
|---|---|---|
| `v1_06` | [`v1_06/apps.go`](v1_06/apps.go) | [`v1_06/map.go`](v1_06/map.go) — `acceptedApps.VerifyApp(app)`, before a byte is emitted |

**Read those files. Do not infer the set from the path, and do not trust a set
restated in this README** — a README drifts, `apps.go` is what `Encode` actually
enforces. Where they disagree, `apps.go` is right and this file is stale.

Two properties hold across the codecs, and both are enforced rather than assumed:

- **An application version is accepted by no more than one codec.**
  `appver.VerifyDisjoint` ([`appver/appver.go`](appver/appver.go)) checks it over
  every codec's declaration, and `xmlio/codecs.go` runs it at `init` and panics
  on an overlap. With one codec it guards only against that codec naming a
  version twice, but it is the check a second codec would be held to.
- **An application version may be accepted by no codec at all.** A build nothing
  supports is named by no set, and that is not an error — it is a rejection. A
  later build on the same schema is **added to `v1_06/apps.go`, not given a
  package**.

The sets gate **encoding**. `Decode` does not consult them: the decoder in
[`../decoder.go`](../decoder.go) routes on the file's `map/@release` attribute.
`release="2025"` goes to `v1_06`; a file with no `@release` is refused, as a
classic map when its `@version` is `1.x` and as unsupported metadata otherwise.

---

## 1. The package path is the codec version

A codec version is *our* identifier for a parse/emit pair. It is not on disk, it
is not a Go module version, and it is not the application build that wrote the
file. It names the code.

The `_` stands in for the `.`: a Go package name is an identifier and cannot
contain one. Read `v1_06` as codec version `1.06`.

## 2. The convention: the codec version matches the schema the file states

`v1_06` implements the schema a file states as `schema="1.06"`. That is the
whole convention.

| codec version | the schema a file states | application versions accepted |
|---|---|---|
| `v1_06` | `schema="1.06"` | declared in [`v1_06/apps.go`](v1_06/apps.go) |

The schema names the codec; it does not select it. The encoder is selected by
the **application version** the caller names: `xmlio/codecs.go` builds its
registry by asking each codec which application versions it accepts (issue #45
Decision 8), and the codec writes the schema it declares.

## 3. Why it is a convention and not a definition

The codec version is ours. Mirroring Inkwell's schema numbering is deliberate —
it makes the path a fact you can check against a file instead of a code name you
have to look up — but it is **not binding**.

Binding the path to a value Inkwell controls would leave us stuck if they ever
change their numbering conventions: the model would have to be renamed to follow
a decision made somewhere else, about something that is not the model. Keeping
the codec version as our own identifier that *currently* mirrors theirs lets us
pivot without renaming anything.

**If the two ever diverge, the table in section 2 carries the mapping.** That is
this README's standing job.

## Why these packages are under `internal/`

`xmlio/internal/` is a visibility boundary, and its position is the point: only
the dispatcher directly above it may pick a codec (issue #41 requirement 5). A
caller naming a codec could pair any codec with any identity — #41 documents what
that bought, W2025 content emitted under a classic identity.

Root-level `internal/` would have been looser: it would still let `cmd/*` tools
call an encoder directly, and those are precisely the callers the rule exists to
stop. The exception survives by construction — Go's internal rule is
directory-based, so `xmlio`'s external test package (`package xmlio_test`, but
physically inside `xmlio/`) may still import these packages and choose an
encoder. See the package comment on [`codec/codec.go`](codec/codec.go).

## Adding a codec

1. Add the package, named for the schema it implements per section 2.
2. Declare its accepted application versions and the single schema it writes, in
   its own `apps.go`, and give it a `Codec_t` that implements `codec.Codec`.
3. Add it to `codecs()` in `xmlio/codecs.go`. **A codec missing from that list
   is not checked against the others and is not reachable.**
4. Give its schema an arm in `downgradeLoss` (`xmlio/downgrade.go`) stating what
   that schema cannot express. A schema with no arm is an encode error, never a
   silent "no loss".

Supporting a further application version on an **existing** schema is **not a new
package**: it is an entry in that codec's `apps.go`. See the caution above.

## Not this: encoder ordinals

An earlier proposal numbered the encoders (`/v1`, `/v2`) and mapped those onto
schema versions. It is **rejected** (#41): a `/v2` path element means a module
major version to Go and to every Go reader; it invents a third version axis with
no on-disk referent, after two ADRs spent their length removing exactly one such
coinage; and its premise — that schema versions are strictly additive — is an
assumption we have not verified. A codec version mirrors a real schema version.
An ordinal would only have counted the encoders we happened to write.

## See also

- [`docs/adr/0004-version-struct-and-release-registry.md`](../../docs/adr/0004-version-struct-and-release-registry.md)
  — `Dotted`, `{App, Schema}`, and the supported-release registry.
- [`docs/adr/0002-version-identity.md`](../../docs/adr/0002-version-identity.md)
  — the verbatim-output guarantee, which survives 0004 unchanged.
- [`v1_06/COVERAGE.md`](v1_06/COVERAGE.md) — per-element read/write coverage.
