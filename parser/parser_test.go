package parser

import (
	"testing"

	"github.com/Anuj-72/formul8/ast"
	"github.com/Anuj-72/formul8/errs"
	"github.com/Anuj-72/formul8/glyph"
	"github.com/Anuj-72/formul8/token"
)

// testRegistry carries the subset of the reference registry
// (reference/txm/src/lib.rs:55-160) whose arity the parser has to know: two
// required arguments, one optional argument, a string argument, a limits
// operator and an alphabet mapping. The full registry is #7.
func testRegistry() *glyph.SymbolRegistry {
	r := glyph.NewSymbolRegistry()
	r.Register("frac", glyph.FracGlyph{})
	r.Register("sqrt", glyph.SqrtGlyph{})
	r.Register("color", glyph.TextColorGlyph{})
	r.Register("sum", glyph.SummationGlyph{})
	r.Register("mathbb", glyph.AlphabetGlyph{Map: glyph.ToBB})
	return r
}

func parseInput(t *testing.T, input string) (ast.Expr, error) {
	t.Helper()

	tokens, err := token.Tokenize(input)
	if err != nil {
		t.Fatalf("token.Tokenize(%q) returned %v", input, err)
	}
	return New(input, tokens, testRegistry()).ParseExpr()
}

func asCommand(t *testing.T, got ast.Expr) *ast.Command {
	t.Helper()

	cmd, ok := got.(*ast.Command)
	if !ok {
		t.Fatalf("want *ast.Command, got %T", got)
	}
	return cmd
}

func asIdent(t *testing.T, got ast.Expr) string {
	t.Helper()

	ident, ok := got.(*ast.Ident)
	if !ok {
		t.Fatalf("want *ast.Ident, got %T", got)
	}
	return ident.Name
}

func asNumber(t *testing.T, got ast.Expr) string {
	t.Helper()

	num, ok := got.(*ast.Number)
	if !ok {
		t.Fatalf("want *ast.Number, got %T", got)
	}
	return num.Value
}

func asMatrix(t *testing.T, got ast.Expr) *ast.Matrix {
	t.Helper()

	m, ok := got.(*ast.Matrix)
	if !ok {
		t.Fatalf("want *ast.Matrix, got %T", got)
	}
	return m
}

