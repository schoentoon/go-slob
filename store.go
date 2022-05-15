package goslob

import (
	"bytes"
	"io"
)

type Store struct {
	slob          *Slob
	info          *itemListInfo
	decompressor  decompressor
	content_types []string
}

type Item struct {
	ContentType string
	Content     []byte
}

func (s *Store) GetRef(ref *Ref) (*Item, error) {
	return s.Get(ref.binIndex, ref.itemIndex)
}

func (s *Store) Get(binIndex uint32, itemIndex uint16) (*Item, error) {
	return nil, nil
}

func (s *Store) readPointer(i uint32) (int64, error) {
	pos := s.info.posOffset + int64(i*POS_SIZE)

	_, err := s.slob.reader.Seek(pos, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return read_long(s.slob.reader)
}

func (s *Store) getBlob(binIndex uint32) (*blob, error) {
	pos, err := s.readPointer(binIndex)
	if err != nil {
		return nil, err
	}
	return s.readBlob(s.info.dataOffset + int64(pos))
}

func (s *Store) readBlob(pos int64) (*blob, error) {
	_, err := s.slob.reader.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, err
	}

	binCount, err := read_int(s.slob.reader)
	if err != nil {
		return nil, err
	}

	content_types := make([]uint8, 0, binCount)
	for i := 0; i < int(binCount); i++ {
		typ, err := read_byte(s.slob.reader)
		if err != nil {
			return nil, err
		}
		content_types = append(content_types, typ)
	}

	compressed_len, err := read_int(s.slob.reader)
	if err != nil {
		return nil, err
	}

	limitedReader := io.LimitReader(s.slob.reader, int64(compressed_len))
	buf := bytes.Buffer{}

	err = s.decompressor.Decompress(limitedReader, &buf)
	if err != nil {
		return nil, err
	}

	return newBlob(content_types, buf.Bytes())
}

func (s *Store) Size() uint32 {
	return s.info.count
}
