package boom

import (
	"encoding/binary"
	"hash/fnv"
	"testing"
)

func BenchmarkHashKernel(b *testing.B) {
	hsh := fnv.New64()
	var data [4]byte

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint32(data[:], uint32(i))
		hashKernel(data[:], hsh)
	}
}
