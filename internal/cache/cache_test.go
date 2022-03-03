package cache

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	dir    = "cache_testing_dir"
	bucket = "bucket_abc"
)

func TestCachePutGet(t *testing.T) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	defer t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})

	c, err := NewCache(dir, defaultCacheOptions)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Stop(); err != nil {
			t.Error(err)
		}
	}()

	if err := c.CreateBucketIfNotExists(bucket); err != nil {
		t.Error(err)
	}

	key := []byte("key_01")
	value := []byte("value_01")

	if err := c.Put(&Data{bucket, key, value}); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

	if v, err := c.Get(bucket, key); err != nil {
		t.Error(err)
	} else {
		t.Log(string(v))
	}
}

func TestCacheForEachNotClean(t *testing.T) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	defer t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})

	c, err := NewCache(dir, defaultCacheOptions)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Stop(); err != nil {
			t.Error(err)
		}
	}()

	if err := c.CreateBucketIfNotExists(bucket); err != nil {
		t.Error(err)
	}

	var (
		key    = []byte("key_01")
		value  = []byte("value_01")
		key2   = []byte("key_02")
		value2 = []byte("value_02")

		clean = false
	)

	if err := c.Put(&Data{bucket, key, value}); err != nil {
		t.Error(err)
	}
	if err := c.Put(&Data{bucket, key2, value2}); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

	handle := func(key, value []byte) error {
		t.Logf("key: %s value: %s", string(key), string(value))
		return nil
	}

	if err := c.ForEach(bucket, handle, clean); err != nil {
		t.Error(err)
	}

	// 验证，如果被 clean，则此处 get 为空
	if v, err := c.Get(bucket, key); err != nil {
		t.Error(err)
	} else {
		t.Logf("key: %s value: %v", key, v)
	}
}

func TestCacheForEachAndClean(t *testing.T) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	defer t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})

	c, err := NewCache(dir, defaultCacheOptions)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Stop(); err != nil {
			t.Error(err)
		}
	}()

	if err := c.CreateBucketIfNotExists(bucket); err != nil {
		t.Error(err)
	}

	var (
		key    = []byte("key_01")
		value  = []byte("value_01")
		key2   = []byte("key_02")
		value2 = []byte("value_02")

		clean = true
	)

	if err := c.Put(&Data{bucket, key, value}); err != nil {
		t.Error(err)
	}
	if err := c.Put(&Data{bucket, key2, value2}); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

	handle := func(key, value []byte) error {
		t.Logf("key: %s value: %s", string(key), string(value))
		return nil
	}

	if err := c.ForEach(bucket, handle, clean); err != nil {
		t.Error(err)
	}

	// 验证，如果被 clean，则此处 get 为空
	if v, err := c.Get(bucket, key); err != nil {
		t.Error(err)
	} else {
		t.Logf("key: %s value: %v", key, v)
	}
}

func TestCacheMove(t *testing.T) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	defer t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})

	c, err := NewCache(dir, defaultCacheOptions)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := c.Stop(); err != nil {
			t.Error(err)
		}
	}()

	if err := c.CreateBucketIfNotExists(bucket); err != nil {
		t.Error(err)
	}

	var (
		key    = []byte("key_01")
		value  = []byte("value_01")
		key2   = []byte("key_02")
		value2 = []byte("value_02")
	)

	if err := c.Put(&Data{bucket, key, value}); err != nil {
		t.Error(err)
	}
	if err := c.Put(&Data{bucket, key2, value2}); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

	swap := c.dbfile + ".swap"
	swapdb, err := initDB(swap, c.Buckets())
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := os.RemoveAll(swap); err != nil {
			t.Error(err)
		}
		if err := swapdb.Close(); err != nil {
			t.Error(err)
		}
	}()

	n, err := moveData(c.db, swapdb, c.Buckets())
	if err != nil {
		t.Errorf("move cache failed: %s", err.Error())
	}

	t.Log(n)
}

func TestCacheForEachForTemp(t *testing.T) {
	tmpdir := "tmp"
	c, err := NewCache(tmpdir, defaultCacheOptions)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpdir); err != nil {
			t.Error(err)
		}
	})

	defer func() {
		if err := c.Stop(); err != nil {
			t.Error(err)
		}
	}()

	if err := c.CreateBucketIfNotExists(bucket); err != nil {
		t.Error(err)
	}

	fn := func(k, v []byte) error {
		fmt.Printf("Key:%s value:%s", string(k), string(v))
		return nil
	}

	if err := c.ForEach(bucket, fn, false); err != nil {
		t.Error(err)
	}
}

func TestLoadHist(t *testing.T) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	defer t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})

	c1, err := NewCache(dir, defaultCacheOptions)
	if err != nil {
		t.Fatal(err)
	}
	if err := c1.CreateBucketIfNotExists(bucket); err != nil {
		t.Error(err)
	}

	var (
		key    = []byte("key_01")
		value  = []byte("value_01")
		key2   = []byte("key_02")
		value2 = []byte("value_02")

		clean = false
	)

	if err := c1.Put(&Data{bucket, key, value}); err != nil {
		t.Error(err)
	}
	if err := c1.Put(&Data{bucket, key2, value2}); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)
	_ = c1.Stop()

	//

	c2, err := NewCache(dir, defaultCacheOptions)
	if err != nil {
		t.Fatal(err)
	}
	if err := c2.CreateBucketIfNotExists(bucket); err != nil {
		t.Error(err)
	}

	handle := func(key, value []byte) error {
		t.Logf("key: %s value: %s", string(key), string(value))
		return nil
	}

	if err := c2.ForEach(bucket, handle, clean); err != nil {
		t.Error(err)
	}
	_ = c2.Stop()
}
