// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cache"
)

const (
	checkpointFileName = "logtail_checkpoint.bolt"
	checkpointBucket   = "logtail_checkpoint"
)

var (
	defaultCachePath = datakit.DataDir
	globalCheckpoint *logCheckpoint
	mu               sync.Mutex
)

type logCheckpoint struct {
	cache *cache.Cache
}

func newLogCheckpoint(cachePath string) (*logCheckpoint, error) {
	c, err := cache.NewCache(cachePath, cache.DefaultCacheOptions)
	if err != nil {
		return nil, err
	}
	if err := c.CreateBucketIfNotExists(checkpointBucket); err != nil {
		return nil, err
	}

	g.Go(func(ctx context.Context) error {
		<-datakit.Exit.Wait()
		time.Sleep(time.Second * 2)
		_ = stopLogCheckpint()
		return nil
	})
	return &logCheckpoint{cache: c}, nil
}

type logCheckpointData struct {
	Offset int64
}

func updateLogCheckpoint(key string, d *logCheckpointData) error {
	c, err := getGlobalCheckpoint()
	if err != nil {
		return err
	}

	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return c.cache.Put(&cache.Data{
		Bucket: checkpointBucket,
		Key:    []byte(key),
		Value:  data,
	})
}

func getLogCheckpoint(key string) (*logCheckpointData, error) {
	c, err := getGlobalCheckpoint()
	if err != nil {
		return nil, err
	}

	data, err := c.cache.Get(checkpointBucket, []byte(key))
	if err != nil {
		return nil, err
	}

	var x logCheckpointData
	if err := json.Unmarshal(data, &x); err != nil {
		return nil, err
	}

	return &x, err
}

func stopLogCheckpint() error {
	mu.Lock()
	defer mu.Unlock()

	if globalCheckpoint == nil || globalCheckpoint.cache == nil {
		return nil
	}

	if err := globalCheckpoint.cache.Stop(); err != nil {
		return err
	}

	globalCheckpoint = nil
	return nil
}

// func getInfoFromLogCheckpoint() string {
// 	c, err := getGlobalCheckpoint()
// 	if err != nil {
// 		return ""
// 	}
// 	return c.cache.String()
// }

func getGlobalCheckpoint() (*logCheckpoint, error) {
	mu.Lock()
	defer mu.Unlock()

	if globalCheckpoint != nil {
		return globalCheckpoint, nil
	}

	var err error
	globalCheckpoint, err = newLogCheckpoint(filepath.Join(defaultCachePath, checkpointFileName))
	return globalCheckpoint, err
}
