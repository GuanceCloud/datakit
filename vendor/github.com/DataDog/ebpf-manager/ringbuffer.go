package manager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
)

type RingBufferOptions struct {
	// RingBufferSize - Size in bytes of the ring buffer. Defaults to the manager value if not set.
	RingBufferSize int

	// ErrChan - Reader error channel
	ErrChan chan error

	// DataHandler - Callback function called when a new sample was retrieved from the perf
	// ring buffer.
	DataHandler func(CPU int, data []byte, ringBuffer *RingBuffer, manager *Manager)
}

type RingBuffer struct {
	manager    *Manager
	ringReader *ringbuf.Reader
	wgReader   sync.WaitGroup

	// Map - A PerfMap has the same features as a normal Map
	Map
	RingBufferOptions
}

// loadNewRingBuffer - Creates a new ring buffer map instance, loads it and sets up the ring buffer reader
func loadNewRingBuffer(spec ebpf.MapSpec, options MapOptions, ringBufferOptions RingBufferOptions) (*RingBuffer, error) {
	// Create underlying map
	innerMap, err := loadNewMap(spec, options)
	if err != nil {
		return nil, err
	}

	// Create the new map
	ringBuffer := RingBuffer{
		Map:               *innerMap, //nolint:govet
		RingBufferOptions: ringBufferOptions,
	}
	return &ringBuffer, nil
}

// init - Initialize a ring buffer
func (rb *RingBuffer) init(manager *Manager) error {
	rb.manager = manager

	if rb.DataHandler == nil {
		return fmt.Errorf("no DataHandler set for %s", rb.Name)
	}

	// Set default values if not already set
	if rb.RingBufferSize == 0 {
		rb.RingBufferSize = manager.options.DefaultRingBufferSize
	}

	// Initialize the underlying map structure
	if err := rb.Map.init(manager); err != nil {
		return err
	}
	return nil
}

// Start - Starts fetching events on a perf ring buffer
func (rb *RingBuffer) Start() error {
	rb.stateLock.Lock()
	defer rb.stateLock.Unlock()
	if rb.state == running {
		return nil
	}
	if rb.state < initialized {
		return ErrMapNotInitialized
	}

	// Create and start the perf map
	var err error
	if rb.ringReader, err = ringbuf.NewReader(rb.array); err != nil {
		return err
	}
	// Start listening for data
	rb.wgReader.Add(1)

	go func() {
		var record ringbuf.Record
		var err error

		for {
			if err = rb.ringReader.ReadInto(&record); err != nil {
				if isRingBufferClosed(err) {
					rb.wgReader.Done()
					return
				}
				if rb.ErrChan != nil {
					rb.ErrChan <- err
				}
				continue
			}
			rb.DataHandler(0, record.RawSample, rb, rb.manager)
		}
	}()

	rb.state = running
	return nil
}

// Stop - Stops the perf ring buffer
func (rb *RingBuffer) Stop(cleanup MapCleanupType) error {
	rb.stateLock.Lock()
	defer rb.stateLock.Unlock()
	if rb.state <= stopped {
		return nil
	}
	rb.state = stopped

	// close ring reader
	err := rb.ringReader.Close()

	rb.wgReader.Wait()

	// close underlying map
	if errTmp := rb.Map.close(cleanup); errTmp != nil {
		if err == nil {
			err = errTmp
		} else {
			err = fmt.Errorf("%s: %w", err.Error(), errTmp)
		}
	}

	return err
}

func isRingBufferClosed(err error) bool {
	return errors.Is(err, ringbuf.ErrClosed)
}
