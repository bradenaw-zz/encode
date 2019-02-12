package encode

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrdUvarint64(t *testing.T) {
	checkRoundtrip := func(x uint64) {
		enc := New(OrdUvarint64(&x))
		x2 := x
		b := enc.Encode()
		t.Logf("%d: %s\n", x, hex.EncodeToString(b))
		x = ^x
		err := enc.Decode(b)
		require.NoError(t, err)
		require.Equal(t, x2, x)
	}

	checkOrdering := func(x uint64, x2 uint64) {
		checkRoundtrip(x)
		checkRoundtrip(x2)

		enc := New(OrdUvarint64(&x))
		b := enc.Encode()

		enc2 := New(OrdUvarint64(&x2))
		b2 := enc2.Encode()

		require.True(
			t,
			bytes.Compare(b, b2) < 0,
			"%d < %d but %s >= %s",
			x, x2, hex.EncodeToString(b), hex.EncodeToString(b2),
		)
	}

	checkOrdering(0, 1)
	checkOrdering(1, 14)
	checkOrdering(16, 123)
	checkOrdering(16, 128)
	checkOrdering(1235, 1239)
	checkOrdering(1231, 123151)
	checkOrdering(1231241, 1230123102)
	checkOrdering(1<<7, 1<<14)
	checkOrdering(1<<14, 1<<21)
	checkOrdering(1<<21, 1<<28)
	checkOrdering(1<<28, 1<<35)
	checkOrdering(1<<35, 1<<42)
	checkOrdering(1<<42, 1<<49)
	checkOrdering(1<<49, 1<<56)
	checkOrdering(1<<56, 1<<63)
	checkOrdering(1<<63, 1<<63+15)
	checkOrdering(1<<63+15, 1<<63+1<<62)
	checkOrdering(1231241, 1<<63+1231023105915)
	checkOrdering(1<<63+1<<62, ^uint64(0))
}
