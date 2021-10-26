package cmds

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"

	"github.com/influxdata/influxdb1-client/models"
)

var (
	flagCSV   = "/Users/macbook/go/datakit/x.csv"
	flagForce = true
)

func TestWriteToCsv(t *testing.T) {
	series := []*models.Row{}
	body, _ := os.ReadFile("test.json")
	_ = json.Unmarshal(body, &series)
	file, err := os.OpenFile(flagCSV, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		t.Fatalf("err:%v", err)
	}
	if err := writeToCsv(series, flagCSV); err != nil {
		t.Fatalf("err:%v", err)
	}
	defer file.Close()
}

func TestConvertToString(t *testing.T) {
	series := []*models.Row{}
	body, _ := os.ReadFile("test.json")
	_ = json.Unmarshal(body, &series)

	for _, v := range series[0].Values {
		res := convertStrings(v)
		for _, value := range res {
			t.Logf(value)
			tu.Equals(t, "string", reflect.TypeOf(value).String())
		}
	}
}
