// Package parser is the recursive-descent parser. Macro argument arity is
// driven by the symbol registry, not hardcoded per macro.
//
// Ports reference/txm/src/parser.rs.
package parser

import (
	"fmt"

	"github.com/Anuj-72/formul8/ast"
	"github.com/Anuj-72/formul8/errs"
	"github.com/Anuj-72/formul8/glyph"
	"github.com/Anuj-72/formul8/token"
)

// Parser holds the token slice, the original input (for error snippets),
// the current position, and the registry. Ports parser.rs:9-28.

// Cursor helpers, parser.rs:30-60:
//   peek, currentSpan, advance, expect

// String-argument scanning, parser.rs:62-149:
//   parseContinuousString  -- rejects any gap between token spans
//   parseStringArg         -- inserts a space where a gap appears

// The grammar, parser.rs:151-568:
//   ParseExpr        -> parseBinop
//   parseBinop       :155  left-assoc over + - = *
//   canStartAtom     :185
//   parseJuxtapose   :220
//   parseScripted    :265  see the note below
//   parseAtom        :338
//   parseLeftDelimited :420
//   parseDelimitedArg  :487  scans forward for the matching closer, then
//                            recurses on the token sub-slice
//   isOpenerFor, tokenEq :513, :522
//   parseCommand     :531  arity comes from the registry
//   parseBegin       :570  matrix environments only
//   parseMatrixBody  :624  splits on & and \\ at depth 0
//   parseTokens      :664  builds a sub-parser over a token slice

// parseScripted (parser.rs:295-324) reopens an already-parsed Command to
// attach a trailing brace group when the glyph HasLimits. Rust had to
// destructure and rebuild; in Go just append to the *ast.Command in place.

// Rust let-chains have no Go equivalent. Extract one helper for the
// repeated "base is a Command whose glyph has limits" test and use it here
// and in render/render.go.

// parseLeftDelimited has an unreachable!() at parser.rs:472 that is in fact
// reachable from user input. Return an error there instead of panicking.

type Parser struct {
	tokens   []token.Token
	input    string
	pos      int
	registry *glyph.SymbolRegistry
}

func New(input string, tokens []token.Token, registry *glyph.SymbolRegistry) *Parser {
	return &Parser{
		tokens:   tokens,
		input:    input,
		pos:      0,
		registry: registry,
	}
}

// Cursor Primitives
func (p *Parser) peek() (token.Token, bool) {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos], true
	}
	return token.Token{}, false
}

func (p *Parser) currentSpan() (errs.Span, bool) {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos].Span, true
	}
	return errs.Span{}, false
}

func (p *Parser) advance() token.Token {
	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *Parser) expect(expected token.Kind) error {
	actual, ok := p.peek()
	if !ok {
		return errs.ExpectedToken(token.Token{Kind: expected}.Debug()).
			AtEOF(p.input)
	}

	if actual.Kind == expected {
		p.advance()
		return nil
	}

	return errs.ExpectedTokenGot(token.Token{Kind: expected}.Debug(), actual.Debug()).
		At(actual.Span, p.input)
}

func (p *Parser) parseContinuousString(delim token.Kind) (string, error) {
	var name string
	var lastEnd *int

	for {
		tok, ok := p.peek()
		if ok && tok.Kind == token.KindIdent {
			span, okay := p.currentSpan()
			if !okay {
				return "", errs.UnexpectedEOF()
			}
			if lastEnd != nil && *lastEnd != span.Start {
				return "", errs.UnexpectedWhitespace().At(span, p.input)
			}
			name += tok.Text
			lastEnd = &span.End
			p.advance()
		} else if ok && tok.Kind == delim {
			break
		} else {
			span, okay := p.currentSpan()
			if !okay {
				return "", errs.UnexpectedEOF()
			}
			return "", errs.ExpectedString().At(
				span, p.input,
			)
		}
	}
	return name, nil
}

