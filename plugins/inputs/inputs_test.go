package inputs

import (
	"fmt"
	"os"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
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
			os.Setenv(k, v)
		}

		envs := getEnvs()
		for k, v := range tc.expect {
			val, ok := envs[k]
			tu.Assert(t, val == v && ok, fmt.Sprintf("%s <> %s && %v", v, val, ok))
		}
	}
}
