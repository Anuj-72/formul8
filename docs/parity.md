# Parity with txm

formul8 is a port of [txm](https://github.com/thatmagicalcat/txm) (Terminal TeX
Math). This page records, module by module, how much of that reference exists
here. It is checked against the tree at every pull request that touches it, so
"done" means the code is present and tested, not that it is planned.

Ported baseline: commit `26e6d3cfef4d5d040711f4c49d9147aabd350d4d`
("add C/C++ bindings", 2026-07-19). The reference is a local clone at
`reference/txm`, never committed.

## Modules

| txm | formul8 | State |
| --- | --- | --- |
| `src/token.rs` (156) | `token/token.go`, `token/lexer.go` | Ported in #1. |
| `src/error.rs` (92) | `errs/errors.go` | Ported in #1. |
| `src/ast.rs` (36) | `ast/expr.go` | Ported in #1, plus a `Matrix` node that #10 needs. |
| `src/style.rs` (299) | `style/style.go`, `style/color.go` | Ported. |
| `src/glyph.rs` (537) | `glyph/glyph.go`, `glyph/glyphs.go`, `glyph/alphabet.go`, `glyph/registry.go` | Glyph types, alphabet tables and the registry container are ported. The ~200-entry `buildRegistry` is missing, in #7. |
| `src/layout_tree.rs` (697) | `layout/node.go`, `layout/build.go`, `layout/scriptchars.go` | Ported. |
| `src/parser.rs` (678) | `parser/parser.go`, `parser/parser_test.go` | Ported in #1. Truncated input still panics; #13 tracks it. |
| `src/render.rs` (241) | `render/render.go` | Missing, in #3. The file documents the port and defines no behaviour. |
| `src/backend.rs` (14) | `backend/target.go` | Not ported, deliberately. See divergences. |
| `src/backends/terminal.rs` (104) | `backend/terminal.go` | Missing, in #4. |
| `src/backends/generic_backend.rs` (495) | `backend/generic.go` | Missing, in #5 and #6. |
| `src/lib.rs` (256) | `txm.go` | Missing, in #7. No `Layout` and no `Render` yet. |
| `src/main.rs` (185) | `cmd/formul8` | Missing, in #8. |
| `src/ratatui.rs` (286) | - | Not ported. The ratatui frontend is out of scope; formul8 targets plain terminal output. |
| `tests/smoke.rs` (239) | golden fixtures | Missing, in #9. |

Line counts in parentheses are the reference file sizes at the baseline commit.

## What works today

| Capability | State |
| --- | --- |
| Tokenising | Yes. |
| Parsing to an AST | Yes, including `matrix`, `bmatrix` and `pmatrix`. |
| Building a sized layout tree | Yes as a package, unreachable from outside it: no registry and no public entry point. |
| Rendering an AST into a layout | No. |
| Serialising a layout into characters | No. |
| Drawing that on a terminal | No. |

So `go build ./...` succeeds and `go test ./...` covers the tokeniser, parser,
style and layout packages, while nothing can turn an expression into output
yet. A release is not close until #3 through #7 land.

## Deliberate divergences

These are choices, not gaps. Each one is either a Rust-only mechanism with no
Go counterpart or a reference behaviour formul8 refuses to reproduce.

- **The `Backend` trait is not ported.** `backend.rs` defines `Backend` with an
  associated `Output` type. Go interfaces cannot carry associated types
  cleanly, and with the ratatui backend out of scope there is one output
  format. `RenderTarget`, the real polymorphism point, stays an interface.
- **A zero-width buffer returns an empty string.** The reference collects rows
  with `chunks_exact(0)`, which panics on empty input. formul8 guards `Width`
  at zero instead.
- **`Matrix` node and `matrix`-family environments are extension groundwork.**
  The reference parses `matrix`, `bmatrix` and `pmatrix`. The `Bmatrix`,
  `vmatrix` and `Vmatrix` environments are added by #10 and are the only
  intended difference in behaviour, tracked in milestone v0.2.0.

## Checking this page

```bash
git ls-files 'reference' '.freebuff'
```

prints nothing when the local material is properly excluded.

```bash
grep -c '^func ' render/render.go backend/*.go txm.go
```

counts the exported and unexported functions in the files marked missing; a
non-zero count on this page's "missing" rows means the page is stale.
