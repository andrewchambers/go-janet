package janet

type JanetParserStatus int

const (
	JANET_PARSE_ROOT = iota
	JANET_PARSE_ERROR
	JANET_PARSE_PENDING
	JANET_PARSE_DEAD
)

const (
	PFLAG_CONTAINER     = 0x100
	PFLAG_BUFFER        = 0x200
	PFLAG_PARENS        = 0x400
	PFLAG_SQRBRACKETS   = 0x800
	PFLAG_CURLYBRACKETS = 0x1000
	PFLAG_STRING        = 0x2000
	PFLAG_LONGSTRING    = 0x4000
	PFLAG_READERMAC     = 0x8000
	PFLAG_ATSYM         = 0x10000
	PFLAG_COMMENT       = 0x20000
	PFLAG_TOKEN         = 0x40000
)

type Parser struct {
	args     []Value
	states   []ParseState
	buf      []byte
	line     int
	column   int
	pending  int
	lookback byte
	flag     int
	err      string
}

type ParseState struct {
	counter  int
	argn     int
	flags    int
	line     int
	column   int
	consumer Consumer
}

type Consumer func(parser *Parser, state *ParseState, c byte) int

func isWhitespace(c byte) bool {
	return c == ' ' ||
		c == '\t' ||
		c == '\n' ||
		c == '\r' ||
		c == '\v' ||
		c == '\f' ||
		c == 0
}

/* Code generated by cjanet/tools/symcharsgen.c.
 * The table contains 256 bits, where each bit is 1
 * if the corresponding ascii code is a symbol char, and 0
 * if not. The upper characters are also considered symbol
 * chars and are then checked for utf-8 compliance. */
var symchars = [8]uint32{
	0x00000000, 0xf7ffec72, 0xc7ffffff, 0x07fffffe,
	0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff,
}

/* Check if a character is a valid symbol character
 * symbol chars are A-Z, a-z, 0-9, or one of !$&*+-./:<=>@\^_~| */
func isSymbolChar(c byte) bool {
	return (symchars[c>>5] & (uint32(1) << (c & 0x1F))) != 0
}

func checkEscape(c byte) int {
	switch c {
	default:
		return -1
	case 'x':
		return 1
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	case '0':
		return 0
	case 'z':
		return 0
	case 'f':
		return '\f'
	case 'v':
		return '\v'
	case 'e':
		return 27
	case '"':
		return '"'
	case '\\':
		return '\\'
	}
}

/* Get hex digit from a letter */
func toHex(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c) - '0'
	} else if c >= 'A' && c <= 'F' {
		return 10 + int(c) - 'A'
	} else if c >= 'a' && c <= 'f' {
		return 10 + int(c) - 'a'
	} else {
		return -1
	}
}

func (parser *Parser) checkDead() {
	if parser.flag != 0 {
		JanetPanic("parser is dead, cannot consume")
	}
	if parser.err != "" {
		JanetPanic("parser has unchecked error, cannot consume")
	}
}

func (parser *Parser) Init() {
	parser.args = []Value{}
	parser.states = []ParseState{}
	parser.buf = []byte{}
	parser.err = ""
	parser.lookback = 0
	parser.line = 1
	parser.column = 0
	parser.pending = 0
	parser.flag = 0
	parser.pushState(root, PFLAG_CONTAINER)
}

func (parser *Parser) pushState(consumer Consumer, flags int) {
	state := ParseState{
		counter:  0,
		argn:     0,
		flags:    flags,
		consumer: consumer,
		line:     parser.line,
		column:   parser.column,
	}
	parser.states = append(parser.states, state)
}

