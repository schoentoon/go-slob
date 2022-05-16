package goslob

import (
	"os"
	"testing"
	"time"
)

func TestStoreGetBlob(t *testing.T) {
	units := []struct {
		in     string
		expect string
	}{
		{
			in:     "testdata/lzma2.slob",
			expect: "Hello, Earth!",
		},
		{
			in:     "testdata/zlib.slob",
			expect: "Hello, Earth!",
		},
	}

	for _, unit := range units {
		t.Run(unit.in, func(t *testing.T) {
			f, err := os.Open(unit.in)
			if err != nil {
				if os.IsNotExist(err) {
					t.Skip(err)
					return
				}
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

			item, err := blob.get(0)
			if err != nil {
				t.Fatal(err)
			}

			if unit.expect != string(item.Content) {
				t.Fatalf("content not as expected: %s", item.Content)
			}

			ch, errCh := slob.ref_list.Iterate()

			timeout := time.NewTimer(time.Second * 10)
			for {
				select {
				case <-timeout.C:
					t.Fatal("Timeout after 10 seconds")
				case err := <-errCh:
					if err != nil {
						t.Fatal(err)
					} else {
						return
					}
				case <-ch:
				}
			}
		})
	}
}
