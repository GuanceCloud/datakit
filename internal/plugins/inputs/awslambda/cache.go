// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

type lambdaCtx struct {
	funDurationMs float64
	flag          int
}

const (
	flagStart = 1 << iota
	flagRuntimeDone
	flagReport
	flagAll = 0b111
)

type lambdaCtxID struct {
	ID string
	*lambdaCtx
}

func (c *lambdaCtxCache) Get(id string) (*lambdaCtx, bool) {
	v, ok := c.q.ContainsFunc(func(ctx *lambdaCtxID) bool {
		return ctx.ID == id
	})
	if ok {
		return v.lambdaCtx, ok
	}
	return nil, false
}

func (c *lambdaCtxCache) Set(id string, ctx *lambdaCtx) {
	_, _ = c.q.DeleteFunc(func(ctx *lambdaCtxID) bool {
		return ctx.ID == id
	})
	c.q.Enqueue(&lambdaCtxID{
		ID:        id,
		lambdaCtx: ctx,
	})
}

func (c *lambdaCtxCache) Del(id string) {
	_, _ = c.q.DeleteFunc(func(ctx *lambdaCtxID) bool {
		return ctx.ID == id
	})
}

func newLambdaCtxCache() *lambdaCtxCache {
	// https://docs.aws.amazon.com/zh_cn/lambda/latest/dg/lambda-concurrency.html
	// a lambda will only have one request at a time, so it doesn't need much space.
	return &lambdaCtxCache{
		q: NewCircularQueueDefaultCap[*lambdaCtxID](),
	}
}

type lambdaCtxCache struct {
	q *CircularQueue[*lambdaCtxID]
}