func (p *Parser) popState(val Value) {
	for {
		top := p.states[len(p.states)-1]
		p.states = p.states[:len(p.states)-1]
		newtop := &p.states[len(p.states)-1]
		if (newtop.flags & PFLAG_CONTAINER) != 0 {
			switch val := val.(type) {
			case *Tuple:
				val.Line = top.line
				val.Column = top.column
			default:
			}
			newtop.argn += 1
			/* Keep track of number of values in the root state */
			if len(p.states) == 1 {
				p.pending += 1
			}
			p.args = append(p.args, val)
			return
		} else if (newtop.flags & PFLAG_READERMAC) != 0 {
			which := "<unknown>"
			t := NewTuple(2, 2)
			c := newtop.flags & 0xFF

			switch c {
			case '\'':
				which = "quote"
			case ',':
				which = "unquote"
			case ';':
				which = "splice"
			case '|':
				which = "short-fn"
			case '~':
				which = "quasiquote"
			default:
			}
			t.Vals[0] = Symbol(which)
			t.Vals[1] = val
			/* Quote source mapping info */
			t.Line = newtop.line
			t.Column = newtop.column
			val = t
		} else {
			return
		}
	}
}

func (p *Parser) closeArray(state *ParseState) *Array {
	array := NewArray(state.argn, state.argn)
	for i := state.argn - 1; i >= 0; i = i - 1 {
		array.Data[i] = p.args[len(p.args)-1]
		p.args = p.args[:len(p.args)-1]
	}
	return array
}

func (p *Parser) closeTuple(state *ParseState, flags int) *Tuple {
	tup := NewTuple(state.argn, state.argn)
	tup.Flags = flags
	for i := state.argn - 1; i >= 0; i = i - 1 {
		tup.Vals[i] = p.args[len(p.args)-1]
		p.args = p.args[:len(p.args)-1]
	}
	return tup
}

func (p *Parser) closeStruct(state *ParseState) *Struct {
	/*
	   JanetKV *st = janet_struct_begin(state->argn >> 1);
	   for (int32_t i = state->argn; i > 0; i -= 2) {
	       Janet value = p->args[--p->argcount];
	       Janet key = p->args[--p->argcount];
	       janet_struct_put(st, key, value);
	   }
	   return janet_wrap_struct(janet_struct_end(st));
	*/
	panic("XXX")
}

func (p *Parser) closeTable(state *ParseState) *Table {
	/*
	   JanetTable *table = janet_table(state->argn >> 1);
	   for (int32_t i = state->argn; i > 0; i -= 2) {
	       Janet value = p->args[--p->argcount];
	       Janet key = p->args[--p->argcount];
	       janet_table_put(table, key, value);
	   }
	   return janet_wrap_table(table);
	*/
	panic("XXX")
}

func (parser *Parser) Consume(c byte) {
	consumed := 0
	parser.checkDead()
	if c == '\r' {
		parser.line += 1
		parser.column = 0
	} else if c == '\n' {
		parser.column = 0
		if parser.lookback != '\r' {
			parser.line += 1
		}
	} else {
		parser.column += 1
	}

	for consumed == 0 && parser.err == "" {
		state := &parser.states[len(parser.states)-1]
		consumed = state.consumer(parser, state, c)
	}

	parser.lookback = c
}

