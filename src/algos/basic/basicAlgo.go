package basic

import (
	"connectedComponents/src/algos"
	"connectedComponents/src/graph"
	"connectedComponents/src/utils/ioEdges"
)

// максимально базовый алгоритм
func basicCCSearch(nodesNum uint32, edges1 []uint32, edges2 []uint32) []uint32 {
	f := make([]uint32, nodesNum)
	for i := 0; i < len(f); i++ {
		f[i] = uint32(i)
	}

	changed := true
	for changed {
		changed = false

		for i := 0; i < len(edges1); i++ {
			u, v := edges1[i], edges2[i]
			pu, pv := f[u], f[v]
			if pu < pv {
				f[v] = pu
				changed = true
			} else if pu > pv {
				f[u] = pv
				changed = true
			}
		}
	}
	return f
}

func BasicCCSearch(g *graph.Graph) map[uint32]uint32 {
	f := basicCCSearch(g.NodesNum, g.Edges1, g.Edges2)
	res := make(map[uint32]uint32)
	for i, x := range f {
		res[uint32(i)] = x
	}
	return res
}

func BasicCCSearchRetArray(g *graph.Graph) []uint32 {
	return basicCCSearch(g.NodesNum, g.Edges1, g.Edges2)
}

func BasicCCSearchFromFile(filename string) map[uint32]uint32 {
	g := ioEdges.LoadGraph(filename)
	return BasicCCSearch(g)

}

func Adapter(conf *algos.RunConfig) map[uint32]uint32 {
	return BasicCCSearchFromFile(conf.TestFile)
}
