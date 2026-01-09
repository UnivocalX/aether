package pipelines

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

type StageFunc[In any, Out any] func(In) (Out, error)

type Artifacts[T any] chan T

type Stage[In any, Out any] struct {
	name string
	fn   StageFunc[In, Out]

	Input  Artifacts[In]
	Output Artifacts[Out]

}

func NewStage[In any, Out any](
	name string,
	fn StageFunc[In, Out],
) *Stage[In, Out] {
	return &Stage[In, Out]{
		name:   name,
		fn:     fn,
		Output: make(Artifacts[Out]),
	}
}

// WithInput makes this stage a source stage.
func (s *Stage[In, Out]) WithInput(inputs ...In) *Stage[In, Out] {
	if len(inputs) == 0 {
		panic(fmt.Sprintf("stage %q: empty input", s.name))
	}

	ch := make(Artifacts[In], len(inputs))
	s.Input = ch

	go func() {
		defer close(ch)
		for _, in := range inputs {
			ch <- in
		}
	}()

	return s
}

// WithStream connects this stage to an upstream stage.
func (s *Stage[In, Out]) WithStream(stream Artifacts[In]) *Stage[In, Out] {
	if stream == nil {
		panic(fmt.Sprintf("stage %q: input stream is nil", s.name))
	}
	s.Input = stream
	return s
}


// Run executes a single stage instance.
func (s *Stage[In, Out]) Run(ctx context.Context) error {
	if s.Input == nil {
		return fmt.Errorf("stage %q: input not configured", s.name)
	}

	slog.Debug("stage starting", "name", s.name)
	g, ctx := errgroup.WithContext(ctx)

	for input := range s.Input {
		in := input

		g.Go(func() error {
			select {
			case <-ctx.Done():
				slog.Debug("stage cancelled before fn", "name", s.name)
				return ctx.Err()
			default:
			}

			out, err := s.fn(in)
			if err != nil {
				slog.Debug("stage function error", "name", s.name, "error", err)
				return fmt.Errorf("stage %q failed: %w", s.name, err)
			}

			select {
			case s.Output <- out:
				return nil
			case <-ctx.Done():
				slog.Debug("stage cancelled during send", "name", s.name)
				return ctx.Err()
			}
		})
	}

	err := g.Wait()
	close(s.Output)

	if err != nil {
		slog.Debug("stage completed with error", "name", s.name)
		return err
	}

	slog.Debug("stage completed successfully", "name", s.name)
	return nil
}

