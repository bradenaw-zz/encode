package encode

type TupleItem interface {
	Item
	EncodeTuple(buf []byte, last bool)
	DecodeTuple(buf []byte, last bool) error
	SizeTuple(last bool) int
	OrderPreserving()
}

type Tuple struct {
	items []TupleItem
}

func NewTuple(items ...TupleItem) Tuple {
	return Tuple{items: items}
}
func (t Tuple) Encode() []byte {
	return t.EncodePrefix(len(t.items))
}
func (t Tuple) EncodePrefix(n int) []byte {
	size := 0
	for i := 0; i < n; i++ {
		item := t.items[i]
		size += item.SizeTuple(i == len(t.items)-1)
	}
	buf := make([]byte, size)
	j := 0
	for i := 0; i < n; i++ {
		item := t.items[i]
		size := item.SizeTuple(i == n-1)
		item.EncodeTuple(buf[j:j+size], i == n-1)
		j += size
	}
	return buf
}
func (t Tuple) Decode(buf []byte) error {
	return t.DecodePrefix(buf, len(t.items))
}
func (t Tuple) DecodePrefix(buf []byte, n int) error {
	j := 0
	for i := 0; i < n; i++ {
		item := t.items[i]
		err := item.DecodeTuple(buf[j:], i == n-1)
		if err != nil {
			return err
		}
		j += item.Size()
	}
	return nil
}
