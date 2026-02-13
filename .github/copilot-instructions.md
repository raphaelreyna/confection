# Copilot Instructions for confection

## Build & Test

```sh
go build ./...
go test ./...
go test -race ./...                     # verify thread safety
go test ./dynamic -run TestDataSource   # single test
```

No Makefile, linter config, or CI pipeline exists.

## Architecture

Confection is a typed configuration library for Go that maps YAML config blocks to concrete implementations via a registry pattern. The core flow is:

1. **Register an interface** — `RegisterInterface[I]()` registers an interface type (must embed `confection.Interface`) into a global (or scoped) `Confection` registry.
2. **Register a factory** — `RegisterFactory()` binds a `@type` string name to a strongly-typed factory function (`Factory[Config, Impl]`). The factory's `Implementation` type must be a pointer-to-struct that embeds the target interface. Registration auto-discovers which interfaces the impl satisfies by inspecting embedded interface fields.
3. **Make** — `Make[I]()` / `MakeCtx[I]()` looks up the factory by `@type` from a `TypedConfig`, deserializes the YAML node into the factory's config type, and returns the constructed implementation.

`TypedConfig` supports nesting (a factory's config can contain child `TypedConfig` fields resolved via `Make` inside the factory) and slices (`[]TypedConfig`) for pipeline-style configs.

The `dynamic` sub-package provides `DataSource`, a YAML-unmarshallable `io.ReadCloser` that resolves to registered source types (file, env, string, bytes, or user-registered custom sources) at parse time. It uses its own `Registry` with the same dual pattern.

## Key Conventions

- **Dual API pattern**: Every public function accepts an optional `*Confection` (or `*Registry`) parameter; passing `nil` uses the package-level `Global` singleton. This allows both simple global usage and scoped/testable registries.
- **Thread safety**: All registries use `sync.RWMutex` (write-lock for registration, read-lock for Make/lookup) and `sync.Once` for global init.
- **Generics for type safety**: `RegisterInterface`, `RegisterFactory`, `Make`, and `MakeCtx` are all generic functions — the type parameter is the contract, not a runtime argument.
- **Struct tag `confection:"implement"` / `confection:"-"`**: Controls which embedded interfaces a factory implementation satisfies. Without tags, all embedded `confection.Interface` fields are matched. Use `"implement"` for opt-in or `"-"` for opt-out.
- **`TypedConfig` YAML shape**: Expects `name` + `typed_config` with a `@type` discriminator field inside `typed_config`. The `@type` field is stripped before decoding into the factory's config struct.
- **Panics on registration errors**: `RegisterInterface` and `RegisterFactory` panic on duplicate or invalid registrations (this is intentional — registration is expected at init time).
- **Line numbers in errors**: All error messages from TypedConfig parsing, Make, and DataSource include the YAML line number for debugging.
