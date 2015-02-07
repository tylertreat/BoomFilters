package boom

import (
	"encoding/binary"
	"hash"
	"hash/fnv"

	"github.com/willf/bitset"
)

// BloomFilter implements a classic Bloom filter. A Bloom filter has a non-zero
// probability of false positives and a zero probability of false negatives.
type BloomFilter struct {
	array *bitset.BitSet // filter data
	hash  hash.Hash64    // hash function (kernel for all k functions)
	m     uint           // filter size
	k     uint           // number of hash functions
}

// NewBloomFilter creates a new Bloom filter optimized to store n items with a
// specified target false-positive rate.
func NewBloomFilter(n uint, fpRate float64) *BloomFilter {
	m := OptimalM(n, fpRate)
	return &BloomFilter{
		array: bitset.New(m),
		hash:  fnv.New64(),
		m:     m,
		k:     OptimalK(fpRate),
	}
}

// Capacity returns the Bloom filter capacity, m.
func (b *BloomFilter) Capacity() uint {
	return b.m
}

// K returns the number of hash functions.
func (b *BloomFilter) K() uint {
	return b.k
}

// FillRatio returns the ratio of set bits.
func (b *BloomFilter) FillRatio() float64 {
	return float64(b.array.Count()) / float64(b.m)
}

// Test will test for membership of the data and returns true if it is a
// member, false if not. This is a probabilistic test, meaning there is a
// non-zero probability of false positives but a zero probability of false
// negatives.
func (b *BloomFilter) Test(data []byte) bool {
	lower, upper := b.hashKernel(data)

	// If any of the K bits are not set, then it's not a member.
	for i := uint(0); i < b.k; i++ {
		if !b.array.Test((uint(lower) + uint(upper)*i) % b.m) {
			return false
		}
	}

	return true
}

// Add will add the data to the Bloom filter. It returns the filter to allow
// for chaining.
func (b *BloomFilter) Add(data []byte) *BloomFilter {
	lower, upper := b.hashKernel(data)

	// Set the K bits.
	for i := uint(0); i < b.k; i++ {
		b.array.Set((uint(lower) + uint(upper)*i) % b.m)
	}

	return b
}

// TestAndAdd is equivalent to calling Test followed by Add. It returns true if
// the data is a member, false if not.
func (b *BloomFilter) TestAndAdd(data []byte) bool {
	lower, upper := b.hashKernel(data)
	member := true

	// If any of the K bits are not set, then it's not a member.
	for i := uint(0); i < b.k; i++ {
		idx := (uint(lower) + uint(upper)*i) % b.m
		if !b.array.Test(idx) {
			member = false
		}
		b.array.Set(idx)
	}

	return member
}

// Reset restores the Bloom filter to its original state. It returns the filter
// to allow for chaining.
func (b *BloomFilter) Reset() *BloomFilter {
	b.array.ClearAll()
	return b
}

// hashKernel returns the upper and lower base hash values from which the k
// hashes are derived.
func (b *BloomFilter) hashKernel(data []byte) (uint32, uint32) {
	b.hash.Write(data)
	sum := b.hash.Sum(nil)
	b.hash.Reset()
	return binary.BigEndian.Uint32(sum[4:8]), binary.BigEndian.Uint32(sum[0:4])
}
