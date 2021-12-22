// Package cache wraps boltDB functions
package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	bolt "go.etcd.io/bbolt"
)

const (
	LocalBufferSize = 256

	CacheFileName = "cache.bolt"
)

var (
	defaultCacheOptions = &Options{
		SwapInterval: 30 * time.Second,
		BatchSize:    128,
		MaxDiskSize:  int64(1024 * 1024 * 1024), // 1GB
	}

	l = logger.DefaultSLogger("cache")

	ErrNotFoundBucket = errors.New("not found bucket")
)

type Options struct {
	SwapInterval time.Duration
	BatchSize    int
	MaxDiskSize  int64
}

func complateOptions(opt *Options) *Options {
	if opt == nil {
		opt = defaultCacheOptions
	}
	if !opt.ok() {
		l.Warnf("invalid cache option: %s, use default: %s",
			opt, defaultCacheOptions)
		opt = defaultCacheOptions
	}
	return opt
}

func (opt *Options) String() string {
	return fmt.Sprintf("swap at: %v, batch size: %d, disk size: %s",
		opt.SwapInterval,
		opt.BatchSize,
		humanize.IBytes(uint64(opt.MaxDiskSize)))
}

func (opt *Options) ok() bool {
	return opt.SwapInterval >= time.Second &&
		opt.BatchSize >= 1 &&
		opt.MaxDiskSize >= 1024*1024
}

type Data struct {
	Bucket string
	Key    []byte
	Value  []byte
}

func initDB(dbfile string, buckets []string) (*bolt.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbfile), os.ModePerm); err != nil {
		return nil, err
	}

	db, err := bolt.Open(dbfile, 0o600, &bolt.Options{Timeout: time.Second * 3})
	if err != nil {
		return nil, err
	}

	if len(buckets) != 0 {
		fn := func(tx *bolt.Tx) error {
			for _, bucket := range buckets {
				_, err := tx.CreateBucketIfNotExists([]byte(bucket))
				if err != nil {
					return fmt.Errorf("tx.CreateBucketIfNotExists: %w", err)
				}
			}
			return nil
		}
		if err := db.Update(fn); err != nil {
			return nil, err
		}
	}

	return db, nil
}

type Cache struct {
	Directory string

	opt *Options

	db       *bolt.DB
	dbfile   string
	swapTick *time.Ticker

	// stat info
	diskSize             int64
	swapCnt              int
	totalCacheCnt        int64
	totalFlushed         int64
	totalCacheBytes      int64
	totalDroppedCacheCnt int64

	localBuffer chan *Data
	stopChannle chan struct{}
}

func NewCache(dir string, opt *Options) (*Cache, error) {
	l = logger.SLogger("cache")

	opt = complateOptions(opt)

	c := Cache{
		Directory: dir,
		opt:       opt,

		dbfile:   filepath.Join(dir, CacheFileName),
		swapTick: time.NewTicker(opt.SwapInterval),

		localBuffer: make(chan *Data, LocalBufferSize),
		stopChannle: make(chan struct{}),
	}

	if c.Directory != "" {
		if err := os.MkdirAll(c.Directory, os.ModePerm); err != nil {
			return nil, err
		}
	}

	db, err := initDB(c.dbfile, nil)
	if err != nil {
		return nil, err
	}

	c.db = db
	_ = c.loadHist()

	g := datakit.G("cache")
	g.Go(func(ctx context.Context) error {
		c.start()
		return nil
	})

	return &c, nil
}

func (c *Cache) String() string {
	buckets := c.Buckets()
	return fmt.Sprintf(
		`Dir: %s, CacheFile: %s, buckets: %v, cache count: %d, flushed: %d, 
		dropped: %d, total cache size: %s, disk: %s, swapped: %d, opt:(%s)`,
		c.Directory,
		c.dbfile,
		buckets,
		c.totalCacheCnt,
		c.totalFlushed,
		c.totalDroppedCacheCnt,
		humanize.IBytes(uint64(c.totalCacheBytes)),
		humanize.IBytes(uint64(c.diskSize)),
		c.swapCnt,
		c.opt.String(),
	)
}

