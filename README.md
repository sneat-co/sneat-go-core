# sneat-go-core
Core packages for Go backend of Sneat apps

[![Coverage Status](https://coveralls.io/repos/github/sneat-co/sneat-go-core/badge.svg?branch=main&kill-cache=1)](https://coveralls.io/github/sneat-co/sneat-go-core?branch=main) - [help with code coverage](https://github.com/sneat-co/sneat-go-core/issues/2) needed.

<!-- dev-approach:v1 -->
## Our approach to development

We build with our own tooling:

- **[SpecScore](https://specscore.md)** — specify requirements as `SpecScore.md` artifacts
- **[SpecStudio](https://specscore.studio)** — author & manage specs across their lifecycle
- **[inGitDB](https://ingitdb.com)** — store structured data in Git where applicable
- **[DALgo](https://dalgo.io)** — data access layer for Go
- **[cover100.dev](https://cover100.dev)** — drive toward 100% test coverage
- **[DataTug](https://datatug.io)** — query & explore data
<!-- /dev-approach -->

## Migration notes

### `facade.GetSneatDB` is a func, not a var (since v0.60.3 of sneat-co/sneat-go-core)

Before v0.60.3, `facade.GetSneatDB` was a package-level `var` that tests
reassigned directly:

```go
facade.GetSneatDB = func(ctx context.Context) (dal.DB, error) { ... } // no longer compiles
```

Since v0.60.3 it is a plain `func` and can no longer be reassigned. Tests must
inject a DB through the context instead:

```go
ctx = facade.WithSneatDB(ctx, db) // or facade.WithSneatDBProvider(ctx, provider)
```

A call site that instead did `var getSneatDB = facade.GetSneatDB` — binding
the function *value* to a local and reassigning the local in tests — keeps
compiling either way, because that pattern doesn't care whether the upstream
symbol is a var or a func.

See the doc comments on `facade.GetSneatDB` / `facade.WithSneatDB` in
[`facade/get_sneat_db.go`](facade/get_sneat_db.go) for details.
