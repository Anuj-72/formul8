## Context

`txm.go` is a 36-line design note. There is no default registry, no `Layout` and no `Render`, so every LaTeX command currently renders as its literal name even once the parser and backends exist. The registry is what makes the library a math renderer rather than a text formatter.

## Scope

- [ ] Build the default registry from `reference/txm/src/lib.rs:55-256`, keeping the reference's grouping and names.
- [ ] Register the lowercase Greek letters and `vee`/`wedge` (`lib.rs:58-87`).
- [ ] Register the uppercase Greek letters (`lib.rs:89-102`).
- [ ] Register the function names as text glyphs: `sin`, `cos`, `tan`, `cot`, `sec`, `csc`, `arcsin`, `arccos`, `arctan`, `sinh`, `cosh`, `coth`, `tanh`, `log`, `ln`, `lg`, `det`, `dim`, `hom`, `ker`, `exp`, `deg`, `gcd`, `lcm`, `sup`, `inf`, `max`, `min`, `arg`, `Pr`, `mod`, `adj` (`lib.rs:104-110`).
- [ ] Register `binom`, `frac`, `sqrt`, `lim`, `int`, `sum`, `prod` and `color` (`lib.rs:112-120`).
- [ ] Register the operators, relations, arrows, set symbols, delimiters, dots and letterlike symbols (`lib.rs:122-192`).
- [ ] Register `abs` and the bare pipe (`lib.rs:194-195`).
- [ ] Register `mathbb` (`lib.rs:197`).
- [ ] Register `mathbf`, `mathrm`, `mathsf`, `mathit` and `boldsymbol` with their style modifiers and character maps (`lib.rs:199-211`).
- [ ] Register `text`, `textbf`, `textit` and `textrm` (`lib.rs:213-220`).
- [ ] Register `colorbox` (`lib.rs:222`).
- [ ] Register the unstretched accents and the stretched `overline`, `widehat` and `widetilde` (`lib.rs:224-253`).
- [ ] Replace the Rust `OnceLock` with a package-level `defaultRegistry` initialised by `buildRegistry`; package initialisation is single-threaded, so a plain variable is sufficient.
- [ ] Implement `Layout(input string) (*layout.Node, error)` as tokenize, parse, then render with a root context whose style is `style.New().Italic()` and depth zero (`lib.rs:31-41`).
- [ ] Implement `Render(input string) (string, error)` as `Layout` followed by `backend.RenderTerminal` (`lib.rs:44-48`).
- [ ] Keep the `compactSimpleFractionalExponents` constant and wire it into the renderer's power helper, even though it is false (`lib.rs:28`).
- [ ] Add package documentation that states which upstream module each package ports.

## Acceptance criteria

- [ ] `formul8.Render("E = mc^2")` returns a non-empty ANSI string with an italic `E`, an `=`, an italic `m`, an upright `c` and a superscript `2`.
- [ ] A test asserts that every command name from `lib.rs:55-256` is present in the registry; the count and the names must not silently drift.
- [ ] An unknown command still renders its name as literal text.
- [ ] `\mathbb{R}` renders `ℝ`, `\mathbf{x}` renders a bold math `x`, and `\text{x}` renders an upright `x`.
- [ ] `go test ./...` passes.

## Notes

There is no Go equivalent of Rust's `OnceLock` requirement here, but the registry must be immutable after construction. Do not expose the internal map; expose lookup through `Get`. If a later feature needs a second registry, add a constructor rather than mutating the default one.

## Reference

`reference/txm/src/lib.rs:28-256` contains the constant, the two public entry points and the full registry. `reference/txm/src/glyph.rs:56-74` defines the registry type. `reference/txm/src/style.rs:243-299` defines the color parser.

## Out of scope

The CLI, which is #8, and conformance fixtures, which are #9. This issue ends when the library API can render a formula from a Go program.

## Dependencies

Depends on #1, #3, #4, #5 and #6. Blocks #8 and #9.
