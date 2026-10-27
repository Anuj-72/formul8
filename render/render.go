// Package render walks the AST and builds the sized layout tree.
//
// Ports reference/txm/src/render.rs.
package render

// Render(e ast.Expr, reg *glyph.SymbolRegistry, ctx *glyph.RenderCtx)
//   (*layout.Node, error)
//
// One type-switch case per ast.Expr variant, porting render.rs:6-183.
//
// For Command (render.rs:49-66): look the name up in the registry. On a hit,
// build the eval closure, increment ctx.Depth, call glyph.RenderMacro (the
// free function, never a method), decrement Depth. On a miss, render the
// macro name as literal text.
//
// For Superscript, Subscript and BothScripts (render.rs:68-118): if the base
// is a Command whose glyph HasLimits, build a Limits node with empty
// stand-ins for the missing side. Otherwise fall through to the ordinary
// script constructors. Rust used let-chains; use the shared helper.

// renderPower ports render.rs:185-208, including the
// compactSimpleFractionalExponents flag read from the root package.
// That constant is false in the reference, so the branch is dead -- port it
// anyway to keep the files aligned.
