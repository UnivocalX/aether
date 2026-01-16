package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/UnivocalX/aether/pkg/universe"
)

func TestMap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create source: 5000 "hello" strings
	sayHello := func(size int) []string {
		hellos := make([]string, size)
		for i := 0; i < size; i++ {
			hellos[i] = "hello"
		}
		return hellos
	}

	// Create pipeline and run
	upper := universe.TransformValueAdapter((strings.ToUpper))
	source := universe.Source(ctx, sayHello(5000)...)
	out := universe.NewPipeline(universe.Map(upper)).Run(ctx, source)

	// Checker: ensure transformation worked
	expected := "HELLO"
	checker := universe.ConsumeAdapter(
		func(s string) error {
			if s != expected {
				return fmt.Errorf("map transform failed: expected %q, got %q", expected, s)
			}
			return nil
		},
	)

	// Consume all results and fail test if any error occurs
	if err := universe.Consume(ctx, out, checker); err != nil {
		assertNoError(t, err, "Pipeline Map")
	}

	t.Log("Pipeline Map test passed successfully!")
}

// Test Filter operator alone with integer source
func TestFilter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	nums := func(size int) []int {
		out := make([]int, size)
		for i := 0; i < size; i++ {
			out[i] = i
		}
		return out
	}

	// Keep only odd numbers
	odd := universe.Filter(universe.PredicateAdapter(func(n int) bool { return n%2 == 1 }))

	source := universe.Source(ctx, nums(100)...)
	out := universe.NewPipeline(odd).Run(ctx, source)

	count := 0
	checker := universe.ConsumeAdapter(func(n int) error {
		if n%2 == 0 {
			return fmt.Errorf("filter failed: even value %d passed", n)
		}
		count++
		return nil
	})

	if err := universe.Consume(ctx, out, checker); err != nil {
		assertNoError(t, err, "Pipeline Filter")
	}

	assertEqual(t, 50, count, "Pipeline Filter count")

	t.Log("Pipeline Filter test passed successfully!")
}

// Test Tap operator alone using observation side-effects
func TestTap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create source: 200 "hello" strings
	sayHello := func(size int) []string {
		hellos := make([]string, size)
		for i := 0; i < size; i++ {
			hellos[i] = "hello"
		}
		return hellos
	}

	// Buffered channel to collect observations without blocking
	observed := make(chan string, 200)

	tap := universe.Tap(universe.ObserveAdapter(func(s string) { observed <- s }))

	source := universe.Source(ctx, sayHello(200)...)
	out := universe.NewPipeline(tap).Run(ctx, source)

	// Consume results (no-op) and rely on `observed` for side-effect verification
	checker := universe.ConsumeAdapter(func(s string) error { return nil })

	if err := universe.Consume(ctx, out, checker); err != nil {
		assertNoError(t, err, "Pipeline Tap")
	}

	assertEqual(t, 200, len(observed), "Tap observations")

	t.Log("Pipeline Tap test passed successfully!")
}
