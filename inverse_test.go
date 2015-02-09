/*
Original work Copyright (c) 2012 Jeff Hodges. All rights reserved.
Modified work Copyright (c) 2015 Tyler Treat. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Jeff Hodges nor the names of this project's
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package boom

import (
	"strconv"
	"testing"
)

func TestTheBasics(t *testing.T) {
	f, _ := NewInverseBloomFilter(2)
	twentyNineId := []byte{27, 28, 29}
	thirtyId := []byte{27, 28, 30}
	thirtyFiveId := []byte{27, 28, 35}
	shouldNotContain(t, "nothing should be contained at all", f, twentyNineId)
	shouldContain(t, "now it should", f, twentyNineId)
	shouldNotContain(t, "false unless the hash collides", f, thirtyId)
	shouldContain(t, "original should still return true", f, twentyNineId)
	shouldContain(t, "new array should still return true", f, thirtyId)

	// Handling collisions. {27, 28, 35} and {27, 28, 30} hash to the same
	// index using the current hash function inside InverseBloomFilter.
	shouldNotContain(t, "colliding array returns false", f, thirtyFiveId)
	shouldContain(t,
		"colliding array returns true in second call", f, thirtyFiveId)
	shouldNotContain(t, "original colliding array returns false", f, thirtyId)
	shouldContain(t, "original colliding array returns true", f, thirtyId)
	shouldNotContain(t, "colliding array returns false", f, thirtyFiveId)
}

func TestSizeRounding(t *testing.T) {
	f, _ := NewInverseBloomFilter(3)
	if f.Size() != 4 {
		t.Errorf("3 should round to 4, rounded to: ", f.Size())
	}
	f, _ = NewInverseBloomFilter(4)
	if f.Size() != 4 {
		t.Errorf("4 should round to 4", f.Size())
	}
	f, _ = NewInverseBloomFilter(129)
	if f.Size() != 256 {
		t.Errorf("129 should round to 256", f.Size())
	}
}

func TestTooLargeSize(t *testing.T) {
	size := (1 << 30) + 1
	f, err := NewInverseBloomFilter(size)
	if err == nil {
		t.Errorf("did not error out on a too-large filter size")
	}
	if f != nil {
		t.Errorf("did not return nil on a too-large filter size")
	}
}

func TestTooSmallSize(t *testing.T) {
	f, err := NewInverseBloomFilter(0)
	if err == nil {
		t.Errorf("did not error out on a too small filter size")
	}
	if f != nil {
		t.Errorf("did not return nil on a too small filter size")
	}
}

func BenchmarkObserve(b *testing.B) {
	f, _ := NewInverseBloomFilter(100000)
	for n := 0; n < b.N; n++ {
		f.Observe([]byte(strconv.Itoa(n)))
	}
}

func shouldContain(t *testing.T, msg string, f *InverseBloomFilter, id []byte) {
	if !f.Observe(id) {
		t.Errorf("should contain, %s: id %v, array: %v", msg, id, f.array)
	}
}

func shouldNotContain(t *testing.T, msg string, f *InverseBloomFilter, id []byte) {
	if f.Observe(id) {
		t.Errorf("should not contain, %s: %v", msg, id)
	}
}
