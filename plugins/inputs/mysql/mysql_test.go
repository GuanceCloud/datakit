package mysql

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollect(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.1",
			Port: 3306,
			User: "root",
			Pass: "test",
			Tags: make(map[string]string),
		}

		input.initCfg()
		input.Collect()
	})

	t.Run("error", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.2",
			Port: 3306,
			User: "root",
			Pass: "test",
			Tags: make(map[string]string),
		}

		input.initCfg()
		input.Collect()
	})
}

func TestGetDsnString(t *testing.T) {
	m := MysqlMonitor{}
	m.Host = "127.0.0.1"
	m.Port = 3306
	m.User = "root"
	m.Pass = "test"

	expected := "root:test@tcp(127.0.0.1:3306)/"
	actual := m.getDsnString()
	assert.Equal(t, expected, actual)
}

func TestRun(t *testing.T) {
	input := &Input{
		Host: "127.0.0.1",
		Port: 3306,
		User: "root",
		Pass: "test",
		Tags: make(map[string]string),
	}

	if err := input.initCfg(); err != nil {
		assert.Error(t, err, "collect data err")
	}

	input.Collect()
}

func TestInnodbCollect(t *testing.T) {
	t.Run("bin log off", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.1",
			Port: 3306,
			User: "root",
			Pass: "test",
			Tags: make(map[string]string),
		}

		err := input.initCfg()
		if err != nil {
			assert.Error(t, err, "collect data err")
		}

		resData, err := input.collectInnodbMeasurement()
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
	})
}

func TestBaseCollect(t *testing.T) {
	t.Run("bin log off", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.1",
			Port: 3306,
			User: "root",
			Pass: "test",
			Tags: make(map[string]string),
		}

		err := input.initCfg()
		if err != nil {
			assert.Error(t, err, "collect data err")
		}

		resData, err := input.collectBaseMeasurement()
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
	})

	t.Run("bin log on", func(t *testing.T) {
		input := &Input{
			Host: "rm-bp15268nefz6870hg.mysql.rds.aliyuncs.com",
			Port: 3306,
			User: "cc_monitor",
			Pass: "Zyadmin123",
			Tags: make(map[string]string),
		}

		err := input.initCfg()
		if err != nil {
			assert.Error(t, err, "collect data err")
		}

		resData, err := input.collectBaseMeasurement()
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
	})
}

func TestSchemaCollect(t *testing.T) {
	input := &Input{
		Host: "127.0.0.1",
		Port: 3306,
		User: "root",
		Pass: "test",
		Tags: make(map[string]string),
	}

	err := input.initCfg()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	resData, err := input.collectSchemaMeasurement()
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

func TestCustomSchemaMeasurement(t *testing.T) {
	input := &Input{
		Host: "127.0.0.1",
		Port: 3306,
		User: "root",
		Pass: "test",
		Tags: make(map[string]string),
	}

	input.Query = []*customQuery{
		&customQuery{
			sql:    "select id, namespace,email, username, value from core_stone.biz_main_account",
			metric: "cutomer-metric",
			tags:   []string{"id"},
			fields: []string{},
		},
	}

	err := input.initCfg()
	if err != nil {
		assert.Error(t, err, "collect data err")
	}

	resData, err := input.customSchemaMeasurement()
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
