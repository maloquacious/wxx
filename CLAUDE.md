# CLAUDE.md

Read [AGENTS.md](./AGENTS.md) first. It is the project guide: the format
background, repository layout, codec and CLI conventions, coding conventions,
build commands and validation. This file adds only the rules below.

## Bugs Before Features

**Open bugs outrank feature work.** When choosing what to do next, or when asked to start a
feature while bugs labeled `bug` are open, say so and propose the bugs first. Do not begin
feature work on the assumption that the bug list will be dealt with later — that is how the
identity bugs (#28, #32, #41, #45) accumulated behind each other.

This is a rule about **what to work on next**, not a hard interlock on every commit:

- A bug fix, a test that pins a bug, a doc correction, or a refactor in service of a bug fix
  is never blocked.
- Feature work proceeds only when the maintainer says so explicitly, having been told what is
  still open. "Do the feature anyway" is a legitimate answer — the requirement is that it is a
  decision made with the bug list in view, not one made by default.
- Do not use "it is unrelated to the open bugs" as self-authorization. Independence is an
  argument to put to the maintainer, not a reason to skip asking.

When proposing bugs, rank them by whether they can produce a wrong file on disk. A bug that
writes silently-wrong output (e.g. #45) outranks one that loses a field (#35) or misnames a
diagnostic (#44).

## Tag Every Version Change

When a merge to `main` changes `version.go`, tag that commit `v<version>` (lightweight) and
push the tag. See CONTRIBUTING.md step 7.

## Reasoning About the File Format

When a claim is about what a WXX file contains, cite `schema/1.06.rnc` (the pattern
and its `# observed` / `# inferred` label) and check the fixture bytes
(`gunzip -c f.wxx | iconv -f UTF-16BE -t UTF-8`). Do not infer the format from Go
structs: `Map_t` and the codec's schema types show what wxx models, not what the file
holds. `# inferred` and `TBD` rules are hypotheses. See `schema/README.md`.