func (p *Parser) parseStringArg() (string, error) {
	if err := p.expect(token.KindLBrace); err != nil {
		return "", err
	}
	var name string
	var lastEnd *int

	defaultError := func() error {
		span, ok := p.currentSpan()
		if !ok {
			return errs.UnexpectedEOF()
		}
		return errs.ExpectedString().At(span, p.input)
	}
Loop:
	for {
		tok, ok := p.peek()
		if !ok {
			return "", defaultError()
		}

		switch {
		case tok.Kind == token.KindIdent || tok.Kind == token.KindNumber:
			span, err := p.currentSpan()
			if !err {
				return "", errs.UnexpectedEOF()
			}
			if lastEnd != nil && *lastEnd != span.Start {
				name += " "
			}
			name += tok.Text
			lastEnd = &span.End
			p.advance()
		case tok.Kind == token.KindRBrace:
			break Loop
		default:
			if ch, ok := tok.TextChar(); ok {
				span, err := p.currentSpan()
				if !err {
					return "", errs.UnexpectedEOF()
				}
				if lastEnd != nil && *lastEnd != span.Start {
					name += " "
				}
				name += string(ch)
				lastEnd = &span.End
				p.advance()
			} else {
				return "", defaultError()
			}
		}
	}

	p.advance()
	return name, nil
}

func (p *Parser) canStartAtom() bool {
	tok, ok := p.peek()
	if !ok {
		return false
	}

	switch tok.Kind {
	case token.KindLBrace, token.KindLParen, token.KindRParen, token.KindLBracket, token.KindRBracket,
		token.KindNumber, token.KindIdent, token.KindCommand, token.KindEscape, token.KindBang,
		token.KindPipe, token.KindMinus, token.KindPlus, token.KindAmpersand, token.KindSlash,
		token.KindComma, token.KindDot, token.KindColon, token.KindSemicolon, token.KindLess,
		token.KindGreater, token.KindQuestion, token.KindAt, token.KindHash, token.KindDollar,
		token.KindPercent, token.KindTilde, token.KindQuote, token.KindBacktick:
		return true
	default:
		return false
	}
}

func (p *Parser) parseAtom() (ast.Expr, error) {
	next := p.advance()
	switch next.Kind {
	case token.KindNumber:
		return &ast.Number{Value: next.Text}, nil
	case token.KindIdent:
		return &ast.Ident{Name: next.Text}, nil
	case token.KindLBrace:
		inner, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		er := p.expect(token.KindRBrace)
		if er != nil {
			return nil, er
		}
		return inner, nil
	case token.KindLParen:
		return &ast.Ident{Name: "("}, nil
	case token.KindRParen:
		return &ast.Ident{Name: ")"}, nil
	case token.KindLBracket:
		return &ast.Ident{Name: "["}, nil
	case token.KindRBracket:
		return &ast.Ident{Name: "]"}, nil
	case token.KindCommand:
		name := next.Text
		if name == "begin" {
			return p.parseBegin()
		} else if name == "left" {
			return p.parseLeftDelimited()
		} else if name == "right" {
			return nil, errs.UnexpectedRightWithoutLeft()
		}
		return p.parseCommand(name)
	case token.KindEscape:
		s := next.Text
		return &ast.Escape{Ch: s}, nil
	case token.KindPipe:
		inner, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		er := p.expect(token.KindPipe)
		if er != nil {
			return nil, er
		}
		return &ast.Command{
			Name: "|",
			Opts: []ast.Expr{},
			Args: []ast.Expr{inner},
		}, nil
	case token.KindMinus:
		if p.canStartAtom() {
			expr, err := p.parseAtom()
			if err != nil {
				return nil, err
			}
			return &ast.Neg{Inner: expr}, nil
		}
		return &ast.Ident{Name: "-"}, nil
	case token.KindBang:
		return &ast.Ident{Name: "!"}, nil
	case token.KindAmpersand:
		return &ast.Ident{Name: "&"}, nil
	case token.KindPlus:
		return &ast.Ident{Name: "+"}, nil
		// Punctuation and relations that carry no special layout: render the
		// literal symbol. Without these the tokens are dropped or collide
		// with a closing-delimiter expectation (e.g. `(3,0)`, `x^{a/b}`).
	case token.KindSlash:
		return &ast.Ident{Name: "/"}, nil
	case token.KindComma:
		return &ast.Ident{Name: ","}, nil
	case token.KindDot:
		return &ast.Ident{Name: "."}, nil
	case token.KindColon:
		return &ast.Ident{Name: ":"}, nil
	case token.KindSemicolon:
		return &ast.Ident{Name: ";"}, nil
	case token.KindLess:
		return &ast.Ident{Name: "<"}, nil
	case token.KindGreater:
		return &ast.Ident{Name: ">"}, nil
	case token.KindQuestion:
		return &ast.Ident{Name: "?"}, nil
	case token.KindAt:
		return &ast.Ident{Name: "@"}, nil
	case token.KindHash:
		return &ast.Ident{Name: "#"}, nil
	case token.KindDollar:
		return &ast.Ident{Name: "$"}, nil
	case token.KindPercent:
		return &ast.Ident{Name: "%"}, nil
	case token.KindTilde:
		return &ast.Ident{Name: "~"}, nil
	case token.KindQuote:
		return &ast.Ident{Name: "\""}, nil
	case token.KindBacktick:
		return &ast.Ident{Name: "`"}, nil
	default:
		return nil, errs.UnexpectedToken(
			p.pos-1,
			fmt.Sprintf("%+v", next.Debug()),
		)
	}
}

