package aliyunlog

import (
	//"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/influxdata/toml"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

var (
	TestAliyunSLS = false
)

func TestConfig(t *testing.T) {

	cfg := &ConsumerInstance{
		Endpoint:        "1cn-hangzhou.log.aliyuncs.com",
		AccessKeyID:     "xxx",
		AccessKeySecret: "xxx",
		Projects: []*LogProject{
			&LogProject{
				Name: "project1",
				Stores: []*LogStoreCfg{
					&LogStoreCfg{
						MetricName:        "aliyunala_slb",
						Tags:              []string{"aa"},
						Fields:            []string{"int:dd,ee", "float:ff,gg"},
						Name:              "store1",
						ConsumerGroupName: "cgroup",
						ConsumerName:      "cname",
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

func TestLoadConfig(t *testing.T) {
	var cfg ConsumerInstance

	var data []byte

	var err error
	data, err = ioutil.ReadFile("test.conf")
	if err != nil {
		t.Errorf("%s", err)
	}

	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		log.Printf("ok")
	}
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

	/*option := consumerLibrary.LogHubConfig{
		Endpoint:          "1cn-hangzhou.log.aliyuncs.com",
		AccessKeyID:       "",
		AccessKeySecret:   "",
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
	}*/
}

func TestService(t *testing.T) {

	TestAliyunSLS = true

	ag := NewAgent()

	data, err := ioutil.ReadFile("./test.conf")
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = toml.Unmarshal(data, ag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	ag.Run()

}
