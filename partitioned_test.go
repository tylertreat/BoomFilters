package boom

import (
	"strconv"
	"testing"
)

// Ensures that Capacity returns the number of bits, m, in the Bloom filter.
func TestPartitionedBloomCapacity(t *testing.T) {
	f := NewPartitionedBloomFilter(100, 0.1)

	if capacity := f.Capacity(); capacity != 480 {
		t.Errorf("Expected 480, got %d", capacity)
	}
}

// Ensures that K returns the number of hash functions in the Stable Bloom
// Filter.
func TestPartitionedBloomK(t *testing.T) {
	f := NewPartitionedBloomFilter(100, 0.1)

	if k := f.K(); k != 4 {
		t.Errorf("Expected 4, got %d", k)
	}
}

// Ensures that Test, Add, and TestAndAdd behave correctly.
func TestPartitionedBloomTestAndAdd(t *testing.T) {
	f := NewPartitionedBloomFilter(100, 0.01)

	// `a` isn't in the filter.
	if f.Test([]byte(`a`)) {
		t.Error("`a` should not be a member")
	}

	if f.Add([]byte(`a`)) != f {
		t.Error("Returned PartitionedBloomFilter should be the same instance")
	}

	// `a` is now in the filter.
	if !f.Test([]byte(`a`)) {
		t.Error("`a` should be a member")
	}

	// `a` is still in the filter.
	if !f.TestAndAdd([]byte(`a`)) {
		t.Error("`a` should be a member")
	}

	// `b` is not in the filter.
	if f.TestAndAdd([]byte(`b`)) {
		t.Error("`b` should not be a member")
	}

	// `a` is still in the filter.
	if !f.Test([]byte(`a`)) {
		t.Error("`a` should be a member")
	}

	// `b` is now in the filter.
	if !f.Test([]byte(`b`)) {
		t.Error("`b` should be a member")
	}

	// `c` is not in the filter.
	if f.Test([]byte(`c`)) {
		t.Error("`c` should not be a member")
	}

	for i := 0; i < 1000000; i++ {
		f.TestAndAdd([]byte(strconv.Itoa(i)))
	}

	// `x` should be a false positive.
	if !f.Test([]byte(`x`)) {
		t.Error("`x` should be a member")
	}
}

// Ensures that Reset sets every bit to zero.
func TestPartitionedBloomReset(t *testing.T) {
	f := NewPartitionedBloomFilter(100, 0.1)
	for i := 0; i < 1000; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	if f.Reset() != f {
		t.Error("Returned BloomFilter should be the same instance")
	}

	for _, partition := range f.partitions {
		if partition.Any() {
			t.Error("Expected all bits to be unset")
		}
	}
}

func BenchmarkPartitionedBloomAdd(b *testing.B) {
	b.StopTimer()
	f := NewPartitionedBloomFilter(100000, 0.1)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.Add(data[n])
	}
}

func BenchmarkPartitionedBloomTest(b *testing.B) {
	b.StopTimer()
	f := NewPartitionedBloomFilter(100000, 0.1)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.Test(data[n])
	}
}

func BenchmarkPartitionedBloomTestAndAdd(b *testing.B) {
	b.StopTimer()
	f := NewPartitionedBloomFilter(100000, 0.1)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.TestAndAdd(data[n])
	}
}
