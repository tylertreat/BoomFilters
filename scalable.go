package boom

import (
	"hash"
	"hash/fnv"
)

// ScalableBloomFilter implements a Scalable Bloom Filter as described by
// Almeida, Baquero, Preguica, and Hutchison in Scalable Bloom Filters:
//
// http://gsd.di.uminho.pt/members/cbm/ps/dbloom.pdf
//
// A Scalable Bloom Filter dynamically adapts to the number of elements in the
// data set while enforcing a tight upper bound on the false-positive rate.
// This works by adding Bloom filters with geometrically decreasing
// false-positive rates as filters become full. The tightening ratio, r,
// controls the filter growth.
//
// Scalable Bloom Filters are useful for cases where the size of the data set
// isn't known a priori and memory constraints aren't of particular concern.
// For situations where memory is bounded, refer to Inverse and Stable Bloom
// Filters.
type ScalableBloomFilter struct {
	filters []*PartitionedBloomFilter // filters with geometrically decreasing error rates
	hash    hash.Hash                 // hash function (kernel for all k functions)
	r       float32                   // tightening ratio
	p       float64                   // target false-positive rate
	n       uint                      // filter size hint
}

// NewScalableBloomFilter creates a new Scalable Bloom Filter with the
// specified target false-positive rate and tightening ratio.
func NewScalableBloomFilter(n uint, fpRate float64, r float32) *ScalableBloomFilter {
	s := &ScalableBloomFilter{
		filters: make([]*PartitionedBloomFilter, 0, 1),
		hash:    fnv.New64(),
		r:       r,
		p:       fpRate,
	}

	s.addBloomFilter()
	return s
}

// addBloomFilter adds a new Bloom filter with a restricted false-positive rate
// to the Scalable Bloom Filter
func (s *ScalableBloomFilter) addBloomFilter() {
	// TODO
}
