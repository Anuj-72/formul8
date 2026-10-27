// Package formul8 renders LaTeX math expressions to a styled 2D rune grid
// for terminal output.
//
// Ports reference/txm/src/lib.rs.
package formul8

// compactSimpleFractionalExponents ports lib.rs:28. False in the reference.

// Layout(input string) (*layout.Node, error) ports lib.rs:31-41:
// tokenize -> parse -> render, with the root RenderCtx starting at
// Style{}.Italic().

// Render(input string) (string, error) ports lib.rs:44-48: Layout, then the
// terminal backend.

// defaultRegistry replaces the OnceLock at lib.rs:50-53. The registry is
// immutable after construction and Go runs package variable initialisers
// before main, single-threaded, so a plain package-level var suffices:
//
//   var defaultRegistry = buildRegistry()

// buildRegistry ports lib.rs:55-256 verbatim -- roughly 200 entries:
//   lowercase Greek + vee/wedge         lib.rs:58-87
//   uppercase Greek                     :89-102
//   function names as TextGlyph         :104-110
//   binom frac sqrt lim int sum prod    :112-118
//   color                               :120
//   operators, relations, arrows, sets,
//     delimiters, dots, letterlike      :122-192
//   abs and the bare pipe               :194-195
//   mathbb                              :197
//   mathbf mathrm mathsf mathit boldsymbol   :199-211
//   text textbf textit textrm           :213-220
//   colorbox                            :222
//   accents, unstretched                :224-243
//   overline widehat widetilde, stretched    :245-253
