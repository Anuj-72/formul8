## Context

The Rust reference supports three matrix environments: `matrix`, `bmatrix` and `pmatrix`. This port adds three more: `Bmatrix` for braces, `vmatrix` for single vertical bars and `Vmatrix` for double vertical bars. The layout constructor already contains the delimiter mapping, but the parser rejects the extras and the backend has no tested strategy for a tall double bar, so the feature is not usable yet.

## Scope

- [ ] Extend the parser's environment whitelist from the three reference names to all six, preserving the `UnknownEnvironment` error for anything else (`reference/txm/src/parser.rs:576` is the reference line to diverge from).
- [ ] Keep the existing layout delimiters: `matrix` uses spaces, `bmatrix` uses `[` `]`, `pmatrix` uses `(` `)`, `Bmatrix` uses `{` `}`, `vmatrix` uses `|` `|` and `Vmatrix` uses `‖` `‖` (`layout/build.go`, `MatrixOf`).
- [ ] Add an explicit `‖`/`‖` case to delimiter resolution instead of relying on the fill fallback, so the intent is visible and testable (`reference/txm/src/backends/generic_backend.rs:183-220`).
- [ ] Decide and document the tall `Vmatrix` appearance: use `‖` on every row so the double bar reads as a norm, rather than inventing box-drawing corners that the single-row path cannot match.
- [ ] Ensure ragged rows still fail with `MismatchedMatrixRows` for all six environments.
- [ ] Add golden and semantic tests for one-row and multi-row `Bmatrix`, `vmatrix` and `Vmatrix`, including a cell taller than one line.
- [ ] Add a matrix table to the README with one rendered example per environment and a note that the last three are formul8 extensions.

## Acceptance criteria

- [ ] `\begin{Bmatrix}a&b\\c&d\end{Bmatrix}` renders with brace sides.
- [ ] `\begin{vmatrix}a&b\\c&d\end{vmatrix}` renders with single vertical bars.
- [ ] `\begin{Vmatrix}a&b\\c&d\end{Vmatrix}` renders with a double vertical bar on every row.
- [ ] `\begin{unknown}x\end{unknown}` still returns `UnknownEnvironment`.
- [ ] A ragged matrix in any of the six environments returns `MismatchedMatrixRows`.
- [ ] The conformance suite covers all six, and the README examples match the tests.
- [ ] `go test ./...` passes.

## Notes

This is the only feature that intentionally goes beyond the reference, so it is the only place where a golden fixture cannot be generated from the Rust binary. Store its fixtures as Go-authored snapshots and mark them as extension fixtures, not parity fixtures. Do not weaken the parity corpus to accommodate them.

## Reference

`reference/txm/src/parser.rs:570-622` for environment parsing, `reference/txm/src/layout_tree.rs:554-631` for matrix metrics and the original three delimiters, and `reference/txm/src/backends/generic_backend.rs:360-420` for placement. The extension itself has no reference; its behavior is defined by this issue and its tests.

## Dependencies

Depends on #6 and #9. Belongs to the `v0.2.0` milestone, not the port parity milestone.
