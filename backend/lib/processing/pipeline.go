package processing

import (
	"context"
	"errors"
	"log"
)

var (
	ErrNoOutputChannel error = errors.New("no output channel available")
)

type PipelineHandler[Input any, Output any] interface {
	Shutdown()
	Handle(Input, chan Output) error
}

type PipelineComponent[Input any, Output any] struct {
	handler    PipelineHandler[Input, Output]
	OutputChan chan Output
}

func NewPipelineComponent[Input any, Output any](handler PipelineHandler[Input, Output]) *PipelineComponent[Input, Output] {
	return &PipelineComponent[Input, Output]{
		handler:    handler,
		OutputChan: make(chan Output),
	}
}

func (p PipelineComponent[Input, Output]) Start(ctx context.Context, inputChan <-chan Input, doneChan chan struct{}) {
	go func() {
		defer func() {
			close(p.OutputChan)
			doneChan <- struct{}{}
		}()
		for {
			select {
			case <-ctx.Done():
				p.handler.Shutdown()
				return
			case input := <-inputChan:
				err := p.handler.Handle(input, p.OutputChan)
				if err != nil {
					log.Println(err)
				}

			}
		}
	}()
}
