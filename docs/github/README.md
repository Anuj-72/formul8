# Running formul8 on GitHub

This is the operating manual for finishing the Go port and learning the GitHub surfaces while doing it. It is written to be executed in order. Every command that changes state on GitHub is marked **REMOTE** and is run by the repository owner, not by an assistant.

## The mental model

| Surface | What it is for | What it is not for |
| --- | --- | --- |
| Issue | One unit of work with a testable definition of done | A diary, a wish list, or a place to plan several PRs |
| Label | Cross-cutting classification: type, area, parity vs extension, priority | Workflow state |
| Milestone | Release-shaped bucket of issues, with a due date and completion percentage | A sprint board |
| Project | Day-to-day status, sizing, views and automation across issues and PRs | A replacement for issues |
| Branch | Short-lived workspace for exactly one issue | Long-running integration branches |
| Pull request | The reviewable diff that closes one issue | A merge queue for unrelated work |
| Release and tag | An immutable snapshot of `main` after a milestone | A branch |
| Actions | Automated checks that gate merges | A deployment platform for this project |
| Wiki | Nothing here; documentation lives in `docs/` | A second documentation site |
| Discussions | Nothing here initially; use issues | A support forum for a one-person project |

The working rule is: **one issue, one branch, one pull request, one squash merge**. `main` should stay releasable after every merge.

## Phase 0: preflight

Run these from the repository root and confirm the outputs before changing anything.

```bash
git --version
gh --version
gh auth status
git config user.name
git config user.email
git status --short --branch
```

`gh auth status` currently reports the `repo` and `workflow` scopes. The Projects commands later need the `project` scope, which is added in Phase 6.

The repository has no commits yet. Everything is untracked, and `reference/txm` is a nested clone that must never be committed.

## Phase 1: bootstrap the first commit

The first commit imports the existing work as-is. It is the one deliberate exception to "everything goes through a PR", because there is no history to branch from yet. It does not compile; issue #1 fixes that before CI exists.

Rename the unborn branch, stage only the port files, and inspect what will be committed:

```bash
git symbolic-ref HEAD refs/heads/main
git add .gitignore go.mod txm.go ast backend errs glyph layout parser render style token
git status --short
```

The staged list must contain exactly these paths and nothing from `reference/` or `.freebuff/`:

```text
.gitignore
go.mod
txm.go
ast/expr.go
backend/generic.go
backend/target.go
backend/terminal.go
errs/errors.go
glyph/alphabet.go
glyph/glyph.go
glyph/glyphs.go
glyph/registry.go
layout/build.go
layout/node.go
layout/scriptchars.go
parser/parser.go
render/render.go
style/color.go
style/style.go
token/lexer.go
token/token.go
```

Only after that review is approved:

```bash
git commit -m "chore: import initial formul8 port skeleton"
```

The commit contains no generated files, no reference sources, no editor settings and no attribution trailers.

## Phase 2: create the remote and push

**REMOTE.** Create an empty public repository, attach it as `origin`, and push the bootstrap commit.

```bash
gh repo create Anuj-72/formul8 --public --description "Terminal LaTeX math renderer in Go, ported from txm"
git remote add origin https://github.com/Anuj-72/formul8.git
git push -u origin main
```

Do not use `gh repo create --push` here: creating the repository and publishing the first commit are separate actions.

## Phase 3: repository settings

**REMOTE.** Set the description and topics, keep issues and projects on, turn the wiki off, and allow squash merges only.

```bash
gh repo edit Anuj-72/formul8 \
  --description "Terminal LaTeX math renderer in Go, ported from txm" \
  --add-topic latex --add-topic terminal --add-topic math --add-topic go --add-topic renderer \
  --enable-issues --enable-projects --enable-wiki=false \
  --enable-squash-merge --enable-merge-commit=false --enable-rebase-merge=false \
  --delete-branch-on-merge --enable-auto-merge
```

Why squash only: one PR becomes one commit on `main`, the PR title becomes the commit subject, and `git log --oneline` stays readable. Rebase and merge commits are disabled so the history cannot drift into three shapes.

