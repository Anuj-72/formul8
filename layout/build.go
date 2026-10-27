package layout

import (
	"github.com/Anuj-72/formul8/ast"
	"github.com/Anuj-72/formul8/errs"
	"github.com/Anuj-72/formul8/style"
)

// Node constructors. This file is the sizing math and is the single most
// fidelity-critical part of the port. Ports layout_tree.rs:113-631.
//
//   NewEmpty                       layout_tree.rs:114
//   NewText(content []rune)        :124   Width = len(content), rune count
//   NewTextWithStyle               :135
//   NewFromChar                    :146
//   NewTextStr(s string)           :150   []rune(s), never len(string)
//   HStackOf(children, spacing)    :154   1 child returns it directly,
//                                         0 children returns empty -- both
//                                         short-circuits are load-bearing
//   VStackOf(top, bottom, line, style)  :187
//   InfixOf(lhs, op, rhs)          :208
//   SuperscriptOf(base, exp, style):229   inline path if exp.Height == 1 and
//                                         every rune maps to a superscript
//   SubscriptOf(base, sub, style)  :265
//   BothScriptsOf                  :300
//   StretchyDelimOf                :324   separate path for Height <= 1
//   AccentOf                       :364
//   LimitsOf                       :381
//   SqrtOf(inner, index)           :399
//   SummationOf(inner)             :416   three size branches
//   ProductOf(inner)               :454
//   IntegralOf(inner)              :488   three size branches
//   NegateOf                       :525
//   PrimeOf                        :541
//   MatrixOf(name, rows, style)    :554   only constructor that can fail

// Helpers this package needs:
//   satSub(a, b int) int   -- Rust saturating_sub. Use ONLY where the
//                             reference wrote saturating_sub. Every plain
//                             '-' in the reference is proven non-negative
//                             (Rust would have panicked otherwise), so port
//                             those as plain '-'.
//   isPowerOfTwo(n int)    -- n > 0 && n&(n-1) == 0
//
// Go 1.26 has builtin max/min, so Rust's .max()/.min() chains port directly.
// (1.5 * h as f32) as usize  ->  int(1.5 * float64(h)); both truncate.

func NewEmpty() *Node {
	return &Node{
		Width:    0,
		Height:   0,
		Baseline: 0,
		Style:    style.New(),
		Kind:     Empty{},
	}
}

func NewText(content []rune) *Node {
	return &Node{
		Width:    len(content),
		Height:   1,
		Baseline: 0,
		Style:    style.New(),
		Kind:     Text{Content: content},
	}
}

func NewTextWithStyle(content []rune, st style.Style) *Node {
	return &Node{
		Width:    len(content),
		Height:   1,
		Baseline: 0,
		Style:    st,
		Kind:     Text{Content: content},
	}
}

func NewFromChar(char rune) *Node {
	return NewText([]rune{char})
}

func NewTextStr(s string) *Node {
	return NewText([]rune(s))
}

func satSub(a, b int) int {
	if a < b {
		return 0
	}
	return a - b
}

func HStackOf(children []*Node, spacing int) *Node {
	if len(children) == 0 {
		return NewEmpty()
	}
	if len(children) == 1 {
		return children[0]
	}

	newBaseline := 0
	for _, c := range children {
		newBaseline = max(newBaseline, c.Baseline)
	}

	newHeight := 0
	newWidth := 0
	for _, c := range children {
		newHeight = max(newHeight, newBaseline+satSub(c.Height, c.Baseline))
		newWidth += c.Width
	}

	newWidth += spacing * (len(children) - 1)

	childrenCopy := make([]*Node, len(children))
	copy(childrenCopy, children)
	newKind := HStack{Children: childrenCopy, Spacing: spacing}

	return &Node{
		Width:    newWidth,
		Height:   newHeight,
		Baseline: newBaseline,
		Style:    style.New(),
		Kind:     newKind,
	}
}

func VStackOf(top, bottom *Node, line LineStyle, st style.Style) *Node {
	innerW := max(top.Width, bottom.Width)
	pad := 1
	maxH := max(top.Height, bottom.Height)
	w := innerW + 2*pad
	h := maxH*2 + 1
	baseline := maxH

	newKind := VStack{Top: top, Bottom: bottom, Line: line}

	return &Node{
		Width:    w,
		Height:   h,
		Baseline: baseline,
		Style:    st,
		Kind:     newKind,
	}
}

func InfixOf(lhs *Node, op ast.BinOp, rhs *Node) *Node {
	baseline := max(lhs.Baseline, rhs.Baseline)
	lhsY := baseline - lhs.Baseline
	rhsY := baseline - rhs.Baseline

	width := lhs.Width + 3 + rhs.Width
	height := max(lhsY+lhs.Height, rhsY+rhs.Height)

	newKind := Infix{Lhs: lhs, Rhs: rhs, Op: op}

	return &Node{
		Width:    width,
		Height:   height,
		Baseline: baseline,
		Style:    style.New(),
		Kind:     newKind,
	}
}

