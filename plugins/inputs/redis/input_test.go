package redis

import (
	"fmt"
	"testing"
)

func TestCollect(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
	}

	input.initCfg()

	input.Collect()

	for _, obj := range input.collectCache {
		fmt.Println("obj =====>", obj)
	}
}
