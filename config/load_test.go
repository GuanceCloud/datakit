package config

import (
	"testing"
)

func initGitRepos() {
	repos := []*GitRepository{
		{
			Enable:                true,
			URL:                   "ssh://git@gitlab.jiagouyun.com:40022/jack/conf.git",
			SSHPrivateKeyPath:     "id.rsa",
			SSHPrivateKeyPassword: "",
			Branch:                "master",
		},
		{
			Enable:                false,
			URL:                   "ssh://git@gitlab.jiagouyun.com:40022/jack/conf2.git",
			SSHPrivateKeyPath:     "id.rsa",
			SSHPrivateKeyPassword: "",
			Branch:                "master",
		},
		{
			Enable:                true,
			URL:                   "ssh://git@gitlab.jiagouyun.com:40022/jack/conf3.git",
			SSHPrivateKeyPath:     "id.rsa",
			SSHPrivateKeyPassword: "",
			Branch:                "master",
		},
		{
			Enable:                true,
			URL:                   "ssh://git@gitlab.jiagouyun.com:40022/jack/not_exist.git",
			SSHPrivateKeyPath:     "id.rsa",
			SSHPrivateKeyPassword: "",
			Branch:                "master",
		},
	}

	Cfg.GitRepos.Repos = append(Cfg.GitRepos.Repos, repos...)
}

// go test -v -timeout 30s -run ^TestGetPipelinePath$ gitlab.jiagouyun.com/cloudcare-tools/datakit/config
func TestGetPipelinePath(t *testing.T) {
	initGitRepos()

	cases := []struct {
		name     string
		pipeline string
	}{
		{
			name:     "absolute_path",
			pipeline: "/usr/local/datakit/gitrepos/conf/nginx.p",
		},
		{
			name:     "relative_path",
			pipeline: "nginx.p",
		},
		{
			name:     "absolute_path_not_exist",
			pipeline: "/usr/local/datakit/gitrepos/conf/not_exist.p",
		},
		{
			name:     "relative_path_not_exist",
			pipeline: "not_exist.p",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss, err := GetPipelinePath(tc.pipeline)
			if err != nil {
				t.Logf("GetPipelinePath failed: %v", err)
				return
			}

			t.Log(ss)
		})
	}
}
