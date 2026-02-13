package dynamic_test

import (
	"io"
	"strings"
	"testing"

	"github.com/raphaelreyna/confection/dynamic"
	"gopkg.in/yaml.v3"
)

func TestRegisterSource_CustomType(t *testing.T) {
	reg := dynamic.NewRegistry()

	// Register a custom "upper" source that uppercases the value
	dynamic.RegisterSource(reg, "upper", func(value string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(strings.ToUpper(value))), nil
	})
	// Also register "string" so we can test alongside built-ins
	dynamic.RegisterSource(reg, "string", func(value string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(value)), nil
	})

	input := `
source:
  upper: "hello world"
`
	var cfg struct {
		Source dynamic.DataSource `yaml:"source"`
	}
	cfg.Source.Registry = reg

	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	data, err := io.ReadAll(&cfg.Source)
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	if string(data) != "HELLO WORLD" {
		t.Errorf("expected 'HELLO WORLD', got %q", string(data))
	}
}

func TestRegisterSource_ScopedRegistryDoesNotAffectGlobal(t *testing.T) {
	reg := dynamic.NewRegistry()
	dynamic.RegisterSource(reg, "custom", func(value string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(value)), nil
	})

	// "custom" should not be recognized by a DataSource using the global registry
	input := `
source:
  custom: "test"
`
	var cfg struct {
		Source dynamic.DataSource `yaml:"source"`
	}
	// Registry is nil → uses global, which doesn't have "custom"
	err := yaml.Unmarshal([]byte(input), &cfg)
	if err == nil {
		t.Fatal("expected error for unregistered source on global registry, got nil")
	}
	if !strings.Contains(err.Error(), "unknown data source type") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestRegisterSource_PanicsOnDuplicate(t *testing.T) {
	reg := dynamic.NewRegistry()
	dynamic.RegisterSource(reg, "dup", func(value string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(value)), nil
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on duplicate registration, got none")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "already registered") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	dynamic.RegisterSource(reg, "dup", func(value string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(value)), nil
	})
}

func TestRegisterSource_GlobalRegistry(t *testing.T) {
	// The global registry should have the built-in types available
	input := `
source:
  string: "global test"
`
	var cfg struct {
		Source dynamic.DataSource `yaml:"source"`
	}

	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	data, err := io.ReadAll(&cfg.Source)
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	if string(data) != "global test" {
		t.Errorf("expected 'global test', got %q", string(data))
	}
}
