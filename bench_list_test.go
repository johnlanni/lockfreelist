package lockfreelist

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
)

func BenchmarkList(b *testing.B) {
	length := 1 << 12
	inputs := make([]int, length)
	for i := 0; i < length; i++ {
		inputs = append(inputs, rand.Int())
	}

	for _, cpus := range []int{4, 32, 1024} {
		runtime.GOMAXPROCS(cpus)
		var list LockfreeList
		b.Run(strconv.Itoa(cpus), func(b *testing.B) {
			b.ResetTimer()
			var c int64
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					i := int(atomic.AddInt64(&c, 1)-1) % length
					v := inputs[i]
					if v >= 0 {
						list.PushBack(v)
					} else {
						list.PopFront()
					}
				}
			})
		})
	}
}
