// Code generated by "enumer -type=Op -transform title-lower"; DO NOT EDIT.

package op

import (
	"fmt"
	"strings"
)

const _OpName = "invalidbuiltincallchangeInterfacechangeTypeconstconvertcopyextractfieldfieldAddrfreeVarfuncglobalindexindexAddrlocallookupmakeInterfacemakeSlicenextnewpanicparameterphiphiCopyrangeregslicesliceToArrayPointerstoretypeAssertaddsubmuldivremandorxorshiftLeftshiftRightandNotequalnotEquallesslessEqualgreatergreaterEqualnotnegateloadinvert"

var _OpIndex = [...]uint16{0, 7, 14, 18, 33, 43, 48, 55, 59, 66, 71, 80, 87, 91, 97, 102, 111, 116, 122, 135, 144, 148, 151, 156, 165, 168, 175, 180, 183, 188, 207, 212, 222, 225, 228, 231, 234, 237, 240, 242, 245, 254, 264, 270, 275, 283, 287, 296, 303, 315, 318, 324, 328, 334}

const _OpLowerName = "invalidbuiltincallchangeinterfacechangetypeconstconvertcopyextractfieldfieldaddrfreevarfuncglobalindexindexaddrlocallookupmakeinterfacemakeslicenextnewpanicparameterphiphicopyrangeregsliceslicetoarraypointerstoretypeassertaddsubmuldivremandorxorshiftleftshiftrightandnotequalnotequallesslessequalgreatergreaterequalnotnegateloadinvert"

