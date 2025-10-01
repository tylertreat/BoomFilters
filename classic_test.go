package boom

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"testing"

	"github.com/d4l3k/messagediff"
)

// Ensures that Capacity returns the number of bits, m, in the Bloom filter.
func TestBloomCapacity(t *testing.T) {
	f := NewBloomFilter(100, 0.1)

	if capacity := f.Capacity(); capacity != 480 {
		t.Errorf("Expected 480, got %d", capacity)
	}
}

// Ensures that K returns the number of hash functions in the Bloom Filter.
func TestBloomK(t *testing.T) {
	f := NewBloomFilter(100, 0.1)

	if k := f.K(); k != 4 {
		t.Errorf("Expected 4, got %d", k)
	}
}

// Ensures that Count returns the number of items added to the filter.
func TestBloomCount(t *testing.T) {
	f := NewBloomFilter(100, 0.1)
	for i := 0; i < 10; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	if count := f.Count(); count != 10 {
		t.Errorf("Expected 10, got %d", count)
	}
}

// Ensures that EstimatedFillRatio returns the correct approximation.
func TestBloomEstimatedFillRatio(t *testing.T) {
	f := NewBloomFilter(100, 0.5)
	for i := 0; i < 100; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	if ratio := f.EstimatedFillRatio(); ratio > 0.5 {
		t.Errorf("Expected less than or equal to 0.5, got %f", ratio)
	}
}

// Ensures that FillRatio returns the ratio of set bits.
func TestBloomFillRatio(t *testing.T) {
	f := NewBloomFilter(100, 0.1)
	f.Add([]byte(`a`))
	f.Add([]byte(`b`))
	f.Add([]byte(`c`))

	if ratio := f.FillRatio(); ratio != 0.025 {
		t.Errorf("Expected 0.025, got %f", ratio)
	}
}

// Ensures that Test, Add, and TestAndAdd behave correctly.
func TestBloomTestAndAdd(t *testing.T) {
	f := NewBloomFilter(100, 0.01)

	// `a` isn't in the filter.
	if f.Test([]byte(`a`)) {
		t.Error("`a` should not be a member")
	}

	if f.Add([]byte(`a`)) != f {
		t.Error("Returned BloomFilter should be the same instance")
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
func TestBloomReset(t *testing.T) {
	f := NewBloomFilter(100, 0.1)
	for i := 0; i < 1000; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	if f.Reset() != f {
		t.Error("Returned BloomFilter should be the same instance")
	}

	for i := uint(0); i < f.buckets.Count(); i++ {
		if f.buckets.Get(i) != 0 {
			t.Error("Expected all bits to be unset")
		}
	}
}

// Ensures that BloomFilter can be serialized and deserialized without errors.
func TestBloomFilter_EncodeDecode(t *testing.T) {
	f := NewBloomFilter(1000, 0.1)

	for i := 0; i < 1000; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(f); err != nil {
		t.Error(err)
	}

	f2 := &BloomFilter{}
	if err := gob.NewDecoder(&buf).Decode(f2); err != nil {
		t.Error(err)
	}

	if diff, equal := messagediff.PrettyDiff(f, f2); !equal {
		t.Errorf("BloomFilter Gob Encode and Decode = %+v; not %+v\n%s", f2, f, diff)
	}

	if len(f.buckets.data) != len(f2.buckets.data) {
		t.Errorf("BloomFilter has different sized data after encode/decode")
	}

	for i := 0; i < len(f.buckets.data); i++ {
		if f.buckets.data[i] != f2.buckets.data[i] {
			t.Errorf("BloomFilter has different data after encode/decode")
		}
	}
}

// TestBloomFilter_ReadFrom tests that ReadFrom correctly deserializes
// a bloom filter and initializes all necessary fields including hash
func TestBloomFilter_ReadFrom(t *testing.T) {
	// Create and populate a bloom filter
	f := NewBloomFilter(1000, 0.01)
	f.Add([]byte("test1"))
	f.Add([]byte("test2"))
	f.Add([]byte("test3"))
	
	// Serialize using WriteTo
	var buf bytes.Buffer
	_, err := f.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	
	// Deserialize using ReadFrom
	f2 := &BloomFilter{}
	_, err = f2.ReadFrom(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	
	// Verify the deserialized filter works correctly
	if !f2.Test([]byte("test1")) || !f2.Test([]byte("test2")) || !f2.Test([]byte("test3")) {
		t.Error("ReadFrom failed to properly restore filter state")
	}
	
	if f2.Test([]byte("test4")) {
		t.Error("ReadFrom produced false positive for item not in original")
	}
}

func BenchmarkBloomAdd(b *testing.B) {
	b.StopTimer()
	f := NewBloomFilter(100000, 0.1)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.Add(data[n])
	}
}

func BenchmarkBloomTest(b *testing.B) {
	b.StopTimer()
	f := NewBloomFilter(100000, 0.1)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.Test(data[n])
	}
}

func BenchmarkBloomTestAndAdd(b *testing.B) {
	b.StopTimer()
	f := NewBloomFilter(100000, 0.1)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		f.TestAndAdd(data[n])
	}
}
