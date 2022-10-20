// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package workerpool used to control concurrency goroutines
package workerpool

import (
	"context"
	"errors"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

type WorkerPoolConfig struct {
	Buffer  int `json:"buffer"`  // worker pool channel buffer size
	Threads int `json:"threads"` // goroutines count in total
}

type Process func(input interface{}) (output interface{})

type ProcessCallback func(input, output interface{}, cost time.Duration)

type JobOption func(job *Job)

func WithInput(input interface{}) JobOption {
	return func(job *Job) {
		job.input = input
	}
}

func WithProcess(p Process) JobOption {
	return func(job *Job) {
		job.p = p
	}
}

func WithProcessCallback(cb ProcessCallback) JobOption {
	return func(job *Job) {
		job.cb = cb
	}
}

type Job struct {
	input interface{}
	p     Process
	cb    ProcessCallback
}

func NewJob(options ...JobOption) (*Job, error) {
	job := &Job{}
	for i := range options {
		options[i](job)
	}
	if job.p == nil {
		return nil, errors.New("process can not be nil")
	}

	return job, nil
}

type WorkerPool struct {
	wkpConf *WorkerPoolConfig
	log     *logger.Logger
	jobs    chan *Job
	exit    *cliutils.Sem
	enabled bool
}

func NewWorkerPool(config *WorkerPoolConfig, log *logger.Logger) (*WorkerPool, error) {
	if config == nil || config.Buffer < 0 || config.Threads < 1 {
		return nil, errors.New("worker-pool config error")
	}

	wkp := &WorkerPool{
		wkpConf: config,
		log:     log,
		jobs:    make(chan *Job, config.Buffer),
		exit:    cliutils.NewSem(),
		enabled: false,
	}
	if wkp.log == nil {
		wkp.log = logger.DefaultSLogger("worker-pool")
	}

	return wkp, nil
}

func (wkp *WorkerPool) Start() error {
	if wkp.enabled {
		return errors.New("worker-pool is already enabled")
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "internal_trace"})
	for i := 0; i < wkp.wkpConf.Threads; i++ {
		g.Go(func(ctx context.Context) error {
			wkp.worker()

			return nil
		})
	}
	wkp.enabled = true

	return nil
}

func (wkp *WorkerPool) Shutdown() {
	close(wkp.jobs)
	wkp.enabled = false
}

func (wkp *WorkerPool) Enabled() bool {
	return wkp.enabled
}

func (wkp *WorkerPool) MoreJob(job *Job) error {
	if !wkp.enabled {
		return errors.New("woker-pool not start")
	}
	if job == nil {
		return errors.New("job is nil")
	}

	select {
	case wkp.jobs <- job:
		return nil
	default:
		return errors.New("worker-pool busy")
	}
}

func (wkp *WorkerPool) worker() {
	if wkp == nil {
		return
	}

	for {
		select {
		case <-wkp.exit.Wait():
			wkp.log.Infof("on exit, worker-pool worker exits")

			return
		case <-datakit.Exit.Wait():
			wkp.log.Infof("on datakit exit, woker-pool worker exits")

			return
		default:
		}

		job, ok := <-wkp.jobs
		if !ok {
			break
		}
		if job == nil || job.p == nil {
			continue
		}

		start := time.Now()
		rslt := job.p(job.input)
		if job.cb != nil {
			job.cb(job.input, rslt, time.Since(start))
		}
	}
}
