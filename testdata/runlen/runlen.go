// Run length encoding example
// from: https://rosettacode.org/wiki/Run-length_encoding#Go
package main

// encoding scheme:
// encode to byte array
// byte value < 26 means single character: byte value + 'A'
// byte value 26..255 means (byte value - 24) copies of next byte
func rllEncode(s string) (r []byte) {
	if s == "" {
		return
	}
	c := s[0]
	if c < 'A' || c > 'Z' {
		panic("invalid")
	}
	nc := byte(1)
	for i := 1; i < len(s); i++ {
		d := s[i]
		switch {
		case d != c:
		case nc < (255 - 24):
			nc++
			continue
		}
		if nc > 1 {
			r = append(r, nc+24)
		}
		r = append(r, c-'A')
		if d < 'A' || d > 'Z' {
			panic("invalid")
		}
		c = d
		nc = 1
	}
	if nc > 1 {
		r = append(r, nc+24)
	}
	r = append(r, c-'A')
	return
}

func main() {
	s := "WWWWWWWWWWWWBWWWWWWWWWWWWBBBWWWWWWWWWWWWWWWWWWWWWWWWBWWWWWWWWWWWWWW"
	println("source: ", len(s), "bytes:", s)
	e := rllEncode(s)
	println("encoded:", len(e), "bytes:", e)
	d := rllDecode(e)
	println("decoded:", len(d), "bytes:", d)
	println("decoded = source:", d == s)
}

func rllDecode(e []byte) string {
	var c byte
	var d []byte
	for i := 0; i < len(e); i++ {
		b := e[i]
		if b < 26 {
			c = 1
		} else {
			c = b - 24
			i++
			b = e[i]
		}
		for c > 0 {
			d = append(d, b+'A')
			c--
		}
	}
	return string(d)
}
