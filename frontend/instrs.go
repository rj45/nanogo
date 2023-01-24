package frontend

import (
	"go/constant"
	"go/token"
	"go/types"
	"log"

	"github.com/rj45/nanogo/ir2"
	"github.com/rj45/nanogo/ir2/op"
	"golang.org/x/tools/go/ssa"
)

func (fe *FrontEnd) translateInstrs(irBlock *ir2.Block, ssaBlock *ssa.BasicBlock) {
	for _, instr := range ssaBlock.Instrs {
		var store *ssa.Store
		_ = store

		var opcode ir2.Op
		var typ types.Type
		var con ir2.Const
		var arg *ir2.Value

		// ops = instr.Operands(ops[:0])
		switch ins := instr.(type) {
		case *ssa.DebugRef:
		case *ssa.If:
			opcode = op.If
		case *ssa.Jump:
			opcode = op.Jump
		case *ssa.Return:
			opcode = op.Return
		case *ssa.Panic:
			opcode = op.Panic
		case *ssa.Phi:
			fe.translateBlockParams(irBlock, ins)
		case *ssa.Store:
			typ = ins.Val.Type()
			opcode = op.Store
			store = ins
		case *ssa.Alloc:
			con = ir2.ConstFor(ins.Comment)
			if ins.Heap {
				opcode = op.New
			} else {
				opcode = op.Local
			}
		case *ssa.Slice:
			opcode = op.Slice
		case *ssa.Call:
			opcode = op.Call
			switch call := ins.Call.Value.(type) {
			case *ssa.Function:
				retType := call.Signature.Results()
				typ = retType
				if retType.Len() == 1 {
					typ = retType.At(0).Type()
				}

			case *ssa.Builtin:
				opcode = op.CallBuiltin
				retType := call.Type().(*types.Signature).Results()
				typ = retType
				if retType.Len() == 1 {
					typ = retType.At(0).Type()
				}
				// name := genName("builtin", call.Name())
				// builtin := irBlock.Func().Package().LookupFunc(name)
				// if builtin != nil {
				// 	builtin.Referenced = true
				// }
				// con = constant.MakeString(name)
				// typ = call.Type()
			default:
				log.Fatalf("unsupported call type: %#v", ins.Call.Value)
			}

		case *ssa.Convert:
			opcode = op.Convert
		case *ssa.MakeInterface:
			opcode = op.MakeInterface
		case *ssa.IndexAddr:
			opcode = op.IndexAddr
		case *ssa.FieldAddr:
			opcode = op.FieldAddr
			con = ir2.ConstFor(ins.Field)
		case *ssa.Range:
			opcode = op.Range
		case *ssa.Next:
			opcode = op.Next
		case *ssa.Extract:
			// extract is used to pull values from tuples,
			// the IR doesn't have tuples, instead it has multiple defs (results)
			// so this instruction is redundant. We just wire the result of the
			// extract up to the proper def
			mulret := fe.val2instr[ins.Tuple]
			fe.val2val[ins] = mulret.Def(ins.Index)

		case *ssa.Lookup:
			opcode = op.Lookup
			if ins.CommaOk {
				// these should be a separate instruction as they have
				// different semantics
				log.Fatal("Comma lookups not yet implented")
			}
		case *ssa.BinOp:
			switch ins.Op {
			case token.ADD:
				opcode = op.Add
			case token.SUB:
				opcode = op.Sub
			case token.MUL:
				opcode = op.Mul
			case token.QUO:
				opcode = op.Div
			case token.REM:
				opcode = op.Rem
			case token.AND:
				opcode = op.And
			case token.OR:
				opcode = op.Or
			case token.XOR:
				opcode = op.Xor
			case token.SHL:
				opcode = op.ShiftLeft
			case token.SHR:
				opcode = op.ShiftRight
			case token.AND_NOT:
				opcode = op.AndNot
			case token.EQL:
				opcode = op.Equal
			case token.NEQ:
				opcode = op.NotEqual
			case token.LSS:
				opcode = op.Less
			case token.LEQ:
				opcode = op.LessEqual
			case token.GTR:
				opcode = op.Greater
			case token.GEQ:
				opcode = op.GreaterEqual
			default:
				log.Fatalf("unsupported binop: %#v", ins)
			}
		case *ssa.UnOp:
			switch ins.Op {
			case token.NOT:
				opcode = op.Not
			case token.SUB:
				opcode = op.Negate
			case token.MUL:
				opcode = op.Load
			case token.XOR:
				opcode = op.Invert
			default:
				log.Fatalf("unsupported unop: %#v", ins)
			}

		case *ssa.RunDefers:
			// ignore
		default:
			log.Fatalf("unknown instruction type %#v", instr)
		}

		if opcode == nil {
			// skip this instruction
			continue
		}

		if typ == nil {
			if typed, ok := instr.(interface{ Type() types.Type }); ok {
				typ = typed.Type()
			}
		}

		ins := irBlock.Func().NewInstr(opcode, typ)
		if con != nil {
			ins.InsertArg(-1, irBlock.Func().ValueFor(typ, con))
		}

		ins.Pos = getPos(instr)

		if arg != nil {
			ins.InsertArg(-1, arg)
		}

		fe.translateArgs(irBlock, ins, instr)

		if ins == nil {
			log.Panicf("ins is nil! %s", instr)
		}

		irBlock.InsertInstr(-1, ins)

		if vin, ok := instr.(ssa.Value); ok {
			fe.val2instr[vin] = ins

			if ins.NumDefs() == 1 {
				fe.val2val[vin] = ins.Def(0)
			}
		}
	}

	fe.translateBlockArgs(irBlock, ssaBlock)
}

