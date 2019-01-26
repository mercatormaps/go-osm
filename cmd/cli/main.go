package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mercatormaps/go-osm/pbf"
	"golang.org/x/text/message"
	"gopkg.in/cheggaaa/pb.v1" 
)

func main() {
	path := "/home/joe/Downloads/oberbayern-latest.osm.pbf"

	stat, err := os.Stat(path)
	exitErr(err)

	r, err := os.Open(path)
	exitErr(err)

	s := pbf.NewScanner(r)

	ctx := context.Background()
	progressDone := make(chan struct{})
	showProgress(ctx, stat.Size(), s, progressDone)

	err = s.Scan()
	exitErr(err)

	numNodes, numWays, numRels := 0, 0, 0
	for {
		obj := s.Object()
		if obj == nil {
			break
		}

		switch obj.(type) {
		case *pbf.Node:
			numNodes++
		case *pbf.Way:
			numWays++
		case *pbf.Relation:
			numRels++
		}
	}

	<-progressDone
	close(progressDone)

	if err := s.Err(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Done!")

	p := message.NewPrinter(message.MatchLanguage("en"))
	p.Printf("%d nodes, %d ways, and %d relations\n", numNodes, numWays, numRels)
}

func showProgress(ctx context.Context, size int64, s *pbf.Scanner, done chan<- struct{}) {
	ticker := time.NewTicker(10 * time.Millisecond)

	bar := pb.New(int(size)).SetUnits(pb.U_BYTES)
	bar.ShowPercent = true
	bar.ShowElapsedTime = true
	bar.ShowTimeLeft = true
	bar.Start()

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			select {
			case <-ticker.C:
				b := s.Bytes()
				bar.Set(int(b))
				if int64(b) == size {
					bar.Finish()
					return
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func exitErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
