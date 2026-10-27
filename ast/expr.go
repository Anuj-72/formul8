// Package ast defines the expression tree produced by the parser.
//
// Ports reference/txm/src/ast.rs.
//
// Rust's single sum-type `enum Expr` becomes a sealed Go interface plus one
// struct per variant. Sealed = the interface has an unexported marker method,
// so no package outside ast can add a variant. Variants are used as pointers
// (*Command, *Ident, ...) so the parser can mutate them in place.
package ast

// Expr is the sealed sum type. Marker method: isExpr().

// Variants, porting ast.rs:2-28:
//   Ident{Name string}
//   Number{Text string}
//   Delimiter{Left, Right rune; Inner Expr}
//   Neg{Inner Expr}
//   Command{Name string; Opts, Args []Expr}
//   Superscript{Base, Sup Expr}
//   Subscript{Base, Sub Expr}
//   BothScripts{Base, Sub, Sup Expr}
//   Prime{Base Expr; Count int}
//   BinOpExpr{Lhs Expr; Op BinOp; Rhs Expr}
//   Juxtapose{Items []Expr}
//   Escape{Text string}
//   Empty{}
//   Matrix{Name string; Rows [][]Expr}

// BinOp is the iota enum porting ast.rs:31-36: Add, Sub, Eq, Mul.

type Expr interface {
	isExpr()
}

// -- Variants, ref to ast.rs ---

type Ident struct{ Name string }

func (*Ident) isExpr() {}

type Number struct{ Value string }

func (*Number) isExpr() {}

type Delimiter struct {
	Left  rune
	Right rune
	Inner Expr
}

func (*Delimiter) isExpr() {}

type Neg struct{ Inner Expr }

func (*Neg) isExpr() {}

type Command struct {
	Name string
	Opts []Expr
	Args []Expr
}

func (*Command) isExpr() {}

type Superscript struct {
	Base Expr
	Exp  Expr
}

func (*Superscript) isExpr() {}

type Subscript struct {
	Base Expr
	Sub  Expr
}

func (*Subscript) isExpr() {}

type BothScripts struct {
	Base Expr
	Sub  Expr
	Sup  Expr
}

func (*BothScripts) isExpr() {}

type Prime struct {
	Base  Expr
	Count int
}

func (*Prime) isExpr() {}

type BinOpExpr struct {
	Left  Expr
	Op    BinOp
	Right Expr
}

func (*BinOpExpr) isExpr() {}

type Juxtapose struct{ Items []Expr }

func (*Juxtapose) isExpr() {}

type Escape struct{ Ch string }

func (*Escape) isExpr() {}

type Empty struct{}

func (*Empty) isExpr() {}

type Matrix struct {
	Name string
	Rows [][]Expr
}

func (*Matrix) isExpr() {}

type BinOp int

const (
	Add BinOp = iota
	Sub
	Eq
	Mul
)
