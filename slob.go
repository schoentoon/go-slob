package goslob

import (
	"bytes"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Slob struct {
	reader        io.ReaderAt
	Uuid          uuid.UUID
	encoding      string
	Tags          map[string]string
	Content_types []string

	blob_count   uint32
	store_offset int64
	size         int64

	ref_list *ref_list
	store    *store
}

type itemListInfo struct {
	count      uint32
	posOffset  int64
	dataOffset int64
}

var magic = []byte{'!', '-', '1', 'S', 'L', 'O', 'B', 0x1F}
var valid_compression = []string{"bz2", "zlib", "lzma2"}

func SlobFromReader(f io.ReaderAt) (*Slob, error) {
	pos := int64(0)
	magicbuf := make([]byte, len(magic))

	n, err := f.ReadAt(magicbuf, 0)
	if err != nil {
		return nil, err
	}
	if n != len(magic) {
		return nil, fmt.Errorf("Input too short")
	}

	if !bytes.Equal(magicbuf, magic) {
		return nil, fmt.Errorf("No magic match: %#v", magicbuf)
	}

	pos += int64(n)

	out := &Slob{
		reader: f,
	}

	uuidbuf := make([]byte, 16)

	n, err = f.ReadAt(uuidbuf, pos)
	if err != nil {
		return nil, err
	}
	if n != 16 {
		return nil, fmt.Errorf("Input too short")
	}
	err = out.Uuid.UnmarshalBinary(uuidbuf)
	if err != nil {
		return nil, err
	}

	pos += int64(n)

	n, encoding, err := read_byte_string(out.reader, pos)
	if err != nil {
		return nil, err
	}
	out.encoding = string(encoding)
	if out.encoding != "utf-8" {
		return nil, fmt.Errorf("Invalid encoding: %s", out.encoding)
	}

	pos += int64(n)

	n, compression, err := read_tiny_text(out.reader, pos)
	if err != nil {
		return nil, err
	}
	valid := false
	for _, comp := range valid_compression {
		if comp == compression {
			valid = true
		}
	}
	if !valid {
		return nil, fmt.Errorf("Invalid compression: %s", compression)
	}

	pos += int64(n)

	n, err = out.read_tags(pos)
	if err != nil {
		return nil, err
	}

	pos += int64(n)

	n, err = out.read_content_types(pos)
	if err != nil {
		return nil, err
	}

	pos += int64(n)

	out.blob_count, err = read_int(out.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += 4

	out.store_offset, err = read_long(out.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += 8

	out.size, err = read_long(out.reader, pos)
	if err != nil {
		return nil, err
	}

	pos += 8

	_, refList, err := out.read_item_list_info(pos)
	if err != nil {
		return nil, err
	}

	_, storeList, err := out.read_item_list_info(int64(out.store_offset))
	if err != nil {
		return nil, err
	}

	out.ref_list, err = out.init_reflist(refList)
	if err != nil {
		return nil, err
	}

	out.store, err = out.init_store(compression, storeList)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Slob) Keys() (<-chan *Ref, <-chan error) {
	return s.ref_list.Iterate()
}

func (s *Slob) Get(binIndex, itemIndex int) (*Item, error) {
	if binIndex < 0 {
		return nil, fmt.Errorf("Invalid binIndex: %d", binIndex)
	}
	if itemIndex < 0 {
		return nil, fmt.Errorf("Invalid itemIndex: %d", itemIndex)
	}
	return s.store.Get(uint32(binIndex), uint16(itemIndex))
}

func (s *Slob) Size() uint32 {
	return s.store.Size()
}

func (s *Slob) Find(key string) (*Item, error) {
	ref, err := s.ref_list.Find(key)
	if err != nil {
		return nil, err
	}

	return ref.Get()
}

func (s *Slob) read_tags(pos int64) (int, error) {
	count, err := read_byte(s.reader, pos)
	if err != nil {
		return 1, err
	}
	s.Tags = make(map[string]string, count)

	read := 1

	for i := 0; uint8(i) < count; i++ {
		n, key, err := read_tiny_text(s.reader, pos+int64(read))
		if err != nil {
			return read, err
		}

		read += n

		n, value, err := read_tiny_text(s.reader, pos+int64(read))
		if err != nil {
			return read, err
		}
		s.Tags[key] = value

		read += n
	}

	return read, nil
}

func (s *Slob) read_content_types(pos int64) (int, error) {
	count, err := read_byte(s.reader, pos)
	if err != nil {
		return 1, err
	}
	s.Content_types = make([]string, 0, count)

	read := 1

	for i := 0; uint8(i) < count; i++ {
		n, content, err := read_text(s.reader, pos+int64(read))
		if err != nil {
			return read, err
		}
		s.Content_types = append(s.Content_types, content)

		read += n
	}

	return read, nil
}

func (s *Slob) read_item_list_info(pos int64) (int, *itemListInfo, error) {
	count, err := read_int(s.reader, pos)
	if err != nil {
		return 4, nil, err
	}

	posOffset := int64(pos + 4) // the 4 is the size of the encoded count at the start

	return 4, &itemListInfo{
		count:      count,
		posOffset:  posOffset,
		dataOffset: int64(uint32(posOffset) + (count * REF_SIZE)),
	}, nil
}

func (s *Slob) init_reflist(refListInfo *itemListInfo) (*ref_list, error) {
	out := &ref_list{
		slob: s,
		info: refListInfo,
	}

	return out, nil
}

func (s *Slob) init_store(compression string, storeInfo *itemListInfo) (*store, error) {
	decompressor, err := get_decompressor(compression)
	if err != nil {
		return nil, err
	}

	out := &store{
		slob:          s,
		info:          storeInfo,
		decompressor:  decompressor,
		content_types: s.Content_types,
	}

	return out, nil
}
