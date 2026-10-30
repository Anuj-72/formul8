## Context

The core positioner handles text, stacks, scripts, delimiters and accents. The remaining cases are the shapes that give terminal math its character: the three big operators and the matrix grid. These are the most geometry-sensitive parts of the backend, so they get their own checkpoint and their own snapshots.

## Scope

- [ ] Implement summation rendering with all three branches: no inner argument, inner height at most two, and the tall branch (`reference/txm/src/backends/generic_backend.rs:222-288`).
- [ ] In the tall summation branch, reproduce the arrow row selection exactly: use the middle arrow only when the height is not a power of two, and alternate the diagonal direction above and below the centre.
- [ ] Implement product rendering, including the no-inner branch and the vertical-bar body (`reference/txm/src/backends/generic_backend.rs:289-325`).
- [ ] Implement integral rendering with the no-inner, short and tall branches, including the one-row vertical offset for a single-line integrand (`reference/txm/src/backends/generic_backend.rs:326-359`).
- [ ] Implement matrix placement: recompute row baselines, row depths, cell width, cell height, horizontal spacing and row spacing from the rendered cells, then centre each cell in its slot (`reference/txm/src/backends/generic_backend.rs:360-420`).
- [ ] Keep matrix placement independent of the matrix constructor's metrics; the reference deliberately recomputes them, and matching that keeps the two passes honest.
- [ ] Add table-driven snapshots for sigma at heights 1, 2, 3, 4 and 5; product at heights 1, 2 and 4; integral at heights 1, 3 and 5; and matrices with one row, two rows and cells of different heights.

## Acceptance criteria

- [ ] A bare `\sum`, `\prod` and `\int` render the reference glyphs at the expected dimensions.
- [ ] A tall sigma uses the centre arrow only when the height is not a power of two.
- [ ] Matrix columns align across rows even when cells have different widths and baselines.
- [ ] An empty matrix renders nothing and does not panic.
- [ ] Colored big operators and matrices carry the inherited color into every generated cell.
- [ ] `go test ./backend/...` passes with the new snapshots.

## Reference

`reference/txm/src/backends/generic_backend.rs:222-495`. The summation arrow test is at `:275`, product at `:289`, integral at `:326`, and matrix placement at `:360`.

## Out of scope

The extra matrix environments `Bmatrix`, `vmatrix` and `Vmatrix`, which are #10. This issue only needs the three reference environments to render.

## Dependencies

Depends on #5. Blocks #10.
