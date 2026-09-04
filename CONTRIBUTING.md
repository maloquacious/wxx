# Contributing

This describes how work actually lands in this repository. Until now the process
existed only in the git log, discoverable by reading it; this file is the record.

## What to work on next

**Open bugs outrank feature work.** See the "Bugs Before Features" section of
`CLAUDE.md` for the full rule. Rank open bugs by whether they can produce a
wrong file on disk: a bug that writes silently-wrong output outranks one that
loses a field, which outranks one that misnames a diagnostic.

## The path

### 1. An issue first

Every change traces to an issue. The issue is what the branch, the commit
subject, and the PR all name, and it is what closes when the work merges.

Assign it to yourself at creation:

```sh
gh issue create --assignee @me ...
```

### 2. A branch named for the issue

```
<kind>/issue-<number>-<slug>
```

`kind` is `fix`, `docs`, or `feature`. Examples from the history:

```
fix/issue-35-label-dropshadow
docs/issue-28-version-axes
feature/issue-32-version-registry
```

Never commit directly to `main` except for the rare repository-keeping change
(this file, a CLAUDE.md rule) that has no issue behind it.

### 3. One commit per pull request

The branch carries a single commit. Its message is the durable explanation of
the change — the PR description is a copy, and the PR itself will be squashed
away.

```
#<issue>: <lowercase imperative summary, no trailing period>

<Prose. What was broken, and how. What changed, and why that shape. What
proves it — a green suite is not proof. Cite the evidence: fixture hashes
that did or did not move, guards broken deliberately and watched fail,
byte counts. Name what you deliberately did not do.>

Closes #<issue>.

Co-Authored-By: <agent, if one worked on this>
Claude-Session: <session URL, if applicable>
```

`git log fe4b432` (#45) and `git log 63def2e` (#35) are the worked examples.
The body is expected to be long. Prefer explaining the reasoning that is not
recoverable from the diff over restating the diff.

### 4. Verify locally

There is no CI. No `.github/` workflows, no automated pipeline — nothing runs
these checks but you.

```sh
go test ./...
go vet ./...
go fmt ./...
```

A passing suite is the floor, not the proof. For codec work, the standard the
repository holds itself to is a fixture-hash baseline: capture the encode of
every fixture before the first commit, re-verify after every increment, and
state in the commit message exactly which bytes moved and why. Where a guard is
added, break it deliberately and watch it fail — #45 records a real coverage
hole found precisely that way, one the hash baseline structurally could not
detect.

### 5. Open the pull request

```sh
gh pr create --assignee @me ...
```

The body repeats `Closes #<issue>.` so the issue closes on merge. Review is
local: no PR in the recent history carries a GitHub review, and none is
expected. `/code-review` is available for a second pass before you merge.

### 6. Squash merge, delete the branch

```sh
gh pr merge <number> --squash --delete-branch
```

Squash is the rule. `main` stays linear, each commit subject gaining the
`(#NN)` suffix GitHub appends. The repository has `deleteBranchOnMerge` off, so
`--delete-branch` is what keeps the branch list clean; afterwards
`git fetch --prune` clears the stale local ref.

## Reading the older history

The merge style changed. Pull requests through #27 landed as merge commits
(`Merge pull request #27 from ...`, with the branch commit preserved beneath).
From #29 onward — 2026-07-15 — every PR is squashed, which is why `main` is a
straight line above `c5d0490` and a braided one below it.

Both halves are legitimate history. Only the squash half is the current
convention; do not take the older shape as a model.

## Code conventions

These live in `CLAUDE.md` and are summarized here only as a pointer: copyright
header on every `.go` file, package doc comments, `_t` suffix on major types,
the standard `testing` package and no external test frameworks, and a hard
preference for adding no dependency that can be avoided. The list is currently
three — `semver`, `golang.org/x/text`, and `ff/v4`, the last of which goes when
the commands collapse into the Lua script host.

Fixtures a test reads belong in `testdata/`, flat and tracked, so the suite runs
from a clean clone. `scratch/` and `dist/` are git-ignored local output and must
never hold anything a test needs.