func (p *Parser) ParseExpr() (ast.Expr, error) {
	return p.parseBinop()
}

func (p *Parser) parseBegin() (ast.Expr, error) {
	if err := p.expect(token.KindLBrace); err != nil {
		return nil, err
	}

	envName, err := p.parseContinuousString(token.KindRBrace)
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.KindRBrace); err != nil {
		return nil, err
	}

	// Only the three reference environments are accepted here. Bmatrix,
	// vmatrix and Vmatrix belong to the extension in #10 and stay rejected
	// until that checkpoint lands its parser and backend work together.
	if envName != "matrix" && envName != "bmatrix" && envName != "pmatrix" {
		return nil, errs.UnknownEnvironment(envName)
	}

	// Scan forward for the \end that closes this environment. Nested
	// \begin ... \end pairs bump and drop the depth counter, so an inner
	// matrix does not end the outer one.
	bodyStart := p.pos
	depth := 0
	endPos := -1
	for endPos < 0 {
		if p.pos >= len(p.tokens) {
			return nil, errs.UnclosedEnvironment(envName)
		}

		tok := p.tokens[p.pos]
		switch {
		case tok.Kind == token.KindCommand && tok.Text == "begin":
			depth++
			p.pos++
		case tok.Kind == token.KindCommand && tok.Text == "end":
			if depth == 0 {
				endPos = p.pos
			} else {
				depth--
				p.pos++
			}
		default:
			p.pos++
		}
	}

	rows, err := p.parseMatrixBody(p.tokens[bodyStart:endPos])
	if err != nil {
		return nil, err
	}

	p.advance() // eat \end
	if err := p.expect(token.KindLBrace); err != nil {
		return nil, err
	}

	endName, err := p.parseContinuousString(token.KindRBrace)
	if err != nil {
		return nil, err
	}
	if endName != envName {
		return nil, errs.MismatchedEnvironment(envName, endName)
	}

	if err := p.expect(token.KindRBrace); err != nil {
		return nil, err
	}

	return &ast.Matrix{Name: envName, Rows: rows}, nil
}

