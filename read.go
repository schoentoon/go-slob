package goslob

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf8"
)

func read_byte(r io.ReaderAt, pos int64) (uint8, error) {
	buf := make([]byte, 1)

	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return 0, err
	}

	return buf[0], nil
}

func read_short(r io.ReaderAt, pos int64) (uint16, error) {
	buf := make([]byte, 2)

	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint16(buf), nil
}

func read_int(r io.ReaderAt, pos int64) (uint32, error) {
	buf := make([]byte, 4)

	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(buf), nil
}

func read_long(r io.ReaderAt, pos int64) (int64, error) {
	buf := make([]byte, 8)

	_, err := r.ReadAt(buf, pos)
	if err != nil {
		return 0, err
	}

	return int64(binary.BigEndian.Uint64(buf)), nil
}

func read_byte_string(r io.ReaderAt, pos int64) (int, []byte, error) {
	len, err := read_byte(r, pos)
	if err != nil {
		return 1, nil, err
	}

	buf := make([]byte, len)
	n, err := r.ReadAt(buf, pos+1)
	if err != nil {
		return n + 1, nil, err
	}

	// the + 1 is from read_byte
	return n + 1, buf, nil
}

func read_tiny_text(r io.ReaderAt, pos int64) (int, string, error) {
	n, buf, err := read_byte_string(r, pos)
	if err != nil {
		return n, "", err
	}

	if !utf8.Valid(buf) {
		return n, "", fmt.Errorf("Invalid utf-8")
	}

	return n, string(buf), nil
}

func read_text(r io.ReaderAt, pos int64) (int, string, error) {
	len, err := read_short(r, pos)
	if err != nil {
		return 2, "", err
	}

	buf := make([]byte, len)
	n, err := r.ReadAt(buf, pos+2)
	if err != nil {
		return n + 2, "", err
	}

	if !utf8.Valid(buf) {
		return n + 2, "", fmt.Errorf("Invalid utf-8")
	}

	return n + 2, string(buf), nil
}
