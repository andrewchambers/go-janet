package janet

import (
	"bytes"
)

type MarshalState struct {
	buf      *bytes.Buffer
	seen     Table
	rreg     *Table
	seenEnvs **FuncEnv
	seenDefs **FuncDef
	nextid   int32
}

func entryGetVal(envEntry Value) Value {
	switch envEntry := envEntry.(type) {
	case *Table:
		checkVal := envEntry.Get(Keyword("value"))
		if checkVal == nil {
			checkVal = envEntry.Get(Keyword("ref"))
		}
		return checkVal
	case *Struct:
		checkVal := envEntry.Get(Keyword("value"))
		if checkVal == nil {
			checkVal = envEntry.Get(Keyword("ref"))
		}
		return checkVal
	default:
		return nil
	}
}
