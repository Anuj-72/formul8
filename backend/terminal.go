package backend

// The terminal output format: a flat rune + style grid, serialised with
// run-length-grouped ANSI escapes, one reset per row.
//
// Ports reference/txm/src/backends/terminal.rs.

// charBuf is the RenderTarget implementation: parallel Data []rune and
// Styles []style.Style slices over a Width x Height grid, with Set and
// FillRow bounds-checked. Ports terminal.rs:37-53, 90-104.

// String walks row by row, emitting an escape only when the style changes
// from the previous cell and a reset before each change and at end of row.
// Ports terminal.rs:55-88.
//
// Guard Width == 0 and return "": Rust's chunks_exact(0) panics on empty
// input, which is a bug in the reference we do not want to reproduce.

// RenderTerminal(tree *layout.Node) string replaces the Backend impl at
// terminal.rs:25-34: allocate a charBuf at the tree's size, renderNode into
// it at 0,0, return its String.
