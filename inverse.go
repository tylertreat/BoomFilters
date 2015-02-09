/*
Original work Copyright (c) 2012 Jeff Hodges. All rights reserved.
Modified work Copyright (c) 2015 Tyler Treat. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Jeff Hodges nor the names of this project's
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package boom

import (
	"bytes"
	"errors"
	"hash"
	"hash/fnv"
	"math"
	"sync/atomic"
	"unsafe"
)

// maxSize indicates the largest possible filter size.
const maxSize = 1 << 30

// InverseBloomFilter is a concurrent "inverse" Bloom filter, which is
// effectively the opposite of a classic Bloom filter. This was originally
// described and written by Jeff Hodges:
//
// http://www.somethingsimilar.com/2012/05/21/the-opposite-of-a-bloom-filter/
//
// The InverseBloomFilter may report a false negative but can never report a
// false positive. That is, it may report that an item has not been seen when
// it actually has, but it will never report an item as seen which it hasn't
// come across. This behaves in a similar manner to a fixed-size hashmap which
// does not handle conflicts.
//
// An example use case is deduplicating events while processing a stream of
// data. Ideally, duplicate events are relatively close together.
type InverseBloomFilter struct {
	array    []*[]byte
	sizeMask uint32
	hash     *uintHash
}

// NewInverseBloomFilter creates and returns a new InverseBloomFilter with the
// specified capacity. It returns an error if the size is not between 0 and
// 2^30.
func NewInverseBloomFilter(size int) (*InverseBloomFilter, error) {
	if size > maxSize {
		return nil, errors.New("Size too large to round to a power of 2")
	}

	if size <= 0 {
		return nil, errors.New("Size must be greater than 0")
	}

	// Round to the next largest power of two.
	size = int(math.Pow(2, math.Ceil(math.Log2(float64(size)))))
	slice := make([]*[]byte, size)
	sizeMask := uint32(size - 1)
	return &InverseBloomFilter{slice, sizeMask, &uintHash{fnv.New32a()}}, nil
}

// Observe marks a key as observed. It returns true if the key has been
// previously observed and false if the key has possibly not been observed
// yet. It may report a false negative but will never report a false positive.
// That is, it may return false even though the key was previously observed,
// but it will never return true for a key that has never been observed.
func (i *InverseBloomFilter) Observe(key []byte) bool {
	i.hash.Write(key)
	uindex := i.hash.Sum32() & i.sizeMask
	i.hash.Reset()
	oldID := getAndSet(i.array, int32(uindex), key)
	return bytes.Equal(oldID, key)
}

// Size returns the filter length.
func (i *InverseBloomFilter) Size() int {
	return len(i.array)
}

type uintHash struct {
	hash.Hash
}

func (u uintHash) Sum32() uint32 {
	sum := u.Sum(nil)
	x := uint32(sum[0])
	for _, val := range sum[1:3] {
		x = x << 3
		x += uint32(val)
	}
	return x
}

// getAndSet returns the key that was in the slice at the given index after
// putting the new key in the slice at that index, atomically.
func getAndSet(arr []*[]byte, index int32, key []byte) []byte {
	indexPtr := (*unsafe.Pointer)(unsafe.Pointer(&arr[index]))
	keyUnsafe := unsafe.Pointer(&key)
	var oldKey []byte
	for {
		oldKeyUnsafe := atomic.LoadPointer(indexPtr)
		if atomic.CompareAndSwapPointer(indexPtr, oldKeyUnsafe, keyUnsafe) {
			oldKeyPtr := (*[]byte)(oldKeyUnsafe)
			if oldKeyPtr != nil {
				oldKey = *oldKeyPtr
			}
			break
		}
	}
	return oldKey
}
