## Context

`.github/ISSUE_TEMPLATE/bug_report.yml` shipped in #14 and GitHub will not render the form. Opening `New issue` and choosing `Bug report` shows a red banner instead of fields:

```
There is a problem with this template
body[4]: placeholder must be of type String and cannot be empty.
```

The bug-report form is therefore unusable: the template is syntactically valid YAML, passes any YAML linter, and still fails the schema GitHub applies when it builds the form. Found after #14 merged, so it is filed and fixed on its own rather than reopened inside that pull request.

## Reproduction

1. Open `https://github.com/Anuj-72/formul8/issues/new?template=bug_report.yml`.
2. The form is replaced by the banner above.

Confirmed against the file at `9c33a25`. `feature_request.yml` renders correctly, so the defect is confined to one template.

## Cause

`body` is a list, and the failing element is index 4, counting from zero:

| index | type | id |
| --- | --- | --- |
| 0 | markdown | — |
| 1 | textarea | observed |
| 2 | textarea | expected |
| 3 | textarea | reference |
| 4 | input | version |
| 5 | textarea | environment |

The `version` field carried `placeholder: 2857302`. YAML resolves an unquoted run of digits to an integer, so the parsed value is `2857302` of type `int`, while GitHub's schema for `placeholder` is a non-empty `String`. The parser is not guessing: `yaml.safe_load` reports `int` for that node and `str` for every other placeholder in both templates.

GitHub does not coerce here, and its error text names the type rather than the line, which is why the visible banner points at an element index that no line number matches.

## Acceptance criteria

- [ ] `body[4].attributes.placeholder` is a quoted string, and the form renders with all six fields instead of the banner.
- [ ] The version field still suggests `2857302` as placeholder text, and submitting the form produces a string value for it.
- [ ] Every `placeholder`, `label`, `description` and `value` in both issue templates parses as a `str`, and every `id` is unique, asserted by a command whose output is recorded in the pull request.
- [ ] `gofmt -l .`, `go vet ./...`, `go build ./...` and `go test ./...` stay clean.

## Out of scope

A CI step that validates issue-template schema on every push. It would have caught this before merge, but it needs a validator dependency in the workflow and that is a separate decision from restoring the form. Filed separately if wanted.

## Dependencies

Found in #14. Blocks nothing, but the bug-report template is the one future defects are meant to be reported with, so it should land before #3 starts filing follow-ups.
