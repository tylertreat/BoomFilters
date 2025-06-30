package boom

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

// Ensures that TotalCount returns the number of items added to the sketch.
func TestCMSTotalCount(t *testing.T) {
	cms := NewCountMinSketch(0.001, 0.99)

	for i := 0; i < 100; i++ {
		cms.Add([]byte(strconv.Itoa(i)))
	}

	if count := cms.TotalCount(); count != 100 {
		t.Errorf("expected 100, got %d", count)
	}
}

// Ensures that AddN adds n items to the set and Count returns the correct
// approximation.
func TestCMSAddNAndCount(t *testing.T) {
	cms := NewCountMinSketch(0.001, 0.99)

	if cms.AddN([]byte(`a`), 5) != cms {
		t.Error("Returned CountMinSketch should be the same instance")
	}

	cms.AddN([]byte(`b`), 3)
	cms.AddN([]byte(`c`), 1)
	cms.AddN([]byte(`b`), 2)
	cms.AddN([]byte(`d`), 1)
	cms.AddN([]byte(`a`), 2).AddN([]byte(`a`), 1)

	if count := cms.Count([]byte(`a`)); count != 8 {
		t.Errorf("expected 8, got %d", count)
	}

	if count := cms.Count([]byte(`b`)); count != 5 {
		t.Errorf("expected 5, got %d", count)
	}

	if count := cms.Count([]byte(`c`)); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	if count := cms.Count([]byte(`d`)); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	if count := cms.Count([]byte(`x`)); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	if totalCount := cms.TotalCount(); totalCount != 15 {
		t.Errorf("expected total count 15, got %d", totalCount)
	}
}