func root(p *Parser, state *ParseState, c byte) int {
	switch c {
	default:
		if isWhitespace(c) {
			return 1
		}
		if !isSymbolChar(c) {
			p.err = "unexpected character"
			return 1
		}
		p.pushState(tokenchar, PFLAG_TOKEN)
		return 0
	case '\'':
	case ',':
	case ';':
	case '~':
	case '|':
		p.pushState(root, PFLAG_READERMAC|int(c))
		return 1
	case '"':
		p.pushState(stringchar, PFLAG_STRING)
		return 1
	case '#':
		p.pushState(comment, PFLAG_COMMENT)
		return 1
	case '@':
		p.pushState(atsign, PFLAG_ATSYM)
		return 1
	case '`':
		p.pushState(longstring, PFLAG_LONGSTRING)
		return 1
	case ')':
	case ']':
	case '}':
		{
			var ds Value
			if len(p.states) == 1 {
				p.err = "unexpected delimiter"
				return 1
			}
			if (c == ')' && (state.flags&PFLAG_PARENS) != 0) ||
				(c == ']' && (state.flags&PFLAG_SQRBRACKETS) != 0) {
				if (state.flags & PFLAG_ATSYM) != 0 {
					ds = p.closeArray(state)
				} else {
					tupFlags := 0
					if c == ']' {
						tupFlags = JANET_TUPLE_FLAG_BRACKETCTOR
					}
					ds = p.closeTuple(state, tupFlags)
				}
			} else if c == '}' && ((state.flags & PFLAG_CURLYBRACKETS) != 0) {
				if (state.argn % 2) == 1 {
					p.err = "struct and table literals expect even number of arguments"
					return 1
				}
				if (state.flags & PFLAG_ATSYM) != 0 {
					ds = p.closeTable(state)
				} else {
					ds = p.closeStruct(state)
				}
			} else {
				p.err = "mismatched delimiter"
				return 1
			}
			p.popState(ds)
		}
		return 1
	case '(':
		p.pushState(root, PFLAG_CONTAINER|PFLAG_PARENS)
		return 1
	case '[':
		p.pushState(root, PFLAG_CONTAINER|PFLAG_SQRBRACKETS)
		return 1
	case '{':
		p.pushState(root, PFLAG_CONTAINER|PFLAG_CURLYBRACKETS)
		return 1
	}
	panic("unreachable")
}

func strEqBuf(str string, buf []byte) bool {
	if len(str) != len(buf) {
		return false
	}

	for i, b := range buf {
		if str[i] != b {
			return false
		}
	}

	return true
}

func scanNumber(s string, out *double) bool {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	*out = v
	return true
}

func tokenchar(p *Parser, state *ParseState, c byte) int {
	var ret Value
	var numval float64

	if isSymbolChar(c) {
		p.buf = append(p.buf, c)
		if c > 127 {
			state.argn = 1 /* Use to indicate non ascii */
		}
		return 1
	}

	if len(p.buf) == 0 {
		p.err = "empty symbol invalid"
		return 0
	}

	/* Token finished */
	startDig := p.buf[0] >= '0' && p.buf[0] <= '9'
	startNum := startDig || p.buf[0] == '-' || p.buf[0] == '+' || p.buf[0] == '.'
	if p.buf[0] == ':' {
		kwStr := p.buf[1:]
		/* Don't do full utf-8 check unless we have seen non ascii characters. */
		valid := (!state.argn) || utf8.Valid(kwStr)
		if !valid {
			p.err = "invalid utf-8 in keyword"
			return 0
		}
		ret = Keyword(kwStr)
	} else if startNum && !scanNumber(string(p.buf), &numval) {
		ret = Number(numval)
	} else if strEqBuf("nil", p.buf) {
		ret = nil
	} else if strEqBuf("false", p.buf) {
		ret = false
	} else if strEqBuf("true", p.buf) {
		ret = true
	} else {
		if startDig {
			p.err = "symbol literal cannot start with a digit"
			return 0
		} else {
			symStr := string(p.buf)
			/* Don't do full utf-8 check unless we have seen non ascii characters. */
			valid := (!state.argn) || utf8.ValidString(symStr)
			if !valid {
				p.err = "invalid utf-8 in symbol"
				return 0
			}
			ret = Symbol(symStr)
		}
	}
	p.buf = p.buf[:0]
	p.popState(p, ret)
	return 0
}

func escapeh(p *Parser, state *ParseState, c byte) int {
	digit := toHex(c)
	if digit < 0 {
		p.err = "invalid hex digit in hex escape"
		return 1
	}
	state.argn = (state.argn << 4) + digit
	state.counter--
	if state.counter == 0 {
		p.buf = append(p.buf, byte(state.argn&0xFF))
		state.argn = 0
		state.consumer = stringchar
	}
	return 1
}

