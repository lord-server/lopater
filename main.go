package main

import (
	"context"
	"flag"
	"log"
	"sort"
	"sync"
	"sync/atomic"
)

type stats struct {
	NodeCount map[string]uint64
}

func newStats() stats {
	return stats{
		NodeCount: make(map[string]uint64),
	}
}

func (s *stats) combine(stat stats) stats {
	nodeCount := make(map[string]uint64)

	for k, v := range s.NodeCount {
		nodeCount[k] += v
	}

	for k, v := range stat.NodeCount {
		nodeCount[k] += v
	}

	return stats{
		NodeCount: nodeCount,
	}
}

func (s *stats) ProcessBlock(block *Block) {
	for i := 0; i < 16*16*16*2; i += 2 {
		hi := uint16(block.NodeData[i])
		lo := uint16(block.NodeData[i+1])

		id := hi<<8 | lo

		name, ok := block.Mapping[id]
		if !ok {
			return
		}

		s.NodeCount[name]++
	}
}

func (s *stats) Report() {
	type nodeCountPair struct {
		name  string
		count uint64
	}

	var mostFrequentNodes []nodeCountPair
	for name, count := range s.NodeCount {
		mostFrequentNodes = append(mostFrequentNodes, nodeCountPair{
			name:  name,
			count: count,
		})
	}

	sort.Slice(mostFrequentNodes, func(i, j int) bool {
		return mostFrequentNodes[i].count > mostFrequentNodes[j].count
	})

	for _, pair := range mostFrequentNodes {
		log.Printf("%v = %v", pair.name, pair.count)
	}
}

func countNodes(ctx context.Context, w *World, threads int) stats {
	type blockData struct {
		x, y, z int
		data    []byte
	}

	var wg sync.WaitGroup

	blockDataChan := make(chan blockData, 10000)
	statsChan := make(chan stats, threads)

	var blockCount atomic.Uint64

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			stats := newStats()

			for data := range blockDataChan {
				block, err := DecodeBlock(data.data)
				if err != nil {
					panic(err)
				}

				cnt := blockCount.Load()
				if cnt%10000 == 0 {
					log.Printf("cnt=%v", cnt)
				}

				blockCount.Add(1)

				stats.ProcessBlock(block)
			}

			statsChan <- stats

			wg.Done()
		}()
	}

	err := w.storage.GetBlocksData(ctx, func(x, y, z int, data []byte) error {
		blockDataChan <- blockData{
			x:    x,
			y:    y,
			z:    z,
			data: data,
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	close(blockDataChan)

	wg.Wait()

	stats := newStats()

	close(statsChan)

	for stat := range statsChan {
		stats = stats.combine(stat)
	}

	log.Printf("processed %v blocks", blockCount.Load())

	return stats
}

var (
	threadCount = flag.Int("threads", 2, "number of data processing threads")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	if flag.NArg() < 1 {
		log.Fatal("usage: lopater <path/to/world>")
	}

	worldPath := flag.Arg(0)
	log.Printf("world path: %v", worldPath)

	world, err := OpenWorld(ctx, worldPath)
	if err != nil {
		log.Fatalf("failed to open world: %v", err)
	}

	stats := countNodes(ctx, world, *threadCount)

	stats.Report()
}
