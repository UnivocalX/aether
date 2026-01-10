package universe

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

// MaxConcurrency returns the recommended maximum concurrency level.
// For CPU-bound tasks, use runtime.NumCPU() instead.
// For I/O-bound tasks, consider using a higher multiple.
func MaxConcurrency() int {
	return int(math.Ceil(float64(runtime.NumCPU()) * 1.5))
}

type StageState string

const (
	Pending   StageState = "Pending"
	Running   StageState = "Running"
	Aborted   StageState = "Aborted"
	Failed    StageState = "Failed"
	Succeeded StageState = "Succeeded"
)

// StageFunc is the function type that processes stage inputs.
// Implementations should check context cancellation for long-running operations.
type StageFunc[In any, Out any] func(context.Context, In) (Out, error)

// Stage represents a concurrent processing stage in a pipeline.
// 
// Thread-safety: Configuration methods (Durable, Concurrency, etc.) are only safe
// to call before Run(). GetStatus() and IsRunning() are safe to call anytime.
//
// Durable mode: When enabled, errors and panics in individual items don't stop
// the stage. All errors are sent to the Errors channel. When disabled, the first
// error stops processing.
type Stage[In any, Out any] struct {
	name string
	fn   StageFunc[In, Out]

	mu          sync.Mutex
	durable     bool
	concurrency int
	status      StageState

	Queue     chan In
	Artifacts chan Out
	Errors    chan error

	closeOnce sync.Once
	abortOnce sync.Once
}

// NewStage creates a new processing stage with the given name and function.
func NewStage[In any, Out any](name string, fn StageFunc[In, Out]) *Stage[In, Out] {
	return &Stage[In, Out]{
		name:        name,
		fn:          fn,
		concurrency: 1,
		status:      Pending,
		Artifacts:   make(chan Out),
		Errors:      make(chan error, 100),
	}
}

// Durable enables durable mode. In durable mode, errors and panics in processing
// individual items don't stop the stage. All errors are sent to the Errors channel.
// Can only be set before Run() is called.
func (s *Stage[In, Out]) Durable() *Stage[In, Out] {
	if s.GetStatus() == Pending {
		s.durable = true
	}
	return s
}

// Concurrency sets the number of concurrent workers.
// Valid range: 1 to MaxConcurrency(). Values outside this range are clamped.
// Can only be set before Run() is called.
func (s *Stage[In, Out]) Concurrency(c int) *Stage[In, Out] {
	if s.GetStatus() != Pending {
		panic(fmt.Sprintf("stage %q: cannot set concurrency after stage has started", s.name))
	}
	if c < 1 {
		c = 1
	}
	if c > MaxConcurrency() {
		c = MaxConcurrency()
	}
	s.concurrency = c
	return s
}

// BufferedArtifacts sets the buffer size for the Artifacts channel.
// Can only be set before Run() is called.
func (s *Stage[In, Out]) BufferedArtifacts(b int) *Stage[In, Out] {
	if s.GetStatus() != Pending {
		panic(fmt.Sprintf("stage %q: cannot set artifact buffer after stage has started", s.name))
	}
	if b < 0 {
		b = 0
	}
	s.Artifacts = make(chan Out, b)
	return s
}

// BufferedErrors sets the buffer size for the Errors channel.
// When the buffer is full, new errors are dropped.
// Can only be set before Run() is called.
func (s *Stage[In, Out]) BufferedErrors(b int) *Stage[In, Out] {
	if s.GetStatus() != Pending {
		panic(fmt.Sprintf("stage %q: cannot set error buffer after stage has started", s.name))
	}
	if b < 0 {
		b = 0
	}
	s.Errors = make(chan error, b)
	return s
}

// Stream sets the input queue to an existing channel
func (s *Stage[In, Out]) Stream(queue chan In) *Stage[In, Out] {
	if queue == nil {
		panic(fmt.Sprintf("stage %q: input stream is nil", s.name))
	}
	s.Queue = queue
	return s
}

// Populate creates an input queue and populates it with the given items.
// The population happens in a background goroutine.
func (s *Stage[In, Out]) Populate(items ...In) *Stage[In, Out] {
	if len(items) == 0 {
		panic(fmt.Sprintf("stage %q: empty input", s.name))
	}
	
	ch := make(chan In, len(items))
	s.Queue = ch
	
	go func() {
		defer close(ch)
		for _, in := range items {
			ch <- in
		}
	}()
	
	return s
}

// GetStatus returns the current stage status.
// Thread-safe and can be called at any time.
func (s *Stage[In, Out]) GetStatus() StageState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

// IsRunning returns true if the stage is currently running.
// Thread-safe and can be called at any time.
func (s *Stage[In, Out]) IsRunning() bool {
	return s.GetStatus() == Running
}

func (s *Stage[In, Out]) IsRunnable(ctx context.Context) bool {
	if err := s.validate(ctx); err != nil {
		return false
	}
	return true
}

