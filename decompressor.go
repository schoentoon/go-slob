package goslob

import (
	"compress/zlib"
	"fmt"
	"io"
)

var decompressors map[string]decompressor = make(map[string]decompressor)

func init() {
	decompressors["zlib"] = &zlibDecompressor{}
}

func get_decompressor(name string) (decompressor, error) {
	out, ok := decompressors[name]
	if ok {
		return out, nil
	}
	return nil, fmt.Errorf("No decompressor found: %s", name)
}

type decompressor interface {
	Decompress(in io.Reader, out io.Writer) error
}

type zlibDecompressor struct {
}

func (z *zlibDecompressor) Decompress(in io.Reader, out io.Writer) error {
	reader, err := zlib.NewReader(in)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, reader)
	return err
}
