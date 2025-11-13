package boom

import (
	"crypto/rand"
	"hash/fnv"
	"reflect"
	"testing"
)

func TestHash32DefaultFnv(t *testing.T) {
	var b []byte
	rand.Read(b)

	h := fnv.New32()
	h.Write(b)
	expected := h.Sum32()

	got := hash32DefaultFnv(b, nil)

	if expected != got {
		t.Errorf("Expected %x, got %x", expected, got)
	}
}

func TestHash64DefaultFnv(t *testing.T) {
	var b []byte
	rand.Read(b)

	h := fnv.New64()
	h.Write(b)
	expected := h.Sum64()

	got := hash64DefaultFnv(b, nil)

	if expected != got {
		t.Errorf("Expected %x, got %x", expected, got)
	}
}

func TestHash32BytesDefaultFnv(t *testing.T) {
	var b []byte
	rand.Read(b)

	h := fnv.New32()
	h.Write(b)
	expected := h.Sum(nil)

	got := hash32BytesDefaultFnv(b, nil)

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %x, got %x", expected, got)
	}
}
