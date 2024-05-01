package processing_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/processing"
)

type mockPipelineHandler struct {
	InputReceived bool
	IsShutdown    bool
	CmdReceived   bool
}

func (m *mockPipelineHandler) Shutdown() {
	log.Println("Shuttting down MockPipelineHandler")
	m.IsShutdown = true
}

func (m *mockPipelineHandler) ProcessInput(input struct{}, output chan struct{}) error {
	log.Println("Received input for processing")
	m.InputReceived = true
	output <- struct{}{}
	return nil
}

func (m mockPipelineHandler) CommandsToReceive() []processing.PipelineCommandType {
	return []processing.PipelineCommandType{
		processing.Scan,
	}
}

func (m *mockPipelineHandler) ProcessCommand(cmd processing.PipelineCommand, input chan struct{}, output chan struct{}) error {
	log.Printf("Command of type %d processed", cmd.CmdType)
	m.CmdReceived = true
	return nil
}

func TestPipelineStage(t *testing.T) {
	mockHandler := &mockPipelineHandler{}
	stage := processing.NewPipelineStage[struct{}, struct{}](mockHandler)
	doneConsumerChan := make(chan struct{})
	errChan := make(chan error)

	context, cancel := context.WithCancel(context.Background())

	outputReceived := false

	go func() {
		defer func() {
			doneConsumerChan <- struct{}{}
		}()
		select {
		case <-context.Done():
			return
		case <-stage.OutputChan:
			outputReceived = true
			return
		}
	}()
	go stage.Start(context, errChan)
	time.Sleep(2 * time.Second)
	stage.InputChan <- struct{}{}

	<-doneConsumerChan
	cancel()
	<-stage.DoneChan

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

func TestPipelineStageProcessCommand(t *testing.T) {
	mockHandler := &mockPipelineHandler{}
	stage := processing.NewPipelineStage[struct{}, struct{}](mockHandler)
	context, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)

	go stage.Start(context, errChan)
	stage.CommandChan <- processing.PipelineCommand{
		CmdType: processing.Scan,
		Payload: struct{}{},
	}

	time.Sleep(1 * time.Second)
	cancel()
	<-stage.DoneChan

	if !mockHandler.CmdReceived {
		t.Fatal("No cmd processed")
	}
}

// TODO Implement
func TestAudiobookProcessingPipeline(t *testing.T) {
	t.SkipNow()
}
