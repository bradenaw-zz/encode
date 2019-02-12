package encode

import (
	"fmt"
)

func ExampleBitpacked() {
	b := make([]byte, 3)

	threeBits := byte(0x3)
	sixBits := byte(0x2A)
	fourBits := byte(0xD)

	flag1 := true
	flag2 := false
	flag3 := true
	flag4 := true

	Bitpacked(
		Bits8(&threeBits, 3), // 011
		Bits8(&sixBits, 6),   // 101010
		Bits8(&fourBits, 4),  // 1101
		BitPadding(5),        // 00000
		BitFlags(
			&flag1, // 1
			&flag2, // 0
			&flag3, // 1
			&flag4, // 1
		),
	).Encode(b)

	fmt.Printf("%08b %08b %08b\n", b[0], b[1], b[2])

	// Output:
	// 01110101 01101000 00101100
}
