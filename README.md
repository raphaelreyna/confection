# Confection

A configuration library written in Go with a focus on typed configs.

# Example

```go
package main

import (
	"context"

	"github.com/raphaelreyna/confection"
	"gopkg.in/yaml.v3"
)

// In general we just want a person that says a greeting
// and would like to support multiple languages / greetings.
type Person interface {
	confection.Interface
	SayHello() string
}

const config = `
greeting:
  name: spanish
  typed_config:
    "@type": greetings.spanish
    use_formal: true # tu v.s usted (optional, default: false)
`

// Config is the configuration struct that will be used to
// unmarshal the YAML configuration.
// This is the config for our application.
type Config struct {
	Greeting confection.TypedConfig `yaml:"greeting"`
}

// The main function is the entry point of the application.
// It parses the YAML configuration and uses confection to
// create the appropriate Person implementation based on the
// configuration.
func main() {
	confection.RegisterInterface[Person](nil)

	// This line may be called from anywhere in the codebase
	// such as from a blank import or from a different package
	// in an init function, etc.
	// Its included here for the purpose of the example.
	confection.RegisterFactory(nil, "greetings.spanish", SpanishFactory)

	var c Config
	err := yaml.Unmarshal([]byte(config), &c)
	if err != nil {
		panic(err)
	}

	// This is where we use confection to create the Person
	// implementation based on the configuration.
	person, err := confection.Make[Person](nil, c.Greeting)
	if err != nil {
		panic(err)
	}

	// This is where we use the Person implementation
	// to say hello.
	// The implementation will depend on the configuration.
	println(person.SayHello())
}

// ------------------------------------------------------
// The following code is the implementation of the Spanish greeting.
// This is an example of how one could add custom types to the
// main application without having to modify the main code.
// ------------------------------------------------------

// SpanishConfig is a struct that represents the configuration
// for the Spanish greeting.
// Note that we are able to use our custom config type.
type SpanishConfig struct {
	Formal bool `yaml:"use_formal"`
}

// Spanish is a struct that implements the Person interface
// and represents a Spanish greeting.
type Spanish struct {
	Person // confection.Interface that this struct implements
	phrase string
}

// SayHello implements the Person interface
func (s *Spanish) SayHello() string {
	return s.phrase
}

// SpanishFactory is a factory function that creates a Spanish instance
// based on the provided configuration.
// Note that we are able to use our custom config type without
// having to deal with the underlying YAML node or the `any` type.
func SpanishFactory(_ context.Context, config *SpanishConfig) (*Spanish, error) {
	phrase := "Hola, ¿cómo estás?"
	if config.Formal {
		phrase = "Hola, ¿cómo está usted?"
	}
	return &Spanish{
		phrase: phrase,
	}, nil
}
```