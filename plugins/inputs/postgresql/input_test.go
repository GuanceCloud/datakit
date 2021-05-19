// +build !test

package postgresql

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type MockCollectService struct {
	Address  string
	mockData map[string]*interface{}
}

func (m *MockCollectService) GetColumnMap(row scanner, columns []string) (map[string]*interface{}, error) {
	return m.mockData, nil
}

func (m *MockCollectService) Query(query string) (Rows, error) {
	rows := &MockCollectRows{}
	return rows, nil
}

func (m *MockCollectService) Stop() {}
func (m *MockCollectService) Start() error {
	return nil
}
func (m *MockCollectService) SetAddress(address string) {}

type MockCollectRows struct {
	calledNext bool
}

func (m *MockCollectRows) Close() error               { return nil }
func (m *MockCollectRows) Columns() ([]string, error) { return []string{}, nil }
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

func TestCollect(t *testing.T) {
	input := &Input{}
	input.service = &MockCollectService{}
	err := input.Collect()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Nil(t, input.collectCache)

	mockFields := map[string]interface{}{
		"numbackends": int64(1),
	}

	input.service = &MockCollectService{
		mockData: getMockData(mockFields),
	}
	err = input.Collect()
	assert.NoError(t, err)
	assert.Greater(t, len(input.collectCache), 0, "input collectCache should has at least one measurement")
	points, err := input.collectCache[0].LineProto()
	assert.NoError(t, err)
	fields, err := points.Fields()
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(mockFields, fields))

	input.IgnoredDatabases = []string{"a"}
	assert.NoError(t, input.Collect())

	input.Databases = []string{"a"}
	input.IgnoredDatabases = []string{}
	assert.NoError(t, input.Collect())
}

func TestParseUrl(t *testing.T) {
	uri := "postgres://postgres@localhost/test?sslmode=disable"
	parsedUri, err := parseURL(uri)
	assert.NoError(t, err)
	assert.NotNil(t, parsedUri)
}

func TestInput(t *testing.T) {
	input := &Input{}
	sampleMeasurements := input.SampleMeasurement()
	assert.Greater(t, len(sampleMeasurements), 0)
	m := sampleMeasurements[0].(*inputMeasurement)

	assert.Equal(t, m.Info().Name, inputName)

	assert.Equal(t, input.Catalog(), catalogName)
	assert.Equal(t, input.SampleConfig(), sampleConfig)
	assert.Equal(t, input.AvailableArchs(), datakit.AllArch)

	assert.Equal(t, input.PipelineConfig()["postgresql"], pipelineCfg)
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

}
