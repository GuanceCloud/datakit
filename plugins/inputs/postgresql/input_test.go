// +build !test

package postgresql

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/araddon/dateparse"
	"github.com/stretchr/testify/assert"
	dk "gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkInputs "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type MockCollectService struct {
	Address        string
	mockData       map[string]*interface{}
	columnError    int
	columnMapError int
	startError     int
}

func (m *MockCollectService) GetColumnMap(row scanner, columns []string) (map[string]*interface{}, error) {
	if m.columnMapError == 1 {
		return nil, errorMock{}
	}
	return m.mockData, nil
}

func (m *MockCollectService) Query(query string) (Rows, error) {
	if query == "-1" {
		return nil, errorMock{}
	}
	rows := &MockCollectRows{
		columnError: m.columnError,
	}
	return rows, nil
}

func (m *MockCollectService) Stop() error { return nil }
func (m *MockCollectService) Start() error {
	if m.startError == 1 {
		return errorMock{}
	}
	return nil
}
func (m *MockCollectService) SetAddress(address string) {}

type MockCollectRows struct {
	calledNext  bool
	columnError int
}

func (m *MockCollectRows) Close() error { return nil }
func (m *MockCollectRows) Columns() ([]string, error) {
	if m.columnError == 1 {
		return nil, errorMock{}
	}
	return []string{}, nil
}
func (m *MockCollectRows) Next() bool {
	isCall := !m.calledNext
	m.calledNext = true
	return isCall
}
func (m *MockCollectRows) Scan(...interface{}) error { return nil }

func getMockData(points map[string]interface{}) map[string]*interface{} {
	data := make(map[string]*interface{})
	for key, val := range points {
		v := new(interface{})
		*v = val
		data[key] = v
	}
	return data
}

func getTrueData(mockFields map[string]interface{}) map[string]interface{} {
	trueData := make(map[string]interface{})
	for k, v := range mockFields {
		if _, ok := postgreFields[k]; ok {
			switch trueV := v.(type) {
			case []uint8:
				trueData[k] = string(trueV)
			default:
				trueData[k] = v
			}
		}
	}

	return trueData
}

func TestCollect(t *testing.T) {
	input := &Input{}
	input.service = &MockCollectService{}
	err := input.Collect()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Nil(t, input.collectCache)

	mockFields := map[string]interface{}{
		"numbackends":    int64(1),
		"invalid_fields": int64(1),
		"blks_read":      []uint8("blks_read"),
		"datname":        []uint8("datname"),
	}

	input.service = &MockCollectService{
		mockData: getMockData(mockFields),
	}
	input.Tags = map[string]string{"self": "self"}
	err = input.Collect()
	assert.NoError(t, err)
	assert.Greater(t, len(input.collectCache), 0, "input collectCache should has at least one measurement")
	points, err := input.collectCache[0].LineProto()
	assert.NoError(t, err)
	fields, err := points.Fields()
	assert.NoError(t, err)
	trueFields := getTrueData(mockFields)
	assert.True(t, reflect.DeepEqual(trueFields, fields))
	tags := points.Tags()
	assert.Equal(t, tags["self"], "self")
	assert.Equal(t, tags["db"], "datname")

	// work correctly when set IgnoredDatabases and Databases
	input.IgnoredDatabases = []string{"a"}
	assert.NoError(t, input.Collect())
	input.Databases = []string{"a"}
	input.IgnoredDatabases = []string{}
	assert.NoError(t, input.Collect())

	// when start() error
	input.service = &MockCollectService{
		startError: 1,
	}
	assert.Error(t, input.Collect())

	// when getDbMetrics() error
	input.service = &MockCollectService{
		columnMapError: 1,
	}
	assert.Error(t, input.Collect())

}

func TestParseUrl(t *testing.T) {
	uri := "postgres://postgres@localhost/test?sslmode=disable"
	parsedUri, err := parseURL(uri)
	assert.NoError(t, err)
	assert.NotNil(t, parsedUri)

	uri = "postgres://postgres@localhost:[]/test?sslmode=disable"
	parsedUri, err = parseURL(uri)
	assert.Error(t, err)
	assert.Equal(t, parsedUri, "")

	parsedUri, err = parseURL("我们")
	assert.Error(t, err)
	assert.Equal(t, parsedUri, "")

	parsedUri, err = parseURL("postgres://postgres@localhost:8888/test?sslmode=disable")
	assert.NoError(t, err)
	assert.NotEmpty(t, parsedUri)

}

func TestInput(t *testing.T) {
	input := &Input{}
	sampleMeasurements := input.SampleMeasurement()
	assert.Greater(t, len(sampleMeasurements), 0)
	m := sampleMeasurements[0].(*inputMeasurement)

	assert.Equal(t, m.Info().Name, inputName)

	assert.Equal(t, input.Catalog(), catalogName)
	assert.Equal(t, input.SampleConfig(), sampleConfig)
	assert.Equal(t, input.AvailableArchs(), dk.AllArch)

	assert.Equal(t, input.PipelineConfig()["postgresql"], pipelineCfg)

	t.Run("executeQuery", func(t *testing.T) {
		var err error
		input.service = &MockCollectService{}
		err = input.executeQuery("")
		assert.NoError(t, err)

		// when service.Query() error
		err = input.executeQuery("-1")
		assert.Error(t, err)

		// when rows.Columns() error
		input.service = &MockCollectService{
			columnError: 1,
		}
		err = input.executeQuery("")
		assert.Error(t, err)

		// when GetColumnMap() error
		input.service = &MockCollectService{
			columnMapError: 1,
		}
		err = input.executeQuery("")
		assert.Error(t, err)

		//when accRow() error
		input.service = &MockCollectService{}
		input.Address = "postgres://:888localhost"
		err = input.executeQuery("")
		assert.Error(t, err)

	})
}

