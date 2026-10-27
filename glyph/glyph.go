// Package glyph is the macro registry: one implementation per LaTeX macro
// family, fusing parse-time arity information with direct AST-to-layout
// construction.
//
// Ports reference/txm/src/glyph.rs.
package glyph

import (
	"github.com/Anuj-72/formul8/ast"
	"github.com/Anuj-72/formul8/layout"
	"github.com/Anuj-72/formul8/style"
)

// RenderCtx carries the current nesting depth and inherited style.
// Ports glyph.rs:8-12.

// EvalFunc is the callback the renderer hands to a glyph so the glyph can
// recurse into its own arguments. This is what breaks the glyph <-> render
// cycle without an import cycle:
//
//   type EvalFunc func(e ast.Expr, ctx *RenderCtx) (*layout.Node, error)

// Glyph is the interface porting the Rust trait at glyph.rs:14-54:
//
//   RequiredArgs() int
//   HasOptional() bool
//   HasLimits() bool
//   TakesStringArg() bool
//   Render(args, opts []*layout.Node, ctx *RenderCtx) *layout.Node
//
// Note render_macro is deliberately NOT in this interface -- see below.

// Base carries the trait's default method bodies and is embedded by every
// concrete glyph, which then overrides only what it needs.
//
//   RequiredArgs -> 0, HasOptional/HasLimits/TakesStringArg -> false,
//   Render -> NewEmpty()

// MacroRenderer is an OPTIONAL capability interface. Only glyphs that need
// the raw ast.Expr arguments or must mutate ctx implement it:
// TextColorGlyph, BgColorGlyph, StyleModifierGlyph, MappedStyleGlyph.
//
//   RenderMacro(args, opts []ast.Expr, ctx *RenderCtx, eval EvalFunc) (*layout.Node, error)

// RenderMacro is a FREE FUNCTION, not a method, and it replaces the Rust
// trait's default render_macro (glyph.rs:31-49).
//
// Why: the Rust default calls self.render(...) and dispatches dynamically to
// the concrete impl. Go's struct embedding does NOT do virtual dispatch -- a
// promoted Base.RenderMacro could only ever reach Base.Render, never an
// override. This is the single biggest porting trap in the codebase.
//
//   func RenderMacro(g Glyph, args, opts []ast.Expr, ctx *RenderCtx,
//                    eval EvalFunc) (*layout.Node, error)
//
// Body: if g implements MacroRenderer, delegate to it. Otherwise eval every
// arg and every opt, then call g.Render on the results.
//
// render/render.go must call this function, never a method.

type RenderCtx struct {
	Depth        int
	CurrentStyle style.Style
}

type EvalFunc func(e ast.Expr, ctx *RenderCtx) (*layout.Node, error)

type Glyph interface {
	RequiredArgs() int
	HasOptional() bool
	HasLimits() bool
	TakesStringArg() bool
	Render(args, opts []*layout.Node, ctx *RenderCtx) *layout.Node
}

type Base struct{}

func (Base) RequiredArgs() int    { return 0 }
func (Base) HasOptional() bool    { return false }
func (Base) HasLimits() bool      { return false }
func (Base) TakesStringArg() bool { return false }
func (Base) Render(_ []*layout.Node, _ []*layout.Node, _ *RenderCtx) *layout.Node {
	return layout.NewEmpty()
}

type MacroRenderer interface {
	RenderMacro(args, opts []ast.Expr, ctx *RenderCtx, eval EvalFunc) (*layout.Node, error)
}

func RenderMacro(
	g Glyph,
	args,
	opts []ast.Expr,
	ctx *RenderCtx,
	eval EvalFunc,
) (*layout.Node, error) {
	if mr, ok := g.(MacroRenderer); ok {
		return mr.RenderMacro(args, opts, ctx, eval)
	}
	// default path: eval every arg and opt, then call g.Render
	renderedArgs := make([]*layout.Node, len(args))
	for i, arg := range args {
		node, err := eval(arg, ctx)
		if err != nil {
			return nil, err
		}
		renderedArgs[i] = node
	}
	renderedOpts := make([]*layout.Node, len(opts))
	for i, opt := range opts {
		node, err := eval(opt, ctx)
		if err != nil {
			return nil, err
		}
		renderedOpts[i] = node
	}
	return g.Render(renderedArgs, renderedOpts, ctx), nil
}
