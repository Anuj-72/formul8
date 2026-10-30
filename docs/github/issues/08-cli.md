## Context

The library API renders a formula, but the reference project is a terminal command and its behavior is part of the port contract: flag handling, exit codes, the decorative box and the way errors are printed. This issue adds the `formul8` binary and keeps it small enough to test.

## Scope

- [ ] Add `cmd/formul8/main.go` so `go install github.com/Anuj-72/formul8/cmd/formul8@…` produces the binary.
- [ ] Support `--help`, `--version`, `--unboxed` and `--file <path>`, with `-` meaning standard input (`reference/txm/src/main.rs:38-131`).
- [ ] With no expression and no `--file`, print the usage text and exit successfully, matching the reference's behavior (`main.rs:60-62`).
- [ ] With an unknown flag, print the error plus the "For more information" hint to standard error and exit with status 2 (`main.rs:63-66`).
- [ ] With an extra positional argument or a second `--file`, print the reference error and exit with status 2.
- [ ] With a render error, print `error: <message>` to standard error and exit with status 1.
- [ ] Implement boxed output: strip ANSI, measure display width, draw the top border, a blank padded row, every rendered line padded to the maximum width, another blank row and the bottom border (`main.rs:151-184`).
- [ ] Implement unboxed output as the raw rendered string with no extra newline.
- [ ] Use a display-width implementation that handles CJK and combining marks rather than counting runes. `github.com/mattn/go-runewidth` is the smallest dependency that matches the reference's `unicode-width` behavior; add it to `go.mod` and explain the choice in the PR.
- [ ] Replace the reference's `/dev/stdin` special case with `os.Stdin` so the binary works on Windows as well.
- [ ] Print `formul8 <version>` for `--version`, with the version injected through `-ldflags` and defaulting to `dev`.

## Acceptance criteria

- [ ] `go run ./cmd/formul8 "x"` prints a box that starts with `┌` and contains `x`.
- [ ] `go run ./cmd/formul8 --unboxed "x"` prints the rendered line without a box or a trailing newline.
- [ ] `go run ./cmd/formul8 "你你"` keeps the box edges aligned, proving display width rather than rune count.
- [ ] `go run ./cmd/formul8 '\\'` prints `error: invalid token at byte 0` and exits 1.
- [ ] `go run ./cmd/formul8 --bogus` exits 2 and suggests `--help`.
- [ ] A file containing `E = mc^2` renders identically through `--file path` and `--file -`.
- [ ] CLI tests spawn the built binary and assert stdout, stderr and exit status, mirroring `reference/txm/tests/smoke.rs:1-82`.
- [ ] `go test ./...` passes.

## Notes

Keep argument parsing hand-written. The reference has four flags and specific exit codes; a flag library would add a dependency and change the help output for no benefit. The box is presentation, not layout, so it belongs in the CLI rather than the backend.

## Reference

`reference/txm/src/main.rs` is 185 lines. Argument parsing is at `:68-131`, the help text at `:132-150`, ANSI stripping at `:152-166`, and the box at `:168-184`. `reference/txm/tests/smoke.rs` shows the expected CLI behavior.

## Out of scope

A `--ratatui` mode, C bindings and Python bindings. The reference gates the ratatui backend behind a feature; this port does not include it.

## Dependencies

Depends on #7.
