package boom

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math"
	"math/rand"
)

// StableBloomFilter implements a Stable Bloom Filter as described by Deng and
// Rafiei in Approximately Detecting Duplicates for Streaming Data using Stable
// Bloom Filters:
//
// http://webdocs.cs.ualberta.ca/~drafiei/papers/DupDet06Sigmod.pdf
//
// A Stable Bloom Filter (SBF) continuously evicts stale information so that it
// has room for more recent elements. Like traditional Bloom filters, an SBF
// has a non-zero probability of false positives, which is controlled by
// several parameters. Unlike the classic Bloom filter, an SBF has a tight
// upper bound on the rate of false positives while introducing a non-zero rate
// of false negatives. The false-positive rate of a classic Bloom filter
// eventually reaches 1, after which all queries result in a false positive.
// The stable-point property of an SBF means the false-positive rate
// asymptotically approaches a configurable fixed constant. A classic Bloom
// filter is actually a special case of SBF where the eviction rate is zero, so
// this package provides support for them as well.
//
// Stable Bloom Filters are useful for cases where the size of the data set
// isn't known a priori, which is a requirement for traditional Bloom filters.
// For example, an SBF can be used to deduplicate events from an unbounded
// event stream with a specified upper bound on false positives and minimal
// false negatives.
type StableBloomFilter struct {
	cells       []uint8
	hash        hash.Hash64
	m           uint
	p           uint
	k           uint
	max         uint8
	indexBuffer []uint
}

// NewStableBloomFilter creates a new Stable Bloom Filter with m cells and k
// hash functions. P indicates the number of cells to decrement in each
// iteration. Use NewDefaultStableFilter if you don't want to calculate
// these parameters.
func NewStableBloomFilter(size, k, p uint, max uint8) *StableBloomFilter {
	if p > size {
		p = size
	}

	if k > size {
		k = size
	}

	return &StableBloomFilter{
		hash:        fnv.New64(),
		m:           size,
		k:           k,
		p:           p,
		max:         max,
		cells:       make([]uint8, size),
		indexBuffer: make([]uint, k),
	}
}

// NewDefaultStableBloomFilter creates a new Stable Bloom Filter which is
// optimized for cases where there is no prior knowledge of the input data
// stream. The upper bound on the rate of false positives is 0.01.
func NewDefaultStableBloomFilter(size uint) *StableBloomFilter {
	return NewStableBloomFilter(size, 3, 10, 1)
}

// NewUnstableBloomFilter creates a new special case of Stable Bloom Filter
// which is a traditional Bloom filter with k hash functions. Unlike the stable
// variant, data is not evicted and a cell contains a maximum of 1 hash value.
func NewUnstableBloomFilter(size, k uint) *StableBloomFilter {
	return NewStableBloomFilter(size, k, 0, 1)
}

// Cells returns the number of cells in the Stable Bloom Filter.
func (s *StableBloomFilter) Cells() uint {
	return s.m
}

// K returns the number of hash functions.
func (s *StableBloomFilter) K() uint {
	return s.k
}

// StablePoint returns the limit of the expected fraction of zeros in the
// Stable Bloom Filter when the number of iterations goes to infinity. When
// this limit is reached, the Stable Bloom Filter is considered stable.
func (s *StableBloomFilter) StablePoint() float64 {
	var (
		subDenom = float64(s.p) * (1/float64(s.k) - 1/float64(s.m))
		denom    = 1 + 1/subDenom
		base     = 1 / denom
	)

	return math.Pow(base, float64(s.max))
}

// FalsePositiveRate returns the upper bound on false positives when the filter
// has become stable.
func (s *StableBloomFilter) FalsePositiveRate() float64 {
	return math.Pow(1-s.StablePoint(), float64(s.k))
}

// Test will test for membership of the data and returns true if it is a
// member, false if not. This is a probabilistic test, meaning there is a
// non-zero probability of false positives and false negatives.
func (s *StableBloomFilter) Test(data []byte) bool {
	lower, upper := s.hashKernel(data)
	member := true

	// If any of the K cells are 0, then it's not a member.
	for i := uint(0); i < s.k; i++ {
		s.indexBuffer[i] = (uint(lower) + uint(upper)*i) % s.m
		if s.cells[s.indexBuffer[i]] == 0 {
			member = false
		}
	}

	return member
}

// Add will add the data to the Stable Bloom Filter. It returns the filter to
// allow for chaining.
func (s *StableBloomFilter) Add(data []byte) *StableBloomFilter {
	// Randomly decrement p cells to make room for new elements.
	s.decrement()

	lower, upper := s.hashKernel(data)

	// Set the K cells to max.
	for i := uint(0); i < s.k; i++ {
		s.cells[(uint(lower)+uint(upper)*i)%s.m] = s.max
	}

	return s
}

// TestAndAdd is equivalent to calling Test followed by Add. It returns true if
// the data is a member, false if not.
func (s *StableBloomFilter) TestAndAdd(data []byte) bool {
	lower, upper := s.hashKernel(data)
	member := true

	// If any of the K cells are 0, then it's not a member.
	for i := uint(0); i < s.k; i++ {
		s.indexBuffer[i] = (uint(lower) + uint(upper)*i) % s.m
		if s.cells[s.indexBuffer[i]] == 0 {
			member = false
		}
	}

	// Randomly decrement p cells to make room for new elements.
	s.decrement()

	// Set the K cells to max.
	for _, idx := range s.indexBuffer {
		s.cells[idx] = s.max
	}

	return member
}

// Reset restores the Stable Bloom Filter to its original state. It returns the
// filter to allow for chaining.
func (s *StableBloomFilter) Reset() *StableBloomFilter {
	for i := uint(0); i < s.m; i++ {
		s.cells[i] = 0
	}

	return s
}

// decrement will decrement a random cell and (p-1) adjacent cells by 1. This
// is faster than generating p random numbers. Although the processes of
// picking the p cells are not independent, each cell has a probability of p/m
// for being picked at each iteration, which means the properties still hold.
func (s *StableBloomFilter) decrement() {
	r := rand.Intn(int(s.m))
	for i := uint(0); i < s.p; i++ {
		idx := (r + int(i)) % int(s.m)
		if s.cells[idx] >= 1 {
			s.cells[idx]--
		}
	}
}

// hashKernel returns the upper and lower base hash values from which the k
// hashes are derived.
func (s *StableBloomFilter) hashKernel(data []byte) (uint32, uint32) {
	s.hash.Write(data)
	sum := s.hash.Sum(nil)
	s.hash.Reset()
	return binary.BigEndian.Uint32(sum[4:8]), binary.BigEndian.Uint32(sum[0:4])
}
