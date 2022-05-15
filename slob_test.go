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
		t.Fatal(err)
	}
	defer f.Close()

	_, err = SlobFromReader(f)
	if err != nil {
		t.Fatal(err)
	}
}

func FuzzFromReader(f *testing.F) {
	{
		filename := "freedict-eng-nld-0.1.1.slob"
		file, err := os.Open(filename)
		if err != nil {
			f.Fatal(err)
		}
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		reader := bytes.NewReader(data)
		_, _ = SlobFromReader(reader)
	})
}
