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
