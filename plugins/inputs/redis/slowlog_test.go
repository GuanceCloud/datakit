package redis

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSlowlog(t *testing.T) {
	input := &Input{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "dev",
		Service:  "dev-test",
		Tags:     make(map[string]string),
	}

	err := input.initCfg()
	if err != nil {
		t.Log("init cfg err", err)
		return
	}

	resData, err := input.getSlowData()
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

func TestLineProto(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := &slowlogMeasurement{
			tags:              make(map[string]string),
			fields:            make(map[string]interface{}),
			lastTimestampSeen: make(map[string]int64),
			slowlogMaxLen:     128,
		}

		m.name = "dev-test"
		m.tags = map[string]string{
			"key1": "val1",
			"key2": "val3",
		}

		m.fields = map[string]interface{}{
			"field1": 123,
			"field2": "abc",
		}

		pt, err := m.LineProto()
		if err != nil {
			assert.Error(t, err, "collect data err")
		} else {
			t.Log("point data -->", pt.String())
		}
	})

	t.Run("fail", func(t *testing.T) {
		m := &slowlogMeasurement{
			tags:              make(map[string]string),
			fields:            make(map[string]interface{}),
			lastTimestampSeen: make(map[string]int64),
			slowlogMaxLen:     128,
		}

		m.name = "dev-test"
		m.tags = map[string]string{
			"key1": "val1",
			"key2": "val3",
		}

		pt, err := m.LineProto()
		if err != nil {
			assert.Error(t, err, "collect data err")
		} else {
			t.Log("point data -->", pt.String())
		}
	})
}

func TestInfo(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		m := &slowlogMeasurement{
			tags:              make(map[string]string),
			fields:            make(map[string]interface{}),
			lastTimestampSeen: make(map[string]int64),
			slowlogMaxLen:     128,
		}

		m.name = "dev-test"
		m.tags = map[string]string{
			"key1": "val1",
			"key2": "val3",
		}

		m.fields = map[string]interface{}{
			"field1": 123,
			"field2": "abc",
		}

		pt := m.Info()
		t.Log("point data -->", pt)
	})

	t.Run("fail", func(t *testing.T) {
		m := &slowlogMeasurement{
			tags:              make(map[string]string),
			fields:            make(map[string]interface{}),
			lastTimestampSeen: make(map[string]int64),
			slowlogMaxLen:     128,
		}

		m.name = "dev-test"
		m.tags = map[string]string{
			"key1": "val1",
			"key2": "val3",
		}

		pt := m.Info()
		t.Log("point data -->", pt)
	})
}