// parseLeftDelimited ports parser.rs:420-485. The \left token itself is the
// first token of the enclosed slice, and every token up to the matching
// \right is re-parsed as a sub-expression.
//
// The reference ends with unreachable!() for a left delimiter outside
// ( [ { |, but as_char also admits ) ] }, so `\left)x\right)` reaches that arm
// from user input. Here it returns ExpectedDelimiter("left") instead of
// panicking.
func (p *Parser) parseLeftDelimited() (ast.Expr, error) {
	tok, ok := p.peek()
	if !ok {
		return nil, errs.ExpectedDelimiter("left")
	}
	left, ok := tok.AsChar()
	if !ok {
		return nil, errs.ExpectedDelimiter("left")
	}

	// Depth scan for the matching \right, so nested pairs do not close each
	// other early.
	innerStart := p.pos
	depth := 0
	matchIdx := -1
	for i := innerStart; i < len(p.tokens) && matchIdx < 0; i++ {
		t := p.tokens[i]
		if t.Kind != token.KindCommand {
			continue
		}

		switch t.Text {
		case "left":
			depth++
		case "right":
			if depth == 0 {
				matchIdx = i
			} else {
				depth--
			}
		}
	}
	if matchIdx < 0 {
		return nil, errs.UnclosedLeftRight()
	}
	if p.tokens[matchIdx].Text != "right" {
		return nil, errs.Internal("mismatched delimiter scan")
	}

	if matchIdx+1 >= len(p.tokens) {
		return nil, errs.ExpectedDelimiter("right")
	}
	right, ok := p.tokens[matchIdx+1].AsChar()
	if !ok {
		return nil, errs.ExpectedDelimiter("right")
	}

	var expectedRight rune
	switch left {
	case '(':
		expectedRight = ')'
	case '[':
		expectedRight = ']'
	case '{':
		expectedRight = '}'
	case '|':
		expectedRight = '|'
	default:
		return nil, errs.ExpectedDelimiter("left")
	}
	if right != expectedRight {
		return nil, errs.MismatchedDelimiters(left, right)
	}

	inner, err := p.parseTokens(p.tokens[innerStart:matchIdx])
	if err != nil {
		return nil, err
	}
	p.pos = matchIdx + 2
	return &ast.Delimiter{Left: left, Right: right, Inner: inner}, nil
}

// parseDelimitedArg ports parser.rs:487-511. The opener was consumed by the
// caller; this scans forward for its closer at the right nesting depth and
// re-parses the enclosed slice.
func (p *Parser) parseDelimitedArg(closeKind token.Kind) (ast.Expr, error) {
	p.advance()
	innerStart := p.pos
	depth := 1
	matchIdx := -1
	for i := innerStart; i < len(p.tokens) && matchIdx < 0; i++ {
		t := p.tokens[i]
		if isOpenerFor(t.Kind, closeKind) {
			depth++
		} else if tokenEq(t.Kind, closeKind) {
			if depth == 1 {
				matchIdx = i
			} else {
				depth--
			}
		}
	}
	if matchIdx < 0 {
		return nil, errs.UnclosedLeftRight()
	}

	argTokens := p.tokens[innerStart:matchIdx]
	p.pos = matchIdx + 1
	return p.parseTokens(argTokens)
}

// isOpenerFor ports is_opener_for (parser.rs:513) and tokenEq ports token_eq
// (parser.rs:522). The reference takes &Token and inspects only the variant,
// so the port takes the kinds directly.
func isOpenerFor(open, closeKind token.Kind) bool {
	return (open == token.KindLBrace && closeKind == token.KindRBrace) ||
		(open == token.KindLParen && closeKind == token.KindRParen) ||
		(open == token.KindLBracket && closeKind == token.KindRBracket)
}

func tokenEq(kind, closeKind token.Kind) bool {
	return (kind == token.KindRBrace && closeKind == token.KindRBrace) ||
		(kind == token.KindRParen && closeKind == token.KindRParen) ||
		(kind == token.KindRBracket && closeKind == token.KindRBracket)
}

