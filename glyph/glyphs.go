package glyph

import (
	"github.com/Anuj-72/formul8/ast"
	"github.com/Anuj-72/formul8/errs"
	"github.com/Anuj-72/formul8/layout"
	"github.com/Anuj-72/formul8/style"
)

// The concrete glyph implementations. Ports every `impl Glyph for ...`
// block in glyph.rs.
//
// Each embeds Base and overrides only the methods it needs. Rust tuple
// structs have no Go equivalent, so UnicodeGlyph(pub char) becomes a struct
// with a named field.
//
//   LimitGlyph{Base}                     glyph.rs:77-93    HasLimits
//   UnicodeGlyph{Base; Ch rune}          glyph.rs:96-102
//   TextGlyph{Base; Text string}         glyph.rs:105-111
//   BinomGlyph{Base}                     glyph.rs:114-131  2 args
//   FracGlyph{Base}                      glyph.rs:134-149  2 args
//   SqrtGlyph{Base}                      glyph.rs:152-169  1 arg + optional
//   SummationGlyph{Base}                 glyph.rs:172-196  limits + optional
//   ProductGlyph{Base}                   glyph.rs:199-219  limits
//   IntegralGlyph{Base}                  glyph.rs:222-244  limits
//   AlphabetGlyph{Base; Map func(rune) rune}   glyph.rs:247-274
//   AbsGlyph{Base}                       glyph.rs:360-379
//   AccentGlyph{Base; Mark rune; Stretch bool} glyph.rs:382-397
//
// These four implement MacroRenderer instead of Render, because they need
// the raw ast.Expr arguments or must push and pop ctx.CurrentStyle:
//
//   TextColorGlyph{Base}                 glyph.rs:400-432
//   StyleModifierGlyph{Base; Modify func(style.Style) style.Style}
//                                        glyph.rs:435-461
//   MappedStyleGlyph{Base; Modify func(style.Style) style.Style;
//                          Map func(rune) rune}
//                                        glyph.rs:464-488
//   BgColorGlyph{Base}                   glyph.rs:507-537

// mapNodeChars rewrites Text runes in place through f, recursing into
// HStack only. Ports glyph.rs:490-504.
//
// Kind payloads are stored as interface values, so "mutating in place"
// means reassigning: n.Kind = Text{Content: mapped}.

type LimitGlyph struct{ Base }

func (LimitGlyph) HasLimits() bool { return true }

func (LimitGlyph) Render(_, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	node := layout.NewTextStr("lim")
	node.Style = ctx.CurrentStyle.UnItalic()
	return node
}

type UnicodeGlyph struct {
	Base
	Ch rune
}

func (u UnicodeGlyph) Render(_, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	return layout.NewTextWithStyle([]rune{u.Ch}, ctx.CurrentStyle.UnItalic())
}

type TextGlyph struct {
	Base
	Text string
}

func (t TextGlyph) Render(_, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	return layout.NewTextWithStyle([]rune(t.Text), ctx.CurrentStyle.UnItalic())
}

type BinomGlyph struct{ Base }

func (BinomGlyph) RequiredArgs() int { return 2 }

func (BinomGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	inner := layout.VStackOf(args[0], args[1], layout.None, ctx.CurrentStyle)
	return layout.StretchyDelimOf(inner, '(', ')', false, ctx.CurrentStyle.UnItalic())
}

type FracGlyph struct{ Base }

func (FracGlyph) RequiredArgs() int { return 2 }

func (FracGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	return layout.VStackOf(args[0], args[1], layout.Solid, ctx.CurrentStyle)
}

type SqrtGlyph struct{ Base }

func (SqrtGlyph) RequiredArgs() int { return 1 }

func (SqrtGlyph) HasOptional() bool { return true }

func (SqrtGlyph) Render(args, opts []*layout.Node, ctx *RenderCtx) *layout.Node {
	var index *layout.Node
	if len(opts) > 0 {
		index = opts[0]
	}
	node := layout.SqrtOf(args[0], index)
	node.Style = ctx.CurrentStyle

	return node
}

type SummationGlyph struct{ Base }

func (SummationGlyph) HasOptional() bool { return true }

func (SummationGlyph) HasLimits() bool { return true }

func (SummationGlyph) RequiredArgs() int { return 1 }

func (SummationGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	var inner *layout.Node
	if len(args) > 0 {
		inner = args[0]
	}
	node := layout.SummationOf(inner)
	node.Style = ctx.CurrentStyle
	return node
}

type ProductGlyph struct{ Base }

func (ProductGlyph) HasLimits() bool { return true }

func (ProductGlyph) RequiredArgs() int { return 1 }

func (ProductGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	var inner *layout.Node
	if len(args) > 0 {
		inner = args[0]
	}
	node := layout.ProductOf(inner)
	node.Style = ctx.CurrentStyle
	return node
}

type IntegralGlyph struct{ Base }

func (IntegralGlyph) HasLimits() bool { return true }

func (IntegralGlyph) RequiredArgs() int { return 1 }

func (IntegralGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	var inner *layout.Node
	if len(args) > 0 {
		inner = args[0]
	}
	node := layout.IntegralOf(inner)
	node.Style = ctx.CurrentStyle
	return node
}

type AlphabetGlyph struct {
	Base
	Map func(rune) rune
}

func (AlphabetGlyph) RequiredArgs() int { return 1 }

