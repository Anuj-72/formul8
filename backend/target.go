// Package backend positions a sized layout tree into a flat character
// buffer and serialises it with ANSI styling.
//
// Ports reference/txm/src/backend.rs and reference/txm/src/backends/.
//
// The generic Backend trait (backend.rs:4-9) is NOT ported: Go interfaces
// cannot carry associated types cleanly, and with the ratatui backend out of
// scope there is exactly one output format. RenderTarget -- the actual
// polymorphism point -- stays an interface.
package backend

// RenderTarget is the write surface the positioner draws into.
// Ports backend.rs:11-14.
//
//   Set(x, y int, ch rune, st style.Style)
//   FillRow(y, xStart, xEnd int, ch rune, st style.Style)
//
// Rust took `impl RenderTarget` for static dispatch; a plain interface
// parameter here is dynamic dispatch, which costs nothing at this scale.
