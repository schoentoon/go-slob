package goslob

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf8"
)

func read_byte(r io.Reader) (uint8, error) {
	var out uint8

	err := binary.Read(r, binary.BigEndian, &out)

	return out, err
}

func read_short(r io.Reader) (uint16, error) {
	var out uint16

	err := binary.Read(r, binary.BigEndian, &out)

	return out, err
}

func read_int(r io.Reader) (uint32, error) {
	var out uint32

	err := binary.Read(r, binary.BigEndian, &out)

	return out, err
}

func read_long(r io.Reader) (int64, error) {
	var out int64

	err := binary.Read(r, binary.BigEndian, &out)

	return out, err
}

func read_byte_string(r io.Reader) ([]byte, error) {
	len, err := read_byte(r)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, len)
	_, err = r.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func read_tiny_text(r io.Reader) (string, error) {
	buf, err := read_byte_string(r)
	if err != nil {
		return "", err
	}

	if !utf8.Valid(buf) {
		return "", fmt.Errorf("Invalid utf-8")
	}

	return string(buf), nil
}

func read_text(r io.Reader) (string, error) {
	len, err := read_short(r)
	if err != nil {
		return "", err
	}

	buf := make([]byte, len)
	_, err = r.Read(buf)
	if err != nil {
		return "", err
	}

	if !utf8.Valid(buf) {
		return "", fmt.Errorf("Invalid utf-8")
	}

	return string(buf), nil
}
