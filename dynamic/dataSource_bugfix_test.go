package dynamic_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/raphaelreyna/confection/dynamic"
	"gopkg.in/yaml.v3"
)

type dsConfig struct {
	Source dynamic.DataSource `yaml:"source"`
}

func TestDataSource_String(t *testing.T) {
	input := `
source:
  string: "hello world"
`
	var cfg dsConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	data, err := io.ReadAll(&cfg.Source)
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
}

func TestDataSource_Bytes(t *testing.T) {
	input := `
source:
  bytes: "raw bytes here"
`
	var cfg dsConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	data, err := io.ReadAll(&cfg.Source)
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	if string(data) != "raw bytes here" {
		t.Errorf("expected 'raw bytes here', got %q", string(data))
	}
}

func TestDataSource_Env(t *testing.T) {
	t.Setenv("CONFECTION_TEST_VAL", "from env")

	input := `
source:
  env: CONFECTION_TEST_VAL
`
	var cfg dsConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	data, err := io.ReadAll(&cfg.Source)
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	if string(data) != "from env" {
		t.Errorf("expected 'from env', got %q", string(data))
	}
}

func TestDataSource_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("file contents"), 0644); err != nil {
		t.Fatal(err)
	}

	input := "source:\n  file: " + path + "\n"
	var cfg dsConfig
	if err := yaml.Unmarshal([]byte(input), &cfg); err != nil {
		t.Fatalf("unmarshal: %s", err)
	}
	data, err := io.ReadAll(&cfg.Source)
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	if string(data) != "file contents" {
		t.Errorf("expected 'file contents', got %q", string(data))
	}
	if err := cfg.Source.Close(); err != nil {
		t.Errorf("close: %s", err)
	}
}

func TestDataSource_UnknownType(t *testing.T) {
	input := `
source:
  redis: "some-key"
`
	var cfg dsConfig
	err := yaml.Unmarshal([]byte(input), &cfg)
	if err == nil {
		t.Fatal("expected error for unknown data source type, got nil")
	}
	if !strings.Contains(err.Error(), "unknown data source type") {
		t.Fatalf("unexpected error: %s", err)
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("expected line number in error, got: %s", err)
	}
}

func TestDataSource_ReadBeforeInit(t *testing.T) {
	var ds dynamic.DataSource
	_, err := ds.Read(make([]byte, 10))
	if err == nil {
		t.Fatal("expected error reading uninitialized DataSource, got nil")
	}
}

func TestDataSource_CloseBeforeInit(t *testing.T) {
	var ds dynamic.DataSource
	err := ds.Close()
	if err == nil {
		t.Fatal("expected error closing uninitialized DataSource, got nil")
	}
}
