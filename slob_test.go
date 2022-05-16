package goslob

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func FuzzFromReader(f *testing.F) {
	seed := func(filename string) {
		file, err := os.Open(filename)
		if err != nil {
			return
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return
		}
		f.Add(data, 0, 0)
	}
	seed("testdata/lzma2.slob")
	seed("testdata/zlib.slob")

	f.Fuzz(func(t *testing.T, data []byte, binIndex, itemIndex int) {
		reader := bytes.NewReader(data)
		slob, err := SlobFromReader(reader)
		if err == nil {
			_, _ = slob.Get(binIndex, itemIndex)
		}
	})
}
