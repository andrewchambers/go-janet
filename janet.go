package janet

import (
	"bytes"
)

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

type Symbol string
type Keyword string
type String string

const JANET_TUPLE_FLAG_BRACKETCTOR = 0x10000

type Tuple struct {
	Flags  int
	Line   int
	Column int
	Vals   []Value
}

func NewTuple(l, cap int) *Tuple {
	return &Tuple{
		Vals: make([]Value, l, cap),
	}
}

type Array struct {
	Data []Value
}

func NewArray(l, cap int) *Array {
	return &Array{
		Data: make([]Value, l, cap),
	}
}

type Struct struct {
}

type Table struct {
}

type Buffer struct {
	Buf bytes.Buffer
}

func NewBuffer(n int) *Buffer {
	b := &Buffer{}
	b.Buf.Grow(n)
	return b
}
