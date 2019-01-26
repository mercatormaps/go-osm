package pbf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"github.com/mercatormaps/go-osm/pbf/OSMPBF"
)

type Scanner struct {
	in    io.Reader
	bytes uint64

	headerOnce sync.Once
	header     Header

	scanOnce  sync.Once
	objectsCh chan Object

	errOnce sync.Once
	err     error
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		in:        r,
		objectsCh: make(chan Object),
	}
}

func (s *Scanner) Header() (*Header, error) {
	var err error
	s.headerOnce.Do(func() {
		var blob *blob
		blob, err = s.blob()
		if err != nil {
			return
		}

		const typ = "OSMHeader"
		if blob.header.GetType() != typ {
			err = fmt.Errorf("invalid header; type != '%s'", typ)
			return
		}

		var h *Header
		if h, err = DecodeHeader(&blob.blob); err != nil {
			return
		}

		s.header = *h
		atomic.AddUint64(&s.bytes, blob.bytes)
	})
	return &s.header, err
}

func (s *Scanner) Scan() error {
	if _, err := s.Header(); err != nil {
		return err
	}

	s.scanOnce.Do(func() {
		blobs := make(chan *blob)

		go func() {
			defer close(blobs)

			for {
				blob, err := s.blob()
				if err == io.EOF {
					blobs <- nil
					return
				} else if err != nil {
					s.setErr(err)
					return
				}

				const typ = "OSMData"
				if blob.header.GetType() != typ {
					s.setErr(fmt.Errorf("invalid blob; type != '%s'", typ))
					return
				}

				blobs <- blob
			}
		}()

		go func() {
			wg := sync.WaitGroup{}

			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					for blob := range blobs {
						if blob == nil {
							return
						}

						s.decodeBlob(&blob.blob)
						atomic.AddUint64(&s.bytes, blob.bytes)
					}
				}(i)
			}

			wg.Wait()
			close(s.objectsCh)
		}()
	})
	return nil
}

func (s *Scanner) Object() Object {
	obj, ok := <-s.objectsCh
	if !ok {
		return nil
	}
	return obj
}

func (s *Scanner) Err() error {
	return nil
}

func (s *Scanner) Bytes() uint64 {
	return atomic.LoadUint64(&s.bytes)
}

func (s *Scanner) decodeBlob(blob *OSMPBF.Blob) {
	block, err := blob.Block()
	if err != nil {
		s.setErr(err)
		return
	}

	for _, group := range block.GetPrimitivegroup() {
		s.decodeNodes(group.GetNodes(), block)
		s.decodeDenseNodes(group.GetDense(), block)
		s.decodeWays(group.GetWays(), block)
		s.decodeRelations(group.GetRelations(), block)
	}
}

func (s *Scanner) decodeNodes(nodes []*OSMPBF.Node, block *OSMPBF.PrimitiveBlock) {
	for _, node := range nodes {
		n, err := DecodeNode(node, block)
		if err != nil {
			s.setErr(err)
			return
		}
		s.objectsCh <- n
	}
}

func (s *Scanner) decodeDenseNodes(nodes *OSMPBF.DenseNodes, block *OSMPBF.PrimitiveBlock) {
	ids := nodes.GetId()

	var state State
	for i := range ids {
		if len(ids) < i+1 {
			s.setErr(nil) // TODO
			return
		}
		id := ids[i]

		n, err := DecodeDenseNode(i, id, nodes, block, &state)
		if err != nil {
			s.setErr(err)
			return
		}
		s.objectsCh <- n
	}
}

func (s *Scanner) decodeWays(ways []*OSMPBF.Way, block *OSMPBF.PrimitiveBlock) {
	for _, way := range ways {
		w, err := DecodeWay(way, block)
		if err != nil {
			s.setErr(err)
			return
		}
		s.objectsCh <- w
	}
}

func (s *Scanner) decodeRelations(rels []*OSMPBF.Relation, block *OSMPBF.PrimitiveBlock) {
	for _, rel := range rels {
		r, err := DecodeRelation(rel, block)
		if err != nil {
			s.setErr(err)
			return
		}
		s.objectsCh <- r
	}
}

func (s *Scanner) setErr(err error) {
	s.errOnce.Do(func() {
		s.err = &Error{} // TODO
	})
}

func (s *Scanner) blob() (*blob, error) {
	out := blob{}

	buf := bytes.Buffer{}
	if n, err := io.CopyN(&buf, s.in, 4); err != nil {
		return nil, err
	} else if n != 4 {
		return nil, fmt.Errorf("%d != 4", n)
	}
	out.bytes += 4

	headerLen := int64(binary.BigEndian.Uint32(buf.Bytes()))

	buf.Reset()
	if n, err := io.CopyN(&buf, s.in, headerLen); err != nil {
		return nil, err
	} else if n != headerLen {
		return nil, fmt.Errorf("%d != %d", n, headerLen)
	}
	out.bytes += uint64(headerLen)

	if err := proto.Unmarshal(buf.Bytes(), &out.header); err != nil {
		return nil, err
	}

	buf.Reset()
	if n, err := io.CopyN(&buf, s.in, int64(out.header.GetDatasize())); err != nil {
		return nil, nil
	} else if n != int64(out.header.GetDatasize()) {
		return nil, fmt.Errorf("%d != %d", n, out.header.GetDatasize())
	}
	out.bytes += uint64(out.header.GetDatasize())

	if err := proto.Unmarshal(buf.Bytes(), &out.blob); err != nil {
		return nil, err
	}
	return &out, nil
}

type blob struct {
	header OSMPBF.BlobHeader
	blob   OSMPBF.Blob
	bytes  uint64
}
