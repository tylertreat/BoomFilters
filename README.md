# Boom Filters

Boom Filters contains probabilistic data structures for processing continuous, unbounded data streams in Go. There are currently two structures implemented: Stable Bloom Filter and Inverse Bloom Filter. This also provides a traditional Bloom filter, which is actually a special case of a Stable Bloom Filter.

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
