## Context

`render/render.go` is currently a 24-line design note with no implementation. The parser can produce an AST but nothing turns that AST into the sized layout tree, so no backend can consume it. This issue is the bridge between parsing and positioning.

## Scope

- [ ] Implement `Render(expr, registry, ctx)` with one type-switch case for every `ast.Expr` variant (`reference/txm/src/render.rs:6-183`).
- [ ] For `Ident`, render the name as text; un-italicise only `(`, `)`, `[` and `]`, matching `render.rs:11-22`.
- [ ] For `Number`, render upright text and leave the style un-italicised.
- [ ] For `Delimiter`, render the inner expression and wrap it in `layout.StretchyDelimOf` with `fill=false` and an un-italicised style.
- [ ] For `Neg`, `Prime` and `BinOpExpr`, build the corresponding layout nodes and apply the current style.
- [ ] For `Command`, look the name up in the registry. On a hit, increment `ctx.Depth`, call the free function `glyph.RenderMacro`, decrement `ctx.Depth`, and propagate the error. On a miss, render the macro name as literal text in the current style.
- [ ] For `Superscript`, `Subscript` and `BothScripts`, use the shared helper that tests whether the base is a command whose glyph reports `HasLimits`; if so, build a `Limits` node with `layout.NewEmpty()` on the missing side; otherwise call the ordinary script constructors.
- [ ] For `Escape`, map `\ ` to four spaces, `\,` to one space, `\:` to two colons, `\;` to three semicolons and `\!` to an empty node, and render anything else literally.
- [ ] For `Juxtapose`, render every item and combine them with a horizontal stack at spacing zero.
- [ ] For `Matrix`, reject rows whose length differs from the first row with `errs.MismatchedMatrixRows`, render every cell, and finish with a matrix node in the un-italicised current style.
- [ ] For `Empty`, return `layout.NewEmpty()`.
- [ ] Port `renderPower` including the `compactSimpleFractionalExponents` branch, even though the flag is false and the branch is currently dead (`reference/txm/src/render.rs:185-208`).
- [ ] Add one helper for the repeated "base is a command whose glyph has limits" test, used by the three script cases, instead of repeating the type assertion.

## Acceptance criteria

- [ ] Every AST variant has a test that asserts the node kind, dimensions and style.
- [ ] A superscript over `\sum` produces a `Limits` node with an empty lower side; a subscript produces an empty upper side; both scripts produce both sides.
- [ ] `\color{red}{x+y}` pushes and restores the render context style, and the restored style is identical to the style before the call.
- [ ] A matrix with ragged rows returns `errs.MismatchedMatrixRows` instead of panicking.
- [ ] An unknown command such as `\doesnotexist` renders its name as upright text rather than failing.
- [ ] `go test ./render/...` passes.

## Reference

`reference/txm/src/render.rs` is 241 lines. The main match is at `:6-183`, the power helper at `:185-208`, and the style rules are visible in the individual arms. `reference/txm/src/glyph.rs:14-54` defines the default `render_macro` that the free function replaces.

## Out of scope

Character placement, ANSI serialization, the registry contents and the CLI. This issue ends when an AST becomes a correctly sized layout tree.

## Dependencies

Depends on #1. Blocks #7 and #9.
