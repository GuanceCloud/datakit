package cache

import (
	"errors"
)

var (
	ErrGlobalCacheNotInitialize = errors.New("global cache is not initialized")
	defaultCache                *Cache
)

func Initialize(dir string, opt *Options) error {
	c, err := NewCache(dir, opt)
	if err != nil {
		return err
	}
	defaultCache = c
	return nil
}

// func Start(dir string) error {
// 	return nil
// }

func Info() string {
	if defaultCache == nil {
		return ""
	}
	return defaultCache.String()
}

func CreateBucketIfNotExists(bucket string) error {
	if defaultCache == nil {
		return ErrGlobalCacheNotInitialize
	}
	return defaultCache.CreateBucketIfNotExists(bucket)
}

func Stop() error {
	if defaultCache == nil {
		return ErrGlobalCacheNotInitialize
	}
	return defaultCache.Stop()
}

func Put(bucket string, key, value []byte) error {
	if defaultCache == nil {
		return ErrGlobalCacheNotInitialize
	}
	return defaultCache.Put(&Data{bucket, key, value})
}

func PutWithData(d *Data) error {
	if defaultCache == nil {
		return ErrGlobalCacheNotInitialize
	}
	return defaultCache.Put(d)
}

func Get(bucket string, key []byte) ([]byte, error) {
	if defaultCache == nil {
		return nil, ErrGlobalCacheNotInitialize
	}
	return defaultCache.Get(bucket, key)
}

func ForEach(bucket string, handle ProcessHandle, clean bool) error {
	if defaultCache == nil {
		return ErrGlobalCacheNotInitialize
	}
	return defaultCache.ForEach(bucket, handle, clean)
}