In the browser, open **Settings → Code security** and enable Dependabot alerts, Dependabot security updates, secret scanning and push protection. These are free for public repositories. Keep the default `GITHUB_TOKEN` permissions read-only under **Settings → Actions → General**.

## Phase 4: labels

**REMOTE.** Create a small taxonomy. Labels answer "what kind of work is this", not "where is it in the workflow"; status belongs to the project board.

| Label | Color | Meaning |
| --- | --- | --- |
| `type: feat` | `0E8A16` | New capability |
| `type: fix` | `D73A4A` | Corrects wrong behavior |
| `type: docs` | `0075CA` | Documentation only |
| `type: test` | `5319E7` | Tests and fixtures only |
| `type: chore` | `6E7781` | Maintenance with no behavior change |
| `area: parser` | `1D76DB` | Lexer and parser |
| `area: layout` | `0052CC` | Sized layout tree |
| `area: render` | `006B75` | AST to layout |
| `area: backend` | `6F42C1` | Character buffer and positioning |
| `area: api` | `C2E0C6` | Registry, public entry points |
| `area: cli` | `BFD4F2` | Command-line interface |
| `area: infra` | `D4C5F9` | CI, docs, repository plumbing |
| `parity` | `FBCA04` | Required to match the Rust reference |
| `extension` | `F9D0C4` | Beyond the reference, added by formul8 |
| `priority: high` | `B60205` | Blocks the milestone |
| `priority: medium` | `FBCA04` | Normal order |
| `priority: low` | `C2E0C6` | Can wait |

```bash
gh label create "type: feat" --repo Anuj-72/formul8 --color 0E8A16 --description "New capability" --force
gh label create "type: fix" --repo Anuj-72/formul8 --color D73A4A --description "Corrects wrong behavior" --force
gh label create "type: docs" --repo Anuj-72/formul8 --color 0075CA --description "Documentation only" --force
gh label create "type: test" --repo Anuj-72/formul8 --color 5319E7 --description "Tests and fixtures only" --force
gh label create "type: chore" --repo Anuj-72/formul8 --color 6E7781 --description "Maintenance with no behavior change" --force
gh label create "area: parser" --repo Anuj-72/formul8 --color 1D76DB --description "Lexer and parser" --force
gh label create "area: layout" --repo Anuj-72/formul8 --color 0052CC --description "Sized layout tree" --force
gh label create "area: render" --repo Anuj-72/formul8 --color 006B75 --description "AST to layout" --force
gh label create "area: backend" --repo Anuj-72/formul8 --color 6F42C1 --description "Character buffer and positioning" --force
gh label create "area: api" --repo Anuj-72/formul8 --color C2E0C6 --description "Registry and public entry points" --force
gh label create "area: cli" --repo Anuj-72/formul8 --color BFD4F2 --description "Command-line interface" --force
gh label create "area: infra" --repo Anuj-72/formul8 --color D4C5F9 --description "CI, docs and repository plumbing" --force
gh label create "parity" --repo Anuj-72/formul8 --color FBCA04 --description "Required to match the Rust reference" --force
gh label create "extension" --repo Anuj-72/formul8 --color F9D0C4 --description "Beyond the reference, added by formul8" --force
gh label create "priority: high" --repo Anuj-72/formul8 --color B60205 --description "Blocks the milestone" --force
gh label create "priority: medium" --repo Anuj-72/formul8 --color FBCA04 --description "Normal order" --force
gh label create "priority: low" --repo Anuj-72/formul8 --color C2E0C6 --description "Can wait" --force
```

Optionally remove the default labels that this scheme replaces:

```bash
for label in bug documentation duplicate enhancement "good first issue" "help wanted" invalid question wontfix; do
  gh label delete "$label" --repo Anuj-72/formul8 --yes || true
done
```

## Phase 5: milestones

**REMOTE.** Create two milestones. `v0.1.0` is port parity; `v0.2.0` is the matrix extension that the reference does not have. Keeping them separate prevents a post-parity feature from holding the port release hostage.

