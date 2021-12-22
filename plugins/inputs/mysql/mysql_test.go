package mysql

/* test: fail
func TestdoCollect(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.1",
			Port: testMysqlPort,
			User: "root",
			Pass: "test",
			Tags: make(map[string]string),
		}

		if err := input.initCfg(); err != nil {
			t.Error(err)
		}

		input.doCollect()
	})

	t.Run("error", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.2",
			Port: testMysqlPort,
			User: "root",
			Pass: "test",
			Tags: make(map[string]string),
		}

		if err := input.initCfg(); err != nil {
			t.Error(err)
		}

		input.doCollect()
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

	input.doCollect()
}

func TestInnodbCollect(t *testing.T) {
	t.Run("bin log off", func(t *testing.T) {
		input := &Input{
			Host: "127.0.0.1",
			Port: testMysqlPort,
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
			Port: testMysqlPort,
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
		Port: testMysqlPort,
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
		Port: testMysqlPort,
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
		Port: testMysqlPort,
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
		Port: testMysqlPort,
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

func TestDbmStatement(t *testing.T) {
	input := &Input{
		Host: "127.0.0.1",
		Port: testMysqlPort,
		User: "root",
		Pass: "123456",
		Tags: make(map[string]string),
	}

	err := input.initCfg()
	assert.NoError(t, err)

	ms, err := input.getDbmMetric()

	assert.NoError(t, err)
	assert.Equal(t, len(ms), 0)

	time.Sleep(5 * time.Second)
	ms, err = input.getDbmMetric()

	assert.GreaterOrEqual(t, len(ms), 0)
	assert.NoError(t, err)
}

func TestDbmStatementSamples(t *testing.T) {
	test := assert.New(t)
	input := &Input{
		Host:      "127.0.0.1",
		Port:      3307,
		User:      "root",
		Pass:      "123456",
		Tags:      make(map[string]string),
		Dbm:       true,
		DbmMetric: dbmMetric{Enabled: true},
		DbmSample: dbmSample{Enabled: true},
	}

	err := input.initCfg()
	test.NoError(err)

	ms, err := input.getDbmSample()
	test.GreaterOrEqual(len(ms), 0)
	test.NoError(err)
	time.Sleep(5 * time.Second)

	ms, err = input.getDbmSample()
	test.NoError(err)
	test.GreaterOrEqual(len(ms), 0)
}

func TestUtil(t *testing.T) {
	input := &Input{
		Host: "127.0.0.1",
		Port: testMysqlPort,
		User: "root",
		Pass: "123456",
		Tags: make(map[string]string),
	}

	err := input.initCfg()
	assert.NoError(t, err)

	t.Run("mysqlVersion", func(t *testing.T) {
		version, err := getVersion(input.db)
		assert.NoError(t, err)

		// mock
		versions := []string{"5.6.36", "5.6.36a"}
		for _, v := range versions {
			version.version = v
			assert.True(t, version.versionCompatible([]int{5, 6, 36}))
			assert.True(t, version.versionCompatible([]int{4, 6, 36}))
			assert.False(t, version.versionCompatible([]int{6, 6, 36}))
			assert.True(t, version.versionCompatible([]int{5, 4, 36}))
			assert.False(t, version.versionCompatible([]int{5, 7, 36}))
			assert.True(t, version.versionCompatible([]int{5, 6, 35}))
			assert.False(t, version.versionCompatible([]int{5, 6, 37}))
		}
	})

	t.Run("canExplain", func(t *testing.T) {
		assert.True(t, canExplain("select * from demo"))
		assert.False(t, canExplain("alter table"))
	})

	t.Run("cacheLimit", func(t *testing.T) {
		cache := cacheLimit{
			Size: 5,
			TTL:  10,
		}
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("key-%v", i)
			assert.True(t, cache.Acquire(key))
		}
		assert.False(t, cache.Acquire("key"))
		time.Sleep(15 * time.Second)
		assert.True(t, cache.Acquire("key"))
	})
} */

// const (
// 	testMysqlHost     = "47.98.119.69"
// 	testMysqlPort     = 3306
// 	testMysqlPassword = "12345@"
// )

// // go test -v -timeout 30s -run ^TestCollect$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
// func TestCollect(t *testing.T) {
// 	input := &Input{
// 		Host: testMysqlHost,
// 		Port: testMysqlPort,
// 		User: "root",
// 		Pass: testMysqlPassword,
// 		Tags: make(map[string]string),
// 	}

// 	pts, err := input.Collect()
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, pts, "collect empty!")

// 	t.Logf("pts = %v", pts)
// }

// // go test -v -timeout 30s -run ^TestMetricCollectMysqlGeneral$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
// func TestMetricCollectMysqlGeneral(t *testing.T) {
// 	input := &Input{
// 		Host: testMysqlHost,
// 		Port: testMysqlPort,
// 		User: "root",
// 		Pass: testMysqlPassword,
// 		Tags: make(map[string]string),
// 	}

// 	err := input.initDBConnect()
// 	assert.NoError(t, err)

// 	input.initDbm()

// 	cases := []struct {
// 		name string
// 		fun  func() ([]*io.Point, error)
// 	}{
// 		{
// 			name: "CollectMysql",
// 			fun:  input.metricCollectMysql,
// 		},
// 		{
// 			name: "CollectMysqlSchema",
// 			fun:  input.metricCollectMysqlSchema,
// 		},
// 		{
// 			name: "CollectMysqlTableSschema",
// 			fun:  input.metricCollectMysqlTableSschema,
// 		},
// 		{
// 			name: "CollectMysqlUserStatus",
// 			fun:  input.metricCollectMysqlUserStatus,
// 		},
// 		// {
// 		// 	name: "metricCollectMysqlCustomQueries",
// 		// 	fun:  input.metricCollectMysqlCustomQueries,
// 		// },
// 		{
// 			name: "CollectMysqlInnodb",
// 			fun:  input.metricCollectMysqlInnodb,
// 		},
// 		{
// 			name: "CollectMysqlDbmMetric",
// 			fun:  input.metricCollectMysqlDbmMetric,
// 		},
// 		{
// 			name: "CollectMysqlDbmSample",
// 			fun:  input.metricCollectMysqlDbmSample,
// 		},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			input.start = time.Now()

// 			if tc.name == "CollectMysqlDbmMetric" {
// 				// 这个测试需要运行两遍 fun() 才能通过
// 				// go test 也要运行两遍
// 				tc.fun() //nolint:errcheck
// 			}

// 			pts, err := tc.fun()
// 			assert.NoError(t, err)
// 			assert.NotEmpty(t, pts, "collect empty!")

// 			t.Logf("pts = %v", pts)
// 		})
// 	}
// }
