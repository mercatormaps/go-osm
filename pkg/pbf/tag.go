package pbf

import (
	"errors"
	"fmt"

	"github.com/joe-mann/osm/pkg/pbf/OSMPBF"
)

type Tag struct {
	Key   string
	Value string
}

type Tags []Tag

type KeyValGetter interface {
	GetKeys() []uint32
	GetVals() []uint32
}

func DecodeTags(g KeyValGetter, block *OSMPBF.PrimitiveBlock) (*Tags, error) {
	var strings [][]byte
	if t := block.GetStringtable(); t != nil {
		strings = t.GetS()
	} else {
		return nil, errors.New("missing string table")
	}

	keyIDs, valIDs := g.GetKeys(), g.GetVals()
	tags := make(Tags, len(keyIDs))
	for i, keyID := range keyIDs {
		if len(strings) < int(keyID+1) {
			return nil, fmt.Errorf("unknown string for key ID %d", keyID)
		}
		key := string(strings[keyID])

		if len(valIDs) < i+1 {
			return nil, fmt.Errorf("unknown value ID %d", i)
		} else if len(strings) < int(valIDs[i]+1) {
			return nil, fmt.Errorf("unknown string for value ID %d", valIDs[i])
		}
		val := string(strings[valIDs[i]])

		tags[i] = Tag{
			Key:   key,
			Value: val,
		}
	}
	return &tags, nil
}

func DecodeDenseTags(nodes *OSMPBF.DenseNodes, block *OSMPBF.PrimitiveBlock, state *tagsState) (*Tags, error) {
	var strings [][]byte
	if t := block.GetStringtable(); t != nil {
		strings = t.GetS()
	} else {
		return nil, errors.New("missing string table")
	}

	kvs := nodes.GetKeysVals()
	tags := Tags{}
	for state.i < len(kvs) {
		if len(kvs) < state.i+1 {
			return nil, fmt.Errorf("unknown key for ID %d", state.i)
		}
		keyID := kvs[state.i]
		state.i++
		if keyID == 0 {
			break
		}

		if len(kvs) < state.i+1 {
			return nil, fmt.Errorf("unknown value for ID %d", state.i)
		}
		valID := kvs[state.i]
		state.i++

		if len(strings) < int(keyID+1) {
			return nil, fmt.Errorf("unknown string for key ID %d", keyID)
		}
		key := string(strings[keyID])

		if len(strings) < int(valID+1) {
			return nil, fmt.Errorf("unknown string for val ID %d", valID)
		}
		val := string(strings[valID])

		tags = append(tags, Tag{
			Key:   key,
			Value: val,
		})
	}
	return &tags, nil
}

type tagsState struct {
	i int
}
