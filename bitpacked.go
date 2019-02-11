package encode

var errBufferOverrun = errors.New("encode: buffer overrun")

type BitpackItem interface {
	encode(w *bitBuffer)
	decode(r *bitBuffer) error
	size() int
}

type bitpacked struct{ items []BitpackItem }

func Bitpacked(items ...BitpackItem) Item {
	return bitpacked{items: items}
}

func (b bitpacked) Encode(buf []byte) {
	bitBuf := bitBuffer{b: buf, i: 0}
	for _, item := range b.items {
		item.encode(&bitBuf)
	}
}
func (b bitpacked) Decode(buf []byte) error {
	bitBuf := bitBuffer{b: buf, i: 0}
	for _, item := range b.items {
		err := item.decode(&bitBuf)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b bitpacked) Size() int {
	sizeBits := 0
	for _, item := range b.items {
		sizeBits += item.size()
	}
	return (sizeBits + 7) / 8
}

type bitItem struct{ v *bool }

func Bit(v *bool) BitpackItem {
	return bitItem{v}
}

func (e bitItem) encode(w *bitBuffer) {
	x := uint64(0)
	if *e.v {
		x = 1
	}
	w.writeBits(x, 1)
}
func (e bitItem) decode(r *bitBuffer) error {
}
func (e bitItem) size() int {
	return 1
}

type bits8 struct {
	v *byte
	n int
}

func Bits8(v *byte, n int) BitpackItem {
	return bits8{v, n}
}

func (e bits8) encode(w *bitBuffer) {
	w.writeBits(uint64(*e.v), e.n)
}
func (e bits8) decode(r *bitBuffer) error {
	*e.v = byte(r.readBits(e.n))
}
func (e bits8) size() int {
	return e.n
}

type bits16 struct {
	v *uint16
	n int
}

func Bits16(v *uint16, n int) BitpackItem {
	return bits16{v, n}
}

func (e bits16) encode(w *bitBuffer) {
	w.writeBits(uint64(*e.v), e.n)
}
func (e bits16) decode(r *bitBuffer) error {
	*e.v = uint16(r.readBits(e.n))
}
func (e bits16) size() int {
	return e.n
}

type bitBuffer struct {
	b []byte
	// The current bit index, where the next bit will be read or written from.
	i int
}

// Write the n lowest-order bits from x into b.
func (b *bitBuffer) writeBits(x uint64, n int) {
	if b.i+n > b.lenBits() {
		panic(errBufferOverrun)
	}

	shiftedX := x << uint(64-n) >> uint(b.i%8)
	for j := 0; n > 0; j++ {
		take := minInt(b.availInByte(), n)
		b.b[b.i/8] |= byte(shiftedX >> uint(56-j*8))
		b.i += take
		n -= take
	}
}

// Read n bits from the buffer and return them as the low order bits of the result.
func (b *bitBuffer) readBits(n int) (uint64, error) {
	if b.i+n > b.lenBits() {
		return 0, io.ErrUnexpectedEOF
	}
	result := uint64(0)
	for n > 0 {
		take := minInt(b.availInByte(), n)
		result <<= take
		mask := byte((uint16(1) << uint(take)) - 1)
		result |= b.b[i/8] & mask
		n -= take
		b.i += take
	}
	return result, nil
}

func (b *bitBuffer) availInByte() int {
	return 8 - (b.i % 8)
}

func (b *bitBuffer) lenBits() int {
	return len(b.b) * 8
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