func escape1(p *Parser, state *ParseState, c byte) int {
	e := checkEscape(c)
	if e < 0 {
		p.err = "invalid string escape sequence"
		return 1
	}
	if c == 'x' {
		state.counter = 2
		state.argn = 0
		state.consumer = escapeh
	} else {
		p.buf = append(p.buf, c)
		state.consumer = stringchar
	}
	return 1
}

func stringend(p *Parser, state *ParseState) int {
	var ret Value
	buf := p.buf
	if (state.flags & PFLAG_LONGSTRING) != 0 {
		/* Check for leading newline character so we can remove it */
		if buf[0] == '\n' {
			buf = buf[1:]
		}
		if len(buf) > 0 && buf[len(buf)-1] == '\n' {
			buf = buf[:len(buf)-1]
		}
	}
	if (state.flags & PFLAG_BUFFER) != 0 {
		b := NewBuffer(len(buf))
		_, _ = b.Buf.Write(buf)
		ret = b
	} else {
		ret = String(buf)
	}
	p.buf = []byte{}
	p.popState(ret)
	return 1
}

func stringchar(p *Parser, state *ParseState, c byte) int {
	/* Enter escape */
	if c == '\\' {
		state.consumer = escape1
		return 1
	}
	/* String end */
	if c == '"' {
		return stringend(p, state)
	}
	/* normal char */
	if c != '\n' && c != '\r' {
		p.buf = append(p.buf, c)
	}
	return 1
}

const PFLAG_INSTRING = 0x100000
const PFLAG_END_CANDIDATE = 0x200000

func longstring(p *Parser, state *ParseState, c byte) int {
	if (state.flags & PFLAG_INSTRING) != 0 {
		/* We are inside the long string */
		if c == '`' {
			state.flags |= PFLAG_END_CANDIDATE
			state.flags &= ^PFLAG_INSTRING
			state.counter = 1 /* Use counter to keep track of number of '=' seen */
			return 1
		}
		p.buf = append(p.buf, c)
		return 1
	} else if (state.flags & PFLAG_END_CANDIDATE) != 0 {
		/* We are checking a potential end of the string */
		if state.counter == state.argn {
			stringend(p, state)
			return 0
		}
		if c == '`' && state.counter < state.argn {
			state.counter += 1
			return 1
		}
		/* Failed end candidate */
		for i := 0; i < state.counter; i++ {
			p.buf = append(p.buf, '`')
		}
		p.buf = append(p.buf, c)
		state.counter = 0
		state.flags &= ^PFLAG_END_CANDIDATE
		state.flags |= PFLAG_INSTRING
		return 1
	} else {
		/* We are at beginning of string */
		state.argn += 1
		if c != '`' {
			state.flags |= PFLAG_INSTRING
			p.buf = append(p.buf, c)
		}
		return 1
	}
}

func comment(p *Parser, state *ParseState, c byte) int {
	if c == '\n' {
		p.states = p.states[:len(p.states)-1]
		p.buf = p.buf[:0]
	} else {
		p.buf = append(p.buf, c)
	}
	return 1
}

func atsign(p *Parser, state *ParseState, c byte) int {
	p.states = p.states[:len(p.states)-1]
	switch c {
	case '{':
		p.pushState(root, PFLAG_CONTAINER|PFLAG_CURLYBRACKETS|PFLAG_ATSYM)
		return 1
	case '"':
		p.pushState(stringchar, PFLAG_BUFFER|PFLAG_STRING)
		return 1
	case '`':
		p.pushState(longstring, PFLAG_BUFFER|PFLAG_LONGSTRING)
		return 1
	case '[':
		p.pushState(root, PFLAG_CONTAINER|PFLAG_SQRBRACKETS|PFLAG_ATSYM)
		return 1
	case '(':
		p.pushState(root, PFLAG_CONTAINER|PFLAG_PARENS|PFLAG_ATSYM)
		return 1
	default:
		break
	}
	p.pushState(tokenchar, PFLAG_TOKEN)
	p.buf = append(p.buf, '@')
	return 0
}
