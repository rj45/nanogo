package regalloc2

import "github.com/rj45/nanogo/ir2"

type iNodeID uint32

type iGraph struct {
	nodes   []iNode
	valNode map[ir2.ID]iNodeID

	maxColour uint16
}

type iNode struct {
	val           ir2.ID
	intervals     []interval
	interferences []iNodeID
	interferes    map[iNodeID]struct{}
	moves         []iNodeID

	card   uint16
	colour uint16
}

const noColour uint16 = 0

func (nd *iNode) pickColour(ig *iGraph) {
	// first try to pick a move colour if that colour doesn't
	// interfere with any others
	for _, mv := range nd.moves {
		moveColour := ig.nodes[mv].colour
		// if the move node has already been assigned a colour
		if moveColour > noColour {
			// check if that colour interferes with any neighbors
			interferes := false
			for _, nb := range nd.interferences {
				if ig.nodes[nb].colour == moveColour {
					interferes = true
					break
				}
			}

			// if it doesn't interfere, then choose that colour
			if !interferes {
				nd.colour = moveColour
				return
			}
		}
	}

	// find the lowest numbered colour that doesn't interfere
	for colour := uint16(1); ; colour++ {
		interferes := false

		// for each neighbor in the interferences
		for _, nb := range nd.interferences {
			// if the neighbor already has this colour
			if ig.nodes[nb].colour == colour {
				// then it interferes and we can't use it
				interferes = true
				break
			}
		}

		// if it doesn't interfere then
		if !interferes {
			// choose the colour
			nd.colour = colour

			// keep track of the largest chosen colour
			if ig.maxColour < colour {
				ig.maxColour = colour
			}

			return
		}
	}
}

func (ra *RegAlloc) buildInterferenceGraph() {
	ig := &ra.iGraph

	ig.nodes = nil

	// initialize the list of nodes
	for _, rng := range ra.liveRanges {
		if _, found := ig.valNode[rng.val]; !found {
			ig.nodes = append(ig.nodes, iNode{
				val: rng.val,
			})

			// storing the index since a pointer would change
			// as we append to the above array and the array
			// gets reallocated
			ig.valNode[rng.val] = iNodeID(len(ig.nodes) - 1)
		}

		node := &ig.nodes[ig.valNode[rng.val]]
		node.intervals = append(node.intervals, rng)
	}

	// for each pair of live ranges
	for i, rngI := range ra.liveRanges {
		for j, rngJ := range ra.liveRanges {
			if i == j {
				continue
			}

			// check if the live ranges overlap
			if rngI.start < rngJ.end && rngI.end > rngJ.start {
				idI := ig.valNode[rngI.val]
				nodeI := &ig.nodes[idI]

				idJ := ig.valNode[rngJ.val]
				nodeJ := &ig.nodes[idJ]

				// if the node hasn't been seen in a previous live range
				if _, found := nodeI.interferes[idJ]; !found {
					// add it to the interferences list
					nodeI.interferences = append(nodeI.interferences, idJ)

					// and the interferes map
					nodeI.interferes[idJ] = struct{}{}
				}

				// and also with nodeJ
				if _, found := nodeJ.interferes[idI]; !found {
					// add it to the interferences list
					nodeJ.interferences = append(nodeJ.interferences, idI)

					// and the interferes map
					nodeJ.interferes[idI] = struct{}{}
				}
			}
		}
	}

}
