package janet

const JANET_TUPLE_FLAG_BRACKETCTOR = 0x10000

type Tuple struct {
	Flags  int
	Line   int
	Column int
	Vals   []Value
}

func (t *Tuple) Hash() uint32 {
	panic("unimplemented")
}

func NewTuple(l, cap int) *Tuple {
	return &Tuple{
		Vals: make([]Value, l, cap),
	}
}
