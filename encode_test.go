package encode

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/bradenaw/trand"
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

	trand.RandomN(t, 10000, func(t *testing.T, r *rand.Rand) {
		x1 := r.Uint64() >> uint(r.Int()%64)
		x2 := r.Uint64() >> uint(r.Int()%64)

		if x1 == x2 {
			x2++
		}
		if x2 < x1 {
			x1, x2 = x2, x1
		}

		checkOrdering(x1, x2)
	})
}

func BenchmarkOrdUvarint64Encode(b *testing.B) {
	bunchaUint64s := make([]uint64, b.N)
	for i := range bunchaUint64s {
		bunchaUint64s[i] = rand.Uint64() >> uint(rand.Int()%64)
	}

	b.ResetTimer()

	var backing [9]byte
	for i := 0; i < b.N; i++ {
		x := bunchaUint64s[i%len(bunchaUint64s)]
		enc := ordUvarint64{&x}
		enc.Encode(backing[:enc.Size()])
	}
}

func BenchmarkOrdUvarint64Decode(b *testing.B) {
	bunchaEncoded := make([][]byte, b.N)
	for i := range bunchaEncoded {
		x := rand.Uint64() >> uint(rand.Int()%64)
		enc := ordUvarint64{&x}
		buf := make([]byte, enc.Size())
		enc.Encode(buf)
		bunchaEncoded[i] = buf
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var x uint64
		enc := ordUvarint64{&x}
		_ = enc.Decode(bunchaEncoded[i%len(bunchaEncoded)])
	}
}