func (i Op) String() string {
	if i < 0 || i >= Op(len(_OpIndex)-1) {
		return fmt.Sprintf("Op(%d)", i)
	}
	return _OpName[_OpIndex[i]:_OpIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _OpNoOp() {
	var x [1]struct{}
	_ = x[Invalid-(0)]
	_ = x[Builtin-(1)]
	_ = x[Call-(2)]
	_ = x[ChangeInterface-(3)]
	_ = x[ChangeType-(4)]
	_ = x[Const-(5)]
	_ = x[Convert-(6)]
	_ = x[Copy-(7)]
	_ = x[Extract-(8)]
	_ = x[Field-(9)]
	_ = x[FieldAddr-(10)]
	_ = x[FreeVar-(11)]
	_ = x[Func-(12)]
	_ = x[Global-(13)]
	_ = x[Index-(14)]
	_ = x[IndexAddr-(15)]
	_ = x[Local-(16)]
	_ = x[Lookup-(17)]
	_ = x[MakeInterface-(18)]
	_ = x[MakeSlice-(19)]
	_ = x[Next-(20)]
	_ = x[New-(21)]
	_ = x[Panic-(22)]
	_ = x[Parameter-(23)]
	_ = x[Phi-(24)]
	_ = x[PhiCopy-(25)]
	_ = x[Range-(26)]
	_ = x[Reg-(27)]
	_ = x[Slice-(28)]
	_ = x[SliceToArrayPointer-(29)]
	_ = x[Store-(30)]
	_ = x[TypeAssert-(31)]
	_ = x[Add-(32)]
	_ = x[Sub-(33)]
	_ = x[Mul-(34)]
	_ = x[Div-(35)]
	_ = x[Rem-(36)]
	_ = x[And-(37)]
	_ = x[Or-(38)]
	_ = x[Xor-(39)]
	_ = x[ShiftLeft-(40)]
	_ = x[ShiftRight-(41)]
	_ = x[AndNot-(42)]
	_ = x[Equal-(43)]
	_ = x[NotEqual-(44)]
	_ = x[Less-(45)]
	_ = x[LessEqual-(46)]
	_ = x[Greater-(47)]
	_ = x[GreaterEqual-(48)]
	_ = x[Not-(49)]
	_ = x[Negate-(50)]
	_ = x[Load-(51)]
	_ = x[Invert-(52)]
}

var _OpValues = []Op{Invalid, Builtin, Call, ChangeInterface, ChangeType, Const, Convert, Copy, Extract, Field, FieldAddr, FreeVar, Func, Global, Index, IndexAddr, Local, Lookup, MakeInterface, MakeSlice, Next, New, Panic, Parameter, Phi, PhiCopy, Range, Reg, Slice, SliceToArrayPointer, Store, TypeAssert, Add, Sub, Mul, Div, Rem, And, Or, Xor, ShiftLeft, ShiftRight, AndNot, Equal, NotEqual, Less, LessEqual, Greater, GreaterEqual, Not, Negate, Load, Invert}

var _OpNameToValueMap = map[string]Op{
	_OpName[0:7]:          Invalid,
	_OpLowerName[0:7]:     Invalid,
	_OpName[7:14]:         Builtin,
	_OpLowerName[7:14]:    Builtin,
	_OpName[14:18]:        Call,
	_OpLowerName[14:18]:   Call,
	_OpName[18:33]:        ChangeInterface,
	_OpLowerName[18:33]:   ChangeInterface,
	_OpName[33:43]:        ChangeType,
	_OpLowerName[33:43]:   ChangeType,
	_OpName[43:48]:        Const,
	_OpLowerName[43:48]:   Const,
	_OpName[48:55]:        Convert,
	_OpLowerName[48:55]:   Convert,
	_OpName[55:59]:        Copy,
	_OpLowerName[55:59]:   Copy,
	_OpName[59:66]:        Extract,
	_OpLowerName[59:66]:   Extract,
	_OpName[66:71]:        Field,
	_OpLowerName[66:71]:   Field,
	_OpName[71:80]:        FieldAddr,
	_OpLowerName[71:80]:   FieldAddr,
	_OpName[80:87]:        FreeVar,
	_OpLowerName[80:87]:   FreeVar,
	_OpName[87:91]:        Func,
	_OpLowerName[87:91]:   Func,
	_OpName[91:97]:        Global,
	_OpLowerName[91:97]:   Global,
	_OpName[97:102]:       Index,
	_OpLowerName[97:102]:  Index,
	_OpName[102:111]:      IndexAddr,
	_OpLowerName[102:111]: IndexAddr,
	_OpName[111:116]:      Local,
	_OpLowerName[111:116]: Local,
	_OpName[116:122]:      Lookup,
	_OpLowerName[116:122]: Lookup,
	_OpName[122:135]:      MakeInterface,
	_OpLowerName[122:135]: MakeInterface,
	_OpName[135:144]:      MakeSlice,
	_OpLowerName[135:144]: MakeSlice,
	_OpName[144:148]:      Next,
	_OpLowerName[144:148]: Next,
	_OpName[148:151]:      New,
	_OpLowerName[148:151]: New,
	_OpName[151:156]:      Panic,
	_OpLowerName[151:156]: Panic,
	_OpName[156:165]:      Parameter,
	_OpLowerName[156:165]: Parameter,
	_OpName[165:168]:      Phi,
	_OpLowerName[165:168]: Phi,
	_OpName[168:175]:      PhiCopy,
	_OpLowerName[168:175]: PhiCopy,
	_OpName[175:180]:      Range,
	_OpLowerName[175:180]: Range,
	_OpName[180:183]:      Reg,
	_OpLowerName[180:183]: Reg,
	_OpName[183:188]:      Slice,
	_OpLowerName[183:188]: Slice,
	_OpName[188:207]:      SliceToArrayPointer,
	_OpLowerName[188:207]: SliceToArrayPointer,
	_OpName[207:212]:      Store,
	_OpLowerName[207:212]: Store,
	_OpName[212:222]:      TypeAssert,
	_OpLowerName[212:222]: TypeAssert,
	_OpName[222:225]:      Add,
	_OpLowerName[222:225]: Add,
	_OpName[225:228]:      Sub,
	_OpLowerName[225:228]: Sub,
	_OpName[228:231]:      Mul,
	_OpLowerName[228:231]: Mul,
	_OpName[231:234]:      Div,
	_OpLowerName[231:234]: Div,
	_OpName[234:237]:      Rem,
	_OpLowerName[234:237]: Rem,
	_OpName[237:240]:      And,
	_OpLowerName[237:240]: And,
	_OpName[240:242]:      Or,
	_OpLowerName[240:242]: Or,
	_OpName[242:245]:      Xor,
	_OpLowerName[242:245]: Xor,
	_OpName[245:254]:      ShiftLeft,
	_OpLowerName[245:254]: ShiftLeft,
	_OpName[254:264]:      ShiftRight,
	_OpLowerName[254:264]: ShiftRight,
	_OpName[264:270]:      AndNot,
	_OpLowerName[264:270]: AndNot,
	_OpName[270:275]:      Equal,
	_OpLowerName[270:275]: Equal,
	_OpName[275:283]:      NotEqual,
	_OpLowerName[275:283]: NotEqual,
	_OpName[283:287]:      Less,
	_OpLowerName[283:287]: Less,
	_OpName[287:296]:      LessEqual,
	_OpLowerName[287:296]: LessEqual,
	_OpName[296:303]:      Greater,
	_OpLowerName[296:303]: Greater,
	_OpName[303:315]:      GreaterEqual,
	_OpLowerName[303:315]: GreaterEqual,
	_OpName[315:318]:      Not,
	_OpLowerName[315:318]: Not,
	_OpName[318:324]:      Negate,
	_OpLowerName[318:324]: Negate,
	_OpName[324:328]:      Load,
	_OpLowerName[324:328]: Load,
	_OpName[328:334]:      Invert,
	_OpLowerName[328:334]: Invert,
}

var _OpNames = []string{
	_OpName[0:7],
	_OpName[7:14],
	_OpName[14:18],
	_OpName[18:33],
	_OpName[33:43],
	_OpName[43:48],
	_OpName[48:55],
	_OpName[55:59],
	_OpName[59:66],
	_OpName[66:71],
	_OpName[71:80],
	_OpName[80:87],
	_OpName[87:91],
	_OpName[91:97],
	_OpName[97:102],
	_OpName[102:111],
	_OpName[111:116],
	_OpName[116:122],
	_OpName[122:135],
	_OpName[135:144],
	_OpName[144:148],
	_OpName[148:151],
	_OpName[151:156],
	_OpName[156:165],
	_OpName[165:168],
	_OpName[168:175],
	_OpName[175:180],
	_OpName[180:183],
	_OpName[183:188],
	_OpName[188:207],
	_OpName[207:212],
	_OpName[212:222],
	_OpName[222:225],
	_OpName[225:228],
	_OpName[228:231],
	_OpName[231:234],
	_OpName[234:237],
	_OpName[237:240],
	_OpName[240:242],
	_OpName[242:245],
	_OpName[245:254],
	_OpName[254:264],
	_OpName[264:270],
	_OpName[270:275],
	_OpName[275:283],
	_OpName[283:287],
	_OpName[287:296],
	_OpName[296:303],
	_OpName[303:315],
	_OpName[315:318],
	_OpName[318:324],
	_OpName[324:328],
	_OpName[328:334],
}

// OpString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func OpString(s string) (Op, error) {
	if val, ok := _OpNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _OpNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Op values", s)
}

// OpValues returns all values of the enum
func OpValues() []Op {
	return _OpValues
}

// OpStrings returns a slice of all String values of the enum
func OpStrings() []string {
	strs := make([]string, len(_OpNames))
	copy(strs, _OpNames)
	return strs
}

// IsAOp returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Op) IsAOp() bool {
	for _, v := range _OpValues {
		if i == v {
			return true
		}
	}
	return false
}
