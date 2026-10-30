# Contributing

formul8 is a small project with a strict process, because the process is half
the point. The rule that shapes everything else is: **one issue, one branch,
one pull request, one squash merge.**

## The loop

1. Pick an open issue from the current milestone. Dependencies are written in
   the issue body as `Depends on #...`, so read them before starting.
2. Branch from an up-to-date `main`.
3. Write the failing test first, then the implementation.
4. Run the local checks until they pass.
5. Commit with a Conventional Commit message.
6. Open a pull request whose body closes exactly that issue.
7. Merge with squash once the `build-and-test` check is green and delete the
   branch.

## Branches

`main` is the only long-lived branch and it should stay releasable after every
merge. Branch names take the type as a prefix and the issue as the subject:

```bash
git switch main
git pull --ff-only
git switch -c feat/matrix-environments
```

| Prefix | Use for |
| --- | --- |
| `feat/` | New capability |
| `fix/` | Correcting wrong output or a crash |
| `docs/` | Documentation only |
| `test/` | Tests and fixtures only |
| `chore/` | Repository plumbing with no behaviour change |

## Commits

Conventional Commits, imperative mood, scope in parentheses:

```text
feat(parser): complete the recursive-descent parser
fix(parser): return a parse error on truncated input
docs(parity): mark the renderer as missing
```

Keep the body to the reason for the change. Do not add attribution trailers or
generated-by lines of any kind: this repository has one author of record and
nothing in it should credit a tool.

## Tests first

Every behaviour change starts as a test that fails for the right reason. A test
that already passes has specified nothing, and a test that was written after
the implementation is a description of what the code happens to do.

Expectations come from the reference binary, not from reading its Rust source:

```bash
reference/txm/target/debug/txm --unboxed '\frac{1}{2}'
```

Record the output and the exit status for each case. Table-driven tests are
preferred: one slice of cases, `t.Run` per row, and `t.Helper()` on shared
assertions. Never weaken an assertion, skip a case or add a suppression to make
the suite green; find the cause or file an issue about it.

## The reference checkout

Parity work needs a local clone of txm. It lives at `reference/txm`, is never
committed, and stays out of Git through the per-clone exclude file rather than
the committed `.gitignore`:

```bash
git clone https://github.com/thatmagicalcat/txm reference/txm
cd reference/txm && git checkout 26e6d3cfef4d5d040711f4c49d9147aabd350d4d
```

Add it to your own clone's exclude file so `git add` can never pick it up:

```bash
echo '/reference/' >> .git/info/exclude
```

`.git/info/exclude` is local: it is not cloned, not committed and not visible to
anyone else. That is the right mechanism for material that exists on one
machine, which is why it is used here and the committed `.gitignore` covers only
what any clone produces, such as build output, profiles and editor files.

## Local checks

All four run in CI as the `build-and-test` job, so a pull request that fails any
of them cannot merge:

```bash
gofmt -l .        # must print nothing
go vet ./...
go build ./...
go test ./...
```

`gofmt -l` exits 0 even when it lists files, so the check is that its output is
empty, not that the command succeeded.

## Scope

If you find a bug that the current issue does not cover, do not fix it in the
same pull request. File an issue that records the exact command, the exact
output and the exit status, link it from the pull request, and keep the diff
focused. An unexplained hunk in a diff is how a small port turns into an
unreviewable one.

Discoveries that are out of scope for a pull request are not out of mind: the
issue is the memory, and the pull request that closes it is the proof.

## Documentation

Documentation is reviewed in the same pull request as the behaviour it
describes, and lives in `docs/` so that review is possible. There is no wiki.
Text written for GitHub - issue bodies, pull request bodies, comments - uses one
line per paragraph and is never hard-wrapped.

## Licence

formul8 is dual-licensed under MIT OR Apache-2.0, matching the reference it
ports. By contributing you agree that your contribution is licensed the same
way, and that the upstream copyright notice stays intact.
