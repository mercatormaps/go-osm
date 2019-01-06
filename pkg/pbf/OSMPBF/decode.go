package OSMPBF

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
)

func (b *Blob) Data() ([]byte, error) {
	switch {
	case b.Raw != nil:
		return b.GetRaw(), nil

	case b.ZlibData != nil:
		r, err := zlib.NewReader(bytes.NewReader(b.GetZlibData()))
		if err != nil {
			return nil, err
		}
		defer r.Close()

		buf := bytes.Buffer{}
		if n, err := buf.ReadFrom(r); err != nil {
			return nil, err
		} else if n != int64(b.GetRawSize()) {
			return nil, fmt.Errorf("%d != %d", n, b.GetRawSize())
		}
		return buf.Bytes(), nil

	default:
		return nil, errors.New("unknown data type")
	}
}

func (b *Blob) Block() (*PrimitiveBlock, error) {
	data, err := b.Data()
	if err != nil {
		return nil, err
	}

	block := PrimitiveBlock{}
	if err := proto.Unmarshal(data, &block); err != nil {
		return nil, err
	}
	return &block, nil
}
