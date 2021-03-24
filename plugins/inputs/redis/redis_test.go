package redis

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCollectInfoMetrics(t *testing.T) {
	r := Reids{}

	r.client = redis.NewClient(
			&redis.Options{
				Addr:      "127.0.0.1",
				Password:  "dev",
				PoolSize:  1,
		},
	)

	r.collectInfoMetrics()
}