package aliyunobjectecs

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func TestApiDescribeInstances(t *testing.T) {
	ak := os.Getenv("AK")
	sk := os.Getenv("SK")

	cli, err := ecs.NewClientWithAccessKey("cn-hangzhou", ak, sk)
	if err != nil {
		t.Error(err)
	}
	req := ecs.CreateDescribeInstancesRequest()
	req.PageSize = requests.NewInteger(100)
	resp, err := cli.DescribeInstances(req)
	if err != nil {
		t.Error(err)
	}

	log.Printf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	for index, inst := range resp.Instances.Instance {
		log.Printf("%d - %s", index, inst.InstanceId)
	}
}

func TestConfig(t *testing.T) {

	var ag objectAgent

	load := true

	if !load {
		ag.ECSObject = []*AliyunCfg{
			{
				RegionID:        "cn-hangzhou",
				AccessKeyID:     "xxx",
				AccessKeySecret: "xxx",
				Tags: map[string]string{
					"key1": "val1",
				},
			},
			{
				RegionID:        "cn-hangzhou",
				AccessKeyID:     "yyy",
				AccessKeySecret: "yyy",
				Tags: map[string]string{
					"key1": "val1",
				},
			},
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
	//datakit.OutputFile = "metrics.txt"
	datakit.GRPCDomainSock = filepath.Join(datakit.InstallDir, "datakit.sock")
	datakit.Exit = cliutils.NewSem()

	config.Cfg.MainCfg = &config.MainConfig{}
	config.Cfg.MainCfg.DataWay = &config.DataWayCfg{}

	config.Cfg.MainCfg.DataWay.Host = "openway.dataflux.cn"
	config.Cfg.MainCfg.DataWay.Token = "tkn_61c438e7786141d8988dcdf92f899b3f"
	config.Cfg.MainCfg.DataWay.Scheme = "https"
	config.Cfg.MainCfg.Interval.Duration = time.Second * 10

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
