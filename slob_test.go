package goslob

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestFromReader(t *testing.T) {
	filename := "freedict-eng-nld-0.1.1.slob"

	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip(err)
			return
		}
		t.Fatal(err)
	}
	defer f.Close()

	_, err = SlobFromReader(f)
	if err != nil {
		t.Fatal(err)
	}
}

func FuzzFromReader(f *testing.F) {
	seed := func() {
		filename := "freedict-eng-nld-0.1.1.slob"
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
	seed()

	f.Fuzz(func(t *testing.T, data []byte, binIndex, itemIndex int) {
		reader := bytes.NewReader(data)
		slob, err := SlobFromReader(reader)
		if err == nil {
			_, _ = slob.Get(binIndex, itemIndex)
		}
	})
}
