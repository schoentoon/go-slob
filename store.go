package goslob

import (
	"bytes"
	"io"
)

type store struct {
	slob          *Slob
	info          *itemListInfo
	decompressor  decompressor
	content_types []string
}

type Item struct {
	ContentType string
	Content     []byte
}

func (s *store) GetRef(ref *Ref) (*Item, error) {
	return s.Get(ref.BinIndex, ref.ItemIndex)
}

func (s *store) Get(binIndex uint32, itemIndex uint16) (*Item, error) {
	blob, err := s.getBlob(binIndex)
	if err != nil {
		return nil, err
	}
	item, err := blob.get(itemIndex)
	if err != nil {
		return nil, err
	}
	item.ContentType = s.content_types[blob.content_types[itemIndex]]

	return item, err
}

func (s *store) readPointer(i uint32) (int64, error) {
	pos := s.info.posOffset + int64(i*POS_SIZE)

	return read_long(s.slob.reader, pos)
}

func (s *store) getBlob(binIndex uint32) (*blob, error) {
	pos, err := s.readPointer(binIndex)
	if err != nil {
		return nil, err
	}
	return s.readBlob(s.info.dataOffset + int64(pos))
}

func (s *store) readBlob(pos int64) (*blob, error) {
	binCount, err := read_int(s.slob.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += 4

	content_types := make([]uint8, 0, binCount)
	for i := 0; i < int(binCount); i++ {
		typ, err := read_byte(s.slob.reader, pos)
		if err != nil {
			return nil, err
		}
		content_types = append(content_types, typ)
		pos++
	}

	compressed_len, err := read_int(s.slob.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += 4

	reader := io.NewSectionReader(s.slob.reader, pos, int64(compressed_len))
	buf := bytes.Buffer{}

	err = s.decompressor.Decompress(reader, &buf)
	if err != nil {
		return nil, err
	}

	return newBlob(content_types, buf.Bytes())
}

func (s *store) Size() uint32 {
	return s.info.count
}
