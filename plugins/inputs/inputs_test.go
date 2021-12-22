package inputs

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestGetEnvs(t *testing.T) {
	cases := []struct {
		envs   map[string]string
		expect map[string]string
	}{
		{
			envs:   map[string]string{"ENV_A": "a", "DK_B": "b", "ABC_X": "x"},
			expect: map[string]string{"ENV_A": "a", "DK_B": "b"},
		},

		{
			envs:   map[string]string{"ENV_A": "===a", "DK_B": "b", "ABC_X": "x"},
			expect: map[string]string{"ENV_A": "===a", "DK_B": "b"},
		},

		{
			envs:   map[string]string{"ENV_A": "===a", "DK_B": "-=-=-=a:b", "ABC_X": "x"},
			expect: map[string]string{"ENV_A": "===a", "DK_B": "-=-=-=a:b"},
		},

		{
			envs:   map[string]string{"ENV_": "===a", "DK_": "-=-=-=a:b", "ABC_X": "x"},
			expect: map[string]string{"ENV_": "===a", "DK_": "-=-=-=a:b"},
		},
	}

	for _, tc := range cases {
		for k, v := range tc.envs {
			if err := os.Setenv(k, v); err != nil {
				t.Error(err)
			}
		}

		envs := getEnvs()
		for k, v := range tc.expect {
			val, ok := envs[k]
			tu.Assert(t, val == v && ok, fmt.Sprintf("%s <> %s && %v", v, val, ok))
		}
	}
}

func TestEventMeasurement(t *testing.T) {
	m := getReporterMeasurement(&io.Reporter{Status: "info", Message: "message"})
	assert.Equal(t, m.tags["source"], "datakit")
	assert.Equal(t, m.tags["status"], "info")
	assert.Equal(t, m.fields["message"], "message")

	_, err := m.LineProto()
	assert.NoError(t, err)
}
