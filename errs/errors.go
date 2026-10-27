// Package errs holds the parse error type and byte spans shared by the
// token, parser, layout and glyph stages.
//
// Ports reference/txm/src/error.rs.
//
// Named "errs" and not "error" because a package named error would shadow
// Go's predeclared error interface in every file that imports it.
package errs

import "fmt"

// Span is a half-open byte range into the input string. Ports Range<usize>.
// Offsets are BYTE offsets, not rune indices: the parser compares
// prevEnd != span.Start to detect whitespace between adjacent tokens, and
// that comparison must use the same units the lexer produced.
type Span struct {
	Start int
	End   int
}

// Kind identifies which ParseError variant this is. Ports the variants of
// the thiserror enum in error.rs, in the same order.
type Kind int

const (
	KindInvalidToken Kind = iota
	KindUnexpectedEOF
	KindUnexpectedToken
	KindUnexpectedWhitespace
	KindUnexpectedRightWithoutLeft
	KindExpectedToken
	KindExpectedTokenGot
	KindExpectedString
	KindExpectedDelimiter
	KindMismatchedDelimiters
	KindUnclosedLeftRight
	KindUnclosedEnvironment
	KindMismatchedEnvironment
	KindUnknownEnvironment
	KindExpectedColorName
	KindInvalidColor
	KindMismatchedMatrixRows
	KindInternal
	KindLocated
	KindLocatedEOF
)

// ParseError is the single error type for the whole pipeline. Msg is the
// fully formatted message, built at construction, so Error() is trivial and
// the strings stay byte-identical to the Rust #[error(...)] attributes.
type ParseError struct {
	Kind    Kind
	Msg     string
	Byte    int
	Snippet string
}

// Error implements the built-in error interface. Any type with this one
// method automatically satisfies `error` in Go -- no explicit "implements"
// declaration needed, unlike Rust's `impl Error for ParseError`.
func (e *ParseError) Error() string {
	return e.Msg
}

// --- Constructors, one per error.rs variant ---

func InvalidToken(b int) *ParseError {
	return &ParseError{
		Kind: KindInvalidToken,
		Msg:  fmt.Sprintf("invalid token at byte %d", b),
		Byte: b,
	}
}

func UnexpectedEOF() *ParseError {
	return &ParseError{
		Kind: KindUnexpectedEOF,
		Msg:  "unexpected end of input",
	}
}

func UnexpectedToken(pos int, tok string) *ParseError {
	return &ParseError{
		Kind: KindUnexpectedToken,
		Msg:  fmt.Sprintf("unexpected token at position %d: %s", pos, tok),
	}
}

func UnexpectedWhitespace() *ParseError {
	return &ParseError{
		Kind: KindUnexpectedWhitespace,
		Msg:  "unexpected whitespace",
	}
}

func UnexpectedRightWithoutLeft() *ParseError {
	return &ParseError{
		Kind: KindUnexpectedRightWithoutLeft,
		Msg:  `unexpected \right without matching \left`,
	}
}

func ExpectedToken(expected string) *ParseError {
	return &ParseError{
		Kind: KindExpectedToken,
		Msg:  fmt.Sprintf("expected %s", expected),
	}
}

func ExpectedTokenGot(expected, got string) *ParseError {
	return &ParseError{
		Kind: KindExpectedTokenGot,
		Msg:  fmt.Sprintf("expected %s, got %s", expected, got),
	}
}

func ExpectedString() *ParseError {
	return &ParseError{
		Kind: KindExpectedString,
		Msg:  "expected a string",
	}
}

func ExpectedDelimiter(side string) *ParseError {
	return &ParseError{
		Kind: KindExpectedDelimiter,
		Msg:  fmt.Sprintf(`expected a delimiter after \%s`, side),
	}
}

func MismatchedDelimiters(left, right rune) *ParseError {
	return &ParseError{
		Kind: KindMismatchedDelimiters,
		Msg:  fmt.Sprintf(`mismatched delimiters: \left%c and \right%c`, left, right),
	}
}

func UnclosedLeftRight() *ParseError {
	return &ParseError{
		Kind: KindUnclosedLeftRight,
		Msg:  `unclosed \left ... \right pair`,
	}
}

func UnclosedEnvironment(name string) *ParseError {
	return &ParseError{
		Kind: KindUnclosedEnvironment,
		Msg:  fmt.Sprintf(`unclosed \begin{%s}`, name),
	}
}

func MismatchedEnvironment(begin, end string) *ParseError {
	return &ParseError{
		Kind: KindMismatchedEnvironment,
		Msg:  fmt.Sprintf(`mismatched \begin{%s} and \end{%s}`, begin, end),
	}
}

func UnknownEnvironment(name string) *ParseError {
	return &ParseError{
		Kind: KindUnknownEnvironment,
		Msg:  fmt.Sprintf("unknown matrix environment: %s", name),
	}
}

func ExpectedColorName() *ParseError {
	return &ParseError{
		Kind: KindExpectedColorName,
		Msg:  "expected a color name",
	}
}

func InvalidColor(name string) *ParseError {
	return &ParseError{
		Kind: KindInvalidColor,
		Msg:  fmt.Sprintf("invalid color name: %s", name),
	}
}

func MismatchedMatrixRows() *ParseError {
	return &ParseError{
		Kind: KindMismatchedMatrixRows,
		Msg:  "matrix rows have different lengths",
	}
}

func Internal(msg string) *ParseError {
	return &ParseError{
		Kind: KindInternal,
		Msg:  fmt.Sprintf("internal parser error: %s", msg),
	}
}

// --- Wrappers, porting error.rs:70-92 ---

// At wraps e with the source span where it occurred, producing a new
// error whose message embeds the original message as a prefix.
func (e *ParseError) At(span Span, input string) *ParseError {
	snippet := input[span.Start:span.End]
	return &ParseError{
		Kind:    KindLocated,
		Snippet: snippet,
		Byte:    span.Start,
		Msg:     fmt.Sprintf("%s near '%s' at byte %d", e.Msg, snippet, span.Start),
	}
}

// AtEOF wraps e for the case where parsing ran out of input. If input is
// empty there is nothing to snippet, so e is returned unchanged.
func (e *ParseError) AtEOF(input string) *ParseError {
	if input == "" {
		return e
	}

	start := len(input) - 5
	if start < 0 {
		start = 0
	}
	snippet := input[start:]

	return &ParseError{
		Kind:    KindLocatedEOF,
		Snippet: snippet,
		Msg:     fmt.Sprintf("%s but reached end of input near '%s'", e.Msg, snippet),
	}
}
