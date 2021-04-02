package aliyuncost

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `aliyuncost`
	moduleLogger *logger.Logger

	historyCacheDir = ""
)

func (*agent) Catalog() string {
	return "aliyun"
}

func (*agent) SampleConfig() string {
	return sampleConfig
}

func (ag *agent) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	if !ag.isDebug() && !ag.isTest() {
		historyCacheDir = filepath.Join(datakit.DataDir, inputName)
		os.MkdirAll(historyCacheDir, 0775)
	}

	limit := rate.Every(60 * time.Millisecond)
	ag.rateLimiter = rate.NewLimiter(limit, 1)

	if ag.AccountInterval.Duration > 0 {
		if ag.AccountInterval.Duration < time.Minute {
			ag.AccountInterval.Duration = time.Minute
		}
		ag.subModules = append(ag.subModules, newCostAccount(ag))
	}

	if ag.BiilInterval.Duration > 0 {
		if ag.BiilInterval.Duration < time.Minute {
			ag.BiilInterval.Duration = time.Minute
		}
		if ag.ByInstance {
			ag.subModules = append(ag.subModules, newCostInstanceBill(ag))
		} else {
			ag.subModules = append(ag.subModules, newCostBill(ag))
		}
	}

	if ag.OrdertInterval.Duration > 0 {
		if ag.OrdertInterval.Duration < time.Minute {
			ag.OrdertInterval.Duration = time.Minute
		}
		ag.subModules = append(ag.subModules, newCostOrder(ag))
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		default:
		}

		var err error
		ag.client, err = bssopenapi.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			if ag.isTest() {
				ag.testError = err
				return
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	//先获取account name
	ag.queryBillOverview(ag.ctx)

	var wg sync.WaitGroup
	for _, m := range ag.subModules {
		wg.Add(1)
		go func(m subModule) {
			defer wg.Done()
			m.run(ag.ctx)
		}(m)
	}

	wg.Wait()
}

func (ag *agent) cacheFileKey(subname string) string {
	m := md5.New()
	m.Write([]byte(ag.AccessKeyID))
	m.Write([]byte(ag.AccessKeySecret))
	m.Write([]byte(ag.RegionID))
	m.Write([]byte(subname))
	return hex.EncodeToString(m.Sum(nil))
}

func newAgent(mode string) *agent {
	ag := &agent{}
	ag.mode = mode
	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent("")
	})
}
