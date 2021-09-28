package kubernetes

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type Discovery struct {
	list map[string]interface{}
	mu   sync.Mutex
}

func NewDiscovery() *Discovery {
	return &Discovery{list: make(map[string]interface{})}
}

func (d *Discovery) TryRun(name, cfg string) error {
	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("invalid inputName")
	}

	existed, md5str := d.IsExist(cfg)
	if existed {
		return nil
	}

	inputList, err := config.LoadInputConfig(cfg, creator)
	if err != nil {
		return err
	}

	d.addList(md5str)

	g := datakit.G("kubernetes-autodiscovery")
	for _, ii := range inputList {
		if ii == nil {
			l.Debugf("skip non-datakit-input %s", name)
			continue
		}

		func(name string, ii inputs.Input) {
			g.Go(func(ctx context.Context) error {
				time.Sleep(time.Duration(rand.Int63n(int64(10 * time.Second))))
				l.Infof("starting input %s ...", name)
				ii.Run()
				l.Infof("input %s exited", name)
				return nil
			})
		}(name, ii)
	}

	return nil
}

func (d *Discovery) IsExist(config string) (bool, string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	h := md5.New()
	h.Write([]byte(config))
	md5Str := hex.EncodeToString(h.Sum(nil))
	_, exist := d.list[md5Str]
	return exist, md5Str
}

func (d *Discovery) addList(md5str string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.list[md5str]; ok {
		return
	}
	d.list[md5str] = nil
}
