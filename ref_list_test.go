package goslob

import (
	"os"
	"testing"
)

func TestIterateRefList(t *testing.T) {
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

	size := slob.ref_list.Size()
	if size != 7717 {
		t.Fatalf("Size isn't correct: %d", size)
	}

	ch, errCh := slob.ref_list.Iterate()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatal(err)
			} else {
				return
			}
		case <-ch:
		}
	}
}
