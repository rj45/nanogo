package typ

import (
	"fmt"
)

const kindStrings = "invalidbii8i16i32i64uu8u16u32u64uptrf32f64complex64complex128strunsafeptrcbcicrcfccomplexcstrcnilcsrmemblankptrfuncslicearraystructmapinterfacechannumtypes"

var _TypeKindIndex = [...]uint8{0, 7, 8, 9, 11, 14, 17, 20, 21, 23, 26, 29, 32, 36, 39, 42, 51, 61, 64, 73, 75, 77, 79, 81, 89, 93, 97, 100, 103, 108, 111, 115, 120, 125, 131, 134, 143, 147, 155}

// String returns the string form of the type
func (i Kind) String() string {
	if i >= NumTypes {
		// todo: replace with different impl
		return fmt.Sprintf("TypeKind(%d)", i)
	}
	return kindStrings[_TypeKindIndex[i]:_TypeKindIndex[i+1]]
}

var kindMap = map[string]Kind{
	kindStrings[0:7]:     Invalid,
	kindStrings[7:8]:     B,
	kindStrings[8:9]:     I,
	kindStrings[9:11]:    I8,
	kindStrings[11:14]:   I16,
	kindStrings[14:17]:   I32,
	kindStrings[17:20]:   I64,
	kindStrings[20:21]:   U,
	kindStrings[21:23]:   U8,
	kindStrings[23:26]:   U16,
	kindStrings[26:29]:   U32,
	kindStrings[29:32]:   U64,
	kindStrings[32:36]:   Uptr,
	kindStrings[36:39]:   F32,
	kindStrings[39:42]:   F64,
	kindStrings[42:51]:   Complex64,
	kindStrings[51:61]:   Complex128,
	kindStrings[61:64]:   Str,
	kindStrings[64:73]:   UnsafePtr,
	kindStrings[73:75]:   CB,
	kindStrings[75:77]:   CI,
	kindStrings[77:79]:   CR,
	kindStrings[79:81]:   CF,
	kindStrings[81:89]:   CComplex,
	kindStrings[89:93]:   CStr,
	kindStrings[93:97]:   CNil,
	kindStrings[97:100]:  CSR,
	kindStrings[100:103]: Mem,
	kindStrings[103:108]: Blank,
	kindStrings[108:111]: Ptr,
	kindStrings[111:115]: Func,
	kindStrings[115:120]: Slice,
	kindStrings[120:125]: Array,
	kindStrings[125:131]: Struct,
	kindStrings[131:134]: Map,
	kindStrings[134:143]: Interface,
	kindStrings[143:147]: Chan,
}

// TypeKindString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func TypeKindString(s string) (Kind, error) {
	if val, ok := kindMap[s]; ok {
		return val, nil
	}

	return 0, fmt.Errorf("%s does not belong to TypeKind values", s)
}
