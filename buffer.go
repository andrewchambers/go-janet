package janet

import (
	"bytes"
)

type Buffer struct {
	Buf bytes.Buffer
}

func NewBuffer(n int) *Buffer {
	b := &Buffer{}
	b.Buf.Grow(n)
	return b
}

func (v *Buffer) Hash() uint32 {
	panic("unimplemented")
}
