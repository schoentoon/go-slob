package goslob

import (
	"io"
)

const POS_SIZE = 8
const REF_SIZE = 8

type RefList struct {
	slob *Slob
	info *itemListInfo
}

type Ref struct {
	Key       string
	binIndex  uint32
	itemIndex uint16
	fragment  string
}

func (r *RefList) Get(i int) (*Ref, error) {
	pos, err := r.readPointer(i)
	if err != nil {
		return nil, err
	}
	return r.readItem(r.info.dataOffset + int64(pos))
}

func (r *RefList) readPointer(i int) (int64, error) {
	pos := r.info.posOffset + int64(i*POS_SIZE)

	_, err := r.slob.reader.Seek(pos, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return read_long(r.slob.reader)
}

func (r *RefList) readItem(pos int64) (*Ref, error) {
	_, err := r.slob.reader.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, err
	}

	key, err := read_text(r.slob.reader)
	if err != nil {
		return nil, err
	}

	binIndex, err := read_int(r.slob.reader)
	if err != nil {
		return nil, err
	}

	itemIndex, err := read_short(r.slob.reader)
	if err != nil {
		return nil, err
	}

	fragment, err := read_tiny_text(r.slob.reader)
	if err != nil {
		return nil, err
	}

	return &Ref{
		Key:       key,
		binIndex:  binIndex,
		itemIndex: itemIndex,
		fragment:  fragment,
	}, nil
}

func (r *RefList) Size() uint32 {
	return r.info.count
}

func (r *RefList) Iterate() (<-chan *Ref, <-chan error) {
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
