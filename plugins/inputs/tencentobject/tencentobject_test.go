package tencentobject

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestInput(t *testing.T) {

	logger.SetGlobalRootLogger("", "debug", logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)

	datakit.InstallDir = "."
	datakit.OutputFile = "metrics.txt"
	datakit.GRPCDomainSock = filepath.Join(datakit.InstallDir, "datakit.sock")
	datakit.Exit = cliutils.NewSem()

	datakit.Cfg.MainCfg = &datakit.MainConfig{}
	datakit.Cfg.MainCfg.DataWay = &datakit.DataWayCfg{}
	datakit.Cfg.MainCfg.DataWay.Host = "openway.dataflux.cn"
	datakit.Cfg.MainCfg.DataWay.Token = "tkn_61c438e7786141d8988dcdf92f899b3f"
	datakit.Cfg.MainCfg.DataWay.Scheme = "https"
	datakit.IntervalDuration = time.Second * 10

	io.Start()

	data, err := ioutil.ReadFile("test.conf")
	if err != nil {
		t.Error(err)
	}
	ag := newAgent()
	if err = toml.Unmarshal(data, &ag); err != nil {
		t.Error(err)
	}
	ag.Run()
}
