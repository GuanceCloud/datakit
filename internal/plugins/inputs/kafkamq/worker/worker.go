// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package worker group
package worker

import (
	"github.com/IBM/sarama"
)

type Worker func(msg *sarama.ConsumerMessage) error

type WorkerPool struct {
	Threads chan Worker
}

func NewWorkerPool(w Worker, threadCount int) *WorkerPool {
	if threadCount <= 0 {
		return nil
	}
	wp := &WorkerPool{
		Threads: make(chan Worker, threadCount),
	}
	for i := 0; i < threadCount; i++ {
		wp.Threads <- w
	}

	return wp
}

func (wp *WorkerPool) GetWorker() Worker {
	return <-wp.Threads
}

func (wp *WorkerPool) PutWorker(w Worker) {
	wp.Threads <- w
}
