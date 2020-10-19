package aliyunobject

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/influxdata/toml"

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
	ag.RegionID = `cn-hangzhou`
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
