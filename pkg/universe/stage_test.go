package universe

import (
	"context"
	"testing"
	"time"
)

// Test helper functions
func square(x int) int {
	return x * x
}

func double(x int) int {
	return x * 2
}

// Basic functionality tests
func TestStage_Naked(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // Always cancel to release resources

	Source(ctx, 1, 2, 3, 4, 5)
}
