package distribution

import (
	"connectedComponents/src/algos/basic"
	"connectedComponents/src/graph"
	"connectedComponents/src/utils"
	"fmt"
	"math"
	"sort"
)

func FindDistribution(iterator *graph.GraphIterator, numSlaves uint32, hashNum uint32) []uint32 {
	var dist Distributor
	if iterator.G.NodesNum < hashNum {
		hashNum = iterator.G.NodesNum + 1
	}
	dist.init(iterator, numSlaves, hashNum)

	multiComponents := dist.findMulticomponents()

	slavesHashes := dist.balanceHashes(multiComponents)

	hashToSlave := dist.makeMapHashToSlave(slavesHashes)
	// это во время выполнения реальной задачи не нужно:
	dist.analyseDistrib(iterator, hashToSlave)

	return hashToSlave
}

func (dist *Distributor) init(iterator *graph.GraphIterator, numSlaves uint32, hashNum uint32) {
	iterator.StartIter()
	// инициализируем объект распределителя
	dist.multiEdges = make(map[multiEdge]struct{})
	dist.hashNum = hashNum
	dist.nodesWeight = make([]uint32, hashNum)
	dist.numSlaves = numSlaves
	dist.nodesNum = iterator.NodesNum()

	fmt.Println()
	fmt.Printf("-> {dist}: gonna distribute %d nodes around %d slaves\n", iterator.NodesNum(), numSlaves)
	fmt.Printf("-> {dist}: create %d hashes, each holds for %f nodes\n", hashNum, float32(iterator.NodesNum())/float32(hashNum))
	fmt.Printf("-> {dist}: %d hashes per slave\n", uint32(float32(dist.hashNum+1)/float32(dist.numSlaves)))
	// задаем мультиграф
	for iterator.HasEdges() {
		dist.addEdge(iterator.GetNextEdge())
	}
}

func (dist *Distributor) findMulticomponents() [][]uint32 {
	// находим связные компоненты мультиграфа
	multiG := dist.toGraph()
	starForest := basic.BasicCCSearchRetArray(multiG)
	multiComponents := utils.StarForestToComponents(starForest)
	// fmt.Println(multiComponesnts)
	fmt.Printf("-> {dist}: detected %v multinodes and %v multiedges\n", float32(dist.hashNum), float32(len(dist.multiEdges)))
	fmt.Printf("-> {dist}: found %v multicomponents\n", len(multiComponents))

	// находим веса связных компонент
	componentsWeight := make([]uint32, len(multiComponents))
	k := uint32(dist.nodesNum / dist.hashNum)
	for i := range multiComponents {
		for _, multinodeNum := range multiComponents[i] {
			componentsWeight[i] += k
			if multinodeNum < dist.nodesNum%dist.hashNum {
				componentsWeight[i]++
			}
		}
	}
	// и сортируем компоненты по весам
	sort.Slice(multiComponents, func(i, j int) bool {
		if componentsWeight[i] > componentsWeight[j] {
			componentsWeight[i], componentsWeight[j] = componentsWeight[j], componentsWeight[i]
			return true
		}
		return false
	})
	fmt.Print("-> {dist}: 10 heaviest components: ")
	for i := 0; i < len(componentsWeight); i++ {
		fmt.Print(componentsWeight[i], " ")
		if i == 10 {
			break
		}
	}
	fmt.Println()

	return multiComponents
}

func (dist *Distributor) balanceHashes(multiComponents [][]uint32) [][]uint32 {
	direction := 1
	i := uint32(0)
	nextI := func() {
		i += uint32(direction)
		if i == dist.numSlaves && direction == 1 {
			i--
			direction = -1
		} else if i > 2000000 && direction == -1 {
			i = 0
			direction = 1
		}
	}

	// создае массив распределений хешей по слейвам
	slavesHashes := make([][]uint32, dist.numSlaves)
	for j := uint32(0); j < uint32(dist.numSlaves); j++ {
		slavesHashes[j] = make([]uint32, 0)
	}

	hashPerSlave := uint32(float32(dist.hashNum)/float32(dist.numSlaves)) + 1

	// "змейкой" распределяем мультикомпоненты по слейвам
	for _, mc := range multiComponents {
		// fmt.Println("mc", mcNum, "has", len(mc), "hashes")
		for len(mc) > 0 {
			min := int(math.Min(
				float64(hashPerSlave)*1.2-float64(len(slavesHashes[i])),
				float64(len(mc))))
			slavesHashes[i] = append(slavesHashes[i], mc[:min]...)
			mc = mc[min:]
			nextI()
		}
	}

	return slavesHashes
}

func (dist *Distributor) makeMapHashToSlave(slavesHashes [][]uint32) []uint32 {
	// возвращаем массив ret[h] = b, где h - хэш, b - номер слейва
	hashToSlave := make([]uint32, dist.hashNum)
	for slaveNum, hashes := range slavesHashes {
		for _, hash := range hashes {
			hashToSlave[hash] = uint32(slaveNum)
		}
	}
	return hashToSlave
}

func (dist *Distributor) analyseDistrib(iterator *graph.GraphIterator, hashToSlave []uint32) {
	connectivityMatrix := make([][]uint32, dist.numSlaves)
	for i := range connectivityMatrix {
		connectivityMatrix[i] = make([]uint32, dist.numSlaves)
	}

	iterator.StartIter()
	for iterator.HasEdges() {
		v1, v2 := iterator.GetNextEdge()
		h1, h2 := H(v1, dist.hashNum), H(v2, dist.hashNum)
		s1, s2 := hashToSlave[h1], hashToSlave[h2]
		connectivityMatrix[s1][s2] += 1
		if s1 != s2 {
			connectivityMatrix[s2][s1] += 1
		}
	}

	// fmt.Println("-> {dist}: матрица A \"смежности\" хешей:")
	// fmt.Println("-> {dist}: A[i,j] = n - значит, что существует n ребер (v1,v2): h(v1) = i & h(v2) = j")
	// for _, row := range connectivityMatrix {
	// 	fmt.Print("    ")
	// 	for _, elem := range row {
	// 		fmt.Printf("%05d ", elem)
	// 	}
	// 	fmt.Println()
	// }
}
