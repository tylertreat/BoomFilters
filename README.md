# Boom Filters

Boom Filters are probabilistic data structures for processing continuous, unbounded data streams in Go. There are currently two structures implemented: Stable Bloom Filter and Inverse Bloom Filter. This also provides a traditional Bloom filter, which is actually a special case of a Stable Bloom Filter.

Classic Bloom filters generally require a priori knowledge of the data set in order to allocate an appropriately sized bit array. This works well for offline processing, but online processing typically involves unbounded data streams. With enough data, a traditional Bloom filter "fills up", after which it has a false-positive probability of 1.

Boom Filters are useful for cases where the size of the data set isn't known ahead of time. For example, a Stable Bloom Filter can be used to deduplicate events from an unbounded event stream with a specified upper bound on false positives and minimal false negatives. Alternatively, an Inverse Bloom Filter is ideal for deduplicating a stream where duplicate events are relatively close together. This results in no false positives and, depending on how close together duplicates are, a small probability of false negatives.

For documentation, see [godoc](http://godoc.org/github.com/tylertreat/BoomFilters).

## Installation 

```
$ go get github.com/tylertreat/BoomFilters
```

## Stable Bloom Filter

This is an implementation of Stable Bloom Filters as described by Deng and Rafiei in [Approximately Detecting Duplicates for Streaming Data using Stable Bloom Filters](http://webdocs.cs.ualberta.ca/~drafiei/papers/DupDet06Sigmod.pdf).

A Stable Bloom Filter (SBF) continuously evicts stale information so that it has room for more recent elements. Like traditional Bloom filters, an SBF has a non-zero probability of false positives, which is controlled by several parameters. Unlike the classic Bloom filter, an SBF has a tight upper bound on the rate of false positives while introducing a non-zero rate of false negatives. The false-positive rate of a classic Bloom filter eventually reaches 1, after which all queries result in a false positive. The stable-point property of an SBF means the false-positive rate asymptotically approaches a configurable fixed constant. A classic Bloom filter is actually a special case of SBF where the eviction rate is zero and the cell size is one, so this provides support for them as well.

### Usage

```go
package main

import (
    "fmt"
    "github.com/tylertreat/BoomFilters"
)

func main() {
    sbf := boom.NewDefaultStableBloomFilter(10000)
    fmt.Println("stable point", sbf.StablePoint())
    
    sbf.Add([]byte(`a`))
    if sbf.Test([]byte(`a`)) {
        fmt.Println("contains a")
    }
    
    if !sbf.TestAndAdd([]byte(`b`)) {
        fmt.Println("doesn't contain b")
    }
    
    if sbf.Test([]byte(`b`)) {
        fmt.Println("now it contains b!")
    }
    
    // Restore to initial state.
    sbf.Reset()
}
```

## Inverse Bloom Filter

An Inverse Bloom Filter, or "the opposite of a Bloom filter", is a concurrent, probabilistic data structure used to test whether an item has been observed or not. This implementation, [originally described and written by Jeff Hodges](http://www.somethingsimilar.com/2012/05/21/the-opposite-of-a-bloom-filter/), replaces the use of MD5 hashing with a non-cryptographic FNV-1a function.

The Inverse Bloom Filter may report a false negative but can never report a false positive. That is, it may report that an item has not been seen when it actually has, but it will never report an item as seen which it hasn't come across. This behaves in a similar manner to a fixed-size hashmap which does not handle conflicts.

This structure is particularly well-suited to streams in which duplicates are relatively close together.

### Usage

```go
package main

import (
    "fmt"
    "github.com/tylertreat/BoomFilters"
)

func main() {
    ibf := boom.NewInverseBloomFilter(10000)
    
    if !ibf.Observe([]byte(`a`)) {
        fmt.Println("haven't observed a")
    }
    
    if ibf.Observe([]byte(`a`)) {
        fmt.Println("observed a")
    }
}
```
