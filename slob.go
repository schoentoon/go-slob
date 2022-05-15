package goslob

import (
	"bytes"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Slob struct {
	reader        io.ReadSeeker
	uuid          uuid.UUID
	encoding      string
	tags          map[string]string
	content_types []string

	blob_count   uint32
	store_offset int64
	size         int64

	ref_list *RefList
	store    *Store
}

type itemListInfo struct {
	count      uint32
	posOffset  int64
	dataOffset int64
}

var magic = []byte{'!', '-', '1', 'S', 'L', 'O', 'B', 0x1F}
var valid_compression = []string{"bz2", "zlib", "lzma2"}

func SlobFromReader(f io.ReadSeeker) (*Slob, error) {
	magicbuf := make([]byte, len(magic))

	n, err := f.Read(magicbuf)
	if err != nil {
		return nil, err
	}
	if n != len(magic) {
		return nil, fmt.Errorf("Input too short")
	}

	if !bytes.Equal(magicbuf, magic) {
		return nil, fmt.Errorf("No magic match: %#v", magicbuf)
	}

	out := &Slob{
		reader: f,
	}

	uuidbuf := make([]byte, 16)

	n, err = f.Read(uuidbuf)
	if err != nil {
		return nil, err
	}
	if n != 16 {
		return nil, fmt.Errorf("Input too short")
	}
	err = out.uuid.UnmarshalBinary(uuidbuf)
	if err != nil {
		return nil, err
	}

	encoding, err := read_byte_string(out.reader)
	if err != nil {
		return nil, err
	}
	out.encoding = string(encoding)
	if out.encoding != "utf-8" {
		return nil, fmt.Errorf("Invalid encoding: %s", out.encoding)
	}

	compression, err := read_tiny_text(out.reader)
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

	err = out.read_tags()
	if err != nil {
		return nil, err
	}

	err = out.read_content_types()
	if err != nil {
		return nil, err
	}

	out.blob_count, err = read_int(out.reader)
	if err != nil {
		return nil, err
	}

	out.store_offset, err = read_long(out.reader)
	if err != nil {
		return nil, err
	}

	out.size, err = read_long(out.reader)
	if err != nil {
		return nil, err
	}

	refList, err := out.read_item_list_info(-1)
	if err != nil {
		return nil, err
	}

	storeList, err := out.read_item_list_info(int64(out.store_offset))
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

func (s *Slob) read_tags() error {
	count, err := read_byte(s.reader)
	if err != nil {
		return err
	}
	s.tags = make(map[string]string, count)

	for i := 0; uint8(i) < count; i++ {
		key, err := read_tiny_text(s.reader)
		if err != nil {
			return err
		}
		value, err := read_tiny_text(s.reader)
		if err != nil {
			return err
		}
		s.tags[key] = value
	}

	return nil
}

func (s *Slob) read_content_types() error {
	count, err := read_byte(s.reader)
	if err != nil {
		return err
	}
	s.content_types = make([]string, 0, count)

	for i := 0; uint8(i) < count; i++ {
		content, err := read_text(s.reader)
		if err != nil {
			return err
		}
		s.content_types = append(s.content_types, content)
	}

	return nil
}

func (s *Slob) read_item_list_info(offset int64) (*itemListInfo, error) {
	if offset < 0 {
		n, err := s.reader.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		offset = n
	} else {
		_, err := s.reader.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}

	count, err := read_int(s.reader)
	if err != nil {
		return nil, err
	}

	posOffset := int64(offset + 4) // the 4 is the size of the encoded count at the start

	return &itemListInfo{
		count:      count,
		posOffset:  posOffset,
		dataOffset: int64(uint32(posOffset) + (count * REF_SIZE)),
	}, nil
}

func (s *Slob) init_reflist(refListInfo *itemListInfo) (*RefList, error) {
	out := &RefList{
		slob: s,
		info: refListInfo,
	}

	return out, nil
}

func (s *Slob) init_store(compression string, storeInfo *itemListInfo) (*Store, error) {
	decompressor, err := get_decompressor(compression)
	if err != nil {
		return nil, err
	}

	out := &Store{
		slob:          s,
		info:          storeInfo,
		decompressor:  decompressor,
		content_types: s.content_types,
	}

	return out, nil
}
