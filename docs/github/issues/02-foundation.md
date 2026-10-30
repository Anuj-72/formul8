## Context

The imported port has no README, no license, no CI and no issue or pull request templates, and its `.gitignore` covers only build output and editor noise. It is also a derivative of `txm`, which is dual-licensed MIT OR Apache-2.0, so preserving the upstream notices is part of correctness rather than optional polish. This issue turns the imported skeleton into a repository that another person could clone, build and understand.

## Scope

- [ ] Keep `/reference/` and `/.freebuff/` out of Git through the per-clone `.git/info/exclude`, and extend the committed `.gitignore` to cover local binaries, coverage output, profiles and editor noise.
- [ ] Add `LICENSE-MIT` and `LICENSE-APACHE` with the upstream copyright notice, plus a `NOTICE` that names the upstream project, URL and the commit used for the port.
- [ ] Write `README.md`: what formul8 is, its relationship to txm, the current parity status, build and install commands, CLI examples, the matrix extension and credits.
- [ ] Write `CONTRIBUTING.md` with branch naming, the one-issue-one-PR rule, Conventional Commits, the local check commands and the no-attribution rule.
- [ ] Add `docs/parity.md` with a table mapping every Rust module to its Go package and marking done, in progress or missing.
- [ ] Add `.github/workflows/ci.yml` with a job named `build-and-test` that runs `gofmt -l .`, `go vet ./...`, `go build ./...` and `go test ./...` on pushes and pull requests to `main`.
- [ ] Add `.github/ISSUE_TEMPLATE/bug_report.yml`, `.github/ISSUE_TEMPLATE/feature_request.yml` and `.github/PULL_REQUEST_TEMPLATE.md`.
- [ ] Add `.github/dependabot.yml` for `gomod` and `github-actions`, weekly.
- [ ] Add `docs/github/` with the workflow guide, the ruleset payload and the issue drafts so the process is reproducible.
- [ ] Set repository description, topics, squash-only merging, and the security features listed in the bootstrap guide.

## Acceptance criteria

- [ ] A fresh clone builds with `go build ./...` and tests with `go test ./...`.
- [ ] CI runs on a pull request and the check is displayed as `build-and-test`.
- [ ] `git ls-files` contains no file under `reference/` or `.freebuff/`.
- [ ] The README explains that formul8 is a port, links the upstream project and names the ported commit.
- [ ] Both upstream license files are present and a reader can tell which license applies to the port.
- [ ] `docs/parity.md` matches the actual state of the tree at merge time.

## Notes

The license choice should be dual MIT OR Apache-2.0 to match upstream and keep redistribution simple. The README should be honest about the state: at the time this lands the library builds but does not yet render, because the renderer and backends are later checkpoints. A status table is more useful than a badge that claims more than the code does.

Do not add a wiki. Documentation lives in `docs/` so it is reviewed in the same pull requests as the code.

## Dependencies

Depends on #1 for a green build. The files can be prepared in parallel, but the pull request should merge after the parser checkpoint so CI starts green.
