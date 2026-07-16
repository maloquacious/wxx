# The codec packages: what the number in the path means

```
xmlio/internal/
├── appver/   # the shape of a codec's accepted-application-version declaration,
│             #   plus the check that no two codecs claim the same version
├── codec/    # the schema -> codec table and selector (ForSchema)
├── v0_77/    # codec version 0.77 -- classic
└── v1_06/    # codec version 1.06 -- W2025 (its DECODER is a work in progress)
```

`v0_77` and `v1_06` are **codec versions expressed as package paths**. They are
the same kind of name: neither is a family year, neither is special-cased, and
neither is the set of application versions the codec accepts. That last point is
the one this document exists for, so it comes first.

---

## CAUTION: the path does not tell you which application versions a codec accepts

**`v0_77` is not the 1.77 codec. It accepts `1.73`, `1.74` *and* `1.77`.**

All three classic builds wrote an identical element vocabulary — one schema, so
one codec serves all three (ADR 0004 Decision 2). The `.77` in the path is a
mnemonic, not a set. Nobody needs to write a `v0_73`, and if you conclude from
the path that somebody does, the path has misled you exactly the way this section
is here to prevent.

The set is the **codec's own knowledge**, and it lives inside the codec:

| codec | its declaration | what `Encode` checks |
|---|---|---|
| `v0_77` | [`v0_77/apps.go`](v0_77/apps.go) | [`v0_77/encode.go`](v0_77/encode.go) — `acceptedApps.VerifyApp(app)`, before a byte is emitted |
| `v1_06` | [`v1_06/apps.go`](v1_06/apps.go) | [`v1_06/map.go`](v1_06/map.go) — same check, same place |

**Read those two files. Do not infer the set from the path, and do not trust the
sets restated in this README** — a README drifts, `apps.go` is what `Encode`
actually enforces. Where they disagree, `apps.go` is right and this file is
stale.

Two properties hold across the codecs, and both are enforced rather than assumed:

- **An application version is accepted by no more than one codec.**
  `appver.VerifyDisjoint` ([`appver/appver.go`](appver/appver.go)) checks it over
  every codec's declaration, and `xmlio/codecs.go` runs it at `init` and panics
  on an overlap. The codec set is a constant of the program, so an overlap is a
  programming error, and the failure it would otherwise produce is a silently
  wrong file.
- **An application version may be accepted by no codec at all.** A build nothing
  supports is named by no set, and that is not an error — it is a rejection. As
  of this writing `v1_06` accepts `2.06`, the supported W2025 baseline, and
  nothing else; a hypothetical future `2.07` on the same schema would be **added
  to `v1_06/apps.go`, not given a package**.

The sets gate **encoding**. `Decode` does not consult them: the decoder in
[`../decoder.go`](../decoder.go) routes on the file's `map/@release` attribute,
with a conservative `1.x` check on `map/@version` for classic so that an unknown
format is not swallowed.

---

## 1. The package path is the codec version

A codec version is *our* identifier for a parse/emit pair. It is not on disk, it
is not a Go module version, and it is not the application build that wrote the
file. It names the code.

The `_` stands in for the `.`: a Go package name is an identifier and cannot
contain one. Read `v0_77` as codec version `0.77` and `v1_06` as codec version
`1.06`.

## 2. The convention: the codec version matches the schema the file states

`v1_06` implements the schema a file states as `schema="1.06"`. That is the
whole convention, and [`codec/codec.go`](codec/codec.go)'s table is where it
becomes mechanical: a schema string selects exactly one codec.

| codec version | the schema a file states | application versions accepted |
|---|---|---|
| `v0_77` | *none* — classic files carry no `map/@schema` at all | declared in [`v0_77/apps.go`](v0_77/apps.go) |
| `v1_06` | `schema="1.06"` | declared in [`v1_06/apps.go`](v1_06/apps.go) |

The schema is the format's identity: it is on disk, and it does not change when
the product is relabelled, so it — not the application version, and not
`map/@release` — is what answers "which code path emits this" (ADR 0004
Decision 4).

Note the middle column is a **schema**, and the left column is a **codec
version**. They are numerically equal for `v1_06` by convention, and for `v0_77`
they are not comparable at all, because there is no schema to compare to.

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
this README's standing job. Until then the mapping is the identity function for
`v1_06`, which is why it is easy to mistake the convention for a definition.

