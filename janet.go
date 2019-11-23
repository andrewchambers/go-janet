package janet

type Value interface {
}

type JanetPanicMarker struct {
	PanicV interface{}
}

func JanetPanic(v interface{}) {
	panic(JanetPanicMarker{
		PanicV: v,
	})
}

type Tuple struct {
	Line   int
	Column int
	Vals   []Value
}

func NewTuple(cap int) *Tuple {
	return &Tuple{
		Vals: make([]Value, 0, cap),
	}
}

type Symbol string
