## Context

The reference has 23 passing tests and this repository has none. A port without a conformance suite is a rewrite that happens to look similar; the only way to claim parity is to encode the reference's behavior as tests and compare output byte for byte where presentation is stable.

## Scope

- [ ] Port every test from `reference/txm/tests/smoke.rs` to table-driven Go tests: usage without arguments, boxed identifiers, wide CJK identifiers, adjacent wide identifiers, raw rendering, brace grouping, invalid lexer input, unknown matrix environment, ragged matrix, CLI error output, `\mathbf`, `\mathbb`, brace-less single-token arguments, accents, `\left`/`\right` parentheses and brackets, unmatched delimiters, inline punctuation, tall bracket pieces, pipe delimiters, text modes and color propagation.
- [ ] Port the reference's unit tests: the optional square-root index, the color propagation check over `\color{red}{x+y}`, `-x`, `x'` and a nested integral, plus the terminal style grouping behavior.
- [ ] Add a golden fixture harness. Store the expected ANSI output for a curated corpus in `testdata/golden/`, one file per expression, and compare exact bytes.
- [ ] Generate the fixtures from the pinned Rust reference at commit `26e6d3c` and record that commit in `testdata/golden/README.md`.
- [ ] Add `scripts/golden.sh` that builds the reference binary from `reference/txm`, runs the corpus and rewrites the fixtures. The script must fail loudly if the reference is absent rather than silently skipping.
- [ ] Add a layout snapshot mode that strips ANSI and compares the visible character grid, so a styling-only change is distinguishable from a geometry change.
- [ ] Wire the conformance suite into CI.

## Acceptance criteria

- [ ] The 23 reference smoke tests have Go equivalents and pass.
- [ ] A curated corpus of at least 20 expressions produces byte-identical unboxed output to the reference binary.
- [ ] `scripts/golden.sh` regenerates the fixtures deterministically on a clean checkout with `reference/txm` present.
- [ ] `go test ./...` passes without `reference/txm` present; golden tests skip only when the fixtures themselves are absent, never when the reference source is absent.
- [ ] A deliberately wrong glyph or spacing change makes a golden test fail with a readable diff.

## Notes

Goldens are the strongest parity signal but also the most brittle. Keep them for the small, stable corpus where the reference's exact output is intentional. Use table-driven semantic assertions for behavior that should hold even if the reference has a cosmetic quirk. When the port intentionally diverges, record the divergence in `testdata/golden/README.md` and in `docs/parity.md` rather than editing the reference output silently.

## Reference

`reference/txm/tests/smoke.rs` is 239 lines and 23 tests. `reference/txm/src/render.rs:210-239` has the square-root test. `reference/txm/src/backends/generic_backend.rs:421-495` has the color propagation test. The reference binary can be built with `cargo build --manifest-path reference/txm/Cargo.toml`.

## Dependencies

Depends on #3, #4, #5, #6, #7 and #8. Blocks #10 and #11.