// parseCommand ports parser.rs:531-568. Arity, optionality, limits and the
// string argument all come from the registry entry, so a name with no glyph
// parses as a bare command with no arguments.
func (p *Parser) parseCommand(name string) (ast.Expr, error) {
	g, found := p.registry.Get(name)

	hasOptional := found && g.HasOptional()
	hasLimits := found && g.HasLimits()
	takesString := found && g.TakesStringArg()
	required := 0
	if found {
		required = g.RequiredArgs()
	}

	opts := []ast.Expr{}
	args := []ast.Expr{}

	if hasOptional {
		if tok, ok := p.peek(); ok && tok.Kind == token.KindLBracket {
			opt, err := p.parseDelimitedArg(token.KindRBracket)
			if err != nil {
				return nil, err
			}
			opts = append(opts, opt)
		}
	}

	// A glyph with limits takes its body after the scripts (\int_a^b{f}), so
	// parseScripted attaches that argument and this must not consume it.
	if !hasLimits {
		if takesString {
			text, err := p.parseStringArg()
			if err != nil {
				return nil, err
			}
			args = append(args, &ast.Ident{Name: text})
			required--
		}

		for i := 0; i < required; i++ {
			tok, ok := p.peek()

			var arg ast.Expr
			var err error
			switch {
			case ok && tok.Kind == token.KindLBrace:
				arg, err = p.parseDelimitedArg(token.KindRBrace)
			case ok && tok.Kind == token.KindLParen:
				arg, err = p.parseDelimitedArg(token.KindRParen)
			case ok && tok.Kind == token.KindLBracket:
				arg, err = p.parseDelimitedArg(token.KindRBracket)
			default:
				arg, err = p.parseAtom()
			}
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
	}

	return &ast.Command{Name: name, Opts: opts, Args: args}, nil
}

// parseMatrixBody ports parser.rs:624-662. Cells split on & and rows on \\,
// but only at brace depth zero and nested-environment depth zero, so a nested
// matrix inside a cell survives intact. Ragged rows are returned as they are:
// the reference reports MismatchedMatrixRows at render time, not here.
func (p *Parser) parseMatrixBody(tokens []token.Token) ([][]ast.Expr, error) {
	rows := [][]ast.Expr{}
	currentRow := []ast.Expr{}
	cellStart := 0
	depth := 0
	envDepth := 0

	for i, t := range tokens {
		switch {
		case t.Kind == token.KindLBrace || t.Kind == token.KindLBracket || t.Kind == token.KindLParen:
			depth++
		case t.Kind == token.KindRBrace || t.Kind == token.KindRBracket || t.Kind == token.KindRParen:
			if depth > 0 {
				depth--
			}
		case t.Kind == token.KindCommand && t.Text == "begin":
			envDepth++
		case t.Kind == token.KindCommand && t.Text == "end":
			if envDepth > 0 {
				envDepth--
			}
		case t.Kind == token.KindAmpersand && depth == 0 && envDepth == 0:
			cell, err := p.parseTokens(tokens[cellStart:i])
			if err != nil {
				return nil, err
			}
			currentRow = append(currentRow, cell)
			cellStart = i + 1
		case t.Kind == token.KindEscape && t.Text == "\\" && depth == 0 && envDepth == 0:
			cell, err := p.parseTokens(tokens[cellStart:i])
			if err != nil {
				return nil, err
			}
			currentRow = append(currentRow, cell)
			rows = append(rows, currentRow)
			currentRow = []ast.Expr{}
			cellStart = i + 1
		}
	}

	if cellStart < len(tokens) {
		cell, err := p.parseTokens(tokens[cellStart:])
		if err != nil {
			return nil, err
		}
		currentRow = append(currentRow, cell)
	}
	if len(currentRow) > 0 || len(rows) == 0 {
		rows = append(rows, currentRow)
	}

	return rows, nil
}

// parseTokens ports parser.rs:664-678: a sub-parser over a token slice, so a
// nested parse keeps its own cursor without disturbing the caller's. An empty
// slice is Expr::Empty, which is how an empty matrix cell is represented.
func (p *Parser) parseTokens(tokens []token.Token) (ast.Expr, error) {
	if len(tokens) == 0 {
		return &ast.Empty{}, nil
	}

	sub := New(p.input, tokens, p.registry)
	return sub.ParseExpr()
}

func (p *Parser) parseScripted() (ast.Expr, error) {
	base, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	var sub, sup ast.Expr
	primes := 0

	for {
		tok, ok := p.peek()
		if !ok {
			break
		}

		if tok.Kind == token.KindUnderscore && sub == nil {
			p.advance()
			newSub, err := p.parseAtom()
			if err != nil {
				return nil, err
			}
			sub = newSub
		} else if tok.Kind == token.KindCaret && sup == nil {
			p.advance()
			newSup, err := p.parseAtom()
			if err != nil {
				return nil, err
			}
			sup = newSup
		} else if tok.Kind == token.KindPrime {
			p.advance()
			primes++
		} else {
			break
		}
	}

	result := base
	if cmd, ok := base.(*ast.Command); ok {
		glyph, found := p.registry.Get(cmd.Name)
		if found && glyph.HasLimits() {
			if next, ok := p.peek(); ok && next.Kind == token.KindLBrace {
				if len(cmd.Args) < glyph.RequiredArgs() {
					p.advance() // eat {
					body, err := p.ParseExpr()
					if err != nil {
						return nil, err
					}
					if err := p.expect(token.KindRBrace); err != nil {
						return nil, err
					}
					cmd.Args = append(cmd.Args, body)
					cmd.Opts = nil
					result = cmd
				}
			}
		}
	}

	if primes > 0 {
		result = &ast.Prime{Base: result, Count: primes}
	}

	switch {
	case sub == nil && sup == nil:
		return result, nil
	case sub != nil && sup == nil:
		return &ast.Subscript{Base: result, Sub: sub}, nil
	case sub == nil && sup != nil:
		return &ast.Superscript{Base: result, Exp: sup}, nil
	default:
		return &ast.BothScripts{Base: result, Sub: sub, Sup: sup}, nil
	}
}

func (p *Parser) parseJuxtapose() (ast.Expr, error) {
	if !p.canStartAtom() {
		return &ast.Empty{}, nil
	}

	first, err := p.parseScripted()
	if err != nil {
		return nil, err
	}

	exprs := []ast.Expr{first}

	// Each iteration parses a fresh parseScripted. The reference appended the
	// node it had already parsed without consuming a token, which looped
	// forever on input as ordinary as "x + y".
	for p.canContinueJuxtapose() {
		next, err := p.parseScripted()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, next)
	}

	if len(exprs) == 1 {
		return exprs[0], nil
	}
	return &ast.Juxtapose{Items: exprs}, nil
}

