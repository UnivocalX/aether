package universe

import (
	"context"
	"testing"
	"time"
)

func increment(x int) int {
	return x + 1
}

func isOdd(x int) bool {
	if x % 2 == 1 {
		return true
	}

	return false
}

// Basic functionality tests
func TestStage_All(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	Incrementor := Map(Adapter(increment))
	OddFilter := Filter(isOdd)

	pipeline := NewPipeline(ctx, Generator(ctx, RandomList(5000)...))
	pipeline.Merge(
		pipeline.
			Then(Incrementor).
			Then(OddFilter).
			UntilDone().
			Scatter(Incrementor, 4)...
	)
}