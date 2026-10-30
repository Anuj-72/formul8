## Context

A milestone is a release contract, and the last checkpoint is making the contract real: a tagged commit, published notes, a verified install path and a closed milestone. This issue is deliberately mostly checklists; the only code change is whatever the release verification uncovers.

## Scope

- [ ] Confirm every `v0.1.0` issue is closed and the milestone shows 100%.
- [ ] Confirm `main` is green on the final commit and no pull request is open against the milestone scope.
- [ ] Update `docs/parity.md` so every reference module is marked done, with any intentional divergences listed.
- [ ] Update the README status table, install command and examples to match the tagged version.
- [ ] Write `docs/releases/v0.1.0.md` with sections for features, fixes, documentation and known limitations. Use one bullet per line.
- [ ] Add the version variable and `-ldflags` injection if it was not already added by the CLI checkpoint, and verify `--version` reports the tag.
- [ ] Verify `go install github.com/Anuj-72/formul8/cmd/formul8@v0.1.0` and run the binary from a clean shell.
- [ ] Run the conformance suite one final time and record the exact commands and results in the release notes.
- [ ] Tag `v0.1.0`, publish the GitHub release and close the milestone.

## Acceptance criteria

- [ ] The tag points at a commit on `main` whose CI run is green.
- [ ] The release notes list the supported syntax, the install command and the known divergences from the reference.
- [ ] The install command works outside the repository and `formul8 --version` prints `formul8 v0.1.0`.
- [ ] The milestone is closed and every issue in it is closed.
- [ ] `docs/parity.md` and the README do not claim anything the released binary cannot do.

## Notes

Do not tag from a dirty tree, from a branch, or before the release pull request is merged. The tag is the first immutable artifact in the project; if a fix is needed after it, release `v0.1.1` rather than moving the tag. The `v0.2.0` matrix milestone stays open.

## Dependencies

Depends on #1 through #9. Blocks nothing; it closes the milestone.
