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

	if envName == "matrix" || envName == "bmatrix" || envName == "pmatrix" ||
		// including our newly created ones
		envName == "Bmatrix" || envName == "vmatrix" || envName == "Vmatrix" {
		return nil, errs.UnknownEnvironment(envName)
	}

	bodyStart, depth := p.pos, int(0)
	switch p.tokens
}
func (p *Parser) parseLeftDelimited() (ast.Expr, error) {
	return nil, fmt.Errorf("parseLeftDelimited: not implemented")
}
func (p *Parser) parseCommand(name string) (ast.Expr, error) {
	return nil, fmt.Errorf("parseCommand: not implemented")
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

	exp, err := p.parseScripted()
	if err != nil {
		return nil, err
	}

	var exprs []ast.Expr
	exprs = append(exprs, exp)

	for {
		tok, ok := p.peek()
		if ok &&
			token.KindLBrace == tok.Kind ||
			token.KindLParen == tok.Kind ||
			token.KindRParen == tok.Kind ||
			token.KindLBracket == tok.Kind ||
			token.KindRBracket == tok.Kind ||
			token.KindNumber == tok.Kind ||
			token.KindIdent == tok.Kind ||
			token.KindCommand == tok.Kind ||
			token.KindEscape == tok.Kind ||
			token.KindBang == tok.Kind ||
			token.KindAmpersand == tok.Kind ||
			token.KindSlash == tok.Kind ||
			token.KindComma == tok.Kind ||
			token.KindDot == tok.Kind ||
			token.KindColon == tok.Kind ||
			token.KindSemicolon == tok.Kind ||
			token.KindLess == tok.Kind ||
			token.KindGreater == tok.Kind ||
			token.KindQuestion == tok.Kind ||
			token.KindAt == tok.Kind ||
			token.KindHash == tok.Kind ||
			token.KindDollar == tok.Kind ||
			token.KindPercent == tok.Kind ||
			token.KindTilde == tok.Kind ||
			token.KindQuote == tok.Kind ||
			token.KindBacktick == tok.Kind {
			exprs = append(exprs, exp)
			continue
		}
		break
	}

	if len(exprs) == 1 {
		return exprs[0], nil
	} else {
		return &ast.Juxtapose{Items: exprs}, nil
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
