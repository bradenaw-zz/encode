# encode

[![GoDoc](https://godoc.org/github.com/bradenaw/encode?status.svg)](https://godoc.org/github.com/bradenaw/encode)

Provides utilities for encoding and decoding structures into raw bytes.

Normal usage of this package looks like this:

```
type encodableFoo struct {
    a uint16
    b string
    c bool
}

func (e encodableFoo) encoding() encode.Encoding {
    return encode.New(
        encode.BigEndianUint16(&e.a),
        encode.LengthDelimString(&e.b),
        encode.Bool(&e.c),
    )
}

func (e encodableFoo) Encode() []byte {
    return e.encoding().Encode()
}

func (e encodableFoo) Decode(b []byte) error {
    return e.encoding().Decode(b)
}
```
