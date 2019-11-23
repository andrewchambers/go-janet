package janet

import (
	"bytes"
	"math"
	"unsafe"
)

type Value interface {
	Hash() (uint32, error)
}

type JanetPanicMarker struct {
	PanicV interface{}
}

func JanetPanic(v interface{}) {
	panic(JanetPanicMarker{
		PanicV: v,
	})
}

type Bool bool

func (v Bool) Hash() (uint32, error) {
	if bool(v) {
		return 1, nil
	} else {
		return 0, nil
	}
}

type Symbol string

func (v Symbol) Hash() (uint32, error) { return hashString(string(v)), nil }

type Keyword string

func (v Keyword) Hash() (uint32, error) { return hashString(string(v)), nil }

type String string

func (v String) Hash() (uint32, error) { return hashString(string(v)), nil }

type Number float64

func (v Number) Hash() (uint32, error) {
	if isFinite(float64(v)) {
		return uint32(int64(v)), nil
	}
	return 1618033, nil
}

func isFinite(f float64) bool {
	return math.Abs(f) <= math.MaxFloat64
}

const JANET_TUPLE_FLAG_BRACKETCTOR = 0x10000

type Tuple struct {
	Flags  int
	Line   int
	Column int
	Vals   []Value
}

func (t *Tuple) Hash() (uint32, error) {
	// Use same algorithm as Python + starlark.
	var x, mult uint32 = 0x345678, 1000003
	for _, elem := range t.Vals {
		y, err := elem.Hash()
		if err != nil {
			return 0, err
		}
		x = x ^ y*mult
		mult += 82520 + uint32(len(t.Vals)+len(t.Vals))
	}
	return x, nil
}

func NewTuple(l, cap int) *Tuple {
	return &Tuple{
		Vals: make([]Value, l, cap),
	}
}

type Array struct {
	Data []Value
}

func (v *Array) Hash() (uint32, error) { return uint32(uintptr(unsafe.Pointer(v))), nil }

func NewArray(l, cap int) *Array {
	return &Array{
		Data: make([]Value, l, cap),
	}
}

type Struct struct {
}

func (v *Struct) Hash() (uint32, error) { panic("unimplemented") }

type Table struct {
}

func (v *Table) Hash() (uint32, error) { return uint32(uintptr(unsafe.Pointer(v))), nil }

type Buffer struct {
	Buf bytes.Buffer
}

func (v *Buffer) Hash() (uint32, error) { return uint32(uintptr(unsafe.Pointer(v))), nil }

func NewBuffer(n int) *Buffer {
	b := &Buffer{}
	b.Buf.Grow(n)
	return b
}

func Equal(x, y Value) (bool, error) {
	switch x := x.(type) {
	case Number:
		switch y := y.(type) {
		case Number:
			return x == y, nil
		default:
			return false, nil
		}
	case *Array:
		switch y := y.(type) {
		case *Array:
			return x == y, nil
		default:
			return false, nil
		}
	case String:
		switch y := y.(type) {
		case String:
			return x == y, nil
		default:
			return false, nil
		}
	default:
		return false, nil
	}
}