func (a AlphabetGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	src := args[0]
	if t, ok := src.Kind.(layout.Text); ok {
		mapped := make([]rune, len(t.Content))
		for i, c := range t.Content {
			mapped[i] = a.Map(c)
		}
		return &layout.Node{
			Width:    src.Width,
			Height:   src.Height,
			Baseline: src.Baseline,
			Style:    ctx.CurrentStyle.Merge(src.Style),
			Kind:     layout.Text{Content: mapped},
		}
	}
	node := *src
	node.Style = ctx.CurrentStyle.Merge(node.Style)
	return &node
}

type AbsGlyph struct{ Base }

func (AbsGlyph) RequiredArgs() int { return 1 }

func (AbsGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	node := layout.StretchyDelimOf(
		args[0],
		'|', '|',
		false,
		ctx.CurrentStyle.UnItalic(),
	)
	node.Style = ctx.CurrentStyle
	return node
}

type AccentGlyph struct {
	Base
	Mark    rune
	Stretch bool
}

func (AccentGlyph) RequiredArgs() int { return 1 }

func (a AccentGlyph) Render(args, _ []*layout.Node, ctx *RenderCtx) *layout.Node {
	node := layout.AccentOf(args[0], a.Mark, a.Stretch)
	node.Style = ctx.CurrentStyle.UnItalic()
	return node
}

type TextColorGlyph struct{ Base }

func (TextColorGlyph) RequiredArgs() int { return 2 }

func (TextColorGlyph) TakesStringArg() bool { return true }

func (t TextColorGlyph) RenderMacro(
	args, _ []ast.Expr,
	ctx *RenderCtx,
	eval EvalFunc,
) (*layout.Node, error) {
	ident, ok := args[0].(*ast.Ident)
	if !ok {
		return nil, errs.ExpectedColorName()
	}

	prevStyle := ctx.CurrentStyle
	defer func() { ctx.CurrentStyle = prevStyle }()

	parsedColor, err := style.ParseColor(ident.Name)
	if err != nil {
		return nil, err
	}
	ctx.CurrentStyle = ctx.CurrentStyle.Fg(parsedColor)
	node, err := eval(args[1], ctx)

	return node, err
}

type BgColorGlyph struct{ Base }

func (BgColorGlyph) RequiredArgs() int { return 2 }

func (BgColorGlyph) TakesStringArg() bool { return true }

func (t BgColorGlyph) RenderMacro(
	args, _ []ast.Expr,
	ctx *RenderCtx,
	eval EvalFunc,
) (*layout.Node, error) {
	ident, ok := args[0].(*ast.Ident)
	if !ok {
		return nil, errs.ExpectedColorName()
	}

	prevStyle := ctx.CurrentStyle
	defer func() { ctx.CurrentStyle = prevStyle }()

	parsedColor, err := style.ParseColor(ident.Name)
	if err != nil {
		return nil, err
	}
	ctx.CurrentStyle = ctx.CurrentStyle.Bg(parsedColor)
	node, err := eval(args[1], ctx)

	return node, err
}

type StyleModifierGlyph struct {
	Base
	Modify func(style.Style) style.Style
}

func (StyleModifierGlyph) RequiredArgs() int { return 1 }

func (StyleModifierGlyph) TakesStringArg() bool { return true }

func (s StyleModifierGlyph) RenderMacro(
	args, _ []ast.Expr,
	ctx *RenderCtx,
	eval EvalFunc,
) (*layout.Node, error) {
	prev := ctx.CurrentStyle
	defer func() { ctx.CurrentStyle = prev }()

	ctx.CurrentStyle = s.Modify(ctx.CurrentStyle)
	result, err := eval(args[0], ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type MappedStyleGlyph struct {
	Base
	Modify func(style.Style) style.Style
	Map    func(rune) rune
}

func (MappedStyleGlyph) RequiredArgs() int { return 1 }

func (m MappedStyleGlyph) RenderMacro(
	args, _ []ast.Expr,
	ctx *RenderCtx,
	eval EvalFunc,
) (*layout.Node, error) {
	prev := ctx.CurrentStyle
	ctx.CurrentStyle = m.Modify(ctx.CurrentStyle)

	result, err := eval(args[0], ctx)
	if err != nil {
		return nil, err
	}

	ctx.CurrentStyle = prev
	mapNodeChars(result, m.Map) // To be written
	return result, nil
}

func mapNodeChars(n *layout.Node, f func(rune) rune) {
	switch k := n.Kind.(type) {
	case layout.Text:
		newContentChar := make([]rune, len(k.Content))
		for i, c := range k.Content {
			newContentChar[i] = f(c)
		}
		n.Kind = layout.Text{Content: newContentChar}
	case layout.HStack:
		for _, c := range k.Children {
			mapNodeChars(c, f)
		}
		// No Default deliberate
	}
}

var (
	_ Glyph = LimitGlyph{}
	_ Glyph = UnicodeGlyph{}
	_ Glyph = TextGlyph{}
	_ Glyph = BinomGlyph{}
	_ Glyph = FracGlyph{}
	_ Glyph = SqrtGlyph{}
	_ Glyph = SummationGlyph{}
	_ Glyph = ProductGlyph{}
	_ Glyph = IntegralGlyph{}
	_ Glyph = AlphabetGlyph{}
	_ Glyph = AbsGlyph{}
	_ Glyph = AccentGlyph{}
	_ Glyph = TextColorGlyph{}
	_ Glyph = BgColorGlyph{}
	_ Glyph = StyleModifierGlyph{}
	_ Glyph = MappedStyleGlyph{}
)

var (
	_ MacroRenderer = TextColorGlyph{}
	_ MacroRenderer = BgColorGlyph{}
	_ MacroRenderer = StyleModifierGlyph{}
	_ MacroRenderer = MappedStyleGlyph{}
)
