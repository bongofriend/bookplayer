package pipeline

import (
	"context"
	"log"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
	"github.com/bongofriend/bookplayer/backend/lib/data/repo"
	"github.com/bongofriend/bookplayer/backend/lib/processing/audiobooksink"
	"github.com/bongofriend/bookplayer/backend/lib/processing/chaptersplitter"
	"github.com/bongofriend/bookplayer/backend/lib/processing/directorywatcher"
	"github.com/bongofriend/bookplayer/backend/lib/processing/metadataextractor"
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

type AudiobookProcessingPipeline struct {
	doneChan chan struct{}
}

func NewAudiobookProcessingPipeline() AudiobookProcessingPipeline {
	return AudiobookProcessingPipeline{
		doneChan: make(chan struct{}, 4),
	}
}

// TODO How to handle communication between processing stage, pipeline and external event sources?
func (a *AudiobookProcessingPipeline) Start(appContext context.Context, config config.Config) error {
	dbClient, err := repo.NewDbClient(config.Database)
	if err != nil {
		return err
	}
	directoryWatcher := directorywatcher.NewDirectoryWatcher(config)
	metadataExtractor, err := metadataextractor.NewMetadataExtractor()
	if err != nil {
		return err
	}
	chapterSplitter, err := chaptersplitter.NewChapterSplitter(config)
	if err != nil {
		return err
	}
	audiobookSink := audiobooksink.NewAudiobookSink(repo.NewAudiobookRepository(dbClient))

	ticker := time.NewTicker(config.ScanInterval)
	directoryWatcherComponent := NewPipelineComponent(&directoryWatcher)
	metadataExtractorComponent := NewPipelineComponent(metadataExtractor)
	chapterSplitterComponent := NewPipelineComponent(chapterSplitter)
	audiobookSinkComponent := NewPipelineComponent(audiobookSink)

	directoryWatcherComponent.Start(appContext, ticker.C, a.doneChan)
	metadataExtractorComponent.Start(appContext, directoryWatcherComponent.OutputChan, a.doneChan)
	chapterSplitterComponent.Start(appContext, metadataExtractorComponent.OutputChan, a.doneChan)
	audiobookSinkComponent.Start(appContext, chapterSplitterComponent.OutputChan, a.doneChan)

	return nil
}