```bash
gh api --method POST repos/Anuj-72/formul8/milestones \
  -f title="v0.1.0 — Port parity" \
  -f description="Complete the Go port of txm: parser, renderer, terminal backend, generic backend, registry and public API, CLI, conformance tests and docs. Done when the reference smoke corpus renders with matching output." \
  -f due_on="2026-12-31T23:59:59Z"

gh api --method POST repos/Anuj-72/formul8/milestones \
  -f title="v0.2.0 — Matrix environments" \
  -f description="Extend the port with the Bmatrix, vmatrix and Vmatrix environments, including backend support, docs and golden tests." \
  -f due_on="2027-01-31T23:59:59Z"
```

List them and note the numbers:

```bash
gh api repos/Anuj-72/formul8/milestones --jq '.[] | "\(.number)\t\(.title)\t\(.open_issues)/\(.open_issues + .closed_issues) open"'
```

A milestone is a release contract: its issue list is the scope, and reaching 100% is the signal to tag. Do not add nice-to-haves to it mid-flight; create a new milestone or an issue outside it.

## Phase 6: issues

**REMOTE.** Create one issue per checkpoint from the drafts in `docs/github/issues/`. Create them in numeric order so the issue numbers match the draft numbers. Each issue is self-contained: context, scope, acceptance criteria, reference pointers and dependencies.

```bash
gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Complete the parser and make the module build" \
  --body-file docs/github/issues/01-parser.md \
  --label "type: feat" \
  --label "area: parser" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Bootstrap repository hygiene, docs and CI" \
  --body-file docs/github/issues/02-foundation.md \
  --label "type: chore" \
  --label "area: infra" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Implement the AST-to-layout renderer" \
  --body-file docs/github/issues/03-renderer.md \
  --label "type: feat" \
  --label "area: render" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Implement the terminal buffer and ANSI serializer" \
  --body-file docs/github/issues/04-terminal-backend.md \
  --label "type: feat" \
  --label "area: backend" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Implement the generic backend for core layout nodes" \
  --body-file docs/github/issues/05-generic-backend-core.md \
  --label "type: feat" \
  --label "area: backend" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Implement big operators and matrix placement" \
  --body-file docs/github/issues/06-generic-backend-operators.md \
  --label "type: feat" \
  --label "area: backend" \
  --label "area: layout" \
  --label "parity" \
  --label "priority: medium" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Build the symbol registry and public Layout/Render API" \
  --body-file docs/github/issues/07-registry-api.md \
  --label "type: feat" \
  --label "area: api" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Port the CLI with boxed and unboxed output" \
  --body-file docs/github/issues/08-cli.md \
  --label "type: feat" \
  --label "area: cli" \
  --label "parity" \
  --label "priority: medium" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Port the smoke suite and add golden parity fixtures" \
  --body-file docs/github/issues/09-conformance.md \
  --label "type: test" \
  --label "area: infra" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Finish the extra matrix environments" \
  --body-file docs/github/issues/10-matrix-environments.md \
  --label "type: feat" \
  --label "area: parser" \
  --label "area: backend" \
  --label "extension" \
  --label "priority: medium" \
  --milestone "v0.2.0 — Matrix environments"

gh issue create --repo Anuj-72/formul8 --assignee @me \
  --title "Cut v0.1.0" \
  --body-file docs/github/issues/11-release.md \
  --label "type: chore" \
  --label "area: infra" \
  --label "parity" \
  --label "priority: high" \
  --milestone "v0.1.0 — Port parity"
```

Verify:

```bash
gh issue list --repo Anuj-72/formul8 --state all --limit 20
```

The dependency edges are written in the issue bodies (`Depends on #…`, `Blocks #…`). Read them before picking the next issue; the intended order is 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, although 3 and 4 can be developed in parallel once 1 lands.

## Phase 7: project board

**REMOTE.** Projects v2 needs the `project` scope. This command opens a browser and asks you to authorize it.

