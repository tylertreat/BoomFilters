package boom

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"strconv"
	"testing"
)

func inverseStreamFixture(t *testing.T) (*InverseBloomFilter, []byte) {
	t.Helper()
	f := NewInverseBloomFilter(100000)
	for n := 0; n < 1000; n++ {
		f.Add([]byte(strconv.Itoa(n)))
	}
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return f, buf.Bytes()
}

func TestInverseStreamRoundTrip(t *testing.T) {
	original, serialized := inverseStreamFixture(t)
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	if _, err := zw.Write(serialized); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"bytes", "fragmented", "gzip"} {
		for _, method := range []string{"ReadFrom", "ImportElementsFrom"} {
			t.Run(mode+"/"+method, func(t *testing.T) {
				var reader io.Reader = bytes.NewReader(serialized)
				if mode == "fragmented" {
					reader = &inverseChunkReader{r: reader, size: 3}
				}
				if mode == "gzip" {
					zr, err := gzip.NewReader(bytes.NewReader(compressed.Bytes()))
					if err != nil {
						t.Fatal(err)
					}
					defer zr.Close()
					reader = zr
				}
				f := NewInverseBloomFilter(100000)
				if method == "ReadFrom" {
					n, err := f.ReadFrom(reader)
					if err != nil {
						t.Fatal(err)
					}
					if n != int64(len(serialized)) {
						t.Fatalf("read %d bytes, want %d", n, len(serialized))
					}
				} else if _, err := f.ImportElementsFrom(reader); err != nil {
					t.Fatal(err)
				}
				for n := 0; n < 2000; n++ {
					key := []byte(strconv.Itoa(n))
					if f.Test(key) != original.Test(key) {
						t.Fatalf("membership changed for %q", key)
					}
				}
			})
		}
	}
}

type inverseChunkReader struct {
	r    io.Reader
	size int
}

func (r *inverseChunkReader) Read(p []byte) (int, error) {
	if len(p) > r.size {
		p = p[:r.size]
	}
	return r.r.Read(p)
}

type inverseErrorReader struct{ err error }

func (r inverseErrorReader) Read(p []byte) (int, error) { return 0, r.err }

func TestInverseStreamDecodeErrors(t *testing.T) {
	_, serialized := inverseStreamFixture(t)
	sentinel := errors.New("read failed")
	malformed := append([]byte(nil), serialized...)
	for n := 16; n < len(malformed); n++ {
		malformed[n] = 0xff
	}
	var mismatched bytes.Buffer
	binary.Write(&mismatched, binary.BigEndian, uint64(3))
	binary.Write(&mismatched, binary.BigEndian, uint64(len(serialized)-16))
	mismatched.Write(serialized[16:])
	for _, method := range []string{"ReadFrom", "ImportElementsFrom", "GobDecode"} {
		for _, tc := range []struct {
			name string
			data []byte
			tail error
		}{
			{"truncated header", serialized[:7], nil},
			{"truncated payload", serialized[:len(serialized)-1], nil},
			{"malformed gob", malformed, nil},
			{"mismatched capacity", mismatched.Bytes(), nil},
			{"read error", serialized[:16], sentinel},
		} {
			if method == "GobDecode" && tc.tail != nil {
				continue
			}
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				f := NewInverseBloomFilter(7)
				f.Add([]byte("retained"))
				var reader io.Reader = bytes.NewReader(tc.data)
				if tc.tail != nil {
					reader = io.MultiReader(reader, inverseErrorReader{tc.tail})
				}
				var err error
				switch method {
				case "ReadFrom":
					_, err = f.ReadFrom(reader)
				case "ImportElementsFrom":
					_, err = f.ImportElementsFrom(reader)
				case "GobDecode":
					err = f.GobDecode(tc.data)
				}
				if err == nil {
					t.Fatal("expected decode error")
				}
				if tc.tail != nil && err != sentinel {
					t.Fatalf("got %v, want original error", err)
				}
				if f.Capacity() != 7 || !f.Test([]byte("retained")) {
					t.Fatal("failed decode changed filter")
				}
			})
		}
	}
}

func TestInverseStreamReadFromLeavesNextRecord(t *testing.T) {
	original, serialized := inverseStreamFixture(t)
	joined := append(append([]byte(nil), serialized...), serialized...)
	reader := bytes.NewReader(joined)
	for n := 0; n < 2; n++ {
		var f InverseBloomFilter
		if _, err := f.ReadFrom(reader); err != nil {
			t.Fatal(err)
		}
		if f.Capacity() != original.Capacity() || !f.Test([]byte("999")) {
			t.Fatal("record changed")
		}
	}
	if reader.Len() != 0 {
		t.Fatalf("%d bytes remain", reader.Len())
	}
}
