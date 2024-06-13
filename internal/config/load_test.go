// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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

// go test -v -timeout 30s -run ^TestGetPipelinePath$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config
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
			ss, err := GetPipelinePath(point.Logging, tc.pipeline)
			if err != nil {
				t.Logf("GetPipelinePath failed: %v", err)
				return
			}

			t.Log(ss)
		})
	}
}

// go test -v -timeout 30s -run ^Test_getConfRootPaths$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config
func Test_getConfRootPaths(t *testing.T) {
	cases := []struct {
		name                 string
		gitReposRepoName     string
		gitReposRepoFullPath string
		expect               []string
	}{
		{
			name:                 "git_enabled",
			gitReposRepoName:     "conf",
			gitReposRepoFullPath: "/usr/local/datakit/gitrepos/conf",
			expect:               []string{"/usr/local/datakit/gitrepos/conf/conf.d"},
		},
		{
			name:   "git_disabled",
			expect: []string{"/usr/local/datakit/conf.d"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			datakit.GitReposRepoName = tc.gitReposRepoName
			datakit.GitReposRepoFullPath = tc.gitReposRepoFullPath

			out := getConfRootPaths()
			assert.Equal(t, tc.expect, out)
		})
	}
}

func Test_decodeEncs(t *testing.T) {
	// var cryp = "5w1UiRjWuVk53k96WfqEaGUYJ/Oje7zr8xmBeGa3ugI="
	cryStr := "HelloAES9*&."

	type args struct {
		data     []byte
		filepath string
		env      string
		aes      string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test1",
			args: args{data: []byte("abc ENC[this is int]")},
			want: []byte("abc ENC[this is int]"),
		},
		{
			name: "test-file",
			args: args{data: []byte("[[inputs.mysql]]\n  host = \"localhost\"\n  user = \"datakit\"\n  pass = \"ENC[file:///tmp/enc4dk]\"\n  port = 3306"), filepath: "/tmp/enc4dk"},
			want: []byte("[[inputs.mysql]]\n  host = \"localhost\"\n  user = \"datakit\"\n  pass = \"HelloAES9*&.\"\n  port = 3306"),
		},
		{
			name: "test-aes",
			args: args{data: []byte("abc ENC[aes://UtAgMlRRhvCiCE5Q1W8kAKhYerQ78LpY5I4Yx9ICZQ0=]"), aes: "0123456789abcdef"},
			want: []byte("abc " + cryStr),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.filepath != "" {
				f, err := os.OpenFile(tt.args.filepath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o755)
				if err != nil {
					t.Errorf("openfile err=%v", err)
					return
				}
				f.Write([]byte(cryStr))
				f.Write([]byte{'\n'})
				f.Sync()
				f.Close()
			}
			if tt.args.env != "" {
				os.Setenv("TEST_ENV_1", cryStr)
			}
			if tt.args.aes != "" {
				datakit.ConfigAESKey = tt.args.aes
			}
			assert.Equalf(t, tt.want, decodeEncs(tt.args.data), "feedEncs(%v)", tt.args.data)
		})
	}
}
