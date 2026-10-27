// Package layout holds the sized layout tree: a node carries its own
// width, height and baseline, computed eagerly bottom-up at construction.
//
// Ports reference/txm/src/layout_tree.rs.
package layout

import (
	"github.com/Anuj-72/formul8/ast"
	"github.com/Anuj-72/formul8/style"
)

// Node mirrors LayoutNode exactly: shared geometry plus a Kind payload.
// Ports layout_tree.rs:4-11.
//
//   Node{Width, Height, Baseline int; Style style.Style; Kind Kind}
//
// Nodes are passed as *Node and shared, not deep-copied. Every site in the
// reference that mutates a node (node.style = ...) does so on a freshly
// constructed one, so sharing is safe -- but treat any NEW mutation of a
// possibly-shared node as a bug.

// Kind is the sealed payload sum type, porting NodeKind at
// layout_tree.rs:13-105. Marker method: isKind().
//
// Variants:
//   Text{Content []rune}
//   HStack{Children []*Node; Spacing int}
//   VStack{Top, Bottom *Node; Line LineStyle}
//   Infix{Lhs, Rhs *Node; Op ast.BinOp}
//   Superscript{Inline bool; Base, Exp *Node}
//   Subscript{Inline bool; Base, Sub *Node}
//   BothScripts{Base, Sub, Sup *Node}
//   StretchyDelim{Inner *Node; Left, Right rune; Fill bool}
//   Accent{Inner *Node; Mark rune; Stretch bool}
//   Limits{Base, Lower, Upper *Node}
//   Sqrt{Inner *Node; Index *Node}          // Index nil = no optional index
//   Matrix{Name string; Rows [][]*Node}
//   Summation{Inner *Node}                  // nil Inner = bare symbol
//   Product{Inner *Node}
//   Integral{Inner *Node}
//   Neg{Inner *Node}
//   Prime{Base *Node; Count int}
//   Empty{}
//
// Option<Box<LayoutNode>> becomes a plain nil-able *Node.

// LineStyle is the iota enum porting layout_tree.rs:107-111: Solid, None.

// Clone deep-copies a node. Not needed by the port as written, but keep it
// available for any site that starts mutating shared nodes.

type Kind interface{ isKind() }

type Node struct {
	Width    int
	Height   int
	Baseline int
	Style    style.Style
	Kind     Kind
}

type LineStyle int

const (
	Solid LineStyle = iota
	None
)

// Variants
type Text struct{ Content []rune }

func (Text) isKind() {}

type HStack struct {
	Children []*Node
	Spacing  int
}

func (HStack) isKind() {}

type VStack struct {
	Top    *Node
	Bottom *Node
	Line   LineStyle
}

func (VStack) isKind() {}

type Infix struct {
	Lhs *Node
	Rhs *Node
	Op  ast.BinOp
}

func (Infix) isKind() {}

type Superscript struct {
	Inline bool
	Base   *Node
	Exp    *Node
}

func (Superscript) isKind() {}

type Subscript struct {
	Inline bool
	Base   *Node
	Sub    *Node
}

func (Subscript) isKind() {}

type BothScripts struct {
	Base *Node
	Sub  *Node
	Sup  *Node
}

func (BothScripts) isKind() {}

type StretchyDelim struct {
	Inner *Node
	Left  rune
	Right rune
	Fill  bool
}

func (StretchyDelim) isKind() {}

type Accent struct {
	Inner   *Node
	Mark    rune
	Stretch bool
}

func (Accent) isKind() {}

type Limits struct {
	Base  *Node
	Lower *Node
	Upper *Node
}

func (Limits) isKind() {}

type Sqrt struct {
	Inner *Node
	Index *Node
}

func (Sqrt) isKind() {}

type Matrix struct {
	Name string
	Rows [][]*Node
}

func (Matrix) isKind() {}

type Summation struct {
	Inner *Node
}

func (Summation) isKind() {}

type Product struct {
	Inner *Node
}

func (Product) isKind() {}

type Integral struct {
	Inner *Node
}

func (Integral) isKind() {}

type Neg struct {
	Inner *Node
}

func (Neg) isKind() {}

type Prime struct {
	Base  *Node
	Count int
}

func (Prime) isKind() {}

type Empty struct{}

func (Empty) isKind() {}