func (c *Cache) Stop() (err error) {
	if c.swapTick != nil {
		c.swapTick.Stop()
	}
	if c.db != nil {
		err = c.db.Close()
	}

	select {
	case <-c.stopChannle: // closed
	default:
		close(c.stopChannle) // disable write
	}
	return
}

func (c *Cache) Buckets() []string {
	var buckets []string

	fn := func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	}

	if err := c.db.View(fn); err != nil {
		l.Error(err)
		return nil
	}

	return buckets
}

func (c *Cache) CreateBucketIfNotExists(bucket string) error {
	fn := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("tx.CreateBucketIfNotExists: %w", err)
		}
		return nil
	}
	if err := c.db.Update(fn); err != nil {
		return err
	}
	return nil
}

func (c *Cache) Put(d *Data) error {
	if d == nil {
		return nil
	}
	select {
	case <-c.stopChannle:
		return fmt.Errorf("cache failed, cache channel buzy or closed")
	default: // cache openning
	}

	c.localBuffer <- d
	return nil
}

func (c *Cache) Get(bucket string, key []byte) ([]byte, error) {
	var value []byte
	fn := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return fmt.Errorf("%w: %s", ErrNotFoundBucket, bucket)
		}
		value = bkt.Get(key)
		return nil
	}

	if err := c.db.View(fn); err != nil {
		return nil, err
	}
	return value, nil
}

type ProcessHandle func(key []byte, value []byte) error

func (c *Cache) ForEach(bucket string, handle ProcessHandle, clean bool) error {
	keys := [][]byte{}
	// start = time.Now()
	// failed = 0

	fn := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return fmt.Errorf("%w: %s", ErrNotFoundBucket, bucket)
		}
		csor := bkt.Cursor()

		for k, v := csor.First(); k != nil; k, v = csor.Next() {
			select {
			case <-c.stopChannle:
				l.Info("cache exit on sem during upload.")
				return nil
			default:
			}

			if err := handle(k, v); err != nil {
				l.Warnf("handle error: %s", err)
				continue
			} else {
				keys = append(keys, k)
			}

			// no matter repost ok or failed, count it, we can't block here too long
			if len(keys) >= c.opt.BatchSize {
				break
			}
		}
		return nil
	}

	if err := c.db.View(fn); err != nil {
		l.Error(err)
	}

	// l.Debugf("clean %d ok", len(cleanedOKIds))

	// clean cleaned-ok keys
	// if len(cleanedOKIds) > 0 {
	// 	l.Debugf("post %d local cache cost %v, %d failed, try clean %d local cache...",
	// 		len(cleanedOKIds), time.Since(start), failed, len(cleanedOKIds))
	// }

	l.Debugf("clean keys length %d", len(keys))

	if clean {
		c.cleanCache(bucket, keys)
	}
	return nil
}

func (c *Cache) cleanCache(bucket string, keys [][]byte) {
	if len(keys) == 0 {
		return
	}

	fn := func(tx *bolt.Tx) error {
		for _, key := range keys {
			bkt := tx.Bucket([]byte(bucket))
			if bkt == nil {
				return fmt.Errorf("%w: %s", ErrNotFoundBucket, bucket)
			}
			if err := bkt.Delete(key); err != nil {
				return err
			}
		}
		return nil
	}

	if err := c.db.Batch(fn); err != nil {
		l.Error(err)
	}

	if len(keys) > 0 {
		c.totalFlushed += int64(len(keys))
	}
}

func (c *Cache) start() {
	l.Info("cache worker started.")

	for {
		select {
		case data := <-c.localBuffer:
			if err := c.putDB(data.Bucket, data.Key, data.Value); err != nil {
				l.Error(err)
			}

		case <-c.swapTick.C:
			l.Debugf("cache info: %s", c.String())
			if err := c.trySwap(); err != nil {
				l.Warnf("trySwap(): %s", err.Error())
			}

		case <-c.stopChannle:
			c.cleanup() // clean in-channel cache if any
			l.Info("cache exit on sem.")
			return
		}
	}
}