// Ensures that Add adds to the set and Count returns the correct
// approximation.
func TestCMSAddAndCount(t *testing.T) {
	cms := NewCountMinSketch(0.001, 0.99)

	if cms.Add([]byte(`a`)) != cms {
		t.Error("Returned CountMinSketch should be the same instance")
	}

	cms.Add([]byte(`b`))
	cms.Add([]byte(`c`))
	cms.Add([]byte(`b`))
	cms.Add([]byte(`d`))
	cms.Add([]byte(`a`)).Add([]byte(`a`))

	if count := cms.Count([]byte(`a`)); count != 3 {
		t.Errorf("expected 3, got %d", count)
	}

	if count := cms.Count([]byte(`b`)); count != 2 {
		t.Errorf("expected 2, got %d", count)
	}

	if count := cms.Count([]byte(`c`)); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	if count := cms.Count([]byte(`d`)); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	if count := cms.Count([]byte(`x`)); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

// Ensures that Merge combines the two sketches.
func TestCMSMerge(t *testing.T) {
	cms := NewCountMinSketch(0.001, 0.99)
	cms.Add([]byte(`b`))
	cms.Add([]byte(`c`))
	cms.Add([]byte(`b`))
	cms.Add([]byte(`d`))
	cms.Add([]byte(`a`)).Add([]byte(`a`))

	other := NewCountMinSketch(0.001, 0.99)
	other.Add([]byte(`b`))
	other.Add([]byte(`c`))
	other.Add([]byte(`b`))

	if err := cms.Merge(other); err != nil {
		t.Error(err)
	}

	if count := cms.Count([]byte(`a`)); count != 2 {
		t.Errorf("expected 2, got %d", count)
	}

	if count := cms.Count([]byte(`b`)); count != 4 {
		t.Errorf("expected 4, got %d", count)
	}

	if count := cms.Count([]byte(`c`)); count != 2 {
		t.Errorf("expected 2, got %d", count)
	}

	if count := cms.Count([]byte(`d`)); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	if count := cms.Count([]byte(`x`)); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

// Ensures that Reset restores the sketch to its original state.
func TestCMSReset(t *testing.T) {
	cms := NewCountMinSketch(0.001, 0.99)
	cms.Add([]byte(`b`))
	cms.Add([]byte(`c`))
	cms.Add([]byte(`b`))
	cms.Add([]byte(`d`))
	cms.Add([]byte(`a`)).Add([]byte(`a`))

	if cms.Reset() != cms {
		t.Error("Returned CountMinSketch should be the same instance")
	}

	for i := uint(0); i < cms.depth; i++ {
		for j := uint(0); j < cms.width; j++ {
			if x := cms.matrix[i][j]; x != 0 {
				t.Errorf("expected matrix to be completely empty, got %d", x)
			}
		}
	}
}

// Test binary serialization
func TestCMSSerialization(t *testing.T) {
	freq := 73
	epsilon, delta := 0.001, 0.99
	cms := NewCountMinSketch(epsilon, delta)
	a := []byte(`a`)
	for i := 0; i < freq; i++ {
		cms.Add(a)

	}
	if count := cms.Count(a); count != uint64(freq) {
		t.Errorf("expected %d, got %d\n", freq, count)
	}

	buf := new(bytes.Buffer)
	// serialize
	wn, err := cms.WriteDataTo(buf)
	if err != nil {
		t.Errorf("unexpected error bytes written %d: %v", wn, err)
	}

	blankCMS := NewCountMinSketch(epsilon, delta)
	// deserialize
	rn, err := blankCMS.ReadDataFrom(buf)
	if err != nil {
		t.Errorf("readfrom err %s bytes read %d", err, rn)
	}
	if wn != rn {
		t.Errorf("expected %d, got %d\n", wn, rn)
	}
	// check correctness
	if count := blankCMS.Count(a); count != uint64(freq) {
		t.Errorf("expected %d, got %d\n", freq, count)
	}

	// serialize
	wn, err = cms.WriteDataTo(buf)
	if err != nil {
		t.Errorf("unexpected error bytes written %d: %v", wn, err)
	}
	wrongCMS := NewCountMinSketch(epsilon+0.01, delta)
	rn, err = wrongCMS.ReadDataFrom(buf)

	if !strings.Contains(err.Error(), "cms values") {
		t.Errorf("unexpected error %v", err)
	}

}

func TestCMSTestAndRemove(t *testing.T) {
	cms := NewCountMinSketch(0.001, 0.99)

	cms.Add([]byte(`b`))
	cms.Add([]byte(`c`))
	cms.Add([]byte(`b`))
	cms.Add([]byte(`d`))
	cms.Add([]byte(`a`)).Add([]byte(`a`)).Add([]byte(`a`)).Add([]byte(`a`))

	if try := cms.TestAndRemove([]byte("a"), 5); try {
		t.Errorf("expected false, got %t", try)
	}

	if count := cms.Count([]byte("a")); count != 4 {
		t.Errorf("expected 4, got %d", count)
	}

	if try := cms.TestAndRemove([]byte("a"), 3); !try {
		t.Errorf("expected true, got %t", try)
	}

	if count := cms.Count([]byte("a")); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	if try := cms.TestAndRemoveAll([]byte("z")); try {
		t.Errorf("expected false, got %t", try)
	}

	cms.Add([]byte(`a`)).Add([]byte(`a`)).Add([]byte(`a`))

	if try := cms.TestAndRemoveAll([]byte("a")); !try {
		t.Errorf("expected true, got %t", try)
	}

	if count := cms.Count([]byte("a")); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func BenchmarkCMSWriteDataTo(b *testing.B) {
	b.StopTimer()
	freq := 73
	epsilon, delta := 0.001, 0.99
	cms := NewCountMinSketch(epsilon, delta)
	a := []byte(`a`)
	for i := 0; i < freq; i++ {
		cms.Add(a)

	}
	var buf bytes.Buffer
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := cms.WriteDataTo(&buf)
		if err != nil {
			b.Errorf("unexpected error %s\n", err)
		}
	}

}

func BenchmarkCMSReadDataFrom(b *testing.B) {
	b.StopTimer()
	b.N = 10000
	freq := 73
	epsilon, delta := 0.001, 0.99
	cms := NewCountMinSketch(epsilon, delta)
	a := []byte(`a`)
	for i := 0; i < freq; i++ {
		cms.Add(a)

	}
	var buf bytes.Buffer
	_, err := cms.WriteDataTo(&buf)
	if err != nil {
		b.Errorf("unexpected error %s\n", err)
	}
	data := make([]byte, 0)
	for i := 0; i < b.N; i++ {
		data = append(data, buf.Bytes()...)
	}
	rd := bytes.NewReader(data)
	newCMS := NewCountMinSketch(epsilon, delta)
	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := newCMS.ReadDataFrom(rd)
		if err != nil {
			b.Errorf("unexpected error %s\n", err)
		}
	}

}

func BenchmarkCMSAdd(b *testing.B) {
	b.StopTimer()
	cms := NewCountMinSketch(0.001, 0.99)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		cms.Add(data[n])
	}
}

func BenchmarkCMSCount(b *testing.B) {
	b.StopTimer()
	cms := NewCountMinSketch(0.001, 0.99)
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = []byte(strconv.Itoa(i))
		cms.Add([]byte(strconv.Itoa(i)))
	}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		cms.Count(data[n])
	}
}

func BenchmarkCMSReset(b *testing.B) {
	b.StopTimer()
	cms := NewCountMinSketch(0.0001, 0.1)
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		cms.Reset()
	}
}
