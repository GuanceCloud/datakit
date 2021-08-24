package kubernetes

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type PromExporter struct {
	list map[string]interface{}
	mu   sync.Mutex

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewPromExporter() *PromExporter {
	ctx, cancel := context.WithCancel(context.Background())
	return &PromExporter{
		list:       make(map[string]interface{}),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (pe *PromExporter) Stop() {
	pe.cancelFunc()
}

func (pe *PromExporter) TryRun(config string) error {
	if config == "" {
		return nil // skip
	}

	if existed, md5Str := pe.isExist(config); existed {
		return nil // config is already existed, skip
	} else {
		pe.addList(md5Str)
	}

	opt, err := pe.marshalToPromOption(config)
	if err != nil {
		return err
	}

	if opt.IsDisable() {
		return nil // disable export, skip
	}

	p, err := prom.NewProm(opt)
	if err != nil {
		return err
	}

	go pe.do(p)

	return nil
}

func (pe *PromExporter) isExist(config string) (bool, string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	h := md5.New()
	h.Write([]byte(config))
	md5Str := hex.EncodeToString(h.Sum(nil))
	_, exist := pe.list[md5Str]
	return exist, md5Str
}

func (pe *PromExporter) addList(md5Str string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if _, ok := pe.list[md5Str]; ok {
		return
	}
	pe.list[md5Str] = nil
}

func (pe *PromExporter) marshalToPromOption(config string) (*prom.Option, error) {
	opt := &prom.Option{}
	err := json.Unmarshal([]byte(config), opt)
	if err != nil {
		return nil, err
	}
	return opt, nil
}

func (pe *PromExporter) do(p *prom.Prom) {
	source := p.Option().GetSource()
	interval := p.Option().GetIntervalDuration()

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-pe.ctx.Done():
			l.Info("prom exit")
			//return
			break

		case <-tick.C:
			start := time.Now()
			pts, err := p.Collect()
			if err != nil {
				l.Error(err)
				// next
			}

			if len(pts) == 0 {
				continue
			}

			if err := io.Feed(source, datakit.Metric, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
				io.FeedLastError(source, err.Error())
				l.Error(err)
			}
		}
	}
}
