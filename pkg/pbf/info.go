package pbf

import (
	"errors"
	"fmt"
	"time"

	"github.com/joe-mann/osm/pkg/pbf/OSMPBF"
)

type Info struct {
	Version   int32
	Visible   bool
	User      string
	UserID    UserID
	Timestamp time.Time
}

type InfoGetter interface {
	GetInfo() *OSMPBF.Info
}

func DecodeInfo(g InfoGetter, block *OSMPBF.PrimitiveBlock) (*Info, error) {
	info := g.GetInfo()
	if info == nil {
		return &Info{}, nil
	}

	var strings [][]byte
	if t := block.GetStringtable(); t != nil {
		strings = t.GetS()
	} else {
		return nil, errors.New("missing string table")
	}

	sid := info.GetUserSid()
	if len(strings) < int(sid+1) {
		return nil, fmt.Errorf("unknwon string for SID %d", sid)
	}
	user := string(strings[sid])

	d := time.Duration(info.GetTimestamp()*int64(block.GetGranularity())) * time.Millisecond
	return &Info{
		Version:   info.GetVersion(),
		Visible:   info.GetVisible(),
		User:      user,
		UserID:    UserID(info.GetUid()),
		Timestamp: time.Unix(0, d.Nanoseconds()).UTC(),
	}, nil
}

func DecodeDenseInfo(i int, nodes *OSMPBF.DenseNodes, block *OSMPBF.PrimitiveBlock, state *infoState) (*Info, error) {
	info := nodes.GetDenseinfo()
	if info == nil {
		return &Info{}, nil
	}

	var strings [][]byte
	if t := block.GetStringtable(); t != nil {
		strings = t.GetS()
	} else {
		return nil, errors.New("missing string table")
	}

	version, err := version(i, info)
	if err != nil {
		return nil, err
	}

	visible, err := visible(i, info)
	if err != nil {
		return nil, err
	}

	sid, err := userSID(i, info, state)
	if err != nil {
		return nil, err
	}
	if len(strings) < int(sid+1) {
		return nil, fmt.Errorf("unknwon string for SID %d", sid)
	}
	user := string(strings[sid])

	uid, err := uid(i, info, state)
	if err != nil {
		return nil, err
	}

	timestamp, err := timestamp(i, info, block, state)
	if err != nil {
		return nil, err
	}

	return &Info{
		Version:   version,
		Visible:   visible,
		User:      user,
		UserID:    UserID(uid),
		Timestamp: timestamp,
	}, nil
}

func version(i int, info *OSMPBF.DenseInfo) (int32, error) {
	versions := info.GetVersion()
	if len(versions) == 0 {
		return 0, nil
	}

	if len(versions) < i+1 {
		return 0, fmt.Errorf("unknown version for index %d", i)
	}
	return versions[i], nil
}

func visible(i int, info *OSMPBF.DenseInfo) (bool, error) {
	visibles := info.GetVisible()
	if len(visibles) == 0 {
		return true, nil
	}

	if len(visibles) < i+1 {
		return false, fmt.Errorf("unknown visible for index %d", i)
	}
	return visibles[i], nil
}

func userSID(i int, info *OSMPBF.DenseInfo, state *infoState) (int32, error) {
	sids := info.GetUserSid()
	if len(sids) == 0 {
		return 0, nil
	}

	if len(sids) < i+1 {
		return 0, fmt.Errorf("unknown user SID for index %d", i)
	}

	state.userSID += sids[i]
	return state.userSID, nil
}

func uid(i int, info *OSMPBF.DenseInfo, state *infoState) (int32, error) {
	uids := info.GetUid()
	if len(uids) == 0 {
		return 0, nil
	}

	if len(uids) < i+1 {
		return 0, fmt.Errorf("unknown UID for index %d", i)
	}

	state.uid += uids[i]
	return state.uid, nil
}

func timestamp(i int, info *OSMPBF.DenseInfo, block *OSMPBF.PrimitiveBlock, state *infoState) (time.Time, error) {
	stamps := info.GetTimestamp()
	if len(stamps) == 0 {
		return time.Time{}, nil
	}

	if len(stamps) < i+1 {
		return time.Time{}, fmt.Errorf("unknown timestamp for index %d", i)
	}

	state.timestamp += stamps[i]
	d := time.Duration(state.timestamp*int64(block.GetGranularity())) * time.Millisecond
	return time.Unix(0, d.Nanoseconds()).UTC(), nil
}

type infoState struct {
	userSID   int32
	uid       int32
	timestamp int64
}
