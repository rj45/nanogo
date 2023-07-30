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

		for i := 0; i < blk.NumPreds(); i++ {
			pred := blk.Pred(i)
			fmt.Fprintf(dot, "%s -> %s;\n", pred, blk)
		}

		liveInKills := ""
		label := fmt.Sprintf("%s:\\l", blk)
		for i := 0; i < blk.NumInstrs(); i++ {
			instr := blk.Instr(i)

			label += fmt.Sprintf("%s\\l", instr.LongString())
		}

		label = strings.ReplaceAll(label, "\"", "\\\"")
		label = strings.ReplaceAll(label, "{", "\\{")
		label = strings.ReplaceAll(label, "}", "\\}")

		fmt.Fprintf(dot, "%s [label=\"%s\"];\n", blk, label)
		fmt.Fprintln(dot, liveInKills)
	}

	fmt.Fprintln(dot, "}")
}

// WriteGraphvizInterferenceGraph emits a Graphviz dot file of the Interference Graph
func WriteGraphvizInterferenceGraph(ra *RegAlloc) {
	dot, _ := os.Create(ra.fn.Name + ".igraph.dot")
	defer dot.Close()

	fmt.Fprintln(dot, "graph G {")

	edges := map[string]bool{}

	for _, nodeA := range ra.iGraph.nodes {
		label := fmt.Sprintf("%s\n%d:%d", nodeA.val.IDString(), nodeA.order, nodeA.colour)
		fmt.Fprintf(dot, "%s [label=%q];\n", nodeA.val.IDString(), label)

		for nodeBid := range nodeA.interferes {
			nodeB := ra.iGraph.nodes[nodeBid]
			if !edges[nodeB.val.IDString()+"--"+nodeA.val.IDString()] {
				edge := nodeA.val.IDString() + "--" + nodeB.val.IDString()
				fmt.Fprintf(dot, "%s;\n", edge)
				edges[edge] = true
			}
		}

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
	num := uint32(0)

	for ; it.HasNext(); it.Next() {
		newblock := false
		if it.Block() != blk {
			if blk != nil {
				fmt.Fprintf(html, "<tr><td colspan=\"%d\">out:\n", 4)
				for id := range ra.info[blk.Index()].liveOuts {
					fmt.Fprint(html, " ", id.ValueIn(ra.fn).String())
				}
				fmt.Fprintln(html, "</td></tr>")
			}
			blk = it.Block()
			fmt.Fprintf(html, "<tr><td colspan=\"%d\">in:\n", 4)
			for id := range ra.info[blk.Index()].liveIns {
				fmt.Fprint(html, " ", id.ValueIn(ra.fn).String())
			}
			fmt.Fprintln(html, "</td></tr>")

			fmt.Fprintf(html, "<tr><td colspan=\"%d\">%s(", 4, blk)
			for i, val := range blk.Defs() {
				if i != 0 {
					fmt.Fprint(html, ", ")
				}
				fmt.Fprintf(html, "%s", val)
			}
			fmt.Fprintln(html, "):</td></tr>")

			blk.Args()

			newblock = true
		}

		fmt.Fprintln(html, "<tr>")

		if newblock {
			fmt.Fprintln(html, "<td rowspan=\"", blk.NumInstrs()*2, "\">", blk.String(), "</td>")
		}
		fmt.Fprintln(html, "<td>", num, "</td>")

		fmt.Fprintln(html, "<td rowspan=\"2\">")
		fmt.Fprintln(html, it.Instr().LongString())

		fmt.Fprintln(html, "</td>")
		fmt.Fprintln(html, "</tr>")

		num++

		fmt.Fprintln(html, "<tr><td>", num, "</td>")
		fmt.Fprintln(html, "</tr>")
		num++
	}

	fmt.Fprintln(html, "</table>")
	fmt.Fprintln(html, "</body>")
	fmt.Fprintln(html, "</html>")
}

func WriteGraphvizLivenessGraph(ra *RegAlloc) {
	fn := ra.fn

	dot, _ := os.Create(ra.fn.Name + ".vfg.dot")
	defer dot.Close()

	fmt.Fprintln(dot, "digraph G {")
	fmt.Fprintln(dot, "node [fontname=\"Noto Mono\", shape=rect];")

	for b := 0; b < fn.NumBlocks(); b++ {
		blk := fn.Block(b)
		info := &ra.info[blk.Index()]

		fmt.Fprintf(dot, "subgraph cluster_%s {\n", blk.String())

		fmt.Fprintf(dot, "label=\"%s\";\n", blk.String())
		fmt.Fprintln(dot, "labeljust=l;")
		fmt.Fprintln(dot, "color=black;")

		fmt.Fprintln(dot, "node [shape=plaintext];")

		srcs := make(map[ir2.ID]string)
		var lastStr string

		inname := fmt.Sprintf("in_%s", blk)
		ins := maptorecordsrc(fn, inname, info.liveIns, srcs)

		for d := 0; d < blk.NumDefs(); d++ {
			def := blk.Def(d)
			ins += fmt.Sprintf("<td port=\"%s\">%s:%s</td>", def.IDString(), def.IDString(), def.Reg().String())
			srcs[def.ID] = fmt.Sprintf("%s:%s", inname, def.IDString())
		}

		fmt.Fprintf(dot, "%s [label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\"><tr><td port=\"in\">%s in</td>%s</tr></table>>];\n", inname, blk, ins)

		for i := 0; i < blk.NumPreds(); i++ {
			pred := blk.Pred(i)
			fmt.Fprintf(dot, "out_%s_%s:out:s -> %s:in:n;\n", pred, blk, inname)
			lastStr = inname
		}

		if blk.NumInstrs() > 0 && lastStr != "" {
			// force the instructions to be in order
			fmt.Fprintf(dot, "%s -> %s_%s [weight=100, style=invis];\n", lastStr, blk, blk.Instr(0).IDString())
		}

		for i := 0; i < blk.NumInstrs(); i++ {
			instr := blk.Instr(i)

			edges := ""
			name := fmt.Sprintf("%s_%s", blk, instr.IDString())
			lastStr = name

			label := ""
			for d := 0; d < instr.NumDefs(); d++ {
				def := instr.Def(d)
				if def.NeedsReg() {
					label += fmt.Sprintf("<td port=\"%s\">%s</td>", def.IDString(), valname(def))
				}
				srcs[def.ID] = fmt.Sprintf("%s:%s", name, def.IDString())
			}

			label += fmt.Sprintf("<td>%s</td>", instr.Op)
			for j := 0; j < instr.NumArgs(); j++ {
				arg := instr.Arg(j)

				// killed := false
				// for _, kill := range info.kills[instr] {
				// 	if kill == arg {
				// 		killed = true
				// 		break
				// 	}
				// }

				arglabel := valname(arg)
				// if killed {
				// 	arglabel = "[" + arglabel + "]"
				// }

				label += fmt.Sprintf("<td port=\"%s\">%s</td>", arg.IDString(), arglabel)

				if arg.NeedsReg() {
					edges += fmt.Sprintf("%s:s -> %s:%s:n;\n", srcs[arg.ID], name, arg.IDString())
				}

				// chain arrows through uses
				srcs[arg.ID] = fmt.Sprintf("%s:%s", name, arg.IDString())
			}

			fmt.Fprintf(dot, "%s [label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\"><tr>%s</tr></table>>];\n", name, label)

			fmt.Fprint(dot, edges)

			if i < blk.NumInstrs()-1 {
				// force the instructions to be in order
				fmt.Fprintf(dot, "%s -> %s_%s [weight=100, style=invis];\n", name, blk, blk.Instr(i+1).IDString())
			}
		}

		// emit block control instruction
		// {
		// 	name := fmt.Sprintf("%s_ctrl", blk)
		// 	label := ""

		// 	var edges string
		// 	label += fmt.Sprintf("<td>%s</td>", blk.Op)

		// 	for i := 0; i < blk.NumControls(); i++ {
		// 		arg := blk.Control(i)

		// 		arglabel := valname(arg)
		// 		if info.blkKills[arg] {
		// 			arglabel = "[" + arglabel + "]"
		// 		}

		// 		label += fmt.Sprintf("<td port=\"%s\">%s</td>", arg.IDString(), arglabel)

		// 		if arg.NeedsReg() {
		// 			edges += fmt.Sprintf("%s:s -> %s:%s:n;\n", srcs[arg], name, arg.IDString())
		// 		}

		// 		// chain arrows through uses
		// 		srcs[arg] = fmt.Sprintf("%s:%s", name, arg.IDString())
		// 	}

		// 	fmt.Fprintf(dot, "%s [label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\"><tr>%s</tr></table>>];\n", name, label)
		// 	fmt.Fprint(dot, edges)

		// 	if blk.NumInstrs() > 0 {
		// 		// force the instructions to be in order
		// 		fmt.Fprintf(dot, "%s -> %s [weight=100, style=invis];\n", lastStr, name)
		// 	}
		// 	lastStr = name
		// }

		for i := 0; i < blk.NumSuccs(); i++ {
			succ := blk.Succ(i)
			sinfo := &ra.info[succ.Index()]

			outs := maptorecord(fn, sinfo.liveIns)
			// if len(info.phiOuts[succ]) > 0 {
			// 	outs += maptorecord(info.phiOuts[succ])
			// }

			fmt.Fprintf(dot, "out_%s_%s [label=<<table border=\"0\" cellborder=\"1\" cellspacing=\"0\"><tr><td port=\"out\">%s out</td>%s</tr></table>>];\n", blk, succ, blk, outs)
			for v := range sinfo.liveIns {
				fmt.Fprintf(dot, "%s -> out_%s_%s:%s;\n", srcs[v], blk, succ, v.IDString())
			}
			// for v := range info.phiOuts[succ] {
			// 	fmt.Fprintf(dot, "%s -> out_%s_%s:%s;\n", srcs[v], blk, succ, v.IDString())
			// }

			// force the instructions to be in order
			fmt.Fprintf(dot, "%s -> out_%s_%s [weight=100, style=invis];\n", lastStr, blk, succ)
		}

		fmt.Fprintln(dot, "}")
	}

	fmt.Fprintln(dot, "}")
}

func valname(val *ir2.Value) string {
	if val.NeedsReg() {
		return fmt.Sprintf("%s:%s", val.IDString(), val.Reg())
	}
	return val.String()
}

func maptorecord(fn *ir2.Func, l map[ir2.ID]struct{}) string {
	ret := ""
	for v := range l {
		val := v.ValueIn(fn)
		ret += fmt.Sprintf("<td port=\"%s\">%s:%s</td>", val.IDString(), val.IDString(), val.Reg().String())
	}
	return ret
}

func maptorecordsrc(fn *ir2.Func, prefix string, l map[ir2.ID]struct{}, src map[ir2.ID]string) string {
	ret := ""
	for v := range l {
		val := v.ValueIn(fn)
		ret += fmt.Sprintf("<td port=\"%s\">%s:%s</td>", val.IDString(), val.IDString(), val.Reg().String())
		src[v] = fmt.Sprintf("%s:%s", prefix, val.IDString())
	}
	return ret
}
