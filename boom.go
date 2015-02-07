/*
Package boom implements probabilistic data structures for processing
continuous, unbounded data streams. This includes Stable Bloom Filters,
Scalable Bloom Filters, Inverse Bloom Filters, and several variants of
traditional Bloom filters.

Classic Bloom filters generally require a priori knowledge of the data set
in order to allocate an appropriately sized bit array. This works well for
offline processing, but online processing typically involves unbounded data
streams. With enough data, a traditional Bloom filter "fills up", after
which it has a false-positive probability of 1.

Boom Filters are useful for situations where the size of the data set isn't
known ahead of time. For example, a Stable Bloom Filter can be used to
deduplicate events from an unbounded event stream with a specified upper
bound on false positives and minimal false negatives. Alternatively, an
Inverse Bloom Filter is ideal for deduplicating a stream where duplicate
events are relatively close together. This results in no false positives
and, depending on how close together duplicates are, a small probability of
false negatives. Scalable Bloom Filters place a tight upper bound on false
positives while avoiding false negatives but require allocating memory
proportional to the size of the data set.
*/
package boom

import "math"

const fillRatio = 0.5

// OptimalM calculates the optimal Bloom filter size, m, based on the number of
// items and the desired rate of false positives.
func OptimalM(n uint, fpRate float64) uint {
	return uint(math.Ceil(float64(n) / ((math.Log(fillRatio) *
		math.Log(1-fillRatio)) / math.Abs(math.Log(fpRate)))))
}

// OptimalK calculates the optimal number of hash functions to use for a Bloom
// filter based on the desired rate of false positives.
func OptimalK(fpRate float64) uint {
	return uint(math.Ceil(math.Log2(1 / fpRate)))
}
