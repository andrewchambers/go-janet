/*
Copyright (c) 2017 The Bazel Authors.  All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

1. Redistributions of source code must retain the above copyright
   notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright
   notice, this list of conditions and the following disclaimer in the
   documentation and/or other materials provided with the
   distribution.

3. Neither the name of the copyright holder nor the names of its
   contributors may be used to endorse or promote products derived
   from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package janet

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func TestHashtable(t *testing.T) {
	makeTestIntsOnce.Do(makeTestInts)
	testHashtable(t, make(map[int]bool))
}

func BenchmarkStringHash(b *testing.B) {
	for len := 1; len <= 1024; len *= 2 {
		buf := make([]byte, len)
		rand.New(rand.NewSource(0)).Read(buf)
		s := string(buf)

		b.Run(fmt.Sprintf("hard-%d", len), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				hashString(s)
			}
		})
		b.Run(fmt.Sprintf("soft-%d", len), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				softHashString(s)
			}
		})
	}
}

func BenchmarkHashtable(b *testing.B) {
	makeTestIntsOnce.Do(makeTestInts)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testHashtable(b, nil)
	}
}

const testIters = 10000

var (
	// testInts is a zipf-distributed array of Ints and corresponding ints.
	// This removes the cost of generating them on the fly during benchmarking.
	// Without this, Zipf and MakeInt dominate CPU and memory costs, respectively.
	makeTestIntsOnce sync.Once
	testInts         [3 * testIters]struct {
		Int   Number
		goInt int
	}
)

func makeTestInts() {
	zipf := rand.NewZipf(rand.New(rand.NewSource(0)), 1.1, 1.0, 1000.0)
	for i := range &testInts {
		r := int(zipf.Uint64())
		testInts[i].goInt = r
		testInts[i].Int = Number(float64(r))
	}
}

// testHashtable is both a test and a benchmark of hashtable.
// When sane != nil, it acts as a test against the semantics of Go's map.
func testHashtable(tb testing.TB, sane map[int]bool) {
	var i int // index into testInts

	var ht hashtable

	x := Bool(true)

	// Insert 10000 random ints into the map.
	for j := 0; j < testIters; j++ {
		k := testInts[i]
		i++
		if err := ht.insert(k.Int, x); err != nil {
			tb.Fatal(err)
		}
		if sane != nil {
			sane[k.goInt] = true
		}
	}

	// Do 10000 random lookups in the map.
	for j := 0; j < testIters; j++ {
		k := testInts[i]
		i++
		_, found, err := ht.lookup(k.Int)
		if err != nil {
			tb.Fatal(err)
		}
		if sane != nil {
			_, found2 := sane[k.goInt]
			if found != found2 {
				tb.Fatal("sanity check failed")
			}
		}
	}

	// Do 10000 random deletes from the map.
	for j := 0; j < testIters; j++ {
		k := testInts[i]
		i++
		_, found, err := ht.delete(k.Int)
		if err != nil {
			tb.Fatal(err)
		}
		if sane != nil {
			_, found2 := sane[k.goInt]
			if found != found2 {
				tb.Fatal("sanity check failed")
			}
			delete(sane, k.goInt)
		}
	}

	if sane != nil {
		if int(ht.len) != len(sane) {
			tb.Fatal("sanity check failed")
		}
	}
}
