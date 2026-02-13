package confection_test

import (
	"context"
	"strings"
	"testing"

	"github.com/raphaelreyna/confection"
	"gopkg.in/yaml.v3"
)

// --- test types ---

type Greeter interface {
	confection.Interface
	Greet() string
}

type EnglishConfig struct {
	Greeting string `yaml:"greeting"`
}

type English struct {
	Greeter
	phrase string
}

func (e *English) Greet() string { return e.phrase }

func EnglishFactory(_ context.Context, cfg *EnglishConfig) (*English, error) {
	g := cfg.Greeting
	if g == "" {
		g = "Hello"
	}
	return &English{phrase: g}, nil
}

// --- TypedConfig tests ---

func TestTypedConfig_MissingTypedConfig(t *testing.T) {
	input := `name: test`

	var tc confection.TypedConfig
	err := yaml.Unmarshal([]byte(input), &tc)
	if err == nil {
		t.Fatal("expected error for missing typed_config, got nil")
	}
	if !strings.Contains(err.Error(), "typed_config is required") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestTypedConfig_MissingAtType(t *testing.T) {
	input := `
name: test
typed_config:
  foo: bar
`
	var tc confection.TypedConfig
	err := yaml.Unmarshal([]byte(input), &tc)
	if err == nil {
		t.Fatal("expected error for missing @type, got nil")
	}
	if !strings.Contains(err.Error(), "@type not found") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestTypedConfig_ValidParse(t *testing.T) {
	input := `
name: english
typed_config:
  "@type": greetings.english
  greeting: Hi
`
	var tc confection.TypedConfig
	if err := yaml.Unmarshal([]byte(input), &tc); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if tc.Name != "english" {
		t.Errorf("expected name 'english', got %q", tc.Name)
	}
	if tc.Type() != "greetings.english" {
		t.Errorf("expected type 'greetings.english', got %q", tc.Type())
	}
}

// --- end-to-end Make test ---

func TestMake_RoundTrip(t *testing.T) {
	c := confection.NewConfection()
	confection.RegisterInterface[Greeter](c)
	confection.RegisterFactory(c, "greetings.english", EnglishFactory)

	input := `
name: english
typed_config:
  "@type": greetings.english
  greeting: Hey there
`
	var tc confection.TypedConfig
	if err := yaml.Unmarshal([]byte(input), &tc); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}

	greeter, err := confection.Make[Greeter](c, tc)
	if err != nil {
		t.Fatalf("Make: %s", err)
	}
	if got := greeter.Greet(); got != "Hey there" {
		t.Errorf("expected 'Hey there', got %q", got)
	}
}

func TestMakeCtx_RoundTrip(t *testing.T) {
	c := confection.NewConfection()
	confection.RegisterInterface[Greeter](c)
	confection.RegisterFactory(c, "greetings.english", EnglishFactory)

	input := `
name: english
typed_config:
  "@type": greetings.english
  greeting: Howdy
`
	var tc confection.TypedConfig
	if err := yaml.Unmarshal([]byte(input), &tc); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}

	ctx := context.Background()
	greeter, err := confection.MakeCtx[Greeter](ctx, c, tc)
	if err != nil {
		t.Fatalf("MakeCtx: %s", err)
	}
	if got := greeter.Greet(); got != "Howdy" {
		t.Errorf("expected 'Howdy', got %q", got)
	}
}

func TestMake_UnregisteredType(t *testing.T) {
	c := confection.NewConfection()
	confection.RegisterInterface[Greeter](c)

	input := `
name: french
typed_config:
  "@type": greetings.french
  greeting: Bonjour
`
	var tc confection.TypedConfig
	if err := yaml.Unmarshal([]byte(input), &tc); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}

	_, err := confection.Make[Greeter](c, tc)
	if err == nil {
		t.Fatal("expected error for unregistered type, got nil")
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("unexpected error: %s", err)
	}
}
