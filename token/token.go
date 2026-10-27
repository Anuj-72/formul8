// Package token defines the lexical tokens and the hand-rolled scanner that
// replaces the Rust `logos` derive macro.
//
// Ports reference/txm/src/token.rs.
package token

import (
	"fmt"
	"unicode/utf8"

	"github.com/Anuj-72/formul8/errs"
)

// Kind identifies the token type. Ports the variants of logos Token enum in token.rs.
type Kind int

const (
	KindLBrace Kind = iota
	KindRBrace
	KindLBracket
	KindRBracket
	KindCaret
	KindUnderscore
	KindPrime
	KindPipe
	KindBang
	KindPlus
	KindMinus
	KindEquals
	KindStar
	KindComma
	KindLParen
	KindRParen
	KindDot
	KindColon
	KindSemicolon
	KindLess
	KindGreater
	KindSlash
	KindAmpersand
	KindQuestion
	KindAt
	KindHash
	KindDollar
	KindPercent
	KindTilde
	KindQuote
	KindBacktick

	KindCommand
	KindEscape
	KindNumber
	KindIdent
)

// Token represents a single lexed token along with its string payload and source span.
// Ports SpannedToken = (Token, Range<usize>) by folding Span into the struct.
type Token struct {
	Kind Kind
	Text string    // Payload for Command, Escape, Number, Ident (without leading '\' for Command/Escape)
	Span errs.Span // Byte range in the input string
}

// Debug ports Rust's {:?} Debug formatting for Token.
// The parser embeds this string into error messages (ExpectedTokenGot, UnexpectedToken).
func (t Token) Debug() string {
	switch t.Kind {
	case KindLBrace:
		return "LBrace"
	case KindRBrace:
		return "RBrace"
	case KindLBracket:
		return "LBracket"
	case KindRBracket:
		return "RBracket"
	case KindCaret:
		return "Caret"
	case KindUnderscore:
		return "Underscore"
	case KindPrime:
		return "Prime"
	case KindPipe:
		return "Pipe"
	case KindBang:
		return "Bang"
	case KindPlus:
		return "Plus"
	case KindMinus:
		return "Minus"
	case KindEquals:
		return "Equals"
	case KindStar:
		return "Star"
	case KindComma:
		return "Comma"
	case KindLParen:
		return "LParen"
	case KindRParen:
		return "RParen"
	case KindDot:
		return "Dot"
	case KindColon:
		return "Colon"
	case KindSemicolon:
		return "Semicolon"
	case KindLess:
		return "Less"
	case KindGreater:
		return "Greater"
	case KindSlash:
		return "Slash"
	case KindAmpersand:
		return "Ampersand"
	case KindQuestion:
		return "Question"
	case KindAt:
		return "At"
	case KindHash:
		return "Hash"
	case KindDollar:
		return "Dollar"
	case KindPercent:
		return "Percent"
	case KindTilde:
		return "Tilde"
	case KindQuote:
		return "Quote"
	case KindBacktick:
		return "Backtick"
	case KindCommand:
		return fmt.Sprintf("Command(%q)", t.Text)
	case KindEscape:
		return fmt.Sprintf("Escape(%q)", t.Text)
	case KindNumber:
		return fmt.Sprintf("Number(%q)", t.Text)
	case KindIdent:
		return fmt.Sprintf("Ident(%q)", t.Text)
	}
	return "Unknown"
}

// TextChar ports Token::text_char from token.rs:94-132.
// Returns the single character representation if this token acts as literal text.
func (t Token) TextChar() (rune, bool) {
	switch t.Kind {
	case KindBang:
		return '!', true
	case KindQuestion:
		return '?', true
	case KindComma:
		return ',', true
	case KindDot:
		return '.', true
	case KindColon:
		return ':', true
	case KindSemicolon:
		return ';', true
	case KindSlash:
		return '/', true
	case KindLess:
		return '<', true
	case KindGreater:
		return '>', true
	case KindPlus:
		return '+', true
	case KindMinus:
		return '-', true
	case KindEquals:
		return '=', true
	case KindStar:
		return '*', true
	case KindPipe:
		return '|', true
	case KindAmpersand:
		return '&', true
	case KindLParen:
		return '(', true
	case KindRParen:
		return ')', true
	case KindLBracket:
		return '[', true
	case KindRBracket:
		return ']', true
	case KindLBrace:
		return '{', true
	case KindRBrace:
		return '}', true
	case KindPrime:
		return '\'', true
	case KindCaret:
		return '^', true
	case KindUnderscore:
		return '_', true
	case KindAt:
		return '@', true
	case KindHash:
		return '#', true
	case KindDollar:
		return '$', true
	case KindPercent:
		return '%', true
	case KindTilde:
		return '~', true
	case KindQuote:
		return '"', true
	case KindBacktick:
		return '`', true
	case KindIdent, KindEscape:
		if len(t.Text) > 0 {
			r, _ := utf8.DecodeRuneInString(t.Text)
			return r, true
		}
	case KindNumber:
		if len(t.Text) == 1 {
			return rune(t.Text[0]), true
		}
	}
	return 0, false
}

// AsChar ports Token::as_char from token.rs:134-145.
// Checks if the token represents a delimiter character (either literal or escaped).
func (t Token) AsChar() (rune, bool) {
	switch t.Kind {
	case KindLParen:
		return '(', true
	case KindRParen:
		return ')', true
	case KindLBracket:
		return '[', true
	case KindRBracket:
		return ']', true
	case KindLBrace:
		return '{', true
	case KindRBrace:
		return '}', true
	case KindPipe:
		return '|', true
	case KindEscape:
		switch t.Text {
		case "(":
			return '(', true
		case ")":
			return ')', true
		case "[":
			return '[', true
		case "]":
			return ']', true
		case "{":
			return '{', true
		case "}":
			return '}', true
		case "|":
			return '|', true
		}
	}
	return 0, false
}
