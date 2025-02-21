package dynamic

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type DataSource struct {
	ReadCloser io.ReadCloser
	read       func(p []byte) (n int, err error) `yaml:"-"`
	close      func() error                      `yaml:"-"`
}

func (ds *DataSource) UnmarshalYAML(value *yaml.Node) error {
	if ds == nil {
		return fmt.Errorf("data source is nil")
	}

	var readCloser io.ReadCloser
	for idx, content := range value.Content {
		switch content.Value {
		case "file":
			readCloser = &FileDataSource{
				Filename: value.Content[idx+1].Value,
			}
		case "env":
			readCloser = &EnviornmentDataSource{
				Key: value.Content[idx+1].Value,
			}
		case "string":
			readCloser = &StringDataSource{
				Value: value.Content[idx+1].Value,
			}
		case "bytes":
			readCloser = &BytesDataSource{
				Value: []byte(value.Content[idx+1].Value),
			}
		default:
			return fmt.Errorf("unknown data source type %s", content.Value)
		}
		if readCloser != nil {
			break
		}
	}
	if readCloser == nil {
		return fmt.Errorf("data source type not found")
	}

	ds.read = readCloser.Read
	ds.close = readCloser.Close
	ds.ReadCloser = readCloser

	return nil
}

func (ds *DataSource) Read(p []byte) (n int, err error) {
	if ds.read == nil {
		return 0, fmt.Errorf("data source not initialized")
	}
	return ds.read(p)
}

func (ds *DataSource) Close() error {
	if ds.close == nil {
		return fmt.Errorf("data source not initialized")
	}
	return ds.close()
}

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

type EnviornmentDataSource struct {
	Key string
	buf *strings.Reader
}

func (e *EnviornmentDataSource) Read(p []byte) (n int, err error) {
	if e.buf == nil {
		val, ok := os.LookupEnv(e.Key)
		if !ok {
			return 0, fmt.Errorf("environment variable %s not found", e.Key)
		}
		e.buf = strings.NewReader(val)
	}

	return e.buf.Read(p)
}

func (e *EnviornmentDataSource) Close() error {
	e.buf = nil
	return nil
}
