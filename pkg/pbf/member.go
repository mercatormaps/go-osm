package pbf

import (
	"errors"
	"fmt"

	"github.com/joe-mann/osm/pkg/pbf/OSMPBF"
)

type Member struct {
	Type Type
}

type Members []Member

func DecodeMembers(rel *OSMPBF.Relation, block *OSMPBF.PrimitiveBlock) (*Members, error) {
	memIDs := rel.GetMemids()
	types := rel.GetTypes()

	members := make(Members, len(memIDs))
	var id int64
	for i := range memIDs {
		id += memIDs[i]

		if len(types) < i+1 {
			return nil, fmt.Errorf("unknown type for ID %d", i)
		}

		var typ Type
		switch types[i] {
		case OSMPBF.Relation_NODE:
			typ = TypeNode
		case OSMPBF.Relation_WAY:
			typ = TypeRelation
		case OSMPBF.Relation_RELATION:
			typ = TypeRelation
		default:
			return nil, errors.New("unknown type")
		}

		members[i] = Member{
			Type: typ,
		}
	}

	return nil, nil
}
