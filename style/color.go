package style

// Color is a u8 palette index. Ports style.rs:243-299.
//
// Constants NONE=0, BLACK=1, RED=2, GREEN=3, YELLOW=4, BLUE=5, MAGENTA=6,
// CYAN=7, WHITE=8, BRIGHT_BLACK=9 .. BRIGHT_WHITE=16.

// ParseColor maps a color name to a Color, or returns an
// errs.InvalidColor error. Accepts the aliases gray/grey for brightblack.

import (
	"github.com/Anuj-72/formul8/errs"
)

type Color uint8

const (
	None Color = iota
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

func ParseColor(s string) (Color, error) {
	switch s {
	case "black":
		return Black, nil
	case "red":
		return Red, nil
	case "green":
		return Green, nil
	case "yellow":
		return Yellow, nil
	case "blue":
		return Blue, nil
	case "magenta":
		return Magenta, nil
	case "cyan":
		return Cyan, nil
	case "white":
		return White, nil
	case "brightblack", "gray", "grey":
		return BrightBlack, nil
	case "brightred":
		return BrightRed, nil
	case "brightgreen":
		return BrightGreen, nil
	case "brightyellow":
		return BrightYellow, nil
	case "brightblue":
		return BrightBlue, nil
	case "brightmagenta":
		return BrightMagenta, nil
	case "brightcyan":
		return BrightCyan, nil
	case "brightwhite":
		return BrightWhite, nil
	default:
		return None, errs.InvalidColor(s)
	}
}
