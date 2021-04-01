package tencentcost

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	moduleLogger    *logger.Logger
	historyCacheDir string
)

func (_ *TencentCost) Catalog() string {
	return `tencentcloud`
}

func (_ *TencentCost) SampleConfig() string {
	return sampleConfig
}

func (t *TencentCost) Run() {
	moduleLogger = logger.SLogger(inputName)

	historyCacheDir = filepath.Join(datakit.DataDir, inputName)
	os.MkdirAll(historyCacheDir, 0775)

	if t.TransactionInterval.Duration > 0 {
		if t.TransactionInterval.Duration < time.Minute {
			t.TransactionInterval.Duration = time.Minute
		}
		t.subModules = append(t.subModules, newTransaction(t))
	}

	if t.OrderInterval.Duration > 0 {
		if t.OrderInterval.Duration < time.Minute {
			t.OrderInterval.Duration = time.Minute
		}
		t.subModules = append(t.subModules, newOrder(t))
	}

	if t.BillInterval.Duration > 0 {
		if t.BillInterval.Duration < time.Minute {
			t.BillInterval.Duration = time.Minute
		}
		t.subModules = append(t.subModules, newBill(t))
	}

	go func() {
		<-datakit.Exit.Wait()
		t.cancelFun()
	}()

	var wg sync.WaitGroup
	for _, m := range t.subModules {
		wg.Add(1)
		go func(m subModule) {
			defer wg.Done()
			m.run(t.ctx)
		}(m)
	}

	wg.Wait()
}

func (t *TencentCost) appendCustomTags(tags map[string]string) {
	for k, v := range t.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
}

func ensureString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func newAgent() *TencentCost {
	ac := &TencentCost{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
