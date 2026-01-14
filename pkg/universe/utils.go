package universe

import (
	"math"
	"runtime"
)

// MaxConcurrency returns the recommended maximum concurrency level.
// For CPU-bound tasks, use runtime.NumCPU() instead.
// For I/O-bound tasks, consider using a higher multiple.
func MaxConcurrency() int {
	return int(math.Ceil(float64(runtime.NumCPU()) * 1.5))
}
