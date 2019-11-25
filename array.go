package janet

type Array struct {
	Data []Value
}

func (v *Array) Hash() uint32 {
	panic("unimplemented")
}

func NewArray(l, cap int) *Array {
	return &Array{
		Data: make([]Value, l, cap),
	}
}
