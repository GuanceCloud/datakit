package dialtesting

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestJSONTaskFile(t *testing.T) {
	cases := []struct {
		name, j string
		fail    bool
	}{
		{
			j:    `{"a": [{"url": "http://172.16.5.9:13001", "name": "auth北京", "method": "GET", "region": "beijing", "status": "stop", "post_url": "http://testing-openway.cloudcare.cn?token=tkn_2dc438b6693711eb8ff97aeee04b54af", "frequency": "1m", "access_key": "tfAA3qeo5AOB2kEflcZA", "external_id": "dial_0b34e8383ff145e997818c1e2ecd8cf7", "update_time": 1618299681542678, "success_when": [{"header": {"jyd": [{"contains": "cccc"}]}}], "advance_options": {}}]}`,
			name: `normal case`,
		},
	}

	for _, tc := range cases {
		_ = tc
		t.Run(tc.name, func(t *testing.T) {
			i := Input{}
			b, err := i.getLocalJSONTasks([]byte(tc.j))
			if tc.fail {
				tu.NotOk(t, err, "expect err, got none")
			}
			tu.Ok(t, err)

			t.Logf(string(b))
		})
	}
}
