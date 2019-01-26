package pbf

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/mercatormaps/go-osm/pbf/OSMPBF"
)

type Header struct {
	RequiredFeatures          []string
	OptionalFeatures          []string
	WritingProgram            string
	Source                    string
	ReplicationSequenceNumber int64
	ReplicationBaseURL        string
	ReplicationTimestamp      time.Time
	BoundingBox               *BoundingBox
}

type BoundingBox struct {
	Top    float64
	Bottom float64
	Right  float64
	Left   float64
}

func DecodeHeader(blob *OSMPBF.Blob) (*Header, error) {
	data, err := blob.Data()
	if err != nil {
		return nil, err
	}

	h := OSMPBF.HeaderBlock{}
	if err := proto.Unmarshal(data, &h); err != nil {
		return nil, err
	}

	header := &Header{
		RequiredFeatures:          h.GetRequiredFeatures(),
		OptionalFeatures:          h.GetOptionalFeatures(),
		WritingProgram:            h.GetWritingprogram(),
		Source:                    h.GetSource(),
		ReplicationBaseURL:        h.GetOsmosisReplicationBaseUrl(),
		ReplicationSequenceNumber: h.GetOsmosisReplicationSequenceNumber(),
	}

	if h.OsmosisReplicationTimestamp != nil {
		header.ReplicationTimestamp = time.Unix(*h.OsmosisReplicationTimestamp, 0)
	}

	if h.Bbox != nil {
		header.BoundingBox = &BoundingBox{
			Top:    1e-9 * float64(*h.Bbox.Top),
			Bottom: 1e-9 * float64(*h.Bbox.Bottom),
			Right:  1e-9 * float64(*h.Bbox.Right),
			Left:   1e-9 * float64(*h.Bbox.Left),
		}
	}

	return header, nil
}
