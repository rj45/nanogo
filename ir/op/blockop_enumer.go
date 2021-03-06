// Code generated by "enumer -type=BlockOp -transform title-lower"; DO NOT EDIT.

package op

import (
	"fmt"
	"strings"
)

const _BlockOpName = "badBlockjumpifreturnpanicifEqualifNotEqualifLessifLessEqualifGreaterifGreaterEqual"

var _BlockOpIndex = [...]uint8{0, 8, 12, 14, 20, 25, 32, 42, 48, 59, 68, 82}

const _BlockOpLowerName = "badblockjumpifreturnpanicifequalifnotequaliflessiflessequalifgreaterifgreaterequal"

func (i BlockOp) String() string {
	if i < 0 || i >= BlockOp(len(_BlockOpIndex)-1) {
		return fmt.Sprintf("BlockOp(%d)", i)
	}
	return _BlockOpName[_BlockOpIndex[i]:_BlockOpIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _BlockOpNoOp() {
	var x [1]struct{}
	_ = x[BadBlock-(0)]
	_ = x[Jump-(1)]
	_ = x[If-(2)]
	_ = x[Return-(3)]
	_ = x[Panic-(4)]
	_ = x[IfEqual-(5)]
	_ = x[IfNotEqual-(6)]
	_ = x[IfLess-(7)]
	_ = x[IfLessEqual-(8)]
	_ = x[IfGreater-(9)]
	_ = x[IfGreaterEqual-(10)]
}

var _BlockOpValues = []BlockOp{BadBlock, Jump, If, Return, Panic, IfEqual, IfNotEqual, IfLess, IfLessEqual, IfGreater, IfGreaterEqual}

var _BlockOpNameToValueMap = map[string]BlockOp{
	_BlockOpName[0:8]:        BadBlock,
	_BlockOpLowerName[0:8]:   BadBlock,
	_BlockOpName[8:12]:       Jump,
	_BlockOpLowerName[8:12]:  Jump,
	_BlockOpName[12:14]:      If,
	_BlockOpLowerName[12:14]: If,
	_BlockOpName[14:20]:      Return,
	_BlockOpLowerName[14:20]: Return,
	_BlockOpName[20:25]:      Panic,
	_BlockOpLowerName[20:25]: Panic,
	_BlockOpName[25:32]:      IfEqual,
	_BlockOpLowerName[25:32]: IfEqual,
	_BlockOpName[32:42]:      IfNotEqual,
	_BlockOpLowerName[32:42]: IfNotEqual,
	_BlockOpName[42:48]:      IfLess,
	_BlockOpLowerName[42:48]: IfLess,
	_BlockOpName[48:59]:      IfLessEqual,
	_BlockOpLowerName[48:59]: IfLessEqual,
	_BlockOpName[59:68]:      IfGreater,
	_BlockOpLowerName[59:68]: IfGreater,
	_BlockOpName[68:82]:      IfGreaterEqual,
	_BlockOpLowerName[68:82]: IfGreaterEqual,
}

var _BlockOpNames = []string{
	_BlockOpName[0:8],
	_BlockOpName[8:12],
	_BlockOpName[12:14],
	_BlockOpName[14:20],
	_BlockOpName[20:25],
	_BlockOpName[25:32],
	_BlockOpName[32:42],
	_BlockOpName[42:48],
	_BlockOpName[48:59],
	_BlockOpName[59:68],
	_BlockOpName[68:82],
}

// BlockOpString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func BlockOpString(s string) (BlockOp, error) {
	if val, ok := _BlockOpNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _BlockOpNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to BlockOp values", s)
}

// BlockOpValues returns all values of the enum
func BlockOpValues() []BlockOp {
	return _BlockOpValues
}

// BlockOpStrings returns a slice of all String values of the enum
func BlockOpStrings() []string {
	strs := make([]string, len(_BlockOpNames))
	copy(strs, _BlockOpNames)
	return strs
}

// IsABlockOp returns "true" if the value is listed in the enum definition. "false" otherwise
func (i BlockOp) IsABlockOp() bool {
	for _, v := range _BlockOpValues {
		if i == v {
			return true
		}
	}
	return false
}
