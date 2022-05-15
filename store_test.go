package goslob

import (
	"os"
	"testing"
)

func TestStoreGetBlob(t *testing.T) {
	filename := "freedict-eng-nld-0.1.1.slob"

	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	slob, err := SlobFromReader(f)
	if err != nil {
		t.Fatal(err)
	}

	blob, err := slob.store.getBlob(0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = blob.Get(0)
	if err != nil {
		t.Fatal(err)
	}
}
