// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	giturls "github.com/whilp/git-urls"
)

func TestGitURL(t *testing.T) {
	const e = "https://username:password@github.com/username/repository.git"

	u, err := giturls.Parse(e)
	assert.NoError(t, err)

	assert.Equal(t, "username", u.User.Username())
	pass, ok := u.User.Password()
	assert.Equal(t, "password", pass)
	assert.True(t, ok)
}

func TestGetCPUMax(t *testing.T) {
	getNumCPU = func() int { return 4 }

	testcases := []struct {
		inQuota, inPeriod string
		fail              bool
		value             float64
	}{
		{
			inQuota:  "100000",
			inPeriod: "100000",
			value:    1,
		},
		{
			inQuota:  "800000",
			inPeriod: "100000",
			value:    4, // max NumCPU
		},
		{
			inQuota:  "800000",
			inPeriod: "200000",
			value:    4,
		},
		{
			inQuota:  "max",
			inPeriod: "100000",
			value:    4,
		},
		{
			inQuota:  "-1",
			inPeriod: "100000",
			value:    4,
		},
		{
			inQuota:  "invalid",
			inPeriod: "",
			fail:     true,
		},
		{
			inQuota:  "",
			inPeriod: "invalid",
			fail:     true,
		},
		{
			inQuota:  "0",
			inPeriod: "100000",
			fail:     true,
		},
		{
			inQuota:  "-100000",
			inPeriod: "100000",
			fail:     true,
		},
		{
			inQuota:  "100000",
			inPeriod: "0",
			fail:     true,
		},
	}

	for _, tc := range testcases {
		res, err := parseCurrentCPUMax(tc.inQuota, tc.inPeriod)
		if tc.fail {
			assert.Error(t, err)
			continue
		}

		assert.Nil(t, err)
		assert.Equal(t, tc.value, res)
	}
}

// Uncomment this if want to test git clone for really
// func TestGitClone(t *testing.T) {
// 	urls := []string{
// 		// gitee or gitlab, same
// 		"git@gitee.com:username/gitrepo1.git",
// 		"https://gitee.com/username/gitrepo1.git",

// 		// gitlab
// 		"ssh://git@gitlab.jiagouyun.com:40022/username/conf.git",
// 		"http://gitlab.jiagouyun.com/username/conf.git",
// 	}

// 	const key = "/Users/mac/.ssh/id_rsa"

// 	var authMethod transport.AuthMethod
// 	if _, err := os.Stat(key); err != nil {
// 		t.Log(err)
// 		return
// 	}
// 	// Clone the given repository to the given directory
// 	publicKeys, err := ssh.NewPublicKeysFromFile("git", key, "")
// 	if err != nil {
// 		t.Log(err)
// 		return
// 	}

// 	publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey() //nolint:errcheck,gosec
// 	authMethod = publicKeys

// 	for _, v := range urls {
// 		t.Log("\n--------------------------\n")

// 		clonePath := fmt.Sprintf("/Users/mac/Downloads/project/other/test/%d", time.Now().UnixNano())
// 		if _, err := git.PlainClone(clonePath, false, &git.CloneOptions{
// 			// The intended use of a GitHub personal access token is in replace of your password
// 			// because access tokens can easily be revoked.
// 			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
// 			Auth:            authMethod,
// 			URL:             v,
// 			InsecureSkipTLS: true,
// 		}); err != nil {
// 			t.Logf("clone [%s] failed: [%v]", v, err)
// 			continue
// 		}

// 		t.Logf("clone [%s] okey", v)
// 	}

// 	t.Log("\n--------------------------\n")
// 	t.Log("git all ok")
// }

// go test -v -timeout 30s -run ^TestParseUserNamePasswd$ gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit
func TestParseUserNamePasswd(t *testing.T) {
	gitURLs := []string{
		"https://username:password@github.com/username/repository.git",
		"https://username@github.com/username/repository.git",
	}

	for _, v := range gitURLs {
		t.Log("\n--------------------------\n")

		_, err := giturls.Parse(v)
		if err != nil {
			t.Logf("parse [%s] failed: [%v]", v, err)
			continue
		}
		t.Logf("parse [%s] ok", v)
	}

	t.Log("\n--------------------------\n")
	t.Log("parse all ok")
}

// Uncomment this if want to test git clone for really
// func TestGitPull(t *testing.T) {
// 	const gitURL = "http://username:password@gitlab.jiagouyun.com/username/conf.git"
// 	const clonePath = "/Users/mac/Downloads/project/other/test/conf"

// 	uGitURL, err := giturls.Parse(gitURL)
// 	if err != nil {
// 		t.Logf("url.Parse failed: %v, url = %s", err, gitURL)
// 		return
// 	}

// 	var gitPassword string
// 	gitUserName := uGitURL.User.Username()
// 	if password, ok := uGitURL.User.Password(); !ok {
// 		t.Logf("invalid_git_password, url = %s", gitURL)
// 		return
// 	} else {
// 		gitPassword = password
// 	}

// 	authMethod := &http.BasicAuth{
// 		Username: gitUserName,
// 		Password: gitPassword,
// 	}

// 	// We instantiate a new repository targeting the given path (the .git folder)
// 	r, err := git.PlainOpen(clonePath)
// 	if err != nil {
// 		t.Logf("PlainOpen failed: %v", err)
// 		return
// 	}

// 	// Get the working directory for the repository
// 	w, err := r.Worktree()
// 	if err != nil {
// 		t.Logf("Worktree failed: %v", err)
// 		return
// 	}

// 	// Pull the latest changes from the origin remote and merge into the current branch
// 	err = w.Pull(&git.PullOptions{
// 		RemoteName:      "origin",
// 		ReferenceName:   plumbing.NewBranchReferenceName("master"),
// 		Auth:            authMethod,
// 		Force:           true,
// 		InsecureSkipTLS: true,
// 	})
// 	if err != nil {
// 		// ignore specific errors
// 		if errors.Is(err, git.NoErrAlreadyUpToDate) {
// 		} else {
// 			t.Logf("Pull failed: %v", err)
// 			return
// 		}
// 	}

// 	// get branch name
// 	hrf, err := r.Head()
// 	if err != nil {
// 		t.Logf("Head failed: %v", err)
// 		return
// 	}
// 	t.Log(hrf.Name())
// 	if hrf.Name().IsBranch() {
// 		t.Log(hrf.Name())
// 	}

// 	err = w.Checkout(&git.CheckoutOptions{
// 		Branch: plumbing.NewBranchReferenceName("master"),
// 		Force:  true,
// 		Create: true,
// 	})
// 	if err != nil {
// 		t.Logf("Checkout failed: %v", err)
// 		return
// 	}

// 	t.Log("TestGitPull ok")
// }
