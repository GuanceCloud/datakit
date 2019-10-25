package uploader

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/siddontang/go-log/log"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/git"
)

type LogItem struct {
	Log string
}

const maxErrMsgLen = 256

func New(r string) *Uploader {
	u := &Uploader{
		Remote:              r,
		numShards:           2,
		shardCap:            1000,
		maxSamplesPerSecond: 100,
		Compress:            true,
		batchSendDeadline:   time.Second * 10,
		flushDeadline:       1 * time.Minute,
		badSamples:          make(chan LogItem, 100),
	}

	u.shards = u.newShards()

	return u
}

type Uploader struct {
	Compress bool
	Remote   string

	batchSendDeadline   time.Duration
	maxSamplesPerSecond int
	flushDeadline       time.Duration

	numShards int
	shardCap  int

	totalSend int64
	totalRecv int64

	shards *shards

	badSamples chan LogItem
}

func (u *Uploader) SendData(data []byte) error {

	var zBuf bytes.Buffer

	zw := gzip.NewWriter(&zBuf)
	if _, err := zw.Write(data); err != nil {
		return err
	}
	zw.Flush()
	zw.Close()

	httpReq, err := http.NewRequest(http.MethodPost, u.Remote, &zBuf)
	if err != nil {
		log.Errorf("%s", err.Error())
		return err
	}

	httpReq.Header.Set("Content-Encoding", "gzip")
	httpReq.Header.Set("X-Version", git.Version)

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(httpResp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s: %s", httpResp.Status, line)
	}
	if httpResp.StatusCode/100 == 5 {
		return err
	}
	return err
}

func (u *Uploader) Start() {
	u.shards.start(u.numShards)
}

func (u *Uploader) Stop() {
	u.shards.stop()
}

func (u *Uploader) AddLog(l *LogItem) error {
	u.shards.enqueue(u.totalRecv, l)
	return nil
}

func (u *Uploader) Check() {

	tm := time.NewTicker(10 * time.Second)
	defer tm.Stop()

	for {
		select {
		case l, ok := <-u.badSamples:
			if !ok {

			}
			_ = l
		case <-tm.C:
			log.Infof("total send=%v, total recv=%v", u.totalSend, u.totalRecv)
		}
	}
}

func (u *Uploader) newShards() *shards {
	s := &shards{
		u:    u,
		done: make(chan struct{}),
	}
	return s
}

type shards struct {
	u      *Uploader
	queues []chan *LogItem

	hardShutdownCtx context.Context
	hardShutdown    context.CancelFunc
	softShutdown    chan struct{}

	mtx     sync.RWMutex
	done    chan struct{}
	running int32
}

func (s *shards) start(n int) {
	log.Infoln("start uploader shards ...")

	newQueues := make([]chan *LogItem, n)
	for i := 0; i < n; i++ {
		newQueues[i] = make(chan *LogItem, s.u.shardCap)
	}
	s.queues = newQueues
	s.hardShutdownCtx, s.hardShutdown = context.WithCancel(context.Background())
	s.softShutdown = make(chan struct{})
	s.running = int32(n)
	s.done = make(chan struct{})
	for i := 0; i < n; i++ {
		go s.runShard(s.hardShutdownCtx, i, newQueues[i])
	}
}

func (s *shards) runShard(ctx context.Context, i int, queue chan *LogItem) {
	defer func() {

		if e := recover(); e != nil {
			log.Errorf("runShard panic err: %s", e)
			return
		}

		if atomic.AddInt32(&s.running, -1) == 0 {
			close(s.done)
		}
	}()

	var okNum, badNum uint64

	pendingSamples := []*LogItem{}

	timer := time.NewTimer(s.u.batchSendDeadline)
	stop := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}
	defer stop()

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Infof("[info] shard[%d]: okNum=%v, badNum=%v, pendding=%v", i, okNum, badNum, len(pendingSamples))
		case sample, ok := <-queue:
			if !ok {
				if len(pendingSamples) > 0 {
					s.sendSamples(ctx, pendingSamples, i)
					log.Println("done flushing")
				}
				return
			}

			pendingSamples = append(pendingSamples, sample)

			if len(pendingSamples) >= s.u.maxSamplesPerSecond {
				if err := s.sendSamples(ctx, pendingSamples, i); err == nil {
					okNum += uint64(len(pendingSamples))
				} else {
					badNum += uint64(len(pendingSamples))
				}
				pendingSamples = pendingSamples[s.u.maxSamplesPerSecond:]

				stop()
				timer.Reset(s.u.batchSendDeadline)
			}

		case <-timer.C:
			if len(pendingSamples) > 0 {
				if err := s.sendSamples(ctx, pendingSamples, i); err == nil {
					okNum += uint64(len(pendingSamples))
				} else {
					badNum += uint64(len(pendingSamples))
				}
				pendingSamples = pendingSamples[:0]
			}
			timer.Reset(s.u.batchSendDeadline)
		}
	}
}

func (s *shards) enqueue(ref int64, sample *LogItem) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	select {
	case <-s.softShutdown:
		return false
	default:
	}

	shard := uint64(ref) % uint64(len(s.queues))
	select {
	case <-s.softShutdown:
		return false
	case s.queues[shard] <- sample:
		//log.Infoln("enqueue ok")
		s.u.totalRecv++
		return true
	}
}

func (s *shards) stop() {
	s.mtx.RLock()
	close(s.softShutdown)
	s.mtx.RUnlock()

	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, q := range s.queues {
		close(q)
	}

	select {
	case <-s.done:
		return
	case <-time.After(s.u.flushDeadline):
		log.Infof("Failed to flush all samples on shutdown")
	}

	s.hardShutdown()
	<-s.done
}

func (s *shards) sendSamples(ctx context.Context, samples []*LogItem, shardIndex int) error {

	data := generatePacket(samples)

	if err := s.u.SendData(data); err != nil {
		log.Errorf("sendSamples fail: %s", err)
		return err
	}

	atomic.AddInt64(&s.u.totalSend, int64(len(samples)))

	return nil
}

func generatePacket(samples []*LogItem) []byte {

	s := ""

	for _, l := range samples {
		s += l.Log
		s += "\n"
	}

	return []byte(s)
}
