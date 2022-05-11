package goslob

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Slob struct {
	reader        io.ReadSeekCloser
	uuid          uuid.UUID
	encoding      string
	tags          map[string]string
	content_types []string

	blob_count   uint32
	store_offset uint64
	size         uint64
}

var magic = []byte{'!', '-', '1', 'S', 'L', 'O', 'B', 0x1F}
var valid_compression = []string{"bz2", "zlib", "lzma2"}

func SlobFromReader(f io.ReadSeekCloser) (*Slob, error) {
	magicbuf := make([]byte, len(magic))

	n, err := f.Read(magicbuf)
	if err != nil {
		return nil, err
	}
	if n != len(magic) {
		return nil, fmt.Errorf("Input too short")
	}

	if !bytes.Equal(magicbuf, magic) {
		return nil, fmt.Errorf("No magic match: %#v", magicbuf)
	}

	out := &Slob{
		reader: f,
	}

	uuidbuf := make([]byte, 16)

	n, err = f.Read(uuidbuf)
	if err != nil {
		return nil, err
	}
	if n != 16 {
		return nil, fmt.Errorf("Input too short")
	}
	err = out.uuid.UnmarshalBinary(uuidbuf)
	if err != nil {
		return nil, err
	}

	encoding, err := out.read_byte_string()
	if err != nil {
		return nil, err
	}
	out.encoding = string(encoding)
	if out.encoding != "utf-8" {
		return nil, fmt.Errorf("Invalid encoding: %s", out.encoding)
	}

	compression, err := out.read_tiny_text()
	if err != nil {
		return nil, err
	}
	valid := false
	for _, comp := range valid_compression {
		if comp == compression {
			valid = true
		}
	}
	if !valid {
		return nil, fmt.Errorf("Invalid compression: %s", compression)
	}

	err = out.read_tags()
	if err != nil {
		return nil, err
	}

	err = out.read_content_types()
	if err != nil {
		return nil, err
	}

	out.blob_count, err = out.read_int()
	if err != nil {
		return nil, err
	}

	out.store_offset, err = out.read_long()
	if err != nil {
		return nil, err
	}

	out.size, err = out.read_long()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Slob) read_byte() (uint8, error) {
	var out uint8

	err := binary.Read(s.reader, binary.BigEndian, &out)

	return out, err
}

func (s *Slob) read_short() (uint16, error) {
	var out uint16

	err := binary.Read(s.reader, binary.BigEndian, &out)

	return out, err
}

func (s *Slob) read_int() (uint32, error) {
	var out uint32

	err := binary.Read(s.reader, binary.BigEndian, &out)

	return out, err
}

func (s *Slob) read_long() (uint64, error) {
	var out uint64

	err := binary.Read(s.reader, binary.BigEndian, &out)

	return out, err
}

func (s *Slob) read_byte_string() ([]byte, error) {
	len, err := s.read_byte()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, len)
	_, err = s.reader.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *Slob) read_tiny_text() (string, error) {
	buf, err := s.read_byte_string()
	if err != nil {
		return "", err
	}

	if !utf8.Valid(buf) {
		return "", fmt.Errorf("Invalid utf-8")
	}

	return string(buf), nil
}

func (s *Slob) read_text() (string, error) {
	len, err := s.read_short()
	if err != nil {
		return "", err
	}

	buf := make([]byte, len)
	_, err = s.reader.Read(buf)
	if err != nil {
		return "", err
	}

	if !utf8.Valid(buf) {
		return "", fmt.Errorf("Invalid utf-8")
	}

	return string(buf), nil
}

func (s *Slob) read_tags() error {
	count, err := s.read_byte()
	if err != nil {
		return err
	}
	s.tags = make(map[string]string, count)

	for i := 0; uint8(i) < count; i++ {
		key, err := s.read_tiny_text()
		if err != nil {
			return err
		}
		value, err := s.read_tiny_text()
		if err != nil {
			return err
		}
		s.tags[key] = value
	}

	return nil
}

func (s *Slob) read_content_types() error {
	count, err := s.read_byte()
	if err != nil {
		return err
	}
	s.content_types = make([]string, 0, count)

	for i := 0; uint8(i) < count; i++ {
		content, err := s.read_text()
		if err != nil {
			return err
		}
		s.content_types = append(s.content_types, content)
	}

	return nil
}
