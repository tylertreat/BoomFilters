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