func (c *Cache) loadHist() error {
	buckets := c.Buckets()

	fn := func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			bkt := tx.Bucket([]byte(bucket))
			if bkt == nil {
				return fmt.Errorf("%w: %s", ErrNotFoundBucket, bucket)
			}

			csor := bkt.Cursor()
			for k, v := csor.First(); k != nil; k, v = csor.Next() {
				c.totalCacheCnt++
				c.totalCacheBytes += int64(len(v))
			}
		}
		return nil
	}

	if err := c.db.View(fn); err != nil {
		l.Error(err)
		return err
	}

	l.Debugf("cache load hist ok: %s", c.String())
	return nil
}

func (c *Cache) putDB(bucket string, key, value []byte) error {
	fn := func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return fmt.Errorf("%w: %s", ErrNotFoundBucket, bucket)
		}
		return bkt.Put(key, value)
	}

	if err := c.db.Update(fn); err != nil {
		l.Errorf("add cache failed: %w", err)
		return err
	}

	c.totalCacheCnt++
	c.totalCacheBytes += int64(len(value))
	return nil
}

func (c *Cache) cleanup() {
	for {
		select {
		case data := <-c.localBuffer:
			if err := c.putDB(data.Bucket, data.Key, data.Value); err != nil {
				l.Error(err)
			}

		default:
			return
		}
	}
}

func (c *Cache) trySwap() error {
	fi, err := os.Stat(c.dbfile)
	if err != nil {
		return fmt.Errorf("should not been here")
	}

	c.diskSize = fi.Size()

	if c.diskSize > c.opt.MaxDiskSize {
		c.swapCnt++
		return c.doSwap()
	}

	return nil
}

func (c *Cache) doSwap() error {
	buckets := c.Buckets()

	swap := c.dbfile + ".swap"
	swapdb, err := initDB(swap, buckets)
	if err != nil {
		return err
	}

	n, err := moveData(c.db, swapdb, buckets)
	if err != nil {
		l.Errorf("move cache failed: %w", err)
		return err
	}

	if fi, err := os.Stat(swap); err == nil { // after swapping, db size still large...
		if fi.Size() >= c.opt.MaxDiskSize/2 {
			// drop all old cache, create new DB

			l.Warnf("swap db size %d", fi.Size())
			if err := dropDB(swap, swapdb); err != nil {
				return err
			}

			c.totalDroppedCacheCnt += int64(n)

			if err := dropDB(c.dbfile, c.db); err != nil {
				return err
			}

			db, err := initDB(c.dbfile, buckets)
			if err != nil {
				return err
			}
			c.db = db

			return nil
		}
	}

	// drop old db
	l.Debug("drop old db")
	if err := dropDB(c.dbfile, c.db); err != nil {
		return err
	}

	// use swap db
	l.Debug("use swap db")
	_ = swapdb.Close()
	if err := os.Rename(swap, c.dbfile); err != nil {
		return err
	}

	db, err := initDB(c.dbfile, buckets)
	if err != nil {
		return err
	}

	c.db = db

	l.Infof("swapped %d caches ok", n)
	return nil
}

func dropDB(f string, db *bolt.DB) error {
	if err := db.Close(); err != nil {
		l.Errorf("db.Close(): %w", err)
		return err
	}

	if err := os.Remove(f); err != nil {
		l.Errorf("os.Remove(): %w", err)
		return err
	}
	return nil
}

func moveData(from, to *bolt.DB, buckets []string) (int, error) {
	count := 0

	fn := func(tx *bolt.Tx) error {
		for _, bucket := range buckets {
			bkt := tx.Bucket([]byte(bucket))
			if bkt == nil {
				return fmt.Errorf("%w: %s", ErrNotFoundBucket, bucket)
			}
			csor := bkt.Cursor()

			for k, v := csor.First(); k != nil; k, v = csor.Next() {
				if err := to.Update(func(_tx *bolt.Tx) error {
					_bkt := _tx.Bucket([]byte(bucket))
					if _bkt == nil {
						return fmt.Errorf("%w: %s", ErrNotFoundBucket, bucket)
					}
					count++
					return _bkt.Put(k, v)
				}); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := from.View(fn); err != nil {
		l.Error(err)
		return 0, err
	}

	return count, nil
}
