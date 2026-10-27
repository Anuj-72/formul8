package glyph

// SymbolRegistry maps a macro name to its Glyph. Ports glyph.rs:56-74.
//
// Rust's HashMap<String, Box<dyn Glyph>> is just map[string]Glyph here --
// Go interface values are already boxed, there is nothing to add.
//
//   NewSymbolRegistry() *SymbolRegistry
//   (*SymbolRegistry) Register(name string, g Glyph)
//   (*SymbolRegistry) Get(name string) (Glyph, bool)
//
// Option<&dyn Glyph> becomes the comma-ok idiom on Get.

type SymbolRegistry struct {
	m map[string]Glyph
}

func NewSymbolRegistry() *SymbolRegistry {
	return &SymbolRegistry{
		m: make(map[string]Glyph),
	}
}

func (s *SymbolRegistry) Register(name string, g Glyph) {
	s.m[name] = g
}

func (s *SymbolRegistry) Get(name string) (Glyph, bool) {
	g, ok := s.m[name]
	return g, ok
}
