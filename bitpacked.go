package encode

import (
	"errors"
	"io"
)

var errBufferOverrun = errors.New("encode: buffer overrun")

type BitpackItem interface {
	encode(b *bitBuffer)
	decode(b *bitBuffer) error
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

func BitPadding(n int) BitpackItem {
	return bitPadding{n}
}

type bitPadding struct{ n int }

func (e bitPadding) encode(b *bitBuffer) {
	b.writeBits(0, e.n)
}
func (e bitPadding) decode(b *bitBuffer) error {
	_, err := b.readBits(e.n)
	return err
}
func (e bitPadding) size() int {
	return e.n
}

// Encodes each value as a single bit, high-order to low-order.
func BitFlags(v ...*bool) BitpackItem {
	return bitFlags{v}
}

type bitFlags struct{ v []*bool }

func (e bitFlags) encode(b *bitBuffer) {
	for i := range e.v {
		x := uint64(0)
		if *e.v[i] {
			x = 1
		}
		b.writeBits(x, 1)
	}
}
func (e bitFlags) size() int {
	return len(e.v)
}
func (e bitFlags) decode(b *bitBuffer) error {
	for i := range e.v {
		bit, err := b.readBits(1)
		if err != nil {
			return err
		}
		*e.v[i] = bit == 1
	}
	return nil
}

type bitItem struct{ v *bool }

func Bit(v *bool) BitpackItem {
	return bitItem{v}
}

func (e bitItem) encode(b *bitBuffer) {
	x := uint64(0)
	if *e.v {
		x = 1
	}
	b.writeBits(x, 1)
}
func (e bitItem) decode(b *bitBuffer) error {
	bit, err := b.readBits(1)
	if err != nil {
		return err
	}
	*e.v = bit == 1
	return nil
}
func (e bitItem) size() int {
	return 1
}

type bits8 struct {
	v *byte
	n int
}

func Bits8(v *byte, n int) BitpackItem {
	if n <= 0 || n > 8 {
		panic(fmt.Sprintf("invalid n=%d, must be in [1, 8]", n))
	}
	return bits8{v, n}
}

func (e bits8) encode(b *bitBuffer) {
	b.writeBits(uint64(*e.v), e.n)
}
func (e bits8) decode(b *bitBuffer) error {
	bits, err := b.readBits(e.n)
	if err != nil {
		return err
	}
	*e.v = byte(bits)
	return nil
}
func (e bits8) size() int {
	return e.n
}

type bits16 struct {
	v *uint16
	n int
}

func Bits16(v *uint16, n int) BitpackItem {
	if n <= 0 || n > 16 {
		panic(fmt.Sprintf("invalid n=%d, must be in [1, 16]", n))
	}
	return bits16{v, n}
}

func (e bits16) encode(b *bitBuffer) {
	b.writeBits(uint64(*e.v), e.n)
}
func (e bits16) decode(b *bitBuffer) error {
	bits, err := b.readBits(e.n)
	if err != nil {
		return err
	}
	*e.v = uint16(bits)
	return nil
}
func (e bits16) size() int {
	return e.n
}

type bitBuffer struct {
	b []byte
	// The current bit index, where the next bit will be read or written from.
	i int
}

// Write the n lowest-order bits from x into b. High order bits come first.
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

	shift := uint(64 - n - b.i%8)
	mask := ((uint64(1) << uint(n)) - 1) << shift
	result := uint64(0)
	for j := 0; n > 0; j++ {
		take := minInt(b.availInByte(), n)
		result |= (uint64(b.b[b.i/8]) << uint(56-j*8)) & mask
		n -= take
		b.i += take
	}
	return result >> shift, nil
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
