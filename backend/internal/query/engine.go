package query

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"fmt"
	"sync"
)

// CommandMetrics tracks performance per command stage
type CommandMetrics struct {
	EventsIn         uint64
	EventsOut        uint64
	EventsDropped    uint64
	ExecutionTimeMs  int64
}

// Processor is the interface for all SPL command implementations
type Processor interface {
	Init(ctx context.Context) error
	Process(ctx context.Context, in <-chan buffer.Event, out chan<- buffer.Event) error
	Close() error
	Metrics() CommandMetrics
}

// PipelineExecutor orchestrates the execution of a series of processors
type PipelineExecutor struct {
	processors []Processor
	meta       NodeMetadata
}

func NewPipelineExecutor(processors []Processor, meta NodeMetadata) *PipelineExecutor {
	return &PipelineExecutor{processors: processors, meta: meta}
}

// Execute runs the pipeline and returns the final results
func (e *PipelineExecutor) Execute(ctx context.Context) ([]buffer.Event, error) {
	if len(e.processors) == 0 {
		return nil, nil
	}

	// Create channels between stages
	// Every stage has a bounded channel to provide backpressure
	channels := make([]chan buffer.Event, len(e.processors)+1)
	for i := range channels {
		channels[i] = make(chan buffer.Event, 1000)
	}

	// For the source (first) stage, we close the input channel immediately 
	// because sources generate data rather than transforming it.
	close(channels[0])

	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	reportErr := func(err error) {
		if err != nil {
			errOnce.Do(func() {
				firstErr = err
			})
		}
	}

	// Start each processor in its own goroutine
	for i, proc := range e.processors {
		wg.Add(1)
		
		// Initialize the processor
		if err := proc.Init(ctx); err != nil {
			reportErr(fmt.Errorf("processor init failed: %w", err))
			wg.Done()
			// We should still close channels and wind down
			continue
		}

		go func(p Processor, in <-chan buffer.Event, out chan<- buffer.Event) {
			defer wg.Done()
			defer p.Close()
			defer close(out)
			
			if err := p.Process(ctx, in, out); err != nil {
				reportErr(err)
			}
		}(proc, channels[i], channels[i+1])
	}

	// The last channel receives the final results
	resultsChan := channels[len(channels)-1]
	var results []buffer.Event

	// Final collector goroutine isn't needed if we block here
	// But we need to ensure we don't block the producers if results are large
	for ev := range resultsChan {
		results = append(results, ev)
	}

	wg.Wait()

	return results, firstErr
}
// ExecuteStream runs the pipeline and streams results through a channel
func (e *PipelineExecutor) ExecuteStream(ctx context.Context) (<-chan buffer.Event, <-chan error) {
	if len(e.processors) == 0 {
		return nil, nil
	}

	channels := make([]chan buffer.Event, len(e.processors)+1)
	for i := range channels {
		channels[i] = make(chan buffer.Event, 1000)
	}

	close(channels[0])

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for i, proc := range e.processors {
		wg.Add(1)
		
		if err := proc.Init(ctx); err != nil {
			errChan <- fmt.Errorf("processor %d init failed: %w", i, err)
			wg.Done()
			continue
		}

		go func(p Processor, in <-chan buffer.Event, out chan<- buffer.Event) {
			defer wg.Done()
			defer p.Close()
			defer close(out)
			
			if err := p.Process(ctx, in, out); err != nil {
				select {
				case errChan <- err:
				default:
				}
			}
		}(proc, channels[i], channels[i+1])
	}

	// Wait for all processors to finish and then close the error channel
	go func() {
		wg.Wait()
		close(errChan)
	}()

	return channels[len(channels)-1], errChan
}
