package janet

type JanetParserStatus int

const (
	JANET_PARSE_ROOT = iota
	JANET_PARSE_ERROR
	JANET_PARSE_PENDING
	JANET_PARSE_DEAD
)

type Parser struct {
	// Janet *args;
	// const char *error;
	states []*ParseState
	// uint8_t *buf;
	// size_t argcount;
	// size_t argcap;
	// size_t bufcount;
	// size_t bufcap;
	line   int
	column int
	// size_t pending;
	lookback rune
	flag     int
	err      error
}

type ParseState struct {
	counter  int
	argn     int
	flags    int
	line     uint
	column   uint
	consumer Consumer
}

type Consumer func(parser *Parser, state *ParseState, r rune) int

func (parser *Parser) checkDead() {
	if parser.flag != 0 {
		JanetPanic("parser is dead, cannot consume")
	}
	if parser.err != nil {
		JanetPanic("parser has unchecked error, cannot consume")
	}
}

func (parser *Parser) Consume(c rune) {
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

	for consumed == 0 && parser.err == nil {
		state := parser.states[len(parser.states)-1]
		consumed = state.consumer(parser, state, c)
	}

	parser.lookback = c
}
