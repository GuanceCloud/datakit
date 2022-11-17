// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package gitrepo ...
package gitrepo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	giturls "github.com/whilp/git-urls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	ssh2 "golang.org/x/crypto/ssh"
)

var (
	l       = logger.DefaultSLogger("gitrepo")
	runGit  sync.Once
	isFirst = true
)

const (
	prefixSSH  = "ssh://"
	prefixGit  = "git@"
	prefixHTTP = "http"

	prefixGitBranchName = "refs/heads/"

	authUseHTTP = 1
	authUseSSH  = 2

	errTextSameNameNotReally = "unable to authenticate, attempted methods [none publickey], no supported methods remain"
)

type authOpt struct {
	Auth        int
	GitUserName string
	GitPassword string
}

func StartPull() error {
	runGit.Do(func() {
		l = logger.SLogger("gitrepo")
		g := datakit.G("gitrepo")

		g.Go(func(ctx context.Context) error {
			return pullMain(config.Cfg.GitRepos)
		})
	})

	return nil
}

func pullMain(cg *config.GitRepost) error {
	l.Info("start")

	pi, err := time.ParseDuration(cg.PullInterval)
	if err != nil {
		l.Warnf("parse pull interval failed: %v, default to 1 minute", err)
		pi = time.Minute
	}

	tick := time.NewTicker(pi)
	defer tick.Stop()

	for {
		// git start pull immediately
		l.Debug("triggered")
		for _, v := range cg.Repos {
			if !v.Enable {
				continue
			}
			if err = doRun(v); err != nil {
				if isFirst {
					if strings.Contains(err.Error(), "connect: operation timed out") {
						isFirst = false

						if err := inputs.RunInputs(); err != nil {
							l.Error("error running inputs: %v", err)
						} else {
							l.Info("first run inputs succeeded")
						}
					}
				}

				tip := fmt.Sprintf("[gitrepo] failed: %v", err)
				l.Error(tip)
				io.SelfError(tip)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return nil

		case <-tick.C:
			// empty here
		} // select
	} // for
}

func doRun(c *config.GitRepository) error {
	clonePath, err := getGitClonePathFromGitURL(c.URL)
	if err != nil {
		return err
	}

	as, err := getUserNamePasswordFromGitURL(c.URL)
	if err != nil {
		l.Errorf("getUserNamePasswordFromGitURL failed: %v, url = %s", err, c.URL)
		return err
	}

	if as.Auth == authUseSSH && len(c.SSHPrivateKeyPath) == 0 {
		// use ssh to auth
		tip := "ssh need key file"
		l.Error(tip)
		return fmt.Errorf(tip)
	}

	authMethod, err := getAuthMethod(as, c)
	if err != nil {
		l.Errorf("getAuthMethod failed: %v, url = %s", err, c.URL)
		return err
	}

	isUpdate := true

	// check git repo exist
	if path.IsDir(clonePath) {
		// clone in the exist dir
		l.Debug("Pull start")

		isUpdate, err = gitPull(clonePath, c.Branch, authMethod)
		if err != nil {
			return err
		}
	} else {
		// clone a new one
		l.Debug("PlainClone start")

		if err := gitPlainClone(clonePath, c.URL, c.Branch, authMethod); err != nil {
			l.Errorf("gitPlainClone failed: %v", err)
			return err
		}
	}

	if isUpdate || isFirst {
		l.Info("reload")

		if isFirst {
			isFirst = false
		}

		ctxNew, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if _, err = reloadCore(ctxNew); err != nil {
			return err
		}
	}

	l.Debug("completed")
	return nil
}

func gitPull(clonePath, branch string, authMethod transport.AuthMethod) (isUpdate bool, err error) {
	isUpdate = true // default is true

	// We instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(clonePath)
	if err != nil {
		l.Errorf("PlainOpen failed: %v", err)
		return
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		l.Errorf("Worktree failed: %v", err)
		return
	}

	ctxNew, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.PullContext(ctxNew, &git.PullOptions{
		RemoteName:      "origin",
		ReferenceName:   plumbing.NewBranchReferenceName(branch),
		Auth:            authMethod,
		Force:           true,
		InsecureSkipTLS: true,
	})
	if err != nil {
		// ignore specific errors
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			isUpdate = false // NOTE: not continue here
		} else {
			l.Errorf("Pull failed: %v", err)

			if strings.Contains(err.Error(), errTextSameNameNotReally) {
				if err := os.RemoveAll(clonePath); err != nil {
					l.Warnf("failed remove clone path %s with error %v", clonePath, err)
				} else {
					l.Infof("succeeded remove clone path %s", clonePath)
				}
			}
			return
		}
	}

	// get branch name, if branch exists, then create flag is false
	flagCreate := true
	hrf, err := r.Head()
	if err != nil {
		l.Errorf("Head failed: %v", err)
		return
	}
	if hrf.Name().IsBranch() {
		branchName := string(hrf.Name())
		branchName = strings.TrimPrefix(branchName, prefixGitBranchName)
		if branchName == branch {
			flagCreate = false
		}
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  true,
		Create: flagCreate,
	})
	if err != nil {
		l.Errorf("Checkout failed: %v", err)
		return
	}
	return isUpdate, nil
}