## 4. Why classic is `v0_77`

Classic files state **no** schema. There is no on-disk value to name the codec
after, so a name had to be chosen — and `1.00` was rejected, because it is shaped
exactly like a real schema version and would sit among genuine ones.

`0.77` was chosen instead. The **`0` major is the load-bearing part**; the `.77`
is only a mnemonic from `1.77`, the last classic application version.

The `0` buys two properties:

- **It cannot be mistaken for a real on-disk value.** Every Worldographer schema
  version examined is `1.x` (`1.01`, `1.06` — ADR 0003), so `0.x` is a range no
  Worldographer schema occupies.
- **It orders correctly for free.** `wxx.Dotted.Compare` puts `0.77` before
  `1.06`, matching classic-precedes-W2025, with no special case.

### `0.77` is a codec version, NOT a schema version

**No file states `0.77`, and none may.** It must never reach `Release_t.Schema`,
`Version_t.Schema`, `Map_t.Schema`, or a file. Three things keep that honest, and
they are worth knowing because the name is otherwise an invitation:

- **The registry guard.** `NewRegistry` ([`../registry.go`](../registry.go))
  enforces `(Schema == nil) iff (Release == "")` — a release states a schema if
  and only if it states a release. A classic entry carrying `Schema="0.77"`
  states a schema without a release and **fails at `init`**.
- **The absent schema is the identity.** `Version_t.Schema == nil` remains the
  model of classic's missing `@schema` (ADR 0004 Decision 2): `nil` is meaningful
  and identifies the one implicit legacy schema. `v0_77` names the *codec*; `nil`
  names the *identity*. Accordingly `v0_77/apps.go` declares `Schema: ""`, not
  `"0.77"`, and the `Schema` column of `codec.table()` has no `0.77` in it:
  nothing selects a codec by that string, and `codec.ForSchema("0.77")` is an
  unsupported-schema error.
- **Verbatim output is inviolable** (ADR 0002 Decision 2, reinforced by ADR 0004
  Decision 1). We never synthesize a schema that was not on input. The classic
  encoder hard-codes its attribute list and emits no `@schema` — that stays. And
  nothing written to disk is ever re-rendered from a `Dotted`'s components:
  `Raw` is authoritative, or `"2.06"` goes out as `"2.6"` and names a different
  file.

## Why these packages are under `internal/`

`xmlio/internal/` is a visibility boundary, and its position is the point: only
the dispatcher directly above it may pick a codec (issue #41 requirement 5). A
caller naming a codec can pair any codec with any identity — #41 documents what
that buys, W2025 content emitted under a classic identity, which then re-decodes
silently as classic.

Root-level `internal/` would have been looser: it would still let `cmd/*` tools
call an encoder directly, and those are precisely the callers the rule exists to
stop. The exception survives by construction — Go's internal rule is
directory-based, so `xmlio`'s external test package (`package xmlio_test`, but
physically inside `xmlio/`) may still import these packages and choose an
encoder. See the package comment on [`codec/codec.go`](codec/codec.go).

## Adding a codec

1. Add the package, named for the schema it implements per section 2.
2. Declare its accepted application versions and the single schema it writes, in
   its own `apps.go`.
3. Add it to `codec.table()` so the schema selects it.
4. Add it to `codecAppSets()` in `xmlio/codecs.go`. **A codec missing from that
   list is not checked against the others**, which is the one way the
   disjointness guarantee is lost.

Supporting a further application version on an **existing** schema is **not a new
package**: it is an entry in that codec's `apps.go` and one in the release
registry ([`../registry.go`](../registry.go)). See the caution above.

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
  — `Dotted`, `{App, Schema}`, the supported-release registry, and "schema
  selects the codec; the application version is data".
- [`docs/adr/0002-version-identity.md`](../../docs/adr/0002-version-identity.md)
  — the verbatim-output guarantee, which survives 0004 unchanged.
- [`v0_77/COVERAGE.md`](v0_77/COVERAGE.md), [`v1_06/COVERAGE.md`](v1_06/COVERAGE.md)
  — per-element read/write coverage for each codec.
