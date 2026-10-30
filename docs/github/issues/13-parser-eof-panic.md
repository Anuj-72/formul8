## Context

A formula that ends in the middle of an argument aborts the process instead of returning a parse error, so a one-character typo such as `x^` kills the renderer with a stack trace. Found while completing #1, and drafted here rather than patched inside that pull request: the fix changes the parser's error contract, which is a separate decision from completing the grammar.

## Reproduction

Reference, exit status 101:

```console
$ reference/txm/target/debug/txm --unboxed '\frac{1}'
thread 'main' panicked at src/parser.rs:39:17:
index out of bounds: the len is 4 but the index is 4
$ echo $?
101
```

Port, same kind of input, through a `recover()` harness in the `parser` package:

```console
\frac        PANIC runtime error: index out of range [1] with length 1
\frac{1}     PANIC runtime error: index out of range [4] with length 4
\sqrt        PANIC runtime error: index out of range [1] with length 1
\sqrt[3]     PANIC runtime error: index out of range [4] with length 4
\color{red}  PANIC runtime error: index out of range [6] with length 6
\sum_        PANIC runtime error: index out of range [2] with length 2
\int_        PANIC runtime error: index out of range [2] with length 2
\lim_        PANIC runtime error: index out of range [2] with length 2
x^           PANIC runtime error: index out of range [2] with length 2
x_           PANIC runtime error: index out of range [2] with length 2
```

The trigger is one rule in three shapes: the last token asks for something that is not there. A macro's required argument (`\frac{1}`), the body of a script operator (`x^`, `\sum_`), or an atom after a command (`\frac`).

Valid and zero-arity input is unaffected, and one branch already reports end of input properly: `\frac{1}{2}` renders, `\sum` renders, and `\color` returns `error: expected LBrace but reached end of input near 'color'`. That wording is the model a fix should follow.

Two reference panics are not evidence about the port yet: `\mathbf` and `\binom` are missing from the parser test registry (the full registry is #7), so the unregistered-name path returns an empty `Command` with no error before any argument is read. With the real registry `\binom` reaches the same required-argument loop as `\frac`.

## Cause

Rust: `peek()` at `reference/txm/src/parser.rs:31-33` checks `self.tokens.get(self.pos)`, but `advance()` at `:38-42` indexes `self.tokens[self.pos]` directly. `parse_command` (`:531-568`) calls `advance()` once per required argument, treating `glyph.arity()` as a promise that tokens exist rather than a count of arguments the macro wants.

Go: `peek()` at `parser/parser.go:70-75` is the bounds-checked one, and `advance()` at `:84-88` is `t := p.tokens[p.pos]`. `parseCommand` (`parser/parser.go:526`) reaches the panic through its `default: p.parseAtom()` arm when no token is available, and `parseAtom` (`:214`) opens with an unconditional `p.advance()`.

## Why this is not parity

The reference exits 101 on a panic; Go exits 2 on an unrecovered one and prints a goroutine trace to standard error. 2 is also the reference's exit status for a usage error (`reference/txm/src/main.rs:63-66`), so without this fix #8 would report a bad flag and a crashed parser with the same status. The fix therefore diverges deliberately and carries `extension`, not `parity`.

## Acceptance criteria

- [ ] A required argument, script body or atom missing at end of input returns an `error` from `Parser.ParseExpr`, never a panic.
- [ ] The error is reported at end of input and names what was expected, reusing the existing wording `expected LBrace but reached end of input near '...'`.
- [ ] A table-driven test covers `\frac`, `\frac{1}`, `\sqrt`, `\sqrt[3]`, `\color{red}`, `\sum_`, `\int_`, `\lim_`, `x^` and `x_`, plus the unaffected neighbours `\frac{1}{2}`, `\sqrt[3]{8}`, `\color{red}{x}` and `x^2`.
- [ ] A guard test parses every prefix of a set of valid formulas under `recover()`, where a non-nil `recover()` result is the failure, so any future unchecked index fails the suite instead of shipping.
- [ ] The divergence and its rationale are recorded in `docs/parity.md` beside the existing `\left|x\right|` note.
- [ ] `gofmt -l .`, `go vet ./...` and `go test ./...` stay clean.

## Possible approach

Give the cursor one checked primitive that returns `(token.Token, bool)` the way `peek()` does, and have the callers that currently trust the token stream — `parseAtom`, the required-argument loop and the script parser — return the end-of-input error instead of indexing. The narrower alternative is a single `ok` check in `parseCommand`'s loop. The first closes every call site at once, which is why the failing tests above belong before either choice.

## Out of scope

Error recovery, unknown-macro handling, and the reference's own odd message for `\frac{1}{` and `\sqrt[`, which report `unclosed \left ... \right pair`. This issue makes the parser total over truncated input; it does not improve what the parser says about anything else.

## Dependencies

Found in #1. Should land before #11 so the v0.1.0 build does not ship a parser that panics on `x^`, and it shares its exit-status reasoning with #8.
