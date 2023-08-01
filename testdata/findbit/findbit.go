// Find the lowest and highest set bit in an integer
// from: https://rosettacode.org/wiki/Find_first_and_last_set_bit_of_a_long_integer

package main

const (
	mask0, bit0 = (1 << (1 << iota)) - 1, 1 << iota
	mask1, bit1
	mask2, bit2
	mask3, bit3
	mask4, bit4
)

// lowest finds the lowest set bit in `x`
func lowest(x uint32) (out int) {
	if x == 0 {
		return -1
	}
	if x&^mask4 != 0 {
		x >>= bit4
		out |= bit4
	}
	if x&^mask3 != 0 {
		x >>= bit3
		out |= bit3
	}
	if x&^mask2 != 0 {
		x >>= bit2
		out |= bit2
	}
	if x&^mask1 != 0 {
		x >>= bit1
		out |= bit1
	}
	if x&^mask0 != 0 {
		out |= bit0
	}
	return
}

// highest finds the highest set bit in `x`
func highest(x uint32) (out int) {
	if x == 0 {
		return 0
	}
	if x&mask4 == 0 {
		x >>= bit4
		out |= bit4
	}
	if x&mask3 == 0 {
		x >>= bit3
		out |= bit3
	}
	if x&mask2 == 0 {
		x >>= bit2
		out |= bit2
	}
	if x&mask1 == 0 {
		x >>= bit1
		out |= bit1
	}
	if x&mask0 == 0 {
		out |= bit0
	}
	return
}

func main() {
	println("power number lowest highest")
	const base = 42
	n := uint32(1)
	for i := 0; i < 6; i++ {
		println(base, i, n, lowest(n), highest(n))
		n *= base
	}
}
