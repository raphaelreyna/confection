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

type DataSource struct {
	ReadCloser io.ReadCloser
	read       func(p []byte) (n int, err error) `yaml:"-"`
	close      func() error                      `yaml:"-"`
}

func (ds *DataSource) UnmarshalYAML(value *yaml.Node) error {
	if ds == nil {
		return errors.New("data source is nil")
	}

	var readCloser io.ReadCloser
	for idx := 0; idx+1 < len(value.Content); idx += 2 {
		key := value.Content[idx].Value
		val := value.Content[idx+1].Value
		switch key {
		case "file":
			readCloser = &FileDataSource{
				Filename: val,
			}
		case "env":
			readCloser = &EnvironmentDataSource{
				Key: val,
			}
		case "string":
			readCloser = &StringDataSource{
				Value: val,
			}
		case "bytes":
			readCloser = &BytesDataSource{
				Value: []byte(val),
			}
		default:
			return fmt.Errorf("unknown data source type %s", key)
		}
		if readCloser != nil {
			break
		}
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
