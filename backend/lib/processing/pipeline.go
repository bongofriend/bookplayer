package processing

import (
	"context"
	"errors"
)

var (
	ErrNoOutputChannel error = errors.New("no output channel available")
)

type PipelineComponent[Input any, Output any] interface {
	Start(ctx context.Context, inputChan chan Input, doneChan chan struct{})
	Shutdown()
	Output() (chan Output, error)
}
