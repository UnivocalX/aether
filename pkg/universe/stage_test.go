package universe

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Test helper functions

func square(ctx context.Context, x int) (int, error) {
	return x * x, nil
}

func double(ctx context.Context, x int) (int, error) {
	return x * 2, nil
}

func failOn(n int) StageFunc[int, int] {
	return func(ctx context.Context, x int) (int, error) {
		if x == n {
			return 0, fmt.Errorf("failed on %d", n)
		}
		return x, nil
	}
}

func slowProcessor(ctx context.Context, x int) (int, error) {
	time.Sleep(50 * time.Millisecond)
	return x * 2, nil
}

// Basic functionality tests
func TestStage_Naked(t *testing.T) {
	stage := NewStage("square", square)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := stage.Run(ctx); err != nil {
		t.Fatalf("stage failed: %v", err)
	}
}
