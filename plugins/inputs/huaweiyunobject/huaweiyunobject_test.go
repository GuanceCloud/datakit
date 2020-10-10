package huaweiyunobject

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestConfig(t *testing.T) {

	var ag objectAgent

	ag.AccessKeyID = `xxx`
	ag.AccessKeySecret = `yyy`
	ag.Tags = map[string]string{
		"key1": "val1",
	}

	load := false

	if !load {
		ag.Ecs = &Ecs{
			InstancesIDs:       []string{"id1", "id2"},
			ExcludeInstanceIDs: []string{"exid1", "exid2"},
		}

		data, err := toml.Marshal(&ag)
		if err != nil {
			t.Error(err)
		}
		ioutil.WriteFile("test.conf", data, 0777)
	} else {
		data, err := ioutil.ReadFile("test.conf")
		if err != nil {
			t.Error(err)
		}
		if err = toml.Unmarshal(data, &ag); err != nil {
			t.Error(err)
		} else {
			log.Println("ok")
		}
	}

}

func TestInput(t *testing.T) {

	logger.SetGlobalRootLogger("", "debug", logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)

	datakit.InstallDir = "."
	datakit.OutputFile = ""
	datakit.GRPCDomainSock = filepath.Join(datakit.InstallDir, "datakit.sock")
	datakit.Exit = cliutils.NewSem()

	datakit.Cfg.MainCfg.DataWay = &datakit.DataWayCfg{}
	datakit.Cfg.MainCfg.DataWay.Host = "172.16.0.12:32758"
	datakit.Cfg.MainCfg.DataWay.Token = "tkn_2fcba7cfa3b84ab880ac78a92da05bf3"
	datakit.Cfg.MainCfg.DataWay.Scheme = "http"
	datakit.Cfg.MainCfg.Interval = `10s`

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
