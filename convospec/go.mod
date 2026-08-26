// convospec is intentionally a nested module with NO dependencies — not even
// on its parent github.com/sneat-co/sneat-go-core. Extension contract libs
// (ext-<id>/backend) import it to declare conversational capability, and they
// are themselves dependency-free by design; adding a require here would push a
// dependency tree into every one of them.
//
// Keep this file's require block empty.
module github.com/sneat-co/sneat-go-core/convospec

go 1.26.0

toolchain go1.27.0
