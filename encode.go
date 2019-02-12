// Package encode provides utilities for encoding and decoding structures into raw bytes.
//
// Normal usage of this package looks like this:
//
//   type encodableFoo struct {
//   	a uint16
//   	b string
//   	c bool
//   }
//
//   func (e encodableFoo) encoding() encode.Encoding {
//   	return encode.New(
//   		encode.BigEndianUint16(&e.a),
//   		encode.LengthDelimString(&e.b),
//   		encode.Bool(&e.c),
//   	)
//   }
//
//   func (e encodableFoo) Encode() []byte {
//   	return e.encoding().Encode()
//   }
//
//   func (e encodableFoo) Decode(b []byte) error {
//   	return e.encoding().Decode(b)
//   }
package encode

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

var errOverflowVarint = errors.New("encode: invalid varint")

type Item interface {
	// Encodes this item into buf. Buf will be at least Size() bytes.
	Encode(buf []byte)
	// Decodes buf into this item, mutating it to match the representation in buf.
	Decode(buf []byte) error
	// Returns the number of bytes that Encode() will use.
	Size() int
}

type Encoding struct {
	items []Item
}

func New(items ...Item) Encoding {
	return Encoding{items: items}
}

func (enc Encoding) Encode() []byte {
	totalSize := 0
	for _, item := range enc.items {
		totalSize += item.Size()
	}
	buf := make([]byte, totalSize)
	i := 0
	for _, item := range enc.items {
		size := item.Size()
		item.Encode(buf[i : i+size])
		i += size
	}
	return buf
}

func (enc Encoding) Decode(buf []byte) error {
	i := 0
	for _, item := range enc.items {
		err := item.Decode(buf[i:])
		if err != nil {
			return err
		}
		i += item.Size()
	}
	return nil
}

func Byte(v *byte) Item {
	return encByte{v}
}

type encByte struct{ v *byte }

func (e encByte) Encode(buf []byte) {
	buf[0] = *e.v
}
func (e encByte) Size() int {
	return 1
}
func (e encByte) Decode(buf []byte) error {
	if len(buf) < 1 {
		return io.ErrUnexpectedEOF
	}
	*e.v = buf[0]
	return nil
}

func BigEndianUint16(v *uint16) Item {
	return bigEndianUint16{v}
}

type bigEndianUint16 struct{ v *uint16 }

func (e bigEndianUint16) Encode(buf []byte) {
	binary.BigEndian.PutUint16(buf, *e.v)
}
func (e bigEndianUint16) Size() int {
	return 2
}
func (e bigEndianUint16) Decode(buf []byte) error {
	if len(buf) < 2 {
		return io.ErrUnexpectedEOF
	}
	*e.v = binary.BigEndian.Uint16(buf)
	return nil
}

func BigEndianUint32(v *uint32) Item {
	return bigEndianUint32{v}
}

type bigEndianUint32 struct{ v *uint32 }

func (e bigEndianUint32) Encode(buf []byte) {
	binary.BigEndian.PutUint32(buf, *e.v)
}
func (e bigEndianUint32) Size() int {
	return 4
}
func (e bigEndianUint32) Decode(buf []byte) error {
	if len(buf) < 4 {
		return io.ErrUnexpectedEOF
	}
	*e.v = binary.BigEndian.Uint32(buf)
	return nil
}

func BigEndianUint64(v *uint64) Item {
	return bigEndianUint64{v}
}

type bigEndianUint64 struct{ v *uint64 }

func (e bigEndianUint64) Encode(buf []byte) {
	binary.BigEndian.PutUint64(buf, *e.v)
}
func (e bigEndianUint64) Size() int {
	return 8
}
func (e bigEndianUint64) Decode(buf []byte) error {
	if len(buf) < 8 {
		return io.ErrUnexpectedEOF
	}
	*e.v = binary.BigEndian.Uint64(buf)
	return nil
}

// Encodes v using a variable-length encoding, so that smaller numbers use fewer bytes.
//
//   min     max          encoded size in bytes
//   0       2^7 - 1      1
//   2^7     2^14 - 1     2
//   2^14    2^21 - 1     3
//   2^21    2^28 - 1     4
//   2^28    2^32 - 1     5
//
// See more at https://developers.google.com/protocol-buffers/docs/encoding#varints
func Uvarint32(v *uint32) Item {
	return uvarint32{v}
}

