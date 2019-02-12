package encode

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitbufferBasic(t *testing.T) {
	check := func(ss []string) {
		size := 0
		for _, s := range ss {
			size += len(s)
		}

		w := bitBuffer{
			b: make([]byte, (size+7)/8),
			i: 0,
		}

		for _, s := range ss {
			x := uint64(0)
			for _, c := range s {
				x <<= 1
				if c == '1' {
					x |= 1
				}
			}
			w.writeBits(x, len(s))
		}

		r := bitBuffer{b: w.b, i: 0}
		expected := ""
		actual := ""
		for _, s := range ss {
			bits, err := r.readBits(len(s))
			require.NoError(t, err)
			expected += s
			actual += fmt.Sprintf("%0"+strconv.Itoa(len(s))+"b", bits)
		}
		require.Equal(t, expected, actual)
	}

	check([]string{"11011011"})
	check([]string{"1101101"})
	check([]string{"110101"})
	check([]string{"1101"})
	check([]string{"1"})
	check([]string{"1", "0", "1"})
	check([]string{"11", "01", "10", "1"})
	check([]string{"0100", "1", "110101", "1011", "00110000", "1"})
}