func gitPlainClone(clonePath, gitURL, branch string, authMethod transport.AuthMethod) error {
	ctxNew, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := git.PlainCloneContext(ctxNew, clonePath, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:            authMethod,
		URL:             gitURL,
		InsecureSkipTLS: true,
		RemoteName:      "origin",
		ReferenceName:   plumbing.NewBranchReferenceName(branch),
	}); err != nil {
		return err
	}
	return nil
}

func reloadCore(ctx context.Context) (int, error) {
	round := 0 // 循环次数
	for {
		select {
		case <-ctx.Done():
			tip := "reload timeout"
			l.Error(tip)
			return round, fmt.Errorf(tip)
		default:
			switch round {
			case 0:
				l.Info("before ReloadCheckInputCfg")

				_, err := config.ReloadCheckInputCfg()
				if err != nil {
					l.Errorf("ReloadCheckInputCfg failed: %v", err)
					return round, err
				}

				l.Info("before ReloadCheckPipelineCfg")

			case 1:
				l.Info("before StopInputs")

				if err := inputs.StopInputs(); err != nil {
					l.Errorf("StopInputs failed: %v", err)
					return round, err
				}

			case 2:
				l.Info("before ReloadInputConfig")

				if err := config.ReloadInputConfig(); err != nil {
					l.Errorf("ReloadInputConfig failed: %v", err)
					return round, err
				}

			case 3:
				l.Info("before set pipelines")

				plscript.LoadAllScripts2StoreFromPlStructPath(plscript.GitRepoScriptNS,
					filepath.Join(datakit.GitReposRepoFullPath, "pipeline"))

			case 4:
				l.Info("before RunInputs")

				httpd.CleanHTTPHandler()
				if err := inputs.RunInputs(); err != nil {
					l.Errorf("RunInputs failed: %v", err)
					return round, err
				}

			case 5:
				l.Info("before ReloadTheNormalServer")

				httpd.ReloadTheNormalServer()
			} // switch round
		} // select

		round++
		if round > 6 {
			return round, nil // round + 1
		} // if round
	} // for
}

func getGitClonePathFromGitURL(gitURL string) (string, error) {
	repoName, err := path.GetGitPureName(gitURL)
	if err != nil {
		l.Errorf("GetGitPureName failed: %v, url = %s", err, gitURL)
		return "", err
	}
	clonePath, err := config.GetGitRepoDir(repoName)
	if err != nil {
		l.Errorf("GetGitRepoDir failed: %v, repo = %s", err, repoName)
		return "", err
	}
	return clonePath, nil
}

// whether use username & password auth.
func isUserNamePasswordAuth(gitURL string) (bool, error) {
	gURL := strings.ToLower(gitURL)
	switch {
	case strings.HasPrefix(gURL, prefixHTTP):
		return true, nil
	case strings.HasPrefix(gURL, prefixSSH), strings.HasPrefix(gURL, prefixGit):
		return false, nil
	default:
		tip := "invalid git url"
		l.Error(tip)
		return false, fmt.Errorf(tip)
	}
}

func getUserNamePasswordFromGitURL(gitURL string) (*authOpt, error) {
	// http only could use username & password auth
	// ssh only could use private key auth
	uGitURL, err := giturls.Parse(gitURL)
	if err != nil {
		l.Errorf("url.Parse failed: %v, url = %s", err, gitURL)
		return nil, err
	}

	bIsUseAuthUserNamePassword, err := isUserNamePasswordAuth(gitURL)
	if err != nil {
		l.Errorf("isUserNamePasswordAuth failed: %v, url = %s", err, gitURL)
		return nil, err
	}

	var auth int
	var gitUserName, gitPassword string

	if bIsUseAuthUserNamePassword {
		gitUserName = uGitURL.User.Username()
		gitPassword, _ = uGitURL.User.Password()
		auth = authUseHTTP
	} else {
		auth = authUseSSH
	}

	return &authOpt{
		Auth:        auth,
		GitUserName: gitUserName,
		GitPassword: gitPassword,
	}, nil
}

func getAuthMethod(as *authOpt, c *config.GitRepository) (transport.AuthMethod, error) {
	if as == nil {
		return nil, fmt.Errorf("invalid auth struct")
	}

	var authMethod transport.AuthMethod

	switch as.Auth {
	case authUseHTTP:
		// use username & password to auth
		if len(as.GitUserName) == 0 {
			authMethod = nil
		} else {
			authMethod = &http.BasicAuth{
				Username: as.GitUserName,
				Password: as.GitPassword,
			}
		}
		l.Debugf("authMethod = %#v", authMethod)

	case authUseSSH:
		// use ssh to auth
		l.Debug("use ssh to auth")

		if _, err := os.Stat(c.SSHPrivateKeyPath); err != nil {
			l.Errorf("read file %s failed %s\n", c.SSHPrivateKeyPath, err.Error())
			return nil, err
		}
		// Clone the given repository to the given directory
		publicKeys, err := ssh.NewPublicKeysFromFile("git", c.SSHPrivateKeyPath, c.SSHPrivateKeyPassword)
		if err != nil {
			l.Errorf("generate publickeys failed: %s\n", err.Error())
			return nil, err
		}

		publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey() //nolint:errcheck,gosec
		authMethod = publicKeys

	default:
		return nil, fmt.Errorf("not supported auth method")
	} // switch

	return authMethod, nil
}