type uvarint32 struct{ v *uint32 }

func (e uvarint32) Encode(buf []byte) {
	binary.PutUvarint(buf, uint64(*e.v))
}
func (e uvarint32) Size() int {
	return uvarintSize(uint64(*e.v))
}
func (e uvarint32) Decode(buf []byte) error {
	l, n := binary.Uvarint(buf)
	if n == 0 {
		return io.ErrUnexpectedEOF
	}
	if n < 0 {
		return errOverflowVarint
	}
	if n > math.MaxUint32 {
		return errOverflowVarint
	}
	*e.v = uint32(l)
	return nil
}

// Encodes v using a variable-length encoding, so that smaller numbers use fewer bytes.
//
//   min     max          encoded size in bytes
//   0       2^7 - 1      1
//   2^7     2^14 - 1     2
//   2^14    2^21 - 1     3
//   2^21    2^28 - 1     4
//   2^28    2^35 - 1     5
//   2^35    2^42 - 1     6
//   2^42    2^49 - 1     7
//   2^49    2^56 - 1     8
//   2^56    2^63 - 1     9
//   2^63    2^64 - 1     10
//
// See more at https://developers.google.com/protocol-buffers/docs/encoding#varints
func Uvarint64(v *uint64) Item {
	return uvarint64{v}
}

type uvarint64 struct{ v *uint64 }

func (e uvarint64) Encode(buf []byte) {
	binary.PutUvarint(buf, *e.v)
}
func (e uvarint64) Size() int {
	return uvarintSize(*e.v)
}
func (e uvarint64) Decode(buf []byte) error {
	l, n := binary.Uvarint(buf)
	if n == 0 {
		return io.ErrUnexpectedEOF
	}
	if n < 0 {
		return errOverflowVarint
	}
	*e.v = l
	return nil
}

// Encodes v as a uvarint of v's length, followed by v.
func LengthDelimBytes(v *[]byte) Item {
	return lengthDelimBytes{v}
}

type lengthDelimBytes struct{ v *[]byte }

func (e lengthDelimBytes) Encode(buf []byte) {
	n := binary.PutUvarint(buf, uint64(len(*e.v)))
	copy(buf, (*e.v)[n:])
}
func (e lengthDelimBytes) Size() int {
	return uvarintSize(uint64(len(*e.v))) + len(*e.v)
}
func (e lengthDelimBytes) Decode(buf []byte) error {
	l, n := binary.Uvarint(buf)
	if n == 0 {
		return io.ErrUnexpectedEOF
	}
	if n < 0 {
		return errOverflowVarint
	}
	if uint64(len(buf[n:])) < l {
		return io.ErrUnexpectedEOF
	}
	*e.v = make([]byte, l)
	copy(buf[n:], *e.v)
	return nil
}

// Encodes v as a uvarint of v's length, followed by v.
func LengthDelimString(v *string) Item {
	return lengthDelimString{v}
}

type lengthDelimString struct{ v *string }

func (e lengthDelimString) Encode(buf []byte) {
	n := binary.PutUvarint(buf, uint64(len(*e.v)))
	copy(buf, (*e.v)[n:])
}
func (e lengthDelimString) Size() int {
	return uvarintSize(uint64(len(*e.v))) + len(*e.v)
}
func (e lengthDelimString) Decode(buf []byte) error {
	l, n := binary.Uvarint(buf)
	if n == 0 {
		return io.ErrUnexpectedEOF
	}
	if n < 0 {
		return errOverflowVarint
	}
	if uint64(len(buf[n:])) < l {
		return io.ErrUnexpectedEOF
	}
	*e.v = string(buf[n:])
	return nil
}

func uvarintSize(x uint64) int {
	var b [binary.MaxVarintLen64]byte
	return binary.PutUvarint(b[:], x)
}

func varintSize(x int64) int {
	var b [binary.MaxVarintLen64]byte
	return binary.PutVarint(b[:], x)
}
