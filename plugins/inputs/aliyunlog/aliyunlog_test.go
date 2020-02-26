package aliyunlog

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"

	"github.com/influxdata/toml"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	consumerLibrary "github.com/aliyun/aliyun-log-go-sdk/consumer"
)

func TestConfig(t *testing.T) {
	var cfg AliyunLog
	cfg.Consumer = []*ConsumerInstance{
		&ConsumerInstance{
			Endpoint:  "1cn-hangzhou.log.aliyuncs.com",
			AccessKey: "xxx",
			AccessID:  "xxx",
			Projects: []*LogProject{
				&LogProject{
					Name: "project1",
					Stores: []*LogStoreCfg{
						&LogStoreCfg{
							Name:              "store1",
							ConsumerGroupName: "cgroup",
							ConsumerName:      "cname",
						},
					},
				},
			},
		},
	}

	if data, err := toml.Marshal(&cfg); err != nil {
		t.Errorf("%s", err)
	} else {
		log.Printf("%s", string(data))
	}
}

func process(shardId int, logGroupList *sls.LogGroupList) string {
	fmt.Println(shardId, logGroupList)
	return ""
}

func createClient() sls.ClientInterface {
	AccessKeyID := ""
	AccessKeySecret := ""
	Endpoint := "cn-hangzhou.log.aliyuncs.com"
	cli := sls.CreateNormalInterface(Endpoint, AccessKeyID, AccessKeySecret, "")

	return cli
}

func TestProject(t *testing.T) {
	cli := createClient()
	proj, err := cli.GetProject("wqc-demo1")
	if err != nil {
		t.Errorf("%s", err)
	}
	log.Printf("%s", proj.Name)

	_, err = cli.GetLogStore("wqc-demo1", "test")
	if err != nil {
		t.Errorf("%s", err)
	}
}

func TestConsumer(t *testing.T) {

	option := consumerLibrary.LogHubConfig{
		Endpoint:          "1cn-hangzhou.log.aliyuncs.com",
		AccessKeyID:       "LTAI4FkR2SokHHESouUMrkxV",
		AccessKeySecret:   "ht4jybX3IrhQAUgHrUOTJRrkP8dONJ",
		Project:           "wqc-demo1",
		Logstore:          "test",
		ConsumerGroupName: "grp-1",
		ConsumerName:      "wqc",
		// This options is used for initialization, will be ignored once consumer group is created and each shard has been started to be consumed.
		// Could be "begin", "end", "specific time format in time stamp", it's log receiving time.
		CursorPosition: consumerLibrary.BEGIN_CURSOR,
	}

	consumerWorker := consumerLibrary.InitConsumerWorker(option, process)
	ch := make(chan os.Signal)
	signal.Notify(ch)
	consumerWorker.Start()
	if _, ok := <-ch; ok {
		fmt.Println("msg", "get stop signal, start to stop consumer worker", "consumer worker name")
		consumerWorker.StopAndWait()
	}
}
