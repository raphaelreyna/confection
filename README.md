# Confection

Config-driven polymorphism for Go. Register interface implementations with typed factory functions, then construct them from YAML using a `@type` discriminator — similar to how Envoy proxy handles its filter configs.

```
go get github.com/raphaelreyna/confection
```

## How it works

1. Define an interface that embeds `confection.Interface`
2. Register a factory function that maps a config struct to an implementation
3. Put a `@type` field in your YAML to select which factory to use
4. Call `Make` to get back a concrete implementation

The `@type` field is stripped before your factory's config struct is deserialized, so factories receive clean, strongly-typed configs without knowing about the dispatch mechanism.

## Example

```yaml
greeting:
  name: spanish
  typed_config:
    "@type": greetings.spanish
    use_formal: true
```

```go
// Define your interface
type Person interface {
    confection.Interface
    SayHello() string
}

// Define an implementation
type Spanish struct {
    Person
    phrase string
}

func (s *Spanish) SayHello() string { return s.phrase }

// Define a typed factory
func SpanishFactory(_ context.Context, cfg *SpanishConfig) (*Spanish, error) {
    phrase := "Hola, ¿cómo estás?"
    if cfg.Formal {
        phrase = "Hola, ¿cómo está usted?"
    }
    return &Spanish{phrase: phrase}, nil
}

type SpanishConfig struct {
    Formal bool `yaml:"use_formal"`
}
```

```go
// Wire it up
confection.RegisterInterface[Person](nil)
confection.RegisterFactory(nil, "greetings.spanish", SpanishFactory)

// Parse config and construct
var c struct {
    Greeting confection.TypedConfig `yaml:"greeting"`
}
yaml.Unmarshal(configBytes, &c)

person, err := confection.Make[Person](nil, c.Greeting)
person.SayHello() // "Hola, ¿cómo está usted?"
```

## Nested and composed configs

`TypedConfig` fields can be nested — a factory's config struct can itself contain `TypedConfig` fields, which are resolved by calling `Make` inside the factory. Slices of `TypedConfig` also work for pipeline-style configs:

```yaml
filters:
- name: auth
  typed_config:
    "@type": middleware.auth
    provider: oauth2
- name: ratelimit
  typed_config:
    "@type": middleware.ratelimit
    max_rps: 100
```

## Scoped registries

Passing `nil` as the first argument to any function uses a global registry. For testing or isolation, create a scoped one:

```go
c := confection.NewConfection()
confection.RegisterInterface[Person](c)
confection.RegisterFactory(c, "greetings.spanish", SpanishFactory)
person, err := confection.Make[Person](c, tc)
```

## Dynamic data sources

The `dynamic` sub-package provides `DataSource`, a YAML-unmarshallable `io.ReadCloser` for config values that resolve at read time:

```yaml
credentials:
  file: /etc/secrets/key.pem
```

Built-in sources: `file`, `env`, `string`, `bytes`. Register custom ones:

```go
dynamic.RegisterSource(nil, "vault", func(value string) (io.ReadCloser, error) {
    return vault.ReadSecret(value)
})
```

```yaml
credentials:
  vault: secret/my-app/key
```

## Thread safety

All registries are safe for concurrent use. Registration takes a write lock, `Make` and data source lookups take a read lock.