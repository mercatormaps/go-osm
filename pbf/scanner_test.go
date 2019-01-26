package pbf_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mercatormaps/go-osm/pbf"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	r, err := os.Open("/home/joe/Downloads/oberbayern-latest.osm.pbf")
	require.NoError(t, err)

	s := pbf.NewScanner(r)

	err = s.Scan()
	require.NoError(t, err)

	printed := false
	n := 0
	for {
		obj := s.Object()
		if obj == nil {
			break
		}
		n++

		if !printed {
			n := obj.(*pbf.Node)
			if len(n.Tags) != 0 {
				fmt.Printf("%+v\n", *n)
				printed = true
			}
		}
	}

	require.NoError(t, s.Err())

	fmt.Println(n, "blobs")
}
