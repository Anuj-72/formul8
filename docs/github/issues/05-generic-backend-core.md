## Context

The terminal buffer can hold cells, but nothing computes absolute positions for a sized layout tree. The generic backend is the second pass: it walks the tree and writes every rune into the buffer, merging inherited styles as it descends. This issue ports the core node cases; the big operators and matrices follow in #6.

## Scope

- [ ] Implement `renderNode` and `renderInner` with the style merge at the top of every node: `inherited.merge(node.style)` (`reference/txm/src/backends/generic_backend.rs:6-10`).
- [ ] `Text`: write every rune at `x+i, y`.
- [ ] `HStack`: compute each child's y from the parent baseline minus the child baseline, advance x by child width plus spacing.
- [ ] `VStack`: centre both children horizontally, place the top at `maxHeight - top.Height`, the bottom at `y + maxHeight + 1`, and fill the separator row when the line style is solid.
- [ ] `Infix`: align both sides on the larger baseline, write `+`, `-`, `=` or `·` at the operator column, and place the right side after the three-column gap.
- [ ] `Superscript` and `Subscript`: honour the inline flag when computing the vertical shift.
- [ ] `BothScripts`: place the superscript at the top right, the base below it, and the subscript below the base baseline.
- [ ] `StretchyDelim`: use the single-row path when the inner height is at most one, otherwise resolve the per-row delimiter pieces and reserve two columns of padding.
- [ ] `Accent`: repeat the mark across the width when stretched, otherwise centre it.
- [ ] `Limits`: stack the upper, base and lower nodes around the taller of the two limits.
- [ ] `Sqrt`: draw the top bar, the left vertical, the diagonal foot and the radicand, then overlay the optional index at the origin.
- [ ] `Neg`, `Prime` and `Empty`: place the minus at the inner baseline, the primes after the base, and nothing at all.
- [ ] Implement `resolveDelimiterSingle` and `resolveDelimiters`, including the `{`/`}` middle pieces at the baseline row (`reference/txm/src/backends/generic_backend.rs:183-220`).
- [ ] Add tests for style inheritance, vertical centring, inline scripts, and every delimiter family.

## Acceptance criteria

- [ ] Hand-built trees for each node kind render at the expected coordinates.
- [ ] A colored parent reaches every cell its children draw, including generated delimiter and operator cells.
- [ ] Inline superscripts and subscripts do not increase the line height, while block scripts do.
- [ ] Tall parentheses, brackets, braces and pipes use the correct corner and middle pieces.
- [ ] `go test ./backend/...` passes.

## Reference

`reference/txm/src/backends/generic_backend.rs` is 495 lines. The dispatcher is at `:6-182`, delimiter resolution at `:183-220`, and the big-operator and matrix cases at `:222-495`.

## Out of scope

Summation, product, integral and matrix placement, which are #6. The registry and public API are #7.

## Dependencies

Depends on #4. Blocks #6.
