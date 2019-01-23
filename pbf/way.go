package pbf

import (
	"fmt"

	"github.com/joe-mann/go-osm/pbf/OSMPBF"
)

type Way struct {
	ID      WayID
	NodeIDs []NodeID
	Tags    Tags
	Info    Info
}

func DecodeWay(way *OSMPBF.Way, block *OSMPBF.PrimitiveBlock) (*Way, error) {
	tags, err := DecodeTags(way, block)
	if err != nil {
		return nil, err
	}

	info, err := DecodeInfo(way, block)
	if err != nil {
		return nil, err
	}

	refs := way.GetRefs()
	nodeIDs := make([]NodeID, len(refs))
	var id NodeID
	for i := range refs {
		id += NodeID(refs[i])

		if len(nodeIDs) < i+1 {
			return nil, fmt.Errorf("unknown node ID for ref %d", i)
		}
		nodeIDs[i] = id
	}

	return &Way{
		ID:      WayID(way.GetId()),
		NodeIDs: nodeIDs,
		Tags:    *tags,
		Info:    *info,
	}, nil
}

func (n *Way) ObjectID() ObjectID {
	return 0
}

type WayID int64
