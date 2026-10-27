package glyph

import (
	"unicode/utf8"
)

// Unicode math-alphabet mappings. Pure rune arithmetic over the
// Mathematical Alphanumeric Symbols block. Ports glyph.rs:276-357.
//
//   shift(r rune, base, off rune) rune   -- :276, falls back to r if the
//                                           resulting code point is invalid
//   ToBold      -- :280   also maps italic -> bold italic
//   ToBB        -- :293   blackboard bold, with the letterlike-specials
//                         exceptions (C H N P Q R Z) taking priority over
//                         the plain 0x1D538 range
//   ToUpright   -- :320   identity
//   ToItalic    -- :324   with the 'h' -> U+210E special case
//   ToBoldItalic-- :334
//   ToSans      -- :346
//
// Each is a func(rune) rune, stored as a field on AlphabetGlyph and
// MappedStyleGlyph. Go has first-class functions, so the Rust
// fn(char) -> char pointers port directly.

func shift(c rune, base, off rune) rune {
	r := rune(base + off)
	if utf8.ValidRune(r) {
		return r
	}
	return c
}

func ToBold(c rune) rune {
	switch {
	case 'A' <= c && c <= 'Z':
		return shift(c, 0x1D400, (c - 'A'))
	case 'a' <= c && c <= 'z':
		return shift(c, 0x1D41A, (c - 'a'))
	case '0' <= c && c <= '9':
		return shift(c, 0x1D7CE, (c - '0'))
	// Italic uppercase → bold italic uppercase
	case 0x1D434 <= c && c <= 0x1D44D:
		return shift(c, 0x1D468, (c - 0x1D434))
	// Italic lowercase → bold italic lowercase
	case 0x1D44E <= c && c <= 0x1D467:
		return shift(c, 0x1D482, (c - 0x1D44E))
	default:
		return c
	}
}

func ToBB(c rune) rune {
	switch {
	// Standard and Italic uppercase overrides → double-struck
	case c == 'C' || c == 0x1D436:
		return 'ℂ'
	case c == 'H' || c == 0x1D43B:
		return 'ℍ'
	case c == 'N' || c == 0x1D441:
		return 'ℕ'
	case c == 'P' || c == 0x1D443:
		return 'ℙ'
	case c == 'Q' || c == 0x1D444:
		return 'ℚ'
	case c == 'R' || c == 0x1D445:
		return 'ℝ'
	case c == 'Z' || c == 0x1D44D:
		return 'ℤ'

	// Standard Latin characters mapping
	case 'A' <= c && c <= 'Z':
		return shift(c, 0x1D538, (c - 'A'))
	case 'a' <= c && c <= 'z':
		return shift(c, 0x1D552, (c - 'a'))
	case '0' <= c && c <= '9':
		return shift(c, 0x1D7D8, (c - '0'))

	// Italic uppercase range (ignoring the specific overrides handled above)
	case 0x1D434 <= c && c <= 0x1D44D:
		return shift(c, 0x1D538, (c - 0x1D434))

	// Italic lowercase range → double-struck lowercase
	case 0x1D44E <= c && c <= 0x1D467:
		return shift(c, 0x1D552, (c - 0x1D44E))

	default:
		return c
	}
}

func ToUpright(c rune) rune { return c }

func ToItalic(c rune) rune {
	switch {
	case c == 'h':
		return 'ℎ'
	case 'A' <= c && c <= 'Z':
		return shift(c, 0x1D434, c-'A')
	case 'a' <= c && c <= 'z':
		return shift(c, 0x1D44E, c-'a')
	default:
		return c
	}
}

func ToSans(c rune) rune {
	switch {
	case 'A' <= c && c <= 'Z':
		return shift(c, 0x1D5A0, c-'A')
	case 'a' <= c && c <= 'z':
		return shift(c, 0x1D5BA, c-'a')
	case '0' <= c && c <= '9':
		return shift(c, 0x1D7E2, c-'0')
	// Italic uppercase → sans-serif italic uppercase
	case 0x1D434 <= c && c <= 0x1D44D:
		return shift(c, 0x1D5A0+0x60, c-0x1D434)
	// Italic lowercase → sans-serif italic lowercase
	case 0x1D44E <= c && c <= 0x1D467:
		return shift(c, 0x1D5BA+0x60, c-0x1D44E)
	default:
		return c
	}
}
