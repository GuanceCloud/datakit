package cmds

import (
	"encoding/json"
	"testing"
)

func Test_switchToken(t *testing.T) {
	wss = []*workerSpace{
		{Name: "测试01", Token: "token01", ID: "ID"},
		{Name: "测试02", Token: "token02", ID: "ID"},
	}

	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "case1", args: args{s: "use token01"}},
		{name: "case2", args: args{s: "use token02"}},
		{name: "case3", args: args{s: "use"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switchToken(tt.args.s)
		})
	}
}

func Test_cache(t *testing.T) {
	jsonStr := `{
      "series": [
        {
          "name": "show_workspaces",
          "columns": [
            "wsuuid",
            "token",
            "expireAt",
            "createAt",
            "name"
          ],
          "values": [
            [
              "wksp_2dc431d6693711eb8ff97aeee04b54af",
              "tkn_2dc438b6693711eb8ff97aeee04b54af",
              "1641384248",
              "1641282903",
              "开发测试一起用"
            ],
            [
              "wksp_system",
              "tokn_bW47smmgQpoZKP5A2xKuj8W2",
              "",
              "",
              "系统工作空间#"
            ],
            [
              "wksp_1fcd93a0766c11ebad5af2b2c21faf74",
              "tkn_1fcd9a08766c11ebad5af2b2c21faf74",
              "1641283729",
              "1641283729",
              "解决方案中心"
            ]
          ]
        }
      ],
      "cost": "",
      "is_running": false,
      "async_id": ""
    }`
	qr := &queryResult{}
	err := json.Unmarshal([]byte(jsonStr), qr)
	if err != nil {
		t.Errorf("json err=%v", err)
		return
	}
	type args struct {
		c *queryResult
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "case", args: args{c: qr}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache(tt.args.c)
			for _, space := range wss {
				t.Logf("name=%s  token=%s", space.Name, space.Token)
			}
		})
	}
}
