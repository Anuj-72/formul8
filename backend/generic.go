package backend

// The second pass: absolute x/y placement of an already-sized tree.
//
// Ports reference/txm/src/backends/generic_backend.rs. Everything here is
// unexported -- the Rust module was pub(crate).
//
//   renderNode(n, buf, x, y)        generic_backend.rs:6
//   renderInner(n, buf, x, y, inherited)  :10, one case per layout.Kind
//   resolveDelimiterSingle          :183
//   resolveDelimiters               :190  per-row box-drawing pieces
//   renderSummation                 :222  three branches, and note the
//                                         isPowerOfTwo test at :275 that
//                                         picks the arrow row of a tall
//                                         sigma -- port it literally
//   renderProduct                   :289
//   renderIntegral                  :326
//   renderMatrix                    :360  recomputes cell metrics rather
//                                         than reusing MatrixOf's
