package huaweiyunces

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

func genAgent() *agent {
	ag := loadCfg("test.conf").(*agent)
	return ag
}

func TestGetMetric(t *testing.T) {

	//https://support.huaweicloud.com/api-ces/ces_03_0033.html

	moduleLogger = logger.DefaultSLogger(inputName)

	ag := genAgent()

	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, ag.EndPoint, ag.ProjectID, moduleLogger)

	dims := []*Dimension{
		{
			Name:  "instance_id",
			Value: "9a4d6fb6-4de2-422a-b4d3-5d436b79ef09",
		},
	}
	dms := []string{}
	for _, d := range dims {
		dms = append(dms, fmt.Sprintf("%s,%s", d.Name, d.Value))
	}

	now := time.Now().Truncate(time.Minute)
	from := now.Add(-30*time.Minute).Unix() * 1000
	to := now.Unix() * 1000
	resp, err := cli.CESGetMetric("SYS.ECS", "cpu_util", "min", 300, from, to, dms)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%s", string(resp))
}

func TestBatchMetrics(t *testing.T) {

	moduleLogger = logger.DefaultSLogger(inputName)
	ag := genAgent()

	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, ag.EndPoint, ag.ProjectID, moduleLogger)

	dims := []*Dimension{
		{
			Name:  "instance_id",
			Value: "b5d7b7a3-681d-4c08-8e32-f14b640b3e12",
		},
	}

	items := []*metricItem{
		{
			Namespace:  "SYS.ECS",
			MetricName: "cpu_util",
			Dimensions: dims,
		},
		{
			Namespace:  "SYS.ECS",
			MetricName: "disk_write_bytes_rate",
			Dimensions: dims,
		},
	}

	b := &batchReq{
		Period:  "300",
		Filter:  "min",
		From:    time.Now().Add(-1*time.Hour).Unix() * 1000,
		To:      time.Now().Unix() * 1000,
		Metrics: items,
	}

	jdata, _ := json.Marshal(b)
	resp, err := cli.CESGetBatchMetrics(jdata)
	if err == nil {
		result := parseBatchResponse(resp, b.Filter)
		if result != nil {
			for _, item := range result.results {
				log.Printf("%s", item)
			}
		}
	}

}

func TestInput(t *testing.T) {

	ag := genAgent()
	ag.mode = "debug"
	ag.Run()
}

func loadCfg(file string) inputs.Input {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("read config file failed: %s", err.Error())
		return nil
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		log.Fatalf("parse toml file failed, %s", err)
		return nil
	}

	creator := func() inputs.Input {
		return newAgent()
	}

	for field, node := range tbl.Fields {

		switch field {
		case "inputs":
			stbl, ok := node.(*ast.Table)
			if !ok {
				log.Fatalf("ignore bad toml node")
			} else {
				for inputName, v := range stbl.Fields {
					input, err := tryUnmarshal(v, inputName, creator)
					if err != nil {
						log.Fatalf("unmarshal input %s failed: %s", inputName, err.Error())
						return nil
					}
					return input
				}
			}

		default: // compatible with old version: no [[inputs.xxx]] header
			input, err := tryUnmarshal(node, "aa", creator)
			if err != nil {
				log.Fatalf("unmarshal input failed: %s", err.Error())
				return nil
			}
			return input
		}
	}

	return nil
}

func tryUnmarshal(tbl interface{}, name string, creator inputs.Creator) (inputs.Input, error) {

	tbls := []*ast.Table{}

	switch t := tbl.(type) {
	case []*ast.Table:
		tbls = tbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, tbl.(*ast.Table))
	default:
		err := fmt.Errorf("invalid toml format: %v", t)
		return nil, err
	}

	for _, t := range tbls {
		input := creator()

		err := toml.UnmarshalTable(t, input)
		if err != nil {
			log.Fatalf("toml unmarshal failed: %v", err)
			return nil, err
		}
		return input, nil

	}
	return nil, nil
}
