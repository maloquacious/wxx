# Architecture Decision Records

This directory holds Architecture Decision Records (ADRs) for the WXX package —
short documents capturing a significant architectural choice, the context that
forced it, the options weighed, and the consequences.

## Conventions

- Files are numbered and kebab-cased: `NNNN-short-title.md`.
- Each ADR carries a **Status**: `Proposed` → `Accepted` / `Rejected` (and later
  `Superseded by NNNN` if replaced).
- An ADR is a record, not a task. Where an ADR's decision authorizes code
  changes, the changes land as a separate, explicitly gated task.

## Index

| ADR | Title | Status |
|---|---|---|
| [0001](0001-codec-file-organization.md) | Codec file organization: co-located per-element encode/decode | Proposed — gates task B3b |
| [0002](0002-version-identity.md) | Version identity: `DataVersion` = `{familyYear, on-disk dotted revision}` | Superseded by 0004 (2026-07-15) |
| [0003](0003-version-axes.md) | Application version and schema version are independent axes | Proposed — gates #28 |
| [0004](0004-version-struct-and-release-registry.md) | Version identity: `{App, Schema}` plus a supported-release registry | Proposed — gates #28 |
