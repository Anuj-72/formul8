## Context

`go build ./...` fails at `parser/parser.go:346` with a syntax error, and the parser is the only package that does not compile. The lexer, AST, style, layout tree and glyph layer are already in place, so this issue is the gate for every later checkpoint: without a parser there is no AST, and without an AST there is nothing to render.

## What is broken

- `parseJuxtapose` appends the same parsed node on every loop iteration without consuming another token, so any multi-atom input such as `x + y` loops forever (`parser/parser.go:429-482`).
- `parseBegin` rejects the environments it should accept and then stops at a dangling `switch p.tokens` (`parser/parser.go:329-347`).
- `parseLeftDelimited` and `parseCommand` return `not implemented` (`parser/parser.go:348-353`).
- `parseDelimitedArg`, `parseMatrixBody`, `parseTokens`, `isOpenerFor` and `tokenEq` are missing.
- The reference's `parse_left_delimited` has an `unreachable!()` for an unsupported left delimiter; that arm is reachable from user input (`reference/txm/src/parser.rs:472`). Return an error instead of panicking.

## Scope

- [ ] Fix `parseJuxtapose` so each iteration parses a fresh `parseScripted()` expression and appends that result.
- [ ] Implement `parseCommand` with registry-driven arity, optional `[...]` arguments, string arguments for `\color` and the text-style family, and the atom fallback for brace, parenthesis and bracket groups (`reference/txm/src/parser.rs:531-568`).
- [ ] Implement `parseLeftDelimited` with the depth scan for nested `\left`/`\right`, delimiter matching, and the `ExpectedDelimiter`, `MismatchedDelimiters` and `UnclosedLeftRight` errors (`reference/txm/src/parser.rs:420-485`).
- [ ] Implement `parseDelimitedArg`, `isOpenerFor` and `tokenEq` (`reference/txm/src/parser.rs:487-529`).
- [ ] Complete `parseBegin` for `matrix`, `bmatrix` and `pmatrix`: scan to the matching `\end`, parse the body, and verify the closing environment name (`reference/txm/src/parser.rs:570-622`). Keep `Bmatrix`, `vmatrix` and `Vmatrix` rejected here; they are the extension in #10.
- [ ] Implement `parseMatrixBody` so `&` and `\\` split cells and rows only at brace depth zero and nested-environment depth zero (`reference/txm/src/parser.rs:624-662`).
- [ ] Implement `parseTokens` as a sub-parser over a token slice, returning `Empty` for an empty slice (`reference/txm/src/parser.rs:664-678`).
- [ ] Add table-driven parser tests for continuous strings, optional arguments, string arguments, nested delimiters, matrix rows and malformed input.

## Acceptance criteria

- [ ] `go build ./...`, `go vet ./...` and `gofmt -l .` come back clean.
- [ ] `\frac{1}{2}`, `x^2`, `\sqrt[3]{8}`, `\color{red}{x+y}`, `\left(\frac{1}{2}\right)` and `\begin{bmatrix}a&b\\c&d\end{bmatrix}` parse into the expected AST.
- [ ] `\left(x` and `\right)` return the reference error variants instead of hanging or panicking.
- [ ] `\begin{matrix}a&b\\c\end{matrix}` is rejected with `MismatchedMatrixRows` only at render time, not by the parser.
- [ ] `go test ./parser/...` passes.

## Reference

`reference/txm/src/parser.rs` is the source of truth for the grammar. The file is 678 lines; the relevant regions are the cursor helpers at `:30-60`, string scanning at `:62-149`, the expression grammar at `:151-568`, matrix handling at `:570-662`, and the sub-parser at `:664-678`.

## Out of scope

Rendering, backends and CLI output. This issue ends when the parser produces the correct AST and the module builds.

## Dependencies

None. Blocks #3 and #7.
