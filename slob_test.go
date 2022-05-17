package goslob

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSlob(t *testing.T) {
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

			item, err := slob.Find("earth")
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
