// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package worker group
package worker

import (
	"sync"
	"testing"

	"github.com/IBM/sarama"
)

func TestWorkerPool_GetWorker(t *testing.T) {
	w := func(message *sarama.ConsumerMessage) error {
		t.Logf("message id=%d", message.Offset)
		return nil
	}
	wp := NewWorkerPool(w, 16)
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		thisWorker := wp.GetWorker()
		msg := &sarama.ConsumerMessage{Offset: int64(i)}
		go func(message *sarama.ConsumerMessage) {
			_ = thisWorker(message)
			wp.PutWorker(thisWorker)
			wg.Done()
		}(msg)
	}
	wg.Wait()
}