func TestSanitizedAddress(t *testing.T) {
	input := &Input{}

	input.Address = "postgres://xxxxx"
	transAddress, err := input.SanitizedAddress()
	assert.NoError(t, err)
	assert.Equal(t, transAddress, "host=xxxxx")

	address := "address"
	input.Outputaddress = address
	transAddress, err = input.SanitizedAddress()
	assert.NoError(t, err)
	assert.Equal(t, transAddress, address)

	input.Address = "postgres://:888localhost"
	input.Outputaddress = ""
	transAddress, err = input.SanitizedAddress()
	assert.Error(t, err)
	assert.Equal(t, transAddress, "")

}

type DkInputsMock struct {
	tailerError int
}
type TailerMock struct{}

func (TailerMock) Run() {}
func (DkInputsMock) FeedMeasurement(name, category string, measurements interface{}, opt interface{}) error {
	return nil
}

func (d DkInputsMock) NewTailer(opt interface{}) (Tailer, error) {
	if d.tailerError == 1 {
		return nil, errorMock{}
	}
	return TailerMock{}, nil
}

func (DkInputsMock) JoinPipelinePath(op interface{}, defaultPipeline string) {
	dkInputs.JoinPipelinePath(op.(*dkInputs.TailerOption), defaultPipeline)
}
func TestRun(t *testing.T) {
	mockFields := map[string]interface{}{
		"numbackends": int64(1),
	}

	input := NewInput(&MockCollectService{
		mockData: getMockData(mockFields),
	})

	inputs := DkInputsMock{}

	datakit := Datakit{
		ch: make(chan interface{}),
	}
	config.ProtectedInterval = func(min, max, cur time.Duration) time.Duration {
		return min
	}
	datakit.Exit = func() <-chan interface{} {
		return datakit.ch
	}
	go func() {
		input.runService(inputs, datakit)

	}()
	time.Sleep(5 * time.Second)
	assert.Nil(t, input.tail)
	datakit.Close()

	t.Run("Log", func(t *testing.T) {
		datakit.ch = make(chan interface{})
		input.Log = &dkInputs.TailerOption{}
		input.Tags = map[string]string{"tag": "pg"}
		go func() {
			input.runService(inputs, datakit)
		}()
		time.Sleep(5 * time.Second)
		assert.NotNil(t, input.tail)
		datakit.Close()

		t.Run("NewTailer error", func(t *testing.T) {
			datakit.ch = make(chan interface{})
			input.tail = nil
			inputs.tailerError = 1
			go func() {
				input.runService(inputs, datakit)
			}()
			time.Sleep(5 * time.Second)
			assert.Nil(t, input.tail)
			datakit.Close()
		})
	})

}

type DbMock struct {
}

func (DbMock) SetMaxOpenConns(n int)            {}
func (DbMock) SetMaxIdleConns(n int)            {}
func (DbMock) SetConnMaxLifetime(time.Duration) {}
func (DbMock) Close() error {
	return nil
}

func (DbMock) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if query == "-1" {
		return nil, errorMock{}
	}
	return nil, nil
}

type errorMock struct {
}

func (e errorMock) Error() string {
	return "error"
}

type RowScanner struct {
	data []int
}

func (r RowScanner) Scan(dest ...interface{}) error {
	for i := 0; i < len(r.data); i++ {
		d, ok := dest[i].(*interface{})
		if r.data[i] == -1 { // mock error
			return errorMock{}
		}
		if ok {
			*d = r.data[i]
		}
	}
	return nil
}

func TestService(t *testing.T) {
	s := &SqlService{
		MaxIdle:     1,
		MaxOpen:     1,
		MaxLifetime: time.Duration(0),
	}
	s.Open = func(dbType, connStr string) (Db, error) {
		db := &DbMock{}
		return db, nil
	}

	err := s.Start()
	assert.Nil(t, err)

	s.Open = func(dbType, connStr string) (Db, error) {
		db := &DbMock{}
		return db, errorMock{}
	}
	err = s.Start()
	assert.NotNil(t, err)
	assert.Nil(t, s.Stop())

	// Query
	t.Run("Query", func(t *testing.T) {
		rows, err := s.Query("query")
		assert.Nil(t, err)
		assert.Nil(t, rows)

		t.Run("should catch error", func(t *testing.T) {
			rows, err := s.Query("-1")
			assert.Nil(t, rows)
			assert.NotNil(t, err)
		})
	})

	// SetAddress
	t.Run("SetAddress", func(t *testing.T) {
		add := "localhost"
		s.SetAddress(add)
		assert.Equal(t, s.Address, add)
	})

	// GetColumnMap
	t.Run("GetColumnMap", func(t *testing.T) {
		row := RowScanner{
			data: []int{1},
		}
		columns := []string{"a"}
		res, err := s.GetColumnMap(row, columns)
		assert.Nil(t, err)
		assert.Equal(t, row.data[0], *res["a"])

		t.Run("catch error", func(t *testing.T) {
			row.data[0] = -1
			res, err = s.GetColumnMap(row, columns)
			assert.NotNil(t, err)
			assert.Nil(t, res)
		})
	})

}

func TestTime(t *testing.T) {
	ti, err := dateparse.ParseIn("2014-12-16 06:20:00 UTC", time.Local)
	fmt.Println(ti, err)
	fmt.Println(ti.UnixNano())

	fmt.Println("ok")

}
