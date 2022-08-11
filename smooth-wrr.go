package p2c

import "fmt"

type SWRRNode struct {
	Weight int
	CurWeight int
}

var SWRRNodes []SWRRNode

func Pick() {
	var pick,totalWeight int
	max := -1
	for index, node := range SWRRNodes {
		totalWeight += node.Weight
		SWRRNodes[index].CurWeight += SWRRNodes[index].Weight
		if SWRRNodes[index].CurWeight > max {
			pick = index
			max = SWRRNodes[index].CurWeight
		}
	}
	SWRRNodes[pick].CurWeight -= totalWeight
	fmt.Println("pick: ", pick)
}

func Debug() {
	for _, node := range SWRRNodes {
		fmt.Println(node.CurWeight)
	}
	fmt.Println("")
}