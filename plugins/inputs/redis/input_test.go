package redis

import (
	"fmt"
	"log"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestCollect(t *testing.T) {
	input := &Input{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "dev",
		Service:      "dev-test",
		Tags:         make(map[string]string),
		CommandStats: true,
		Slowlog:      false,
		Keys:         []string{"queue"},
	}

	input.initCfg()

	input.Collect()

	for _, obj := range input.collectCache {
		point, err := obj.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestLoadCfg(t *testing.T) {
	arr, err := config.LoadInputConfigFile("./redis.conf", func() inputs.Input {
		return &Input{}
	})

	if err != nil {
		log.Fatalf("%s", err)
	}

	arr[0].(*Input).Run()
}
