package universe

import "fmt"

type StageError struct {
	stage string
	cause error
	kind  StageErrorKind
}

type StageErrorKind uint8

const (
	StageNotRunnable StageErrorKind = iota
	StageExecutionFailed
	StageAborted
	StagePanic
)

func (e *StageError) Error() string {
	switch e.kind {
	case StageNotRunnable:
		return fmt.Sprintf("%q is not runnable: %v", e.stage, e.cause)
	case StageExecutionFailed:
		return fmt.Sprintf("%q failed: %v", e.stage, e.cause)
	case StageAborted:
		return fmt.Sprintf("%q aborted: %v", e.stage, e.cause)
	case StagePanic:
		return fmt.Sprintf("%q panic: %v", e.stage, e.cause)
	default:
		return fmt.Sprintf("%q error: %v", e.stage, e.cause)
	}
}

func (e *StageError) Unwrap() error {
	return e.cause
}

func (e *StageError) Stage() string {
	return e.stage
}

func (e *StageError) Kind() StageErrorKind {
	return e.kind
}

func NewStageError(stage string, cause error, kind StageErrorKind) error {
	return &StageError{
		stage: stage,
		cause: cause,
		kind:  kind,
	}
}