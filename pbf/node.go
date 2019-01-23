package pbf

import (
	"fmt"

	"github.com/joe-mann/go-osm/pbf/OSMPBF"
)

type Node struct {
	ID        NodeID
	Latitude  float64
	Longitude float64
	Tags      Tags
	Info      Info
}

func DecodeNode(node *OSMPBF.Node, block *OSMPBF.PrimitiveBlock) (*Node, error) {
	tags, err := DecodeTags(node, block)
	if err != nil {
		return nil, err
	}

	info, err := DecodeInfo(node, block)
	if err != nil {
		return nil, err
	}

	latOff, longOff := block.GetLatOffset(), block.GetLonOffset()
	gran := int64(block.GetGranularity())

	return &Node{
		ID:        NodeID(node.GetId()),
		Latitude:  1e-9 * float64(latOff+gran*node.GetLat()),
		Longitude: 1e-9 * float64(longOff+gran*node.GetLon()),
		Tags:      *tags,
		Info:      *info,
	}, nil
}

func DecodeDenseNode(i int, id int64, nodes *OSMPBF.DenseNodes, block *OSMPBF.PrimitiveBlock, state *State) (*Node, error) {
	tags, err := DecodeDenseTags(nodes, block, &state.tagsState)
	if err != nil {
		return nil, err
	}

	info, err := DecodeDenseInfo(i, nodes, block, &state.infoState)
	if err != nil {
		return nil, err
	}

	state.id += id

	lats := nodes.GetLat()
	if len(lats) < i+1 {
		return nil, fmt.Errorf("unknown lat for index %d", i)
	}
	state.lat += lats[i]

	longs := nodes.GetLon()
	if len(longs) < i+1 {
		return nil, fmt.Errorf("unknown long for index %d", i)
	}
	state.long += longs[i]

	latOff, longOff := block.GetLatOffset(), block.GetLonOffset()
	gran := int64(block.GetGranularity())

	return &Node{
		ID:        NodeID(state.id),
		Latitude:  1e-9 * float64(latOff+gran*state.lat),
		Longitude: 1e-9 * float64(longOff+gran*state.long),
		Tags:      *tags,
		Info:      *info,
	}, nil
}

func (n *Node) ObjectID() ObjectID {
	return 0
}

type NodeID int64

type State struct {
	id   int64
	lat  int64
	long int64
	tagsState
	infoState
}