// canContinueJuxtapose ports the token set of the while-let loop in
// parse_juxtapose (parser.rs:224-250). It is canStartAtom minus Pipe, Minus
// and Plus: those three can open an atom but never continue one, so parseBinop
// and the unary-minus arm of parseAtom get to claim them.
func (p *Parser) canContinueJuxtapose() bool {
	tok, ok := p.peek()
	if !ok {
		return false
	}

	switch tok.Kind {
	case token.KindLBrace, token.KindLParen, token.KindRParen, token.KindLBracket, token.KindRBracket,
		token.KindNumber, token.KindIdent, token.KindCommand, token.KindEscape, token.KindBang,
		token.KindAmpersand, token.KindSlash,
		token.KindComma, token.KindDot, token.KindColon, token.KindSemicolon, token.KindLess,
		token.KindGreater, token.KindQuestion, token.KindAt, token.KindHash, token.KindDollar,
		token.KindPercent, token.KindTilde, token.KindQuote, token.KindBacktick:
		return true
	default:
		return false
	}
}

func (p *Parser) parseBinop() (ast.Expr, error) {
	lhs, err := p.parseJuxtapose()
	if err != nil {
		return nil, err
	}

Loop:
	for {
		tok, ok := p.peek()
		if !ok {
			break
		}

		switch tok.Kind {
		case token.KindPlus:
			p.advance()
			rhs, err := p.parseJuxtapose()
			if err != nil {
				return nil, err
			}
			lhs = &ast.BinOpExpr{Left: lhs, Op: ast.Add, Right: rhs}
		case token.KindMinus:
			p.advance()
			rhs, err := p.parseJuxtapose()
			if err != nil {
				return nil, err
			}
			lhs = &ast.BinOpExpr{Left: lhs, Op: ast.Sub, Right: rhs}
		case token.KindEquals:
			p.advance()
			rhs, err := p.parseJuxtapose()
			if err != nil {
				return nil, err
			}
			lhs = &ast.BinOpExpr{Left: lhs, Op: ast.Eq, Right: rhs}
		case token.KindStar:
			p.advance()
			rhs, err := p.parseJuxtapose()
			if err != nil {
				return nil, err
			}
			lhs = &ast.BinOpExpr{Left: lhs, Op: ast.Mul, Right: rhs}
		default:
			break Loop
		}
	}

	return lhs, nil
}
