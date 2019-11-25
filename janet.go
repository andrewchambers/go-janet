package janet

type Value interface {
	Hash() uint32
}

type JanetPanicMarker struct {
	PanicV interface{}
}

type FuncEnv struct {
}

const (
	JANET_FUNCDEF_FLAG_VARARG       = 0x10000
	JANET_FUNCDEF_FLAG_NEEDSENV     = 0x20000
	JANET_FUNCDEF_FLAG_HASNAME      = 0x80000
	JANET_FUNCDEF_FLAG_HASSOURCE    = 0x100000
	JANET_FUNCDEF_FLAG_HASDEFS      = 0x200000
	JANET_FUNCDEF_FLAG_HASENVS      = 0x400000
	JANET_FUNCDEF_FLAG_HASSOURCEMAP = 0x800000
	JANET_FUNCDEF_FLAG_STRUCTARG    = 0x1000000
	JANET_FUNCDEF_FLAG_TAG          = 0xFFFF
)

type JanetSourceMapping struct {
	line   int32
	column int32
}

type FuncDef struct {
	environments *int32 /* Which environments to capture from parent. */
	constants    Value
	defs         **FuncDef
	bytecode     []uint32

	/* Various debug information */
	sourcemap *JanetSourceMapping
	source    []byte
	name      []byte

	flags               int32
	slotcount           int32 /* The amount of stack space required for the function */
	arity               int32 /* Not including varargs */
	min_arity           int32 /* Including varargs */
	max_arity           int32 /* Including varargs */
	constants_length    int32
	bytecode_length     int32
	environments_length int32
	defs_length         int32
}

func JanetPanic(v interface{}) {
	panic(JanetPanicMarker{
		PanicV: v,
	})
}

type Bool bool

func (v Bool) Hash() uint32 {
	if bool(v) {
		return 1
	} else {
		return 0
	}
}

func hashString(s string) uint32 {
	panic("unimplemented")
}

type Symbol string

func (v Symbol) Hash() uint32 { return hashString(string(v)) }

type Keyword string

func (v Keyword) Hash() uint32 { return hashString(string(v)) }

type String string

func (v String) Hash() uint32 { return hashString(string(v)) }

type Number float64

func (v Number) Hash() uint32 {
	panic("XXX")
}

func Equal(x, y Value) (bool, error) {
	if x == nil {
		if y == nil {
			return true, nil
		} else {
			return false, nil
		}
	}

	switch x := x.(type) {
	case Bool:
		switch y := y.(type) {
		case Bool:
			return x == y, nil
		default:
			return false, nil
		}
	case Number:
		switch y := y.(type) {
		case Number:
			return x == y, nil
		default:
			return false, nil
		}
	case *Tuple:
		switch y := y.(type) {
		case *Tuple:
			if x == y {
				return true, nil
			}
			if len(x.Vals) != len(y.Vals) {
				return false, nil
			}
			for i, v := range x.Vals {
				isEqual, err := Equal(v, y.Vals[i])
				if err != nil {
					return false, err
				}
				if !isEqual {
					return false, nil
				}
			}
			return true, nil
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
	case Symbol:
		switch y := y.(type) {
		case Symbol:
			return x == y, nil
		default:
			return false, nil
		}
	case Keyword:
		switch y := y.(type) {
		case Keyword:
			return x == y, nil
		default:
			return false, nil
		}
	default:
		return false, nil
	}
}
