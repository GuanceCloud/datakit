// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package bytechannel provides a thread-safe channel for sending and receiving byte slices
package bytechannel

import (
	"context"
	"sync/atomic"
	"time"
)

type ByteChannel interface {
	Send(context.Context, []byte) bool
	SendSync(context.Context, []byte)
	Receive() []byte
	CurrentChannelSize() int64
	CurrentByteSize() int64
}

type option struct {
	chanSize int
	capacity int64
}

type Option func(*option)

func WithChanSize(n int) Option   { return func(opt *option) { opt.chanSize = n } }
func WithCapacity(n int64) Option { return func(opt *option) { opt.capacity = n } }

func defaultOption() *option {
	return &option{
		chanSize: 512,
		capacity: 10 * 1024 * 1024, // 10MB
	}
}

type byteChannel struct {
	ch       chan []byte
	capacity int64

	currentChanSize *atomic.Int64
	currentSize     *atomic.Int64
}

func NewByteChannel(opts ...Option) ByteChannel {
	c := defaultOption()
	for _, opt := range opts {
		opt(c)
	}

	return &byteChannel{
		ch:              make(chan []byte, c.chanSize),
		capacity:        c.capacity,
		currentChanSize: &atomic.Int64{},
		currentSize:     &atomic.Int64{},
	}
}

func (bc *byteChannel) Send(ctx context.Context, data []byte) bool {
	if bc.currentSize.Load()+int64(len(data)) > bc.capacity {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	case bc.ch <- data:
		bc.currentChanSize.Add(1)
		bc.currentSize.Add(int64(len(data)))
		return true
	default:
		return false
	}
}

func (bc *byteChannel) SendSync(ctx context.Context, data []byte) {
	sended := false
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if bc.currentSize.Load()+int64(len(data)) < bc.capacity {
				select {
				case <-ctx.Done():
					return
				case bc.ch <- data:
					bc.currentChanSize.Add(1)
					bc.currentSize.Add(int64(len(data)))
					sended = true
				}
			}
			if !sended {
				time.Sleep(time.Millisecond * 10)
			} else {
				return
			}
		}
	}
}

func (bc *byteChannel) Receive() []byte {
	select {
	case data := <-bc.ch:
		bc.currentChanSize.Add(-1)
		bc.currentSize.Add(-int64(len(data)))
		return data
	default:
		return nil
	}
}

func (bc *byteChannel) CurrentChannelSize() int64 {
	return bc.currentChanSize.Load()
}

func (bc *byteChannel) CurrentByteSize() int64 {
	return bc.currentSize.Load()
}