func SuperscriptOf(base, exp *Node, st style.Style) *Node {
	if exp.Height == 1 {
		if converted := mapContent(exp.Kind, toSuperscriptChar); len(converted) == exp.Width {
			return &Node{
				Width:    base.Width + exp.Width,
				Height:   base.Height,
				Baseline: base.Baseline,
				Style:    st,
				Kind: Superscript{
					Inline: true,
					Base:   base,
					Exp:    NewTextWithStyle(converted, style.New()),
				},
			}
		}
	}

	height := exp.Height + base.Height
	width := exp.Width + base.Width
	baseline := base.Baseline + exp.Height
	inline := false

	return &Node{
		Width:    width,
		Height:   height,
		Baseline: baseline,
		Style:    st,
		Kind: Superscript{
			Inline: inline,
			Base:   base,
			Exp:    exp,
		},
	}
}

func SubscriptOf(base, sub *Node, st style.Style) *Node {
	if sub.Height == 1 {
		if converted := mapContent(sub.Kind, toSubscriptChar); len(converted) == sub.Width {
			return &Node{
				Width:    base.Width + sub.Width,
				Height:   base.Height,
				Baseline: base.Baseline,
				Style:    st,
				Kind: Subscript{
					Inline: true,
					Base:   base,
					Sub:    NewTextWithStyle(converted, st),
				},
			}
		}
	}

	height := base.Height + sub.Height
	width := base.Width + sub.Width
	inline := false

	return &Node{
		Width:    width,
		Height:   height,
		Baseline: base.Baseline,
		Style:    st,
		Kind: Subscript{
			Inline: inline,
			Base:   base,
			Sub:    sub,
		},
	}
}

func BothScriptsOf(base, sub, sup *Node, st style.Style) *Node {
	supH := sup.Height
	height := supH + sub.Height + base.Height

	width := base.Width + max(sub.Width, sup.Width)
	baseline := base.Baseline + supH

	return &Node{
		Width:    width,
		Height:   height,
		Baseline: baseline,
		Style:    st,
		Kind: BothScripts{
			Base: base,
			Sub:  sub,
			Sup:  sup,
		},
	}
}

func StretchyDelimOf(inner *Node, left, right rune, fill bool, st style.Style) *Node {
	if inner.Height <= 1 {
		w := inner.Width + 2
		return &Node{
			Width:    w,
			Height:   1,
			Baseline: 0,
			Style:    st,
			Kind: StretchyDelim{
				Inner: inner,
				Left:  left,
				Right: right,
				Fill:  fill,
			},
		}
	}

	h := inner.Height
	w := inner.Width + 4

	return &Node{
		Width:    w,
		Height:   h,
		Baseline: inner.Baseline,
		Style:    st,
		Kind: StretchyDelim{
			Inner: inner,
			Left:  left,
			Right: right,
			Fill:  fill,
		},
	}
}

func AccentOf(inner *Node, mark rune, stretch bool) *Node {
	width := max(inner.Width, 1)
	height := inner.Height + 1

	return &Node{
		Width:    width,
		Height:   height,
		Baseline: inner.Baseline + 1,
		Style:    style.New(),
		Kind: Accent{
			Inner:   inner,
			Mark:    mark,
			Stretch: stretch,
		},
	}
}

func LimitsOf(base, lower, upper *Node) *Node {
	maxH := max(upper.Height, lower.Height)
	w := max(base.Width, lower.Width, upper.Width) + 1
	h := base.Height + 2*maxH

	return &Node{
		Width:    w,
		Height:   h,
		Baseline: maxH + base.Baseline,
		Style:    style.New(),
		Kind: Limits{
			Base:  base,
			Lower: lower,
			Upper: upper,
		},
	}
}

func SqrtOf(inner *Node, index *Node) *Node {
	h := inner.Height + 1
	w := inner.Width + 3
	baseline := inner.Baseline + 1

	return &Node{
		Width:    w,
		Height:   h,
		Baseline: baseline,
		Style:    style.New(),
		Kind: Sqrt{
			Inner: inner,
			Index: index,
		},
	}
}

func SummationOf(inner *Node) *Node {
	if inner == nil {
		return &Node{
			Width:    5,
			Height:   3,
			Baseline: 1,
			Style:    style.New(),
			Kind:     Summation{Inner: nil},
		}
	}

	if inner.Height <= 2 {
		return &Node{
			Width:    inner.Width + 5,
			Height:   3,
			Baseline: 1,
			Style:    style.New(),
			Kind:     Summation{Inner: inner},
		}
	}

	h := inner.Height
	wSigma := max(int(1.5*float64(h)), h/2+2)
	w := wSigma + 1 + inner.Width
	return &Node{
		Width:    w,
		Height:   h,
		Baseline: inner.Baseline,
		Style:    style.New(),
		Kind:     Summation{Inner: inner},
	}
}

