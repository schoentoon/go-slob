package goslob

import (
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
