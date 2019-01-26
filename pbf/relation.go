package pbf

import (
	"github.com/mercatormaps/go-osm/pbf/OSMPBF"
)

type Relation struct {
	ID   RelationID
	Tags Tags
	Info Info
}

func DecodeRelation(rel *OSMPBF.Relation, block *OSMPBF.PrimitiveBlock) (*Relation, error) {
	tags, err := DecodeTags(rel, block)
	if err != nil {
		return nil, err
	}

	info, err := DecodeInfo(rel, block)
	if err != nil {
		return nil, err
	}

	return &Relation{
		ID:   RelationID(rel.GetId()),
		Tags: *tags,
		Info: *info,
	}, nil
}

func (n *Relation) ObjectID() ObjectID {
	return 0
}

type RelationID int64
