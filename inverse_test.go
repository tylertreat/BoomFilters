package boom

import (
	"bytes"
	"encoding/gob"
	"github.com/d4l3k/messagediff"
	"strconv"
	"testing"
)

// Ensures that Capacity returns the correct filter size.
func TestInverseCapacity(t *testing.T) {
	f := NewInverseBloomFilter(100)

	if c := f.Capacity(); c != 100 {
		t.Errorf("expected 100, got %d", c)
	}
}

// Ensures that TestAndAdd behaves correctly.
func TestInverseTestAndAdd(t *testing.T) {
	f := NewInverseBloomFilter(3)

	if f.TestAndAdd([]byte(`a`)) {
		t.Error("`a` should not be a member")
	}

	if !f.TestAndAdd([]byte(`a`)) {
		t.Error("`a` should be a member")
	}

	// `d` hashes to the same index as `a`.
	if f.TestAndAdd([]byte(`d`)) {
		t.Error("`d` should not be a member")
	}

	// `a` was swapped out.
	if f.TestAndAdd([]byte(`a`)) {
		t.Error("`a` should not be a member")
	}

	if !f.TestAndAdd([]byte(`a`)) {
		t.Error("`a` should be a member")
	}

	// `b` hashes to another index.
	if f.TestAndAdd([]byte(`b`)) {
		t.Error("`b` should not be a member")
	}

	if !f.TestAndAdd([]byte(`b`)) {
		t.Error("`b` should be a member")
	}

	// `a` should still be a member.
	if !f.TestAndAdd([]byte(`a`)) {
		t.Error("`a` should be a member")
	}

	if f.Test([]byte(`c`)) {
		t.Error("`c` should not be a member")
	}

	if f.Add([]byte(`c`)) != f {
		t.Error("Returned InverseBloomFilter should be the same instance")
	}

	if !f.Test([]byte(`c`)) {
		t.Error("`c` should be a member")
	}
}

// Tests that an InverseBloomFilter can be encoded and decoded properly without error
func TestInverseBloomFilter_Encode(t *testing.T) {
	f := NewInverseBloomFilter(10000)

	for i := 0; i < 1000; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(f); err != nil {
		t.Error(err)
	}
	//
	f2 := NewInverseBloomFilter(10000)
	if err := gob.NewDecoder(&buf).Decode(f2); err != nil {
		t.Error(err)
	}

	if f.capacity != f2.capacity {
		t.Errorf("Different capacities")
	}

	if len(f.array) != len(f2.array) {
		t.Errorf("Different data")
	}

	if diff, equal := messagediff.PrettyDiff(f.array, f2.array); !equal {
		t.Errorf("BloomFilter Gob Encode and Decode = %+v; not %+v\n%s", f2, f, diff)
	}

	for i := 0; i < 100000; i++ {
		if f.Test([]byte(strconv.Itoa(i))) != f2.Test([]byte(strconv.Itoa(i))) {
			t.Errorf("Expected both filters to Test the same for %i", i)
		}
	}

}

func BenchmarkInverseAdd(b *testing.B) {
	b.StopTimer()
	f := NewInverseBloomFilter(100000)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.Add(data[n])
	}
}

func BenchmarkInverseTest(b *testing.B) {
	b.StopTimer()
	f := NewInverseBloomFilter(100000)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
		f.Add(data[i])
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.Test(data[n])
	}
}

func BenchmarkInverseTestAndAdd(b *testing.B) {
	b.StopTimer()
	f := NewInverseBloomFilter(100000)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.TestAndAdd(data[n])
	}
}
