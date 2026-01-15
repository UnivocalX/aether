package universe

import (
	"math"
	"math/rand"
	"runtime"
)

// MaxConcurrency returns the recommended maximum concurrency level.
// For CPU-bound tasks, use runtime.NumCPU() instead.
// For I/O-bound tasks, consider using a higher multiple.
func MaxConcurrency() int {
	return int(math.Ceil(float64(runtime.NumCPU()) * 1.5))
}

// RandomList generates a slice of 'size' random integers
func RandomList(size int) []int {
	list := make([]int, size)
	for i := 0; i < size; i++ {
		list[i] = rand.Int()
	}
	return list
}