func ProductOf(inner *Node) *Node {
	if inner == nil {
		return &Node{
			Width:    4,
			Height:   2,
			Baseline: 1,
			Style:    style.New(),
			Kind:     Product{Inner: nil},
		}
	}

	height := max(inner.Height, 3)
	wPi := max(height/2+2, 3)
	if height <= 2 {
		wPi = 3
	}

	width := wPi + 1 + inner.Width
	heightDiff := satSub(height, inner.Height)
	baseline := inner.Baseline + (heightDiff / 2)

	return &Node{
		Width:    width,
		Height:   height,
		Baseline: baseline,
		Style:    style.New(),
		Kind: Product{
			Inner: inner,
		},
	}
}

func IntegralOf(inner *Node) *Node {
	if inner == nil {
		return &Node{
			Width:    2,
			Height:   3,
			Baseline: 1,
			Style:    style.New(),
			Kind:     Integral{Inner: nil},
		}
	}

	if inner.Height <= 3 {
		return &Node{
			Width:    inner.Width + 1,
			Height:   3,
			Baseline: 1,
			Style:    style.New(),
			Kind:     Integral{Inner: inner},
		}
	}

	return &Node{
		Width:    inner.Width + 2,
		Height:   inner.Height,
		Baseline: inner.Baseline,
		Style:    style.New(),
		Kind:     Integral{Inner: inner},
	}
}

func NegateOf(inner *Node) *Node {
	width := inner.Width + 1
	height := inner.Height
	baseline := inner.Baseline

	return &Node{
		Width:    width,
		Height:   height,
		Baseline: baseline,
		Style:    style.New(),
		Kind:     Neg{Inner: inner},
	}
}

func PrimeOf(base *Node, count int) *Node {
	return &Node{
		Width:    base.Width + count,
		Height:   base.Height,
		Baseline: base.Baseline,
		Style:    style.New(),
		Kind: Prime{
			Base:  base,
			Count: count,
		},
	}
}

func MatrixOf(name string, rows [][]*Node, st style.Style) (*Node, error) {
	if len(rows) == 0 || len(rows[0]) == 0 {
		return NewEmpty(), nil
	}

	numRows, numCols := len(rows), len(rows[0])

	rowMaxDepths := make([]int, numRows)
	rowMaxBaselines := make([]int, numRows)
	maxItemWidth := 0

	for i, row := range rows {
		maxB, maxD := 0, 0
		for _, item := range row {
			maxItemWidth = max(maxItemWidth, item.Width)
			maxB = max(maxB, item.Baseline)
			maxD = max(maxD, satSub(item.Height, item.Baseline))
		}
		rowMaxBaselines[i] = maxB
		rowMaxDepths[i] = maxD
	}

	cellWidth, cellHeight := maxItemWidth, 0
	for i := range numRows {
		rowContentHeight := rowMaxBaselines[i] + rowMaxDepths[i]
		cellHeight = max(cellHeight, rowContentHeight)
	}

	hSpacing, vSpacing := 4, 1
	matrixLayoutHeight := numRows*cellHeight + (numRows-1)*vSpacing
	matrixLayoutWidth := numCols*cellWidth + (numCols-1)*hSpacing

	if matrixLayoutHeight%2 == 0 {
		matrixLayoutHeight++
	}

	baseline := matrixLayoutHeight / 2
	var leftDelim, rightDelim rune

	// Three extra added deliberately
	switch name {
	case "matrix":
		leftDelim, rightDelim = ' ', ' '
	case "bmatrix":
		leftDelim, rightDelim = '[', ']'
	case "pmatrix":
		leftDelim, rightDelim = '(', ')'
	case "Bmatrix":
		leftDelim, rightDelim = '{', '}'
	case "vmatrix":
		leftDelim, rightDelim = '|', '|'
	case "Vmatrix":
		leftDelim, rightDelim = '‖', '‖' // Double vertical line rune (\u2016)
	default:
		return nil, errs.UnknownEnvironment(name)
	}

	copiedRows := make([][]*Node, numRows)
	copy(copiedRows, rows)
	inner := &Node{
		Width:    matrixLayoutWidth,
		Height:   matrixLayoutHeight,
		Baseline: baseline,
		Style:    style.New(),
		Kind: Matrix{
			Name: name,
			Rows: copiedRows,
		},
	}

	return StretchyDelimOf(
		inner,
		leftDelim,
		rightDelim,
		true,
		st,
	), nil

}
