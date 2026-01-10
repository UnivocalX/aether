package universe

import "context"


type PipelineStage interface {
	Run(ctx context.Context) error
	IsRunnable(ctx context.Context) bool

	IsRunning() bool
	GetStatus() StageState
}

type Pipeline struct {
	stages []PipelineStage
}