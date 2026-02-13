package dynamic

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DataSource is a YAML-unmarshallable io.ReadCloser that resolves its
// underlying data from a registered source type. Set Registry to use
// a scoped registry; nil uses the global default.
type DataSource struct {
	ReadCloser io.ReadCloser
	Registry   *Registry `yaml:"-"`
	read       func(p []byte) (n int, err error)
	close      func() error
}

func (ds *DataSource) UnmarshalYAML(value *yaml.Node) error {
	if ds == nil {
		return errors.New("data source is nil")
	}

	reg := getRegistry(ds.Registry)

	var readCloser io.ReadCloser
	for idx := 0; idx+1 < len(value.Content); idx += 2 {
		key := value.Content[idx].Value
		val := value.Content[idx+1].Value

		factory, ok := reg.lookup(key)
		if !ok {
			return fmt.Errorf("line %d: unknown data source type %s", value.Content[idx].Line, key)
		}
		rc, err := factory(val)
		if err != nil {
			return fmt.Errorf("line %d: data source %s: %w", value.Content[idx].Line, key, err)
		}
		readCloser = rc
		break
	}
	if readCloser == nil {
		return errors.New("data source type not found")
	}

	ds.read = readCloser.Read
	ds.close = readCloser.Close
	ds.ReadCloser = readCloser

	return nil
}

func (ds *DataSource) Read(p []byte) (n int, err error) {
	if ds.read == nil {
		return 0, errors.New("data source not initialized")
	}
	return ds.read(p)
}

func (ds *DataSource) Close() error {
	if ds.close == nil {
		return errors.New("data source not initialized")
	}
	return ds.close()
}

// FileDataSource reads from a file, opening it lazily on first Read.
type FileDataSource struct {
	Filename string
	file     *os.File
}

func (f *FileDataSource) Read(p []byte) (n int, err error) {
	if f.file == nil {
		f.file, err = os.Open(f.Filename)
		if err != nil {
			return 0, fmt.Errorf("failed to open file %s: %w", f.Filename, err)
		}
	}
	return f.file.Read(p)
}

func (f *FileDataSource) Close() error {
	if f.file != nil {
		err := f.file.Close()
		f.file = nil
		if err != nil {
			return fmt.Errorf("failed to close file %s: %w", f.Filename, err)
		}
	}
	return nil
}

// StringDataSource reads from an in-memory string.
type StringDataSource struct {
	Value string
	buf   *strings.Reader
}

func (s *StringDataSource) Read(p []byte) (n int, err error) {
	if s.buf == nil {
		s.buf = strings.NewReader(s.Value)
	}

	return s.buf.Read(p)
}

func (s *StringDataSource) Close() error {
	s.buf = nil
	return nil
}

// BytesDataSource reads from an in-memory byte slice.
type BytesDataSource struct {
	Value []byte
	buf   *bytes.Reader
}

func (b *BytesDataSource) Read(p []byte) (n int, err error) {
	if b.buf == nil {
		b.buf = bytes.NewReader(b.Value)
	}

	return b.buf.Read(p)
}

func (b *BytesDataSource) Close() error {
	b.buf = nil
	return nil
}

// EnvironmentDataSource reads from an environment variable, resolved lazily on first Read.
type EnvironmentDataSource struct {
	Key string
	buf *strings.Reader
}

func (e *EnvironmentDataSource) Read(p []byte) (n int, err error) {
	if e.buf == nil {
		val, ok := os.LookupEnv(e.Key)
		if !ok {
			return 0, fmt.Errorf("environment variable %s not found", e.Key)
		}
		e.buf = strings.NewReader(val)
	}

	return e.buf.Read(p)
}

func (e *EnvironmentDataSource) Close() error {
	e.buf = nil
	return nil
}

// Built-in source factories used by the default registry.

func fileSource(value string) (io.ReadCloser, error) {
	return &FileDataSource{Filename: value}, nil
}

func envSource(value string) (io.ReadCloser, error) {
	return &EnvironmentDataSource{Key: value}, nil
}

func stringSource(value string) (io.ReadCloser, error) {
	return &StringDataSource{Value: value}, nil
}

func bytesSource(value string) (io.ReadCloser, error) {
	return &BytesDataSource{Value: []byte(value)}, nil
}
