package janet

import (
	"fmt"
	"testing"

	"github.com/kr/pretty"
)

func TestParser(t *testing.T) {
	type testCase struct {
		Input      string
		Expected   Value
		DebugTrace bool
	}

	testCases := []testCase{
		{Input: "nil", Expected: nil},
		{Input: "true", Expected: Bool(true)},
		{Input: "false", Expected: Bool(false)},
		{Input: "hello", Expected: Symbol("hello")},
		{Input: ":hello", Expected: Keyword("hello")},
		{Input: "\"hello\"", Expected: String("hello")},
		{Input: "12.4", Expected: Number(12.4)},
		{Input: "-12.4", Expected: Number(-12.4)},
		{Input: "(1 :hello)", Expected: &Tuple{
			Vals: []Value{Number(1), Keyword("hello")},
		}},
		{Input: "[1 :hello]", Expected: &Tuple{
			Vals: []Value{Number(1), Keyword("hello")},
		}},
	}

	runTest := func(tc *testCase) {
		if tc.DebugTrace {
			fmt.Printf("TRACING PARSE OF %q\n", tc.Input)
		}
		p := &Parser{}
		p.Init()
		for i := 0; i < len(tc.Input); i++ {
			b := tc.Input[i]
			p.Consume(b)
			if tc.DebugTrace {
				fmt.Printf("ADDING: '%c'\n", b)
				pretty.Print(p)
			}
		}
		p.Consume('\n')
		if tc.DebugTrace {
			fmt.Printf("FINAL STATE:\n")
			pretty.Print(p)
		}

		v := p.Produce()
		isEqual, err := Equal(tc.Expected, v)
		if err != nil {
			t.Fatal(err)
		}

		if !isEqual {
			t.Fatalf("Error parsing %s, expected %#v, got %#v", tc.Input, tc.Expected, v)
		}
	}

	for _, tc := range testCases {
		runTest(&tc)
	}
}
