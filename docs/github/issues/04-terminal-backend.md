## Context

The three files in `backend/` are design notes. Nothing allocates a character grid, writes a cell, or emits ANSI escapes, so even a correct layout tree cannot become terminal output. This issue builds the output surface that the generic positioner will draw into.

## Scope

- [ ] Define the `RenderTarget` interface with `Set(x, y int, ch rune, st style.Style)` and `FillRow(y, xStart, xEnd int, ch rune, st style.Style)` (`reference/txm/src/backend.rs:11-14`).
- [ ] Implement the unexported character buffer as parallel `Data []rune` and `Styles []style.Style` slices over a `Width x Height` grid, filling both with spaces and the zero style (`reference/txm/src/backends/terminal.rs:37-53`).
- [ ] Bounds-check `Set` and `FillRow` exactly like the reference: ignore writes outside the grid, and clamp the fill end to the width.
- [ ] Implement `String()` with the reference's style grouping: walk row by row, emit a reset before a style change when the active style was non-empty, emit the new prefix only when the new style is non-empty, and emit a final reset at the end of a row when the active style is non-empty.
- [ ] Guard `Width == 0` and return an empty string. The reference calls `chunks_exact(0)`, which panics on an empty buffer; that is a reference bug the port should not reproduce.
- [ ] Implement `RenderTerminal(tree *layout.Node) string`: allocate a buffer at the tree's dimensions, call the generic positioner at `(0, 0)`, and return the serialized string (`reference/txm/src/backends/terminal.rs:25-34`).
- [ ] Add tests for an empty tree, a one-cell tree, a row of mixed styles, and a style that ends exactly at the row boundary.

## Acceptance criteria

- [ ] A hand-built `Text` node renders its runes at the expected coordinates.
- [ ] Adjacent cells with the same style share one ANSI prefix, and a style change emits a reset before the next prefix.
- [ ] A row whose final cell is unstyled ends with no stray reset.
- [ ] A zero-width tree returns `""` without panicking.
- [ ] Out-of-bounds `Set` and `FillRow` calls leave the buffer unchanged.
- [ ] `go test ./backend/...` passes.

## Reference

`reference/txm/src/backends/terminal.rs` is 104 lines. The buffer is at `:37-53`, the display implementation at `:55-88`, and the target implementation at `:90-104`. `reference/txm/src/backend.rs` is 14 lines and defines the two abstractions.

## Out of scope

The generic positioning algorithm, which is #5. This issue can be developed in parallel with #3 because it only needs the layout package and hand-built nodes.

## Dependencies

Depends on nothing except the existing `layout` package. Blocks #5.
