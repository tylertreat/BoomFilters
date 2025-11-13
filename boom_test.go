package boom

import (
	"encoding/binary"
	"testing"
)

func BenchmarkHashKernel(b *testing.B) {
	var data [4]byte

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint32(data[:], uint32(i))
		hashKernel(data[:], nil)
	}
}