func (fe *FrontEnd) translateArgs(block *ir2.Block, irInstr *ir2.Instr, ssaInstr ssa.Instruction) {
	var valarr [10]*ssa.Value
	vals := ssaInstr.Operands(valarr[:0])

	for _, val := range vals {
		if val == nil {
			continue
		}
		var arg interface{}
		var ok bool
		arg, ok = fe.val2val[*val]
		if !ok {
			arg, ok = fe.val2instr[*val]
		}
		if !ok {
			ok = true
			switch con := (*val).(type) {
			case *ssa.Const:
				if con.Type().Underlying().String() == "string" {
					str := constant.StringVal(con.Value)
					fn := block.Func()
					glob := fn.Package().NewStringLiteral(fn.Name, str)
					glob.Referenced = true
					arg = glob
				} else {
					arg = con.Value
				}

			case *ssa.Function:
				pkg := block.Func().Package()
				pkg = pkg.Program().Package(con.Pkg.Pkg.Path())
				otherFunc := pkg.Func(con.Name())
				if otherFunc == nil {
					log.Fatalf("reference to unknown function %s in function %s", con.Name(), block.Func().Name)
				}
				// ensure it gets loaded
				otherFunc.Referenced = true
				arg = otherFunc
				block.Func().NumCalls++

			case *ssa.Builtin:
				arg = con.Name()

			case *ssa.Global:
				pkg := block.Func().Package()
				pkg = pkg.Program().Package(con.Pkg.Pkg.Path())
				glob := pkg.Global(con.Name())
				if glob == nil {
					log.Fatalf("reference to unknown global %s in function %s", con.Name(), block.Func().Name)
				}
				glob.Referenced = true
				arg = glob

			case nil:
				// slice positions can be nil :-/
				arg = ir2.ConstFor(nil)

			default:
				ok = false
			}
		}
		if ok && arg != nil {
			var typ types.Type
			if *val != nil {
				typ = (*val).Type()
			}
			v := block.Func().ValueFor(typ, arg)
			irInstr.InsertArg(-1, v)
		} else {
			if fe.placeholders == nil {
				fe.placeholders = make(map[string]ssa.Value)
			}
			name := (*val).Name()
			fe.placeholders[name] = *val
			irInstr.InsertArg(-1, block.Func().PlaceholderFor(name))
		}
	}
}

func (fe *FrontEnd) resolvePlaceholders(fn *ir2.Func) {
	for _, label := range fn.PlaceholderLabels() {
		ssaValue := fe.placeholders[label]
		irValue, ok := fe.val2val[ssaValue]
		if !ok {
			log.Fatalf("Unmapped value: %s: %#v for %s\n", label, ssaValue, fn.FullName)
		}
		fn.ResolvePlaceholder(label, irValue)
	}
	fe.placeholders = nil
}
