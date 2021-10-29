package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

		if err := input.initCfg(); err != nil {
			t.Error(err)
		}

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

		if err := input.initCfg(); err != nil {
			t.Error(err)
		}

		input.Collect()
	})
}

func TestGetDsnString(t *testing.T) {
	m := &Input{}
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
		Port: 3309,
		User: "root",
		Pass: "test",
		Tags: make(map[string]string),
	}

	if err := input.initCfg(); err != nil {
		t.Error(err)
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

		if err := input.initCfg(); err != nil {
			t.Error(err)
		}

		resData, err := input.collectInnodbMeasurement()
		if err != nil {
			t.Error(err)
		}

		for _, pt := range resData {
			point, err := pt.LineProto()
			if err != nil {
				t.Error(err)
			} else {
				t.Log(point.String())
			}
		}
	})
}

func TestBaseCollect(t *testing.T) {
	t.Run("bin log off", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.1",
			Port: 3307,
			User: "datakitMonitor",
			Pass: "datakitMonitor",
			Tags: make(map[string]string),
		}

		err := input.initCfg()
		if err != nil {
			t.Error(err)
		}

		resData, err := input.collectBaseMeasurement()
		if err != nil {
			t.Error(err)
		}

		for _, pt := range resData {
			point, err := pt.LineProto()
			if err != nil {
				t.Error(err)
			} else {
				t.Log(point.String())
			}
		}
	})

	t.Run("bin log on", func(t *testing.T) {
		input := &Input{
			Host: "rm-bp15268nefz6870hg.mysql.rds.aliyuncs.com",
			Port: 3306,
			User: "datakitMonitor",
			Pass: "SunxEVJEE75tmUJZU7Eb",
			Tags: make(map[string]string),
		}

		err := input.initCfg()
		if err != nil {
			t.Error(err)
		}

		resData, err := input.collectBaseMeasurement()
		if err != nil {
			t.Error(err)
		}

		for _, pt := range resData {
			point, err := pt.LineProto()
			if err != nil {
				t.Error(err)
			} else {
				t.Log(point.String())
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
		t.Error(err)
	}

	resData, err := input.collectSchemaMeasurement()
	if err != nil {
		t.Error(err)
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(point.String())
		}
	}
}

func TestTbSchemaCollect(t *testing.T) {
	input := &Input{
		Host: "127.0.0.1",
		Port: 3306,
		User: "root",
		Pass: "test",
		Tags: make(map[string]string),
	}

	input.Tables = []string{}

	err := input.initCfg()
	if err != nil {
		t.Error(err)
	}

	resData, err := input.collectTableSchemaMeasurement()
	if err != nil {
		t.Error(err)
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(point.String())
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
		{
			sql:    "select id, namespace,email, username, value from core_stone.biz_main_account",
			metric: "cutomer-metric",
			tags:   []string{"id"},
			fields: []string{},
		},
	}

	err := input.initCfg()
	if err != nil {
		t.Error(err)
	}

	resData, err := input.customSchemaMeasurement()
	if err != nil {
		t.Error(err)
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(point.String())
		}
	}
}

func TestUserMeasurement(t *testing.T) {
	input := &Input{
		Host: "127.0.0.1",
		Port: 3306,
		User: "root",
		Pass: "test",
		Tags: make(map[string]string),
	}

	input.Users = []string{}

	err := input.initCfg()
	if err != nil {
		t.Error(err)
	}

	resData, err := input.collectUserMeasurement()
	if err != nil {
		t.Error(err)
	}

	for _, pt := range resData {
		point, err := pt.LineProto()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(point.String())
		}
	}
}
