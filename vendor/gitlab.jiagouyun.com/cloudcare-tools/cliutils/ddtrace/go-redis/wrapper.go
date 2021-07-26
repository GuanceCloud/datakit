package redis

import (
	"context"

	"github.com/go-redis/redis"
)

type TraceableClient interface {
	redis.Cmdable
	Context() context.Context
	Options() *redis.Options
	SetLimiter(l redis.Limiter) *redis.Client
	PoolStats() *redis.PoolStats
	Pipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	Pipeline() redis.Pipeliner
	TxPipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	TxPipeline() redis.Pipeliner
	Subscribe(channels ...string) *redis.PubSub
	PSubscribe(channels ...string) *redis.PubSub
}