```bash
gh auth refresh -s project
gh project create --owner @me --title "formul8" --format json
```

The create output contains the project number; call it `PROJECT_NUMBER` below. Link it to the repository and add a `Size` field. Do not duplicate labels as fields; labels already carry type, area and priority.

```bash
gh project link PROJECT_NUMBER --owner @me --repo Anuj-72/formul8
gh project field-create PROJECT_NUMBER --owner @me --name "Size" --data-type SINGLE_SELECT --single-select-options "S,M,L"
```

Add every open issue:

```bash
gh issue list --repo Anuj-72/formul8 --limit 100 --json number --jq '.[].number' |
while read -r n; do
  gh project item-add PROJECT_NUMBER --owner @me --url "https://github.com/Anuj-72/formul8/issues/$n"
done
```

Open the board and create the views in the UI:

```bash
gh project view PROJECT_NUMBER --owner @me --web
```

- **Board**: group by the built-in `Status` field. Use `Todo`, `In Progress`, `In Review`, `Done`.
- **Table**: columns `Title`, `Labels`, `Milestone`, `Size`, `Status`.
- **Roadmap**: group by milestone, so `v0.1.0` and `v0.2.0` show as two bars.

In **Project → Workflows**, enable the built-in workflows: auto-add issues and PRs from `Anuj-72/formul8`, set new items to `Todo`, and move an item to `Done` when its issue closes. Drag the first issue to `In Progress` by hand; automation should handle the rest.

## Phase 8: protect `main`

**REMOTE.** Wait until the CI workflow from issue #2 has run at least once, otherwise the required check name does not exist yet. The payload is in `docs/github/main-ruleset.json`.

Confirm the exact name the check reports, because a required check matches that string literally:

```bash
gh api repos/Anuj-72/formul8/commits/main/check-runs \
  --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)"'
```

The `context` in the payload must equal one of those names. If the output shows a
prefixed form such as `CI / build-and-test`, change the payload before applying it:
a required check that matches nothing leaves every pull request waiting on a check
that will never report, and because the payload also requires an up-to-date branch,
`main` cannot be updated at all until the mismatch is fixed.

```bash
gh api --method POST repos/Anuj-72/formul8/rulesets \
  --input docs/github/main-ruleset.json
gh api repos/Anuj-72/formul8/rulesets --jq '.[] | "\(.id)\t\(.name)\t\(.enforcement)"'
```

The ruleset enforces:

- every change arrives through a pull request, with zero required approvals because this is a solo repository;
- the `build-and-test` check must pass and the branch must be up to date;
- squash is the only allowed merge method;
- linear history, no force pushes, no branch deletion.

Zero approvals is a deliberate solo-maintainer setting. Classic branch protection rejects a zero approval count, which is why this uses a repository ruleset; if your account rejects the payload, stop and adjust it rather than falling back to unprotected `main`. As soon as a second maintainer joins, change `required_approving_review_count` to `1`; do not leave a review requirement at zero in a team repository.

## Phase 9: the loop for every issue

This is the cycle to repeat eleven times. The local steps are safe; the marked steps touch GitHub and are run by you.

1. Move the issue to `In Progress` on the project board.
2. Sync and branch:

```bash
git switch main
git pull --ff-only
git switch -c feat/parser-completion
```

3. Implement only that issue. Run the checks locally:

```bash
gofmt -l .
go vet ./...
go test ./...
```

4. Inspect and stage the diff. Present the exact file list and the proposed commit message for approval before committing:

```bash
git status --short
git diff
git add path/to/file1 path/to/file2
git commit -m "feat(parser): complete the recursive-descent parser"
```

5. **REMOTE.** Publish the branch:

```bash
git push -u origin feat/parser-completion
```

6. **REMOTE.** Open the pull request. The body has one paragraph per line and closes exactly one issue:

