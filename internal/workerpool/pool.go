// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package workerpool used to control concurrency goroutines
package workerpool

import (
	"errors"
	"time"
)

type WorkerPoolConfig struct {
	Buffer  int `json:"buffer"`  // worker pool channel buffer size
	Threads int `json:"threads"` // goroutines count in total
	Timeout int `json:"timeout"` // job timeout
}
type Process func(input interface{}) (output interface{})

type ProcessCallback func(input, output interface{}, cost time.Duration, isTimeout bool)

type Job struct {
	input   interface{}
	timeout time.Duration
	p       Process
	cb      ProcessCallback
}

type Option func(job *Job)

func WithInput(input interface{}) Option {
	return func(job *Job) {
		job.input = input
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(job *Job) {
		job.timeout = timeout
	}
}

func WithProcess(p Process) Option {
	return func(job *Job) {
		job.p = p
	}
}

func WithProcessCallback(cb ProcessCallback) Option {
	return func(job *Job) {
		job.cb = cb
	}
}

func NewJob(options ...Option) (*Job, error) {
	job := &Job{}
	for i := range options {
		options[i](job)
	}
	if job.p == nil {
		return nil, errors.New("process can not be nil")
	}

	return job, nil
}

type WorkerPool chan *Job

func NewWorkerPool(buffer int) WorkerPool {
	if buffer < 0 {
		return nil
	}

	return make(WorkerPool, buffer)
}

func (wp WorkerPool) Start(threads int) error {
	if wp == nil {
		return errors.New("worker pool is not ready")
	}
	if threads < 1 {
		return errors.New("worker pool needs at least one thread")
	}

	for i := 0; i < threads; i++ {
		go wp.worker()
	}

	return nil
}

func (wp WorkerPool) Shutdown() {
	close(wp)
}

func (wp WorkerPool) MoreJob(job *Job) error {
	if job == nil {
		return nil
	}

	select {
	case wp <- job:
		return nil
	default:
		return errors.New("worker pool busy")
	}
}

func (wp WorkerPool) worker() {
	if wp == nil {
		return
	}

	for {
		job, ok := <-wp
		if !ok {
			break
		}
		if job == nil {
			continue
		}

		start := time.Now()
		if job.timeout < 1 {
			r := job.p(job.input)
			if job.cb != nil {
				job.cb(job.input, r, time.Since(start), false)
			}
		} else {
			result := make(chan interface{}, 1)
			go func(job *Job, r chan interface{}) {
				r <- job.p(job.input)
				close(r)
			}(job, result)

			tick := time.NewTicker(job.timeout)
			select {
			case <-tick.C:
				if job.cb != nil {
					job.cb(job.input, nil, time.Since(start), true)
				}
			case r := <-result:
				if job.cb != nil {
					job.cb(job.input, r, time.Since(start), false)
				}
				tick.Stop()
			}
		}
	}
}
