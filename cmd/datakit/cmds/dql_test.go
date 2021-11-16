package cmds

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/influxdata/influxdb1-client/models"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

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
