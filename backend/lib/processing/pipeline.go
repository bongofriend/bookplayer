package processing

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
)

// Struct that represents processing pipeline
type Pipeline struct {
	// Dispatch commands that are fanned out to each stage in pipeline
	PipelineCommandChan   chan PipelineCommand
	stageCommandPipelines []chan PipelineCommand
	doneChans             []chan struct{}
}

// A stage in the pipeline
type PipelineStage[Input any, Output any] struct {
	// Channel for external commands
	CommandChan chan PipelineCommand
	// Data to process
	InputChan chan Input
	// Result of data procession
	OutputChan chan Output

	DoneChan chan struct{}

	// Stage specfic logic
	handler PipelineStageHandler[Input, Output]
}

func NewPipelineStage[Input any, Output any](handler PipelineStageHandler[Input, Output]) PipelineStage[Input, Output] {
	return PipelineStage[Input, Output]{
		handler:     handler,
		CommandChan: make(chan PipelineCommand),
		InputChan:   make(chan Input),
		OutputChan:  make(chan Output),
		DoneChan:    make(chan struct{}),
	}
}

// Start stage for processing
func (p PipelineStage[Input, Output]) Start(ctx context.Context) {
	defer func() {
		p.handler.Shutdown()
		p.DoneChan <- struct{}{}
		close(p.CommandChan)
		close(p.InputChan)
		close(p.OutputChan)
	}()
	cmds := p.handler.CommandsToReceive()
	for {
		select {
		// Shutdown by context cancelation fpr whole pipeline
		case <-ctx.Done():
			return
		// Process from input channel
		case input := <-p.InputChan:
			if err := p.handler.ProcessInput(input, p.OutputChan); err != nil {
				log.Println(err)
			}
		// React to external commands
		case cmd := <-p.CommandChan:
			if !slices.Contains(cmds, cmd.CmdType) {
				continue
			}
			if err := p.handler.ProcessCommand(cmd, p.InputChan, p.OutputChan); err != nil {
				log.Println(err)
			}
		}
	}
}

type PipelineStageHandler[Input any, Output any] interface {
	// Handle shutdown
	Shutdown()
	// Process received input
	ProcessInput(Input, chan Output) error
	// Specify commands that this stage may receive
	CommandsToReceive() []PipelineCommandType
	// React to received command
	ProcessCommand(cmd PipelineCommand, inputChan chan Input, outputChan chan Output) error
}

// Command to be dispatched to pipeline stages
type PipelineCommand struct {
	CmdType PipelineCommandType
	Payload interface{}
}

type PipelineCommandType int8

const (
	Scan PipelineCommandType = iota + 1
)

func NewPipeline() Pipeline {
	return Pipeline{
		PipelineCommandChan:   make(chan PipelineCommand),
		stageCommandPipelines: []chan PipelineCommand{},
		doneChans:             []chan struct{}{},
	}

}

func (p Pipeline) initCommandPipeline(context context.Context) {
	defer func() {
		close(p.PipelineCommandChan)
	}()

	for {
		select {
		case <-context.Done():
			return
		case cmd := <-p.PipelineCommandChan:
			for _, ch := range p.stageCommandPipelines {
				ch <- cmd
			}
		}
	}
}

// Assemble and start audiobook processing pipeline
func (p *Pipeline) Start(appContext context.Context, appConfig config.Config, appDoneChan chan struct{}, audiobookRepo repo.AudiobookRepository) {
	context, cancel := context.WithCancel(appContext)
	defer func() {
		cancel()
		for _, ch := range p.doneChans {
			<-ch
		}
		appDoneChan <- struct{}{}
	}()

	// Stage 1: Watch for directory changes every n seconds (as specfied in config)
	watcherHandler := NewDirectoryWatcher(appConfig)
	watcherPipelineStage := NewPipelineStage(&watcherHandler)
	p.stageCommandPipelines = append(p.stageCommandPipelines, watcherPipelineStage.CommandChan)
	p.doneChans = append(p.doneChans, watcherPipelineStage.DoneChan)
	go watcherPipelineStage.Start(context)

	// Stage 2: Extract meta from audiobook file
	metadataExtractorHandler, err := NewMetadataExtractor()
	if err != nil {
		log.Println(err)
		return
	}
	metadataExtractorPipelineStage := NewPipelineStage(metadataExtractorHandler)
	p.stageCommandPipelines = append(p.stageCommandPipelines, metadataExtractorPipelineStage.CommandChan)
	p.doneChans = append(p.doneChans, metadataExtractorPipelineStage.DoneChan)
	go metadataExtractorPipelineStage.Start(context)

	// Stage 3: Split audiobook into seperate chapter files
	chapterSplitterHandler, err := NewChapterSplitter(appConfig)
	if err != nil {
		log.Println(err)
		return
	}
	chapterSplitterPipelineStage := NewPipelineStage(chapterSplitterHandler)
	p.stageCommandPipelines = append(p.stageCommandPipelines, chapterSplitterPipelineStage.CommandChan)
	p.doneChans = append(p.doneChans, chapterSplitterPipelineStage.DoneChan)
	go chapterSplitterPipelineStage.Start(context)

	// Stage 4: Insert processed audiobook information to database
	audiobookSinkHandler := NewAudiobookSink(audiobookRepo)
	audiobookSinkPipelineStage := NewPipelineStage(audiobookSinkHandler)
	p.stageCommandPipelines = append(p.stageCommandPipelines, audiobookSinkPipelineStage.CommandChan)
	p.doneChans = append(p.doneChans, audiobookSinkPipelineStage.DoneChan)
	go audiobookSinkPipelineStage.Start(context)

	ticker := time.NewTicker(appConfig.ScanInterval)
	go func() {
		for {
			select {
			case <-context.Done():
				return
			case <-ticker.C:
				watcherPipelineStage.InputChan <- struct{}{}
			case path := <-watcherPipelineStage.OutputChan:
				metadataExtractorPipelineStage.InputChan <- path
			case metaData := <-metadataExtractorPipelineStage.OutputChan:
				chapterSplitterPipelineStage.InputChan <- metaData
			case processedAudiobook := <-chapterSplitterPipelineStage.OutputChan:
				audiobookSinkPipelineStage.InputChan <- processedAudiobook
			case <-audiobookSinkPipelineStage.OutputChan:
				continue
			}
		}
	}()
	go p.initCommandPipeline(appContext)
}
