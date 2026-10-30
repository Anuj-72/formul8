# formul8

A terminal LaTeX math renderer in Go, ported from
[txm](https://github.com/thatmagicalcat/txm) (Terminal TeX Math).

formul8 turns a LaTeX math expression into a styled two-dimensional grid of
rune cells and serialises that grid as ANSI escapes, so `\frac{1}{2}` and a
tall summation come out as typeset maths in a terminal instead of as source
text. There is no image, no browser and no font file: the output is characters
and colour.

## Status

Work in progress. The tokeniser, parser, layout tree and style system are
ported and tested. The renderer, both backends, the symbol registry and the
command-line interface are not written yet, so **nothing renders today**: the
packages build and the parser produces an AST that nothing consumes.

| Capability | State |
| --- | --- |
| Tokenise an expression | Done |
| Parse to an AST | Done, including `matrix`, `bmatrix` and `pmatrix` |
| Build a sized layout tree | Done as a package, with no public entry point yet |
| Render an AST into a layout | Not yet |
| Serialise a layout to characters | Not yet |
| Draw it in a terminal | Not yet |

[docs/parity.md](docs/parity.md) maps every Rust module of the reference to the
Go package that ports it and marks each one done, in progress or missing. That
page, not this table, is the authoritative status.

## Relationship to txm

formul8 is a port, not a fork. It follows the Rust implementation at commit
`26e6d3cfef4d5d040711f4c49d9147aabd350d4d` ("add C/C++ bindings",
2026-07-19), and treats that reference binary as the correctness oracle:
expected values in the test suite are taken by running the reference, not by
reading its source. Where the two disagree, the source and the binary are
compared and the difference is recorded.

The reference is not redistributed here. A development clone keeps it at
`reference/txm`, excluded from Git, and [CONTRIBUTING.md](CONTRIBUTING.md)
explains how to recreate that.

txm is dual-licensed MIT OR Apache-2.0, and formul8 keeps the same licence and
the upstream copyright notice. [NOTICE](NOTICE) records the provenance, the
baseline commit and the one behavioural difference.

## Packages

| Package | Role |
| --- | --- |
| `token` | Lexer and token types |
| `ast` | Expression tree |
| `parser` | Recursive-descent parser producing an `ast.Expr` |
| `glyph` | Glyph types, alphabet tables and the symbol registry |
| `layout` | Sized layout tree and its constructors, the single most fidelity-critical code |
| `render` | AST to layout tree |
| `backend` | Character buffer, positioning and the ANSI serialiser |
| `style` | Colour and text attributes |
| `errs` | Parse and render errors |
| `formul8` (root) | Public `Layout` and `Render` entry points and the default registry |

## Requirements

Go 1.26.5 or newer. The `go` directive in [go.mod](go.mod) carries that
minimum, and CI reads it from there rather than from a hardcoded version.

## Build and test

```bash
git clone https://github.com/Anuj-72/formul8.git
cd formul8
go build ./...
go test ./...
```

## Usage

No command-line interface exists yet; it is checkpoint #8. The intended
interface mirrors the reference, which takes the expression as an argument or
reads it from a file, and prints a bordered box unless `--unboxed` is given:

```bash
formul8 'E = mc^2'
formul8 --unboxed '\sum_{i=0}^{n} i^2'
formul8 --file expression.tex
printf '\\frac{1}{2}' | formul8 --file -
```

The library entry points arrive with #7 and will look like this:

```go
out, err := formul8.Render(`\frac{1}{2}`)
```

Neither form works today. They are documented here so the shape of the finished
port is visible while it is being built, and this section changes as soon as the
code makes it true.

## Matrix environments

The reference accepts `matrix`, `bmatrix` and `pmatrix`. formul8 adds the other
three LaTeX matrix environments the reference rejects, `Bmatrix`, `vmatrix` and
`Vmatrix`, in milestone v0.2.0 and issue #10. This is the only intended
difference in behaviour between the two projects, and it is labelled `extension`
wherever it appears so the parity milestone stays honest.

## Documentation

- [docs/parity.md](docs/parity.md) - module-by-module port status.
- [docs/github/README.md](docs/github/README.md) - the issue, branch, pull-request
  and release process this repository follows, including the branch ruleset.
- [CONTRIBUTING.md](CONTRIBUTING.md) - how to build, test and propose a change.

## Licence

MIT OR Apache-2.0, at your option.

Copyright (c) 2026 Anuj-72
Copyright (c) 2026 thatmagicalcat, for the ported work

See [LICENSE-MIT](LICENSE-MIT), [LICENSE-APACHE](LICENSE-APACHE) and
[NOTICE](NOTICE).
