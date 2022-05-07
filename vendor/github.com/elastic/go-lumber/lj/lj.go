// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package lj implements common lumberjack types and functions.
package lj

// Batch is an ACK-able batch of events as has been received by lumberjack
// server implemenentations. Batches must be ACKed, for the server
// implementations returning an ACK to it's clients.
type Batch struct {
	Events []interface{}
	ack    chan struct{}
}

// NewBatch creates a new ACK-able batch.
func NewBatch(evts []interface{}) *Batch {
	return &Batch{evts, make(chan struct{})}
}

// ACK acknowledges a batch initiating propagation of ACK to clients.
func (b *Batch) ACK() {
	close(b.ack)
}

// Await returns a channel for waiting for a batch to be ACKed.
func (b *Batch) Await() <-chan struct{} {
	return b.ack
}
