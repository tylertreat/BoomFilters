package stable

import (
	"math"
	"strconv"
	"testing"
)

// Ensures that NewBloomFilter clamps p to size.
func TestNewBloomFilterClampP(t *testing.T) {
	f := NewBloomFilter(5, 3, 10, 1)

	if f.p != f.m {
		t.Errorf("Expected %d, got %d", f.m, f.p)
	}
}

// Ensures that NewBloomFilter clamps k to size.
func TestNewBloomFilterClampK(t *testing.T) {
	f := NewBloomFilter(10, 15, 5, 1)

	if f.k != f.m {
		t.Errorf("Expected %d, got %d", f.k, f.p)
	}
}

// Ensures that Cells returns the number of cells, m, in the Stable Bloom
// Filter.
func TestCells(t *testing.T) {
	f := NewBloomFilter(100, 3, 10, 1)

	if cells := f.Cells(); cells != 100 {
		t.Errorf("Expected 100, got %d", cells)
	}
}

// Ensures that K returns the number of hash functions in the Stable Bloom
// Filter.
func TestK(t *testing.T) {
	f := NewBloomFilter(100, 3, 10, 1)

	if k := f.K(); k != 3 {
		t.Errorf("Expected 3, got %d", k)
	}
}

// Ensures that Test, Add, and TestAndAdd behave correctly.
func TestTestAndAdd(t *testing.T) {
	f := NewDefaultBloomFilter(1000)

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

	// `a` should have been evicted.
	if f.Test([]byte(`a`)) {
		t.Error("`a` should not be a member")
	}
}

// Ensures that StablePoint returns the expected fraction of zeros for large
// iterations.
func TestStablePoint(t *testing.T) {
	f := NewDefaultBloomFilter(1000)
	for i := 0; i < 1000000; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	zeros := 0
	for _, cell := range f.cells {
		if cell == 0 {
			zeros++
		}
	}

	actual := round(float64(zeros)/float64(len(f.cells)), 0.5, 1)
	expected := round(f.StablePoint(), 0.5, 1)

	if actual < expected {
		t.Errorf("Expected zeros rate to be greater than or equal to %f, got %f", expected, actual)
	}
}

// Ensures that Reset sets every cell to zero.
func TestReset(t *testing.T) {
	f := NewDefaultBloomFilter(1000)
	for i := 0; i < 1000; i++ {
		f.Add([]byte(strconv.Itoa(i)))
	}

	if f.Reset() != f {
		t.Error("Returned BloomFilter should be the same instance")
	}

	for _, cell := range f.cells {
		if cell != 0 {
			t.Errorf("Expected zero cell, got %d", cell)
		}
	}
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
