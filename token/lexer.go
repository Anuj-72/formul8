package token

// Tokenize scans the input into tokens with byte spans, or returns
// errs.InvalidToken. Ports token.rs:148-156 plus the logos derive
// attributes at token.rs:11-90.
//
// logos picks the LONGEST match, so the scanner must try rules in this
// order at each position:
//
//   1. '\' followed by [a-zA-Z]+   -> Command, Text = name without backslash
//   2. '\' followed by one non-alpha rune -> Escape, Text = that rune
//      (a lone trailing '\' is an error -- tests/smoke.rs:86 asserts it)
//   3. [0-9]+                      -> Number, greedy: "123" is ONE token
//   4. one [a-zA-Z]                -> Ident
//   6. whitespace run              -> skipped, produces no token
//   5. one rune outside printable ASCII 0x20..0x7E
//                                  -> Ident   (this is what makes alpha and
//                                     CJK characters idents)
//   7. one of the 30 literal punctuation characters -> that Kind
//   8. anything else               -> errs.InvalidToken at that byte offset
//
// Rules 5 and 6 run in the opposite order to the numbering above: logos' \s is
// the Unicode whitespace class, so a rune can satisfy both, and whitespace has
// to win. Numbering follows token.rs so the two files stay comparable.
//
// pos is a byte offset throughout, and every rune is decoded with
// utf8.DecodeRuneInString rather than converted with byte(). Spans must stay
// byte offsets end to end -- parser.rs compares prevEnd != span.Start.

import (
	"unicode"
	"unicode/utf8"

	"github.com/Anuj-72/formul8/errs"
)

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// Tokenize scans the input into tokens with byte spans, or returns
// errs.InvalidToken. Ports token.rs:148-156 plus the logos derive
// attributes at token.rs:11-90.
func Tokenize(input string) ([]Token, error) {
	var tokens []Token
	pos := 0
	n := len(input)

	for pos < n {
		start := pos

		// -----------------------------------------------------------------
		// 1. '\' followed by [a-zA-Z]+ -> Command (Text = name without '\')
		// 2. '\' followed by one non-alpha rune -> Escape (Text = that rune)
		//    (a lone trailing '\' is an error)
		// -----------------------------------------------------------------
		if input[pos] == '\\' {
			pos++ // consume '\'
			if pos >= n {
				// Trailing backslash with nothing after it -> error at start byte
				return nil, errs.InvalidToken(start)
			}

			if isAlpha(input[pos]) {
				// Rule 1: Command
				for pos < n && isAlpha(input[pos]) {
					pos++
				}
				tokens = append(tokens, Token{
					Kind: KindCommand,
					Text: input[start+1 : pos],
					Span: errs.Span{Start: start, End: pos},
				})
			} else {
				// Rule 2: Escape (takes exactly one non-alpha rune)
				r, w := utf8.DecodeRuneInString(input[pos:])
				pos += w
				tokens = append(tokens, Token{
					Kind: KindEscape,
					Text: string(r),
					Span: errs.Span{Start: start, End: pos},
				})
			}
			continue
		}

		// -----------------------------------------------------------------
		// 3. [0-9]+ -> Number, greedy: "123" is ONE token
		// -----------------------------------------------------------------
		if isDigit(input[pos]) {
			for pos < n && isDigit(input[pos]) {
				pos++
			}
			tokens = append(tokens, Token{
				Kind: KindNumber,
				Text: input[start:pos],
				Span: errs.Span{Start: start, End: pos},
			})
			continue
		}

		// -----------------------------------------------------------------
		// 4. one [a-zA-Z] -> Ident
		// -----------------------------------------------------------------
		if isAlpha(input[pos]) {
			pos++
			tokens = append(tokens, Token{
				Kind: KindIdent,
				Text: input[start:pos],
				Span: errs.Span{Start: start, End: pos},
			})
			continue
		}

		// Decode the next rune. Every remaining rule classifies a whole rune,
		// so all of them test r -- never byte(r), which would truncate a
		// multi-byte rune to its low byte and mistake U+0420 for a space.
		r, w := utf8.DecodeRuneInString(input[pos:])

		// -----------------------------------------------------------------
		// 6. whitespace run -> skipped, produces no token
		//
		// Checked before rule 5 because logos' \s is the Unicode whitespace
		// class, so runes above ASCII (U+0085, U+00A0) are whitespace and
		// must be skipped rather than becoming Idents.
		// -----------------------------------------------------------------
		if unicode.IsSpace(r) {
			for pos < n {
				r2, w2 := utf8.DecodeRuneInString(input[pos:])
				if !unicode.IsSpace(r2) {
					break
				}
				pos += w2
			}
			continue
		}

		// -----------------------------------------------------------------
		// 5. one rune outside printable ASCII 0x20..0x7E -> Ident
		//    (this is what makes α and CJK characters Idents)
		//
		// w > 1 admits a literal U+FFFD, which decodes to RuneError but is a
		// real rune; only w == 1 with RuneError means invalid UTF-8, which
		// falls through to rule 8.
		// -----------------------------------------------------------------
		if (r < 0x20 || r > 0x7E) && (w > 1 || r != utf8.RuneError) {
			pos += w
			tokens = append(tokens, Token{
				Kind: KindIdent,
				Text: input[start:pos],
				Span: errs.Span{Start: start, End: pos},
			})
			continue
		}

		// -----------------------------------------------------------------
		// 7. one of the 30 literal punctuation characters -> that Kind
		// -----------------------------------------------------------------
		var kind Kind
		matched := true

		switch input[pos] {
		case '{':
			kind = KindLBrace
		case '}':
			kind = KindRBrace
		case '[':
			kind = KindLBracket
		case ']':
			kind = KindRBracket
		case '^':
			kind = KindCaret
		case '_':
			kind = KindUnderscore
		case '\'':
			kind = KindPrime
		case '|':
			kind = KindPipe
		case '!':
			kind = KindBang
		case '+':
			kind = KindPlus
		case '-':
			kind = KindMinus
		case '=':
			kind = KindEquals
		case '*':
			kind = KindStar
		case ',':
			kind = KindComma
		case '(':
			kind = KindLParen
		case ')':
			kind = KindRParen
		case '.':
			kind = KindDot
		case ':':
			kind = KindColon
		case ';':
			kind = KindSemicolon
		case '<':
			kind = KindLess
		case '>':
			kind = KindGreater
		case '/':
			kind = KindSlash
		case '&':
			kind = KindAmpersand
		case '?':
			kind = KindQuestion
		case '@':
			kind = KindAt
		case '#':
			kind = KindHash
		case '$':
			kind = KindDollar
		case '%':
			kind = KindPercent
		case '~':
			kind = KindTilde
		case '"':
			kind = KindQuote
		case '`':
			kind = KindBacktick
		default:
			matched = false
		}

		if matched {
			pos++
			tokens = append(tokens, Token{
				Kind: kind,
				Text: input[start:pos],
				Span: errs.Span{Start: start, End: pos},
			})
			continue
		}

		// -----------------------------------------------------------------
		// 8. anything else -> errs.InvalidToken at that byte offset
		// -----------------------------------------------------------------
		return nil, errs.InvalidToken(start)
	}

	return tokens, nil
}
