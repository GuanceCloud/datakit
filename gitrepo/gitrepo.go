// Package gitrepo ...
package gitrepo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
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
)

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
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return nil

		case <-tick.C:
			l.Debug("triggered")
			for _, v := range cg.Repos {
				if !v.Enable {
					continue
				}
				if err = pullCore(v); err != nil {
					tip := fmt.Sprintf("[gitrepo] failed: %v", err)
					l.Error(tip)
					io.SelfError(tip)
				}
			}
		} // select
	} // for
}

func pullCore(c *config.GitRepository) error {
	repoName, err := path.GetGitPureName(c.URL)
	if err != nil {
		l.Errorf("GetGitPureName failed: %v, url = %s", err, repoName)
		return err
	}
	clonePath, err := config.GetGitRepoDir(repoName)
	if err != nil {
		l.Errorf("GetGitRepoDir failed: %v, url = %s", err, repoName)
		return err
	}

	// http only could use username & password auth
	// ssh only could use private key auth
	uGitURL, err := giturls.Parse(c.URL)
	if err != nil {
		l.Errorf("url.Parse failed: %v, url = %s", err, c.URL)
		return err
	}

	var bIsUseAuthUserNamePassword bool // whether use username & password auth

	gURL := strings.ToLower(c.URL)
	switch {
	case strings.HasPrefix(gURL, prefixHTTP):
		bIsUseAuthUserNamePassword = true
	case strings.HasPrefix(gURL, prefixSSH), strings.HasPrefix(gURL, prefixGit):
		bIsUseAuthUserNamePassword = false
	default:
		tip := "invalid_git_url"
		l.Error(tip)
		return fmt.Errorf(tip)
	}

	var gitUserName, gitPassword string

	if bIsUseAuthUserNamePassword {
		gitUserName = uGitURL.User.Username()
		if password, ok := uGitURL.User.Password(); !ok {
			tip := "invalid_git_password"
			l.Error(tip)
			return fmt.Errorf(tip)
		} else {
			gitPassword = password
		}

		if gitUserName == "" || gitPassword == "" {
			tip := "http_need_username_password"
			l.Error(tip)
			return fmt.Errorf(tip)
		}
	} else if c.SSHPrivateKeyPath == "" {
		tip := "ssh_need_key_file"
		l.Error(tip)
		return fmt.Errorf(tip)
	}

	var authMethod transport.AuthMethod
	if bIsUseAuthUserNamePassword {
		authMethod = &http.BasicAuth{
			Username: gitUserName,
			Password: gitPassword,
		}
	} else {
		if _, err := os.Stat(c.SSHPrivateKeyPath); err != nil {
			l.Errorf("read file %s failed %s\n", c.SSHPrivateKeyPath, err.Error())
			return err
		}
		// Clone the given repository to the given directory
		publicKeys, err := ssh.NewPublicKeysFromFile("git", c.SSHPrivateKeyPath, c.SSHPrivateKeyPassword)
		if err != nil {
			l.Errorf("generate publickeys failed: %s\n", err.Error())
			return err
		}

		publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey() //nolint:errcheck,gosec
		authMethod = publicKeys
	}

	isUpdate := true

	// check git repo exist
	if path.IsDir(clonePath) {
		// clone in the exist dir
		l.Debug("Pull start")

		// We instantiate a new repository targeting the given path (the .git folder)
		r, err := git.PlainOpen(clonePath)
		if err != nil {
			l.Errorf("PlainOpen failed: %v", err)
			return err
		}

		// Get the working directory for the repository
		w, err := r.Worktree()
		if err != nil {
			l.Errorf("Worktree failed: %v", err)
			return err
		}

		ctxNew, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
		}()
		// Pull the latest changes from the origin remote and merge into the current branch
		err = w.PullContext(ctxNew, &git.PullOptions{
			RemoteName:      "origin",
			ReferenceName:   plumbing.NewBranchReferenceName(c.Branch),
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
				return err
			}
		}

		// get branch name, if branch exists, then create flag is false
		flagCreate := true
		hrf, err := r.Head()
		if err != nil {
			l.Errorf("Head failed: %v", err)
			return err
		}
		if hrf.Name().IsBranch() {
			branchName := string(hrf.Name())
			branchName = strings.TrimPrefix(branchName, prefixGitBranchName)
			if branchName == c.Branch {
				flagCreate = false
			}
		}

		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(c.Branch),
			Force:  true,
			Create: flagCreate,
		})
		if err != nil {
			l.Errorf("Checkout failed: %v", err)
			return err
		}
	} else {
		// clone a new one
		l.Debug("PlainClone start")

		ctxNew, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			cancel()
		}()
		if _, err := git.PlainCloneContext(ctxNew, clonePath, false, &git.CloneOptions{
			// The intended use of a GitHub personal access token is in replace of your password
			// because access tokens can easily be revoked.
			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
			Auth:            authMethod,
			URL:             c.URL,
			InsecureSkipTLS: true,
		}); err != nil {
			l.Errorf("PlainClone failed: %v", err)
			return err
		}
	}

	if isUpdate || isFirst {
		l.Info("reload")

		if isFirst {
			isFirst = false
		}

		ctxNew, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer func() {
			cancel()
		}()
		if err = reloadCore(ctxNew); err != nil {
			return err
		}
	}

	l.Debug("completed")
	return nil
}

func reloadCore(ctx context.Context) error {
	round := 0
	for {
		select {
		case <-ctx.Done():
			l.Error("reload_timeout")
			return fmt.Errorf("reload_timeout")
		default:
			switch round {
			case 0:
				l.Debug("before ReloadCheckInputCfg")

				iputs, err := config.ReloadCheckInputCfg()
				if err != nil {
					l.Errorf("ReloadCheckInputCfg failed: %v", err)
					return err
				}

				l.Debug("before ReloadCheckPipelineCfg")

				opt, err := config.ReloadCheckPipelineCfg(iputs)
				if err != nil {
					if opt != nil {
						l.Errorf("ReloadCheckPipelineCfg failed: %v => Source: %s, Service: %s, Pipeline: %s",
							err, opt.Source, opt.Service, opt.Pipeline)
					} else {
						l.Errorf("ReloadCheckPipelineCfg failed: %v", err)
					}
					return err
				}

			case 1:
				l.Debug("before StopInputs")

				if err := inputs.StopInputs(); err != nil {
					l.Errorf("StopInputs failed: %v", err)
					return err
				}

			case 2:
				l.Debug("before ReloadInputConfig")

				if err := config.ReloadInputConfig(); err != nil {
					l.Errorf("ReloadInputConfig failed: %v", err)
					return err
				}

			case 3:
				l.Debug("before RunInputs")

				if err := inputs.RunInputs(true); err != nil {
					l.Errorf("RunInputs failed: %v", err)
					return err
				}

			case 4:
				l.Debug("before ReloadTheNormalServer")

				httpd.ReloadTheNormalServer()
			}
		}

		round++
		if round > 4 {
			return nil
		}
	}
}
