// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"testing"

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
			ss, err := GetPipelinePath(datakit.Logging, tc.pipeline)
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
