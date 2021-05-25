package redis

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestCollectInfoMeasurement(t *testing.T) {
	input := &Input{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "dev",
		Service:      "dev-test",
		Tags:         make(map[string]string),
	}

	err := input.initCfg()
	if err != nil {
		t.Log("init cfg err", err)
		return
	}

	resData, err := input.collectInfoMeasurement()
	if err != nil {
		t.Log("collect data err", err)
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectClientMeasurement(t *testing.T) {
	input := &Input{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "dev",
		Service:      "dev-test",
		Tags:         make(map[string]string),
	}

	input.initCfg()

	resData, err := input.collectClientMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectCommandMeasurement(t *testing.T) {
	input := &Input{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "dev",
		Service:      "dev-test",
		Tags:         make(map[string]string),
	}

	input.initCfg()

	resData, err := input.collectCommandMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectSlowlogMeasurement(t *testing.T) {
	input := &Input{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "dev",
		Service:      "dev-test",
		Tags:         make(map[string]string),
	}

	err := input.initCfg()
	if err != nil {
		t.Log("init cfg err", err)
		return
	}

	resData, err := input.collectSlowlogMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			fmt.Println("error =======>", err)
		} else {
			fmt.Println("point line =====>", point.String())
		}
	}
}

func TestCollectBigKeyMeasurement(t *testing.T) {
	input := &Input{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "dev",
		Service:      "dev-test",
		Tags:         make(map[string]string),
		Keys:         []string{"queue"},
		DB:           1,
	}

	input.initCfg()

	resData, err := input.collectBigKeyMeasurement()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
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
		t.Fatalf("%s", err)
	}

	arr[0].(*Input).Run()
}