package layout

// Superscript and subscript character tables, plus the flattening walk that
// feeds them. Used by SuperscriptOf and SubscriptOf in build.go to decide
// whether scripts can render inline.
//
// Ports layout_tree.rs:633-697. All unexported.

// superscripts ports the match arms of to_superscript_char,
// layout_tree.rs:636-650. The gaps are Unicode's, not typos: no 'q', and only
// A B D E G H I J M N O P R T U W of the uppercase letters.
var superscripts = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ',
	'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'i': 'ⁱ', 'j': 'ʲ',
	'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'n': 'ⁿ', 'o': 'ᵒ',
	'p': 'ᵖ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ',
	'v': 'ᵛ', 'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ', 'z': 'ᶻ',
	'A': 'ᴬ', 'B': 'ᴮ', 'D': 'ᴰ', 'E': 'ᴱ', 'G': 'ᴳ',
	'H': 'ᴴ', 'I': 'ᴵ', 'J': 'ᴶ', 'M': 'ᴹ', 'N': 'ᴺ',
	'O': 'ᴼ', 'P': 'ᴾ', 'R': 'ᴿ', 'T': 'ᵀ', 'U': 'ᵁ',
	'W': 'ᵂ', '+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽',
	')': '⁾',
}

func toSuperscriptChar(r rune) (rune, bool) {
	val, ok := superscripts[r]
	return val, ok
}

// subscripts ports the match arms of to_subscript_char,
// layout_tree.rs:656-662. Digits, a subset of lowercase, and + - = ( ).
// No uppercase at all.
var subscripts = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
	'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
	'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ',
	'k': 'ₖ', 'l': 'ₗ', 'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ',
	'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ', 'u': 'ᵤ',
	'v': 'ᵥ', 'x': 'ₓ', 'y': 'ᵧ', '+': '₊', '-': '₋',
	'=': '₌', '(': '₍', ')': '₎',
}

func toSubscriptChar(r rune) (rune, bool) {
	val, ok := subscripts[r]
	return val, ok
}

// mapContent walks a Kind and flattens Text and HStack into a []rune, applying
// f to every text rune. Ports layout_tree.rs:668-697.
//
// Callers compare len(result) against the source node's Width to decide whether
// a script renders inline, so every rune counted here is load-bearing.
//
// Three details that must not drift:
//
//   - Rust used map_while (:678): it STOPS at the first rune f rejects. An
//     explicit loop with break, never a filter. Dropping unmappable runes
//     instead of stopping would keep the length up and wrongly allow inline.
//   - That stop is scoped to ONE Text node. The enclosing HStack loop keeps
//     walking later children and keeps appending spacing.
//   - Spacing emits raw ' ' runes, NOT passed through f, and only between
//     children (if i+1 != n), never after the last.
func mapContent(k Kind, f func(rune) (rune, bool)) []rune {
	var result []rune

	// Recursive closure: Go has no nested func declarations, so declare the
	// type first, then assign, or the body cannot refer to itself.
	// Appending to the captured result is how Rust's &mut Vec threading is
	// expressed here -- the assignment is visible to the caller after return.
	var mapContentInternal func(Kind)
	mapContentInternal = func(node Kind) {
		switch n := node.(type) {
		case Text:
			for _, c := range n.Content {
				nextChar, ok := f(c)
				if !ok {
					break // map_while halts this node at the first None
				}
				result = append(result, nextChar)
			}

		case HStack:
			numChildren := len(n.Children)
			for i, child := range n.Children {
				// No nil check on purpose. Rust's Vec<LayoutNode> holds
				// values, so a nil child cannot exist there; skipping one
				// here would still emit its trailing spacing and silently
				// diverge. Let it panic instead -- HStackOf must never
				// admit a nil child.
				mapContentInternal(child.Kind)

				// Spacing goes between children only: if i + 1 != n.
				if i+1 != numChildren {
					for range n.Spacing {
						result = append(result, ' ')
					}
				}
			}

		default:
			// Every other Kind contributes nothing, matching Rust's
			// `_ => {}` at :695. Not an oversight -- a Kind that cannot
			// flatten to plain runes cannot render inline, and the empty
			// result is what makes the width check fail.
		}
	}

	mapContentInternal(k)
	return result
}
