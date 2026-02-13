package confection_test

import (
	"context"
	"fmt"
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
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected line number in error, got: %s", err)
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
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected line number in error, got: %s", err)
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
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected line number in error, got: %s", err)
	}
}

// --- concurrency tests ---

func TestConcurrent_RegisterAndMake(t *testing.T) {
	c := confection.NewConfection()
	confection.RegisterInterface[Greeter](c)
	confection.RegisterFactory(c, "greetings.english", EnglishFactory)

	input := `
name: english
typed_config:
  "@type": greetings.english
  greeting: concurrent
`
	var tc confection.TypedConfig
	if err := yaml.Unmarshal([]byte(input), &tc); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}

	// Run many concurrent Make calls
	const n = 100
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			g, err := confection.Make[Greeter](c, tc)
			if err != nil {
				errs <- err
				return
			}
			if g.Greet() != "concurrent" {
				errs <- fmt.Errorf("expected 'concurrent', got %q", g.Greet())
				return
			}
			errs <- nil
		}()
	}
	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}

// --- slice and nested TypedConfig tests ---

func TestSliceOfTypedConfig(t *testing.T) {
	c := confection.NewConfection()
	confection.RegisterInterface[Greeter](c)
	confection.RegisterFactory(c, "greetings.english", EnglishFactory)

	type Pipeline struct {
		Filters []confection.TypedConfig `yaml:"filters"`
	}

	input := `
filters:
- name: first
  typed_config:
    "@type": greetings.english
    greeting: Hello
- name: second
  typed_config:
    "@type": greetings.english
    greeting: Hi there
- name: third
  typed_config:
    "@type": greetings.english
    greeting: Hey
`
	var pipeline Pipeline
	if err := yaml.Unmarshal([]byte(input), &pipeline); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}

	if len(pipeline.Filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(pipeline.Filters))
	}

	expected := []struct {
		name     string
		typeName string
		greeting string
	}{
		{"first", "greetings.english", "Hello"},
		{"second", "greetings.english", "Hi there"},
		{"third", "greetings.english", "Hey"},
	}

	for i, want := range expected {
		tc := pipeline.Filters[i]
		if tc.Name != want.name {
			t.Errorf("filter[%d]: expected name %q, got %q", i, want.name, tc.Name)
		}
		if tc.Type() != want.typeName {
			t.Errorf("filter[%d]: expected type %q, got %q", i, want.typeName, tc.Type())
		}
		g, err := confection.Make[Greeter](c, tc)
		if err != nil {
			t.Fatalf("filter[%d]: Make: %s", i, err)
		}
		if g.Greet() != want.greeting {
			t.Errorf("filter[%d]: expected %q, got %q", i, want.greeting, g.Greet())
		}
	}
}

// Nested test: a factory config that itself contains a TypedConfig

type Wrapper interface {
	confection.Interface
	Inner() Greeter
}

type WrapperConfig struct {
	Prefix string                 `yaml:"prefix"`
	Child  confection.TypedConfig `yaml:"child"`
}

type WrapperImpl struct {
	Wrapper
	inner  Greeter
	prefix string
}

func (w *WrapperImpl) Inner() Greeter { return w.inner }
func (w *WrapperImpl) Greet() string  { return w.prefix + w.inner.Greet() }

func TestNestedTypedConfig(t *testing.T) {
	c := confection.NewConfection()
	confection.RegisterInterface[Greeter](c)
	confection.RegisterInterface[Wrapper](c)
	confection.RegisterFactory(c, "greetings.english", EnglishFactory)

	wrapperFactory := func(ctx context.Context, cfg *WrapperConfig) (*WrapperImpl, error) {
		inner, err := confection.MakeCtx[Greeter](ctx, c, cfg.Child)
		if err != nil {
			return nil, fmt.Errorf("nested: %w", err)
		}
		return &WrapperImpl{inner: inner, prefix: cfg.Prefix}, nil
	}
	confection.RegisterFactory(c, "wrapper", wrapperFactory)

	input := `
name: outer
typed_config:
  "@type": wrapper
  prefix: "Wrapper says: "
  child:
    name: inner
    typed_config:
      "@type": greetings.english
      greeting: Hola
`
	var tc confection.TypedConfig
	if err := yaml.Unmarshal([]byte(input), &tc); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}

	w, err := confection.Make[Wrapper](c, tc)
	if err != nil {
		t.Fatalf("Make: %s", err)
	}

	if w.Inner().Greet() != "Hola" {
		t.Errorf("inner: expected 'Hola', got %q", w.Inner().Greet())
	}

	// WrapperImpl also implements Greet through composition
	wImpl := w.(*WrapperImpl)
	if wImpl.Greet() != "Wrapper says: Hola" {
		t.Errorf("wrapper: expected 'Wrapper says: Hola', got %q", wImpl.Greet())
	}
}
