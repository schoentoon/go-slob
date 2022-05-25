package goslob

import (
	"bytes"
	"fmt"
)

type blob struct {
	content_types []uint8
	content       *bytes.Reader
	data_offset   int64
}

func newBlob(content_types []uint8, content []byte) (*blob, error) {
	return &blob{
		content_types: content_types,
		content:       bytes.NewReader(content),
		data_offset:   int64(len(content_types) * 4),
	}, nil
}

func (b *blob) readPointer(i uint16) (uint32, error) {
	pos := int64(i * 4)

	return read_int(b.content, pos)
}

func (b *blob) get(itemIndex uint16) (*Item, error) {
	if int(itemIndex) >= len(b.content_types) {
		return nil, fmt.Errorf("blob doesn't contain item with index: %d", itemIndex)
	}

	pointer, err := b.readPointer(itemIndex)
	if err != nil {
		return nil, err
	}

	pos := b.data_offset + int64(pointer)
	len, err := read_int(b.content, pos)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, len)
	_, err = b.content.ReadAt(buf, pos+4)
	if err != nil {
		return nil, err
	}

	return &Item{
		ContentType: "",
		Content:     buf,
	}, nil
}
