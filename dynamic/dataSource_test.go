package dynamic_test

import (
	"io"
	"testing"

	"github.com/raphaelreyna/confection/dynamic"
	"gopkg.in/yaml.v3"
)

type TestConfig struct {
	Name string             `yaml:"name"`
	Foo  dynamic.DataSource `yaml:"foo"`
}

func TestDataSource(t *testing.T) {
	configString := `
name: test
foo:
  string: "hello world"
`

	var config TestConfig
	err := yaml.Unmarshal([]byte(configString), &config)
	if err != nil {
		t.Fatalf("failed to unmarshal config: %s", err)
	}

	if config.Name != "test" {
		t.Errorf("expected config.Name to be 'test', got %s", config.Name)
	}

	fooData, err := io.ReadAll(config.Foo.ReadCloser)
	if err != nil {
		t.Fatalf("failed to read from foo: %s", err)
	}

	if string(fooData) != "hello world" {
		t.Errorf("expected foo to be 'hello world', got %s", string(fooData))
	}
}