// Run starts the stage and blocks until completion, failure, or context cancellation.
// The context must have a deadline set.
// Returns an error if the stage fails or is aborted.
func (s *Stage[In, Out]) Run(ctx context.Context) error {
	slog.Debug("stage starting", "name", s.name, "concurrency", s.concurrency, "durable", s.durable)

	// Ensure stage hasn't started
	if s.GetStatus() != Pending {
		return fmt.Errorf("stage %q already started", s.name)
	}
	s.updateStatus(Running)

	// Validate stage
	if err := s.validate(ctx); err != nil {
		s.updateStatus(Failed)
		return NewStageError(s.name, err, StageNotRunnable)
	}

	// Start processing
	err := s.startProcessing(ctx)

	// Finalize: update status and close channels
	s.finalize(err, ctx)

	return err
}

// Close closes the Artifacts and Errors channels exactly once.
// Can only be called after Run() has been called.
func (s *Stage[In, Out]) Close() error {
	if s.GetStatus() == Pending {
		return fmt.Errorf("stage %q: cannot close - stage not started", s.name)
	}
	
	s.closeOnce.Do(func() {
		close(s.Artifacts)
		close(s.Errors)
	})
	
	return nil
}

// Collect gathers all artifacts into a slice.
func (s *Stage[In, Out]) Collect() []Out {
	var collection []Out
	for a := range s.Artifacts {
		collection = append(collection, a)
	}
	
	return collection
}

// updateStatus updates the stage status in a thread-safe manner.
// Once a stage is Aborted, the status cannot be changed.
func (s *Stage[In, Out]) updateStatus(newStatus StageState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Once Aborted, do not overwrite
	if s.status == Aborted {
		return
	}

	s.status = newStatus
}

// sendError sends an error to the Errors channel.
// If the channel is full, the error is dropped.
// This method never blocks.
func (s *Stage[In, Out]) sendError(err error) {
	select {
	case s.Errors <- err:
		// Sent successfully
	default:
		// Channel full - drop the error
		slog.Warn("dropped error - channel full", "name", s.name)
	}
}

// abortStage handles stage abortion in a thread-safe manner.
func (s *Stage[In, Out]) abortStage(err error) {
	s.updateStatus(Aborted)
	s.abortOnce.Do(func() {
		s.sendError(NewStageError(s.name, err, StageAborted))
		slog.Debug("stage aborted", "name", s.name, "error", err)
	})
}

// recovery handles panic recovery for stage functions.
// In durable mode, panics are logged and sent as errors but don't stop the stage.
// In non-durable mode, panics are re-raised after logging.
func (s *Stage[In, Out]) recovery() {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic in process %q: %v", s.name, r)
		slog.Error("panic recovered", "name", s.name, "panic", r)
		s.sendError(NewStageError(s.name, err, StagePanic))
		if !s.durable {
			panic(r)
		}
	}
}

// startProcessing launches all workers and waits for completion.
func (s *Stage[In, Out]) startProcessing(ctx context.Context) error {
	g := new(errgroup.Group)
	g.SetLimit(s.concurrency)

	for input := range s.Queue {
		currentInput := input
		g.Go(func() error {
			defer s.recovery()

			err := s.process(ctx, currentInput)
			if err != nil {
				s.sendError(err)
				slog.Debug("stage function error", "name", s.name, "error", err)
			}
			if s.durable {
				return nil
			}
			return err
		})
	}

	return g.Wait()
}

// finalize updates the stage status and closes channels based on the result.
func (s *Stage[In, Out]) finalize(err error, ctx context.Context) {
	switch {
	case ctx.Err() != nil:
		s.abortStage(ctx.Err())

	case err != nil:
		s.updateStatus(Failed)
		slog.Debug("stage failed", "name", s.name, "error", err)

	default:
		s.updateStatus(Succeeded)
		slog.Debug("stage completed successfully", "name", s.name)
	}

	// Close channels
	if closeErr := s.Close(); closeErr != nil {
		slog.Debug("error closing stage", "name", s.name, "error", closeErr)
	}
}

// process executes the stage function for one input.
// Checks for context cancellation before and after processing.
func (s *Stage[In, Out]) process(ctx context.Context, in In) error {
	// Check cancellation before processing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Execute stage function with context
	out, err := s.fn(ctx, in)
	if err != nil {
		return err
	}

	// Send output
	select {
	case s.Artifacts <- out:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// validate ensures the stage is properly configured and ready to run.
func (s *Stage[In, Out]) validate(ctx context.Context) error {
	if _, ok := ctx.Deadline(); !ok {
		return fmt.Errorf("deadline required")
	}
	if s.Queue == nil {
		return fmt.Errorf("input queue is nil")
	}
	if s.fn == nil {
		return fmt.Errorf("stage function is nil")
	}
	return nil
}