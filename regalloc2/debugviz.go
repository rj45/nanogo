package regalloc2

import (
	"fmt"
	"os"
	"strings"

	"github.com/rj45/nanogo/ir2"
)

// WriteGraphvizCFG emits a Graphviz dot file of the CFG
func WriteGraphvizCFG(ra *RegAlloc) {
	dot, _ := os.Create(ra.fn.Name + ".cfg.dot")
	defer dot.Close()

	fmt.Fprintln(dot, "digraph G {")
	fmt.Fprintln(dot, "labeljust=l;")
	fmt.Fprintln(dot, "node [shape=record, fontname=\"Noto Mono\", labeljust=l];")

	for bn := 0; bn < ra.fn.NumBlocks(); bn++ {
		blk := ra.fn.Block(bn)

		// info := &ra.blockInfo[blk.ID()]

		for i := 0; i < blk.NumPreds(); i++ {
			pred := blk.Pred(i)
			// pinfo := &ra.blockInfo[pred.ID()]
			// outs := maptolist(pinfo.liveOuts) + " - " + maptolist(pinfo.phiOuts[blk])
			// ins := maptolist(info.liveIns) + " - " + maptolist(info.phiIns[pred])
			fmt.Fprintf(dot, "%s -> %s;\n", pred, blk)
		}

		liveInKills := ""
		label := fmt.Sprintf("%s:\\l", blk)
		for i := 0; i < blk.NumInstrs(); i++ {
			instr := blk.Instr(i)

			kills := ""

			// for i, kill := range info.kills[instr] {
			// 	if i != 0 {
			// 		kills += " "
			// 	}
			// 	kills += kill.IDString()
			// }

			label += fmt.Sprintf("%s [%s]\\l", instr.LongString(), kills)
		}

		// first := true
		// kills := ""
		// for kill := range info.blkKills {
		// 	if !first {
		// 		kills += " "
		// 	}
		// 	first = false
		// 	kills += kill.IDString()
		// }

		// label += fmt.Sprintf("%s [%s]\\l" /* blk.OpString()*/, "", kills)

		label = strings.ReplaceAll(label, "\"", "\\\"")
		label = strings.ReplaceAll(label, "{", "\\{")
		label = strings.ReplaceAll(label, "}", "\\}")

		fmt.Fprintf(dot, "%s [label=\"%s\"];\n", blk, label)
		fmt.Fprintln(dot, liveInKills)
	}

	fmt.Fprintln(dot, "}")
}

func DumpLivenessChart(ra *RegAlloc) {
	html, _ := os.Create(ra.fn.Name + ".liveness.html")
	defer html.Close()

	fmt.Fprintln(html, `
<html>
<head>
	<style>
		td {
			border: 1px solid black;
			font-family: monospace;
			padding: 1px 5px;
		}
	</style>
</head>
<body>
	<table>`)

	it := ra.fn.InstrIter()
	var blk *ir2.Block

	num := 0
	for ; it.HasNext(); it.Next() {
		fmt.Fprintln(html, "<tr>")
		if it.Block() != blk {
			if blk != nil {
				fmt.Fprintln(html, "<tr><td colspan=\"3\">out:")
				for id := range ra.info[blk.Index()].liveOuts {
					fmt.Fprint(html, " ", id.ValueIn(ra.fn).String())
				}
				fmt.Fprintln(html, "</td></tr>")
			}
			blk = it.Block()
			fmt.Fprintln(html, "<tr><td colspan=\"3\">in:")
			for id := range ra.info[blk.Index()].liveIns {
				fmt.Fprint(html, " ", id.ValueIn(ra.fn).String())
			}
			fmt.Fprintln(html, "</td></tr>")

			fmt.Fprintln(html, "<td rowspan=\"", blk.NumInstrs()*2, "\">", blk.String(), "</td>")

		}
		fmt.Fprintln(html, "<td>", num, "</td><td rowspan=\"2\">", it.Instr().LongString(), "</td></tr>")
		num++
		fmt.Fprintln(html, "<tr><td>", num, "</td></tr>")
		num++
	}

	fmt.Fprintln(html, "</table>")
	fmt.Fprintln(html, "</body>")
	fmt.Fprintln(html, "</html>")
}
