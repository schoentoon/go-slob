package goslob

import (
	"fmt"
)

const POS_SIZE = 8
const REF_SIZE = 8

type ref_list struct {
	slob *Slob
	info *itemListInfo
}

type Ref struct {
	Key       string
	BinIndex  uint32
	ItemIndex uint16

	slob     *Slob
	fragment string
}

func (r *Ref) Get() (*Item, error) {
	return r.slob.store.GetRef(r)
}

func (r *ref_list) Get(i int) (*Ref, error) {
	pos, err := r.readPointer(i)
	if err != nil {
		return nil, err
	}
	return r.readItem(r.info.dataOffset + int64(pos))
}

func (r *ref_list) readPointer(i int) (int64, error) {
	pos := r.info.posOffset + int64(i*POS_SIZE)

	return read_long(r.slob.reader, pos)
}

func (r *ref_list) readItem(pos int64) (*Ref, error) {
	n, key, err := read_text(r.slob.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += int64(n)

	binIndex, err := read_int(r.slob.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += 4

	itemIndex, err := read_short(r.slob.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += 2

	_, fragment, err := read_tiny_text(r.slob.reader, pos)
	if err != nil {
		return nil, err
	}

	return &Ref{
		Key:       key,
		BinIndex:  binIndex,
		ItemIndex: itemIndex,
		slob:      r.slob,
		fragment:  fragment,
	}, nil
}

func (r *ref_list) Find(key string) (*Ref, error) {
	size := int(r.Size())
	for i := 0; i < size; i++ {
		pos, err := r.readPointer(i)
		if err != nil {
			return nil, err
		}

		_, text, err := read_text(r.slob.reader, r.info.dataOffset+int64(pos))
		if err != nil {
			return nil, err
		}
		if key == text {
			return r.readItem(r.info.dataOffset + int64(pos))
		}
	}

	return nil, fmt.Errorf("Unable to find item with key: %s", key)
}

func (r *ref_list) Size() uint32 {
	return r.info.count
}

func (r *ref_list) Iterate() (<-chan *Ref, <-chan error) {
	ch := make(chan *Ref)
	errCh := make(chan error)

	go func() {
		size := int(r.Size())
		defer close(ch)
		defer close(errCh)

		for i := 0; i < size; i++ {
			ref, err := r.Get(i)
			if err != nil {
				errCh <- err
				break
			}
			ch <- ref
		}
	}()

	return ch, errCh
}