```bash
gh pr create --repo Anuj-72/formul8 \
  --base main \
  --head feat/parser-completion \
  --title "feat(parser): complete the recursive-descent parser" \
  --body "$(cat <<'EOF'
## Summary
Completes the recursive-descent parser so the module builds and every later checkpoint has an AST to consume.

## Changes
- Fixes `parseJuxtapose`, which appended the same node without consuming tokens.
- Implements `parseCommand`, `parseLeftDelimited`, `parseDelimitedArg`, `parseMatrixBody` and `parseTokens`.
- Completes `parseBegin` for `matrix`, `bmatrix` and `pmatrix`; `Bmatrix`, `vmatrix` and `Vmatrix` stay rejected for #10.
- Adds parser unit tests for arguments, delimiters, matrices and malformed input.

## Verification
- `go build ./...`
- `go vet ./...`
- `go test ./...`

## Issue
Closes #1
EOF
)"
```

7. **REMOTE.** Watch the checks and review your own diff:

```bash
gh pr checks --watch
gh pr diff
gh pr view --web
```

8. **REMOTE.** Merge with squash only after the checks are green:

```bash
gh pr merge --squash --delete-branch
```

9. Sync and confirm the issue closed:

```bash
git switch main
git pull --ff-only --prune
gh issue view 1 --json state,closedAt --jq '"\(.state) \(.closedAt)"'
```

10. Update the project item to `Done` if the workflow did not, and check milestone progress:

```bash
gh api repos/Anuj-72/formul8/milestones --jq '.[] | "\(.title): \(.closed_issues)/\(.open_issues + .closed_issues)"'
```

Rules for the loop:

- One issue per PR. If the work grows, split the issue and open a second PR; do not smuggle unrelated changes in.
- The PR title uses Conventional Commits and becomes the squash commit subject.
- The PR body is one paragraph per line. Blank lines separate blocks; do not hard-wrap.
- Every behavior change carries a test in the same PR.
- Documentation changes land in the same PR as the behavior they describe.
- Never commit `reference/` or `.freebuff/`, never push directly to `main`, never force-push.

## Phase 10: release

**REMOTE.** After every `v0.1.0` issue is closed and `main` is green, write the release notes from the release issue, tag, and publish.

```bash
git switch main
git pull --ff-only
git tag -a v0.1.0 -m "formul8 v0.1.0"
git push origin v0.1.0
gh release create v0.1.0 \
  --repo Anuj-72/formul8 \
  --title "formul8 v0.1.0" \
  --notes-file docs/releases/v0.1.0.md \
  --verify-tag
```

Close the milestone:

```bash
gh api repos/Anuj-72/formul8/milestones --jq '.[] | select(.title | startswith("v0.1.0")) | .number'
gh api --method PATCH repos/Anuj-72/formul8/milestones/MILESTONE_NUMBER -f state=closed
```

Verify the install path, which is the real definition of a usable Go release:

```bash
go install github.com/Anuj-72/formul8/cmd/formul8@v0.1.0
formul8 --version
```

A release is not a branch and not a project column. It is an immutable tag plus notes, and the milestone is the contract that says the scope is complete.

## Phase 11: conventions worth keeping

- Branch names: `feat/…`, `fix/…`, `docs/…`, `test/…`, `chore/…`.
- Commits: Conventional Commits, imperative mood, scope in parentheses. No attribution trailers of any kind.
- Issue titles: outcome, not activity. "Complete the parser", not "Working on parser".
- Acceptance criteria: checkboxes a reviewer can verify without reading the implementation.
- Labels describe the work; the project describes its state; the milestone describes the release.
- Do not open an issue for every keystroke. One issue is one PR-sized outcome. If it needs two PRs, it is two issues.
- When a bug is found outside the current issue's scope, open a new issue, link it, and keep the current PR focused.
- Read `git diff main...HEAD` before opening a PR. If the diff contains a file you cannot explain, stop.

## Why this shape

The repository is a port with a live executable reference, so the work naturally decomposes into vertical checkpoints: parse, render, position, serialize, expose, test, document. Each checkpoint can be verified on its own, which makes it a good issue. The milestone is the release, not the project; the project is the flow of issues through states. That separation is what keeps a real repository readable when the issue count grows.
