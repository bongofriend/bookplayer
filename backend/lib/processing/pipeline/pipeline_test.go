package pipeline_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/processing/pipeline"
)

type mockPipelineHandler struct {
	InputReceived bool
	IsShutdown    bool
}

func (m *mockPipelineHandler) Shutdown() {
	log.Println("Shuttting down MockPipelineHandler")
	m.IsShutdown = true
}

func (m *mockPipelineHandler) Handle(input struct{}, output chan struct{}) error {
	log.Println("Received input for processing")
	m.InputReceived = true
	output <- struct{}{}
	return nil
}

func TestPipelineComponent(t *testing.T) {
	mockHandler := &mockPipelineHandler{}
	component := pipeline.NewPipelineComponent[struct{}, struct{}](mockHandler)
	inputChan := make(chan struct{})
	donePipelineChan := make(chan struct{})
	doneConsumerChan := make(chan struct{})

	context, cancel := context.WithCancel(context.Background())

	outputReceived := false

	go func() {
		defer func() {
			doneConsumerChan <- struct{}{}
		}()
		select {
		case <-context.Done():
			return
		case <-component.OutputChan:
			outputReceived = true
			return
		}
	}()
	component.Start(context, inputChan, donePipelineChan)
	time.Sleep(2 * time.Second)
	inputChan <- struct{}{}

	<-doneConsumerChan
	cancel()
	<-donePipelineChan

	if !outputReceived {
		t.Fatal("Consumer received no output")
	}
	if !mockHandler.IsShutdown {
		t.Fatal("PipelineComponent did not shutdown properly")
	}
	if !mockHandler.InputReceived {
		t.Fatal("PipelineComponent did not receive any input to process")
	}
}

// TODO
func TestAudiobookProcessingPipeline(t *testing.T) {

}