// TestParseExpr checks the shape of the AST for the acceptance inputs of #1.
// Every input here was run through the reference CLI
// (reference/txm/target/debug/txm --unboxed '<input>') and parses there.
func TestParseExpr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, got ast.Expr)
	}{
		{
			name:  "fraction takes two brace arguments",
			input: `\frac{1}{2}`,
			check: func(t *testing.T, got ast.Expr) {
				cmd := asCommand(t, got)
				if cmd.Name != "frac" {
					t.Fatalf("command name = %q, want frac", cmd.Name)
				}
				if len(cmd.Args) != 2 {
					t.Fatalf("len(args) = %d, want 2", len(cmd.Args))
				}
				if num := asNumber(t, cmd.Args[0]); num != "1" {
					t.Errorf("args[0] = %q, want 1", num)
				}
				if num := asNumber(t, cmd.Args[1]); num != "2" {
					t.Errorf("args[1] = %q, want 2", num)
				}
			},
		},
		{
			name:  "superscript",
			input: `x^2`,
			check: func(t *testing.T, got ast.Expr) {
				sup, ok := got.(*ast.Superscript)
				if !ok {
					t.Fatalf("want *ast.Superscript, got %T", got)
				}
				if name := asIdent(t, sup.Base); name != "x" {
					t.Errorf("base = %q, want x", name)
				}
				if num := asNumber(t, sup.Exp); num != "2" {
					t.Errorf("exponent = %q, want 2", num)
				}
			},
		},
		{
			name:  "subscript and superscript",
			input: `x_1^2`,
			check: func(t *testing.T, got ast.Expr) {
				both, ok := got.(*ast.BothScripts)
				if !ok {
					t.Fatalf("want *ast.BothScripts, got %T", got)
				}
				if name := asIdent(t, both.Base); name != "x" {
					t.Errorf("base = %q, want x", name)
				}
				if num := asNumber(t, both.Sub); num != "1" {
					t.Errorf("subscript = %q, want 1", num)
				}
				if num := asNumber(t, both.Sup); num != "2" {
					t.Errorf("superscript = %q, want 2", num)
				}
			},
		},
		{
			name:  "prime",
			input: `x'`,
			check: func(t *testing.T, got ast.Expr) {
				prime, ok := got.(*ast.Prime)
				if !ok {
					t.Fatalf("want *ast.Prime, got %T", got)
				}
				if prime.Count != 1 {
					t.Errorf("count = %d, want 1", prime.Count)
				}
				if name := asIdent(t, prime.Base); name != "x" {
					t.Errorf("base = %q, want x", name)
				}
			},
		},
		{
			// The reference appended the node it had already parsed here, so
			// this input used to loop forever instead of terminating.
			name:  "binop terminates and nests left",
			input: `x + y`,
			check: func(t *testing.T, got ast.Expr) {
				binop, ok := got.(*ast.BinOpExpr)
				if !ok {
					t.Fatalf("want *ast.BinOpExpr, got %T", got)
				}
				if binop.Op != ast.Add {
					t.Errorf("op = %v, want Add", binop.Op)
				}
				if name := asIdent(t, binop.Left); name != "x" {
					t.Errorf("left = %q, want x", name)
				}
				if name := asIdent(t, binop.Right); name != "y" {
					t.Errorf("right = %q, want y", name)
				}
			},
		},
		{
			name:  "juxtaposition collects adjacent atoms",
			input: `xy`,
			check: func(t *testing.T, got ast.Expr) {
				jux, ok := got.(*ast.Juxtapose)
				if !ok {
					t.Fatalf("want *ast.Juxtapose, got %T", got)
				}
				if len(jux.Items) != 2 {
					t.Fatalf("len(items) = %d, want 2", len(jux.Items))
				}
				if name := asIdent(t, jux.Items[0]); name != "x" {
					t.Errorf("items[0] = %q, want x", name)
				}
				if name := asIdent(t, jux.Items[1]); name != "y" {
					t.Errorf("items[1] = %q, want y", name)
				}
			},
		},
		{
			name:  "optional argument",
			input: `\sqrt[3]{8}`,
			check: func(t *testing.T, got ast.Expr) {
				cmd := asCommand(t, got)
				if cmd.Name != "sqrt" {
					t.Fatalf("command name = %q, want sqrt", cmd.Name)
				}
				if len(cmd.Opts) != 1 {
					t.Fatalf("len(opts) = %d, want 1", len(cmd.Opts))
				}
				if num := asNumber(t, cmd.Opts[0]); num != "3" {
					t.Errorf("opts[0] = %q, want 3", num)
				}
				if len(cmd.Args) != 1 {
					t.Fatalf("len(args) = %d, want 1", len(cmd.Args))
				}
				if num := asNumber(t, cmd.Args[0]); num != "8" {
					t.Errorf("args[0] = %q, want 8", num)
				}
			},
		},
		{
			// TakesStringArg glyphs consume a bare string argument first, so
			// only the body is a real argument.
			name:  "string argument",
			input: `\color{red}{x+y}`,
			check: func(t *testing.T, got ast.Expr) {
				cmd := asCommand(t, got)
				if cmd.Name != "color" {
					t.Fatalf("command name = %q, want color", cmd.Name)
				}
				if len(cmd.Args) != 2 {
					t.Fatalf("len(args) = %d, want 2", len(cmd.Args))
				}
				if name := asIdent(t, cmd.Args[0]); name != "red" {
					t.Errorf("args[0] = %q, want red", name)
				}
				binop, ok := cmd.Args[1].(*ast.BinOpExpr)
				if !ok {
					t.Fatalf("args[1] = %T, want *ast.BinOpExpr", cmd.Args[1])
				}
				if binop.Op != ast.Add {
					t.Errorf("args[1] op = %v, want Add", binop.Op)
				}
			},
		},
		{
			// The left delimiter token is part of the inner slice, so the
			// inner expression is a juxtaposition of "(" and the fraction.
			name:  "nested delimiters",
			input: `\left(\frac{1}{2}\right)`,
			check: func(t *testing.T, got ast.Expr) {
				delim, ok := got.(*ast.Delimiter)
				if !ok {
					t.Fatalf("want *ast.Delimiter, got %T", got)
				}
				if delim.Left != '(' || delim.Right != ')' {
					t.Errorf("delimiters = %c%c, want ()", delim.Left, delim.Right)
				}
				jux, ok := delim.Inner.(*ast.Juxtapose)
				if !ok {
					t.Fatalf("inner = %T, want *ast.Juxtapose", delim.Inner)
				}
				if len(jux.Items) != 2 {
					t.Fatalf("len(inner items) = %d, want 2", len(jux.Items))
				}
				if name := asIdent(t, jux.Items[0]); name != "(" {
					t.Errorf("inner items[0] = %q, want (", name)
				}
				cmd := asCommand(t, jux.Items[1])
				if cmd.Name != "frac" {
					t.Errorf("inner items[1] command = %q, want frac", cmd.Name)
				}
			},
		},
		{
			name:  "bmatrix rows and cells",
			input: `\begin{bmatrix}a&b\\c&d\end{bmatrix}`,
			check: func(t *testing.T, got ast.Expr) {
				m := asMatrix(t, got)
				if m.Name != "bmatrix" {
					t.Errorf("environment = %q, want bmatrix", m.Name)
				}
				if len(m.Rows) != 2 {
					t.Fatalf("len(rows) = %d, want 2", len(m.Rows))
				}
				for i, row := range m.Rows {
					if len(row) != 2 {
						t.Fatalf("len(rows[%d]) = %d, want 2", i, len(row))
					}
				}
				if name := asIdent(t, m.Rows[0][0]); name != "a" {
					t.Errorf("rows[0][0] = %q, want a", name)
				}
				if name := asIdent(t, m.Rows[1][1]); name != "d" {
					t.Errorf("rows[1][1] = %q, want d", name)
				}
			},
		},
		{
			// The reference reports ragged rows at render time, not here:
			// `txm '\begin{matrix}a&b\\c\end{matrix}'` fails with
			// "matrix rows have different lengths" after parsing succeeds.
			name:  "ragged rows parse and are left for the renderer",
			input: `\begin{matrix}a&b\\c\end{matrix}`,
			check: func(t *testing.T, got ast.Expr) {
				m := asMatrix(t, got)
				if len(m.Rows) != 2 {
					t.Fatalf("len(rows) = %d, want 2", len(m.Rows))
				}
				if len(m.Rows[0]) != 2 || len(m.Rows[1]) != 1 {
					t.Errorf("row lengths = %d, %d, want 2, 1", len(m.Rows[0]), len(m.Rows[1]))
				}
			},
		},
		{
			// The reference parses this into one empty row and then panics in
			// the renderer; the parser half is what this checkpoint owns.
			name:  "empty environment is one empty row",
			input: `\begin{matrix}\end{matrix}`,
			check: func(t *testing.T, got ast.Expr) {
				m := asMatrix(t, got)
				if len(m.Rows) != 1 {
					t.Fatalf("len(rows) = %d, want 1", len(m.Rows))
				}
				if len(m.Rows[0]) != 0 {
					t.Errorf("len(rows[0]) = %d, want 0", len(m.Rows[0]))
				}
			},
		},
		{
			// `&` inside a nested environment must not split the outer row.
			name:  "nested environment inside a cell",
			input: `\begin{pmatrix}\begin{matrix}a\end{matrix}&b\end{pmatrix}`,
			check: func(t *testing.T, got ast.Expr) {
				m := asMatrix(t, got)
				if m.Name != "pmatrix" {
					t.Errorf("environment = %q, want pmatrix", m.Name)
				}
				if len(m.Rows) != 1 {
					t.Fatalf("len(rows) = %d, want 1", len(m.Rows))
				}
				if len(m.Rows[0]) != 2 {
					t.Fatalf("len(rows[0]) = %d, want 2", len(m.Rows[0]))
				}
				inner := asMatrix(t, m.Rows[0][0])
				if inner.Name != "matrix" {
					t.Errorf("nested environment = %q, want matrix", inner.Name)
				}
				if name := asIdent(t, m.Rows[0][1]); name != "b" {
					t.Errorf("rows[0][1] = %q, want b", name)
				}
			},
		},
		{
			// Whitespace between tokens is dropped by the lexer, so a space
			// before the brace is accepted; only a gap inside the environment
			// name is an error.
			name:  "whitespace before the environment brace",
			input: `\begin {matrix}a\end{matrix}`,
			check: func(t *testing.T, got ast.Expr) {
				m := asMatrix(t, got)
				if m.Name != "matrix" {
					t.Errorf("environment = %q, want matrix", m.Name)
				}
			},
		},
		{
			// A limits glyph is parsed bare, its scripts follow, and the
			// trailing brace group becomes its argument.
			name:  "limits glyph takes its body after the scripts",
			input: `\sum_{i}^{n}{x}`,
			check: func(t *testing.T, got ast.Expr) {
				both, ok := got.(*ast.BothScripts)
				if !ok {
					t.Fatalf("want *ast.BothScripts, got %T", got)
				}
				cmd := asCommand(t, both.Base)
				if cmd.Name != "sum" {
					t.Errorf("base command = %q, want sum", cmd.Name)
				}
				if len(cmd.Args) != 1 {
					t.Fatalf("len(args) = %d, want 1", len(cmd.Args))
				}
				if name := asIdent(t, cmd.Args[0]); name != "x" {
					t.Errorf("args[0] = %q, want x", name)
				}
				if name := asIdent(t, both.Sub); name != "i" {
					t.Errorf("subscript = %q, want i", name)
				}
				if name := asIdent(t, both.Sup); name != "n" {
					t.Errorf("superscript = %q, want n", name)
				}
			},
		},
		{
			name:  "alphabet command with a brace argument",
			input: `\mathbb{R}`,
			check: func(t *testing.T, got ast.Expr) {
				cmd := asCommand(t, got)
				if cmd.Name != "mathbb" {
					t.Errorf("command name = %q, want mathbb", cmd.Name)
				}
				if len(cmd.Args) != 1 {
					t.Fatalf("len(args) = %d, want 1", len(cmd.Args))
				}
				if name := asIdent(t, cmd.Args[0]); name != "R" {
					t.Errorf("args[0] = %q, want R", name)
				}
			},
		},
		{
			name:  "unary minus",
			input: `-x`,
			check: func(t *testing.T, got ast.Expr) {
				neg, ok := got.(*ast.Neg)
				if !ok {
					t.Fatalf("want *ast.Neg, got %T", got)
				}
				if name := asIdent(t, neg.Inner); name != "x" {
					t.Errorf("inner = %q, want x", name)
				}
			},
		},
		{
			name:  "pipe pair becomes the absolute-value command",
			input: `|x|`,
			check: func(t *testing.T, got ast.Expr) {
				cmd := asCommand(t, got)
				if cmd.Name != "|" {
					t.Fatalf("command name = %q, want |", cmd.Name)
				}
				if len(cmd.Args) != 1 {
					t.Fatalf("len(args) = %d, want 1", len(cmd.Args))
				}
				if name := asIdent(t, cmd.Args[0]); name != "x" {
					t.Errorf("args[0] = %q, want x", name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInput(t, tt.input)
			if err != nil {
				t.Fatalf("ParseExpr(%q) returned error %v", tt.input, err)
			}
			tt.check(t, got)
		})
	}
}

// TestParseErrors pins the failure modes against the messages the reference
// CLI prints for the same input, so the port cannot drift into a different
// error variant without a test failing.
func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		kind  errs.Kind
		msg   string
	}{
		{
			name:  "unclosed left",
			input: `\left(x`,
			kind:  errs.KindUnclosedLeftRight,
			msg:   `unclosed \left ... \right pair`,
		},
		{
			name:  "right without left",
			input: `\right)`,
			kind:  errs.KindUnexpectedRightWithoutLeft,
			msg:   `unexpected \right without matching \left`,
		},
		{
			name:  "unclosed optional argument",
			input: `\sqrt[3`,
			kind:  errs.KindUnclosedLeftRight,
			msg:   `unclosed \left ... \right pair`,
		},
		{
			name:  "left delimiter is not a delimiter",
			input: `\left<x\right>`,
			kind:  errs.KindExpectedDelimiter,
			msg:   `expected a delimiter after \left`,
		},
		{
			// The reference panics here via unreachable!(). Returning an error
			// is the deliberate divergence recorded in #1.
			name:  "right delimiter used as a left delimiter errors instead of panicking",
			input: `\left)x\right)`,
			kind:  errs.KindExpectedDelimiter,
			msg:   `expected a delimiter after \left`,
		},
		{
			name:  "mismatched delimiters",
			input: `\left[x\right)`,
			kind:  errs.KindMismatchedDelimiters,
			msg:   `mismatched delimiters: \left[ and \right)`,
		},
		{
			name:  "unclosed environment",
			input: `\begin{matrix}a&b`,
			kind:  errs.KindUnclosedEnvironment,
			msg:   `unclosed \begin{matrix}`,
		},
		{
			name:  "environment name mismatch",
			input: `\begin{matrix}a\end{pmatrix}`,
			kind:  errs.KindMismatchedEnvironment,
			msg:   `mismatched \begin{matrix} and \end{pmatrix}`,
		},
		{
			name:  "extension environment is still rejected",
			input: `\begin{Bmatrix}a\\b\end{Bmatrix}`,
			kind:  errs.KindUnknownEnvironment,
			msg:   `unknown matrix environment: Bmatrix`,
		},
		{
			name:  "vmatrix is still rejected",
			input: `\begin{vmatrix}a\end{vmatrix}`,
			kind:  errs.KindUnknownEnvironment,
			msg:   `unknown matrix environment: vmatrix`,
		},
		{
			name:  "Vmatrix is still rejected",
			input: `\begin{Vmatrix}a\end{Vmatrix}`,
			kind:  errs.KindUnknownEnvironment,
			msg:   `unknown matrix environment: Vmatrix`,
		},
		{
			name:  "whitespace inside the environment name",
			input: `\begin{mat rix}`,
			kind:  errs.KindLocated,
			msg:   `unexpected whitespace near 'r' at byte 11`,
		},
		{
			// expect() wraps with At, so the reported kind is the located
			// one even though the underlying variant is ExpectedTokenGot.
			name:  "missing environment brace",
			input: `\begin matrix`,
			kind:  errs.KindLocated,
			msg:   `expected LBrace, got Ident("m") near 'm' at byte 7`,
		},
		{
			// The left delimiter is part of the inner slice, so the leading
			// pipe opens an absolute-value group that never closes. The
			// reference refuses this input for the same reason.
			name:  "pipe delimiter quirk matches the reference",
			input: `\left|x\right|`,
			kind:  errs.KindLocatedEOF,
			msg:   `expected Pipe but reached end of input near 'ight|'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInput(t, tt.input)
			if err == nil {
				t.Fatalf("ParseExpr(%q) succeeded with %T, want error %q", tt.input, got, tt.msg)
			}

			parseErr, ok := err.(*errs.ParseError)
			if !ok {
				t.Fatalf("ParseExpr(%q) returned %T, want *errs.ParseError", tt.input, err)
			}
			if parseErr.Kind != tt.kind {
				t.Errorf("kind = %v, want %v", parseErr.Kind, tt.kind)
			}
			if parseErr.Msg != tt.msg {
				t.Errorf("message = %q, want %q", parseErr.Msg, tt.msg)
			}
		})
	}
}

// TestParseEmptyCell covers the empty-slice branch of parseTokens: a leading
// & produces an empty cell. The reference renders this input, so the shapes
// asserted here are its shapes.
//
// A trailing & does not add a cell, because parseMatrixBody only appends the
// tail when cell_start < len(tokens). That is why
// `\begin{matrix}&\\&\end{matrix}` yields rows of 2 and 1 and only fails
// later, with MismatchedMatrixRows.
func TestParseEmptyCell(t *testing.T) {
	got, err := parseInput(t, `\begin{matrix}&a\\&b\end{matrix}`)
	if err != nil {
		t.Fatalf("ParseExpr returned error %v", err)
	}

	m := asMatrix(t, got)
	if len(m.Rows) != 2 {
		t.Fatalf("len(rows) = %d, want 2", len(m.Rows))
	}
	for i, row := range m.Rows {
		if len(row) != 2 {
			t.Fatalf("len(rows[%d]) = %d, want 2", i, len(row))
		}
		if _, ok := row[0].(*ast.Empty); !ok {
			t.Errorf("rows[%d][0] = %T, want *ast.Empty", i, row[0])
		}
	}
	if name := asIdent(t, m.Rows[0][1]); name != "a" {
		t.Errorf("rows[0][1] = %q, want a", name)
	}
	if name := asIdent(t, m.Rows[1][1]); name != "b" {
		t.Errorf("rows[1][1] = %q, want b", name)
	}
}
