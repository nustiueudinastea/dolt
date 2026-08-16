# Swarmion-maintained Dolt module

This branch publishes the patched Dolt dependency under
`github.com/nustiueudinastea/dolt/go`. The distinct module path is deliberate:
Go does not inherit a dependency module's `replace` directives, so Swarmion's
public module cannot depend on an unpublished fork that still declares the
upstream `github.com/dolthub/dolt/go` identity.

The branch is based on upstream commit `b23cc55053` and carries Swarmion's
pure-Go Zstandard, WebAssembly, and lock-isolation patches. Self-imports and
protobuf Go package metadata use the maintained module path; no behavioral
code was changed as part of the module-identity rewrite.

When syncing a newer upstream revision, preserve the maintained module path,
reapply the Swarmion patches, regenerate protobuf bindings after changing any
`go_package` option, and run the native pure-Go and browser/WASM gates before
publishing a new commit.
