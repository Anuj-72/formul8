// Package style holds the packed text style bitfield and the color palette.
//
// Ports reference/txm/src/style.rs.
package style

import (
	"strconv"
	"strings"
)

// Style is a u32 bitfield: attribute bits, an 8-bit fg index, an 8-bit bg
// index, and one "lock" bit per field. Ports style.rs:5-241.
//
// Value receivers throughout: this is a 4-byte integer meant to be copied.
//
// Bit layout (style.rs:35-50):
//   BOLD 1<<0  ITALIC 1<<1  UNDERLINE 1<<2  DIM 1<<3
//   FG_MASK 0x000FF0 (shift 4)   BG_MASK 0x0FF000 (shift 12)
//   LOCK_BOLD 1<<20 .. LOCK_BG 1<<25

// Builders: New, Bold, Italic, Underline, Dim, UnBold, UnItalic,
// UnUnderline, UnDim, Fg, Bg. Each sets the value bit AND the lock bit,
// and returns a new Style.

// Queries: IsBold, IsItalic, IsUnderline, IsDim, FgColor, BgColor, IsEmpty.

// Merge (style.rs:131-195) is the lock-bit resolution: other's lock wins,
// then self's lock, then a plain OR. Port the bit arithmetic literally --
// this is what makes \color{red}{...} reach every rendered cell.

// WriteANSIPrefix (style.rs:201-240) emits the SGR escape. Colors go out as
// 38;5;{n} where n is the internal 1..16 index, offset by one from the
// standard 256-color palette. Port as-is; golden parity depends on it.

type Style uint32

const (
	bold Style = 1 << iota
	italic
	underline
	dim
)

const (
	fgMask Style = 0x000FF0
	bgMask Style = 0x0FF000
)

const (
	lockBold      Style = 1 << 20
	lockItalic    Style = 1 << 21
	lockUnderline Style = 1 << 22
	lockDim       Style = 1 << 23
	lockFg        Style = 1 << 24
	lockBg        Style = 1 << 25

	attrMask Style = bold | italic | underline | dim
)

// Constructor for readability
func New() Style { return Style(0) }

// --- Builders ---

func (s Style) Bold() Style      { return s | bold | lockBold }
func (s Style) Italic() Style    { return s | italic | lockItalic }
func (s Style) Underline() Style { return s | underline | lockUnderline }
func (s Style) Dim() Style       { return s | dim | lockDim }

func (s Style) UnBold() Style      { return (s &^ bold) | lockBold }
func (s Style) UnItalic() Style    { return (s &^ italic) | lockItalic }
func (s Style) UnUnderline() Style { return (s &^ underline) | lockUnderline }
func (s Style) UnDim() Style       { return (s &^ dim) | lockDim }

func (s Style) Fg(c Color) Style { return (s &^ fgMask) | (Style(c) << 4) | lockFg }
func (s Style) Bg(c Color) Style { return (s &^ bgMask) | (Style(c) << 12) | lockBg }

// --- Queries ---

func (s Style) IsBold() bool      { return s&bold != 0 }
func (s Style) IsItalic() bool    { return s&italic != 0 }
func (s Style) IsUnderline() bool { return s&underline != 0 }
func (s Style) IsDim() bool       { return s&dim != 0 }
func (s Style) IsEmpty() bool     { return s&(attrMask|fgMask|bgMask) == 0 }

func (s Style) FgColor() uint8 { return uint8((s & fgMask) >> 4) }
func (s Style) BgColor() uint8 { return uint8((s & bgMask) >> 12) }

// Merge function
func (s Style) Merge(other Style) Style {
	var combinedLock = s | other

	var boldOut Style
	if other&lockBold != 0 {
		boldOut = other & bold
	} else if s&lockBold != 0 {
		boldOut = s & bold
	} else {
		boldOut = (s | other) & bold
	}

	var italicOut Style
	if other&lockItalic != 0 {
		italicOut = other & italic
	} else if s&lockItalic != 0 {
		italicOut = s & italic
	} else {
		italicOut = (s | other) & italic
	}

	var underlineOut Style
	if other&lockUnderline != 0 {
		underlineOut = other & underline
	} else if s&lockUnderline != 0 {
		underlineOut = s & underline
	} else {
		underlineOut = (s | other) & underline
	}

	var dimOut Style
	if other&lockDim != 0 {
		dimOut = other & dim
	} else if s&lockDim != 0 {
		dimOut = s & dim
	} else {
		dimOut = (s | other) & dim
	}

	var fg Style
	if other&lockFg != 0 {
		fg = other & fgMask
	} else if s&lockFg != 0 {
		fg = s & fgMask
	} else {
		otherFg := other & fgMask
		fg = otherFg
		if otherFg == 0 {
			fg = s & fgMask
		}
	}

	var bg Style
	if other&lockBg != 0 {
		bg = other & bgMask
	} else if s&lockBg != 0 {
		bg = s & bgMask
	} else {
		otherBg := other & bgMask
		bg = otherBg
		if otherBg == 0 {
			bg = s & bgMask
		}
	}

	var lockPart = combinedLock &
		(lockBold | lockItalic | lockUnderline | lockDim | lockFg | lockBg)

	return Style(lockPart | boldOut | italicOut | underlineOut | dimOut | fg | bg)
}

func (s Style) WriteANSIPrefix() string {
	if s.IsEmpty() {
		return ""
	}

	var b strings.Builder
	b.WriteString("\x1b[")
	first := true

	if s.IsBold() {
		b.WriteString("1")
		first = false
	}
	if s.IsItalic() {
		if !first {
			b.WriteString(";")
		}
		b.WriteString("3")
		first = false
	}
	if s.IsUnderline() {
		if !first {
			b.WriteString(";")
		}
		b.WriteString("4")
		first = false
	}
	if s.IsDim() {
		if !first {
			b.WriteString(";")
		}
		b.WriteString("2")
		first = false
	}

	fg := s.FgColor()
	bg := s.BgColor()

	if fg != 0 {
		if !first {
			b.WriteString(";")
		}
		b.WriteString("38;5;")
		b.WriteString(strconv.Itoa(int(fg)))
		first = false
	}

	if bg != 0 {
		if !first {
			b.WriteString(";")
		}
		b.WriteString("48;5;")
		b.WriteString(strconv.Itoa(int(bg)))
	}

	b.WriteString("m")
	return b.String()
}
