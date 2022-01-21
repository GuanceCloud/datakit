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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
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
				if err = doRun(v); err != nil {
					tip := fmt.Sprintf("[gitrepo] failed: %v", err)
					l.Error(tip)
					io.SelfError(tip)
				}
			}
		} // select
	} // for
}

func doRun(c *config.GitRepository) error {
	clonePath, err := getGitClonePathFromGitURL(c.URL)
	if err != nil {
		return err
	}

	gitUserName, gitPassword, err := getUserNamePasswordFromGitURL(c.URL)
	if err != nil {
		l.Errorf("getUserNamePasswordFromGitURL failed: %v, url = %s", err, c.URL)
		return err
	}

	if gitUserName == "" {
		// use ssh to auth
		if c.SSHPrivateKeyPath == "" {
			tip := "ssh need key file"
			l.Error(tip)
			return fmt.Errorf(tip)
		}
	}

	authMethod, err := getAuthMethod(gitUserName, gitPassword, c)
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

		if err := gitPlainClone(clonePath, c.URL, authMethod); err != nil {
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
		defer func() {
			cancel()
		}()
		if _, err = reloadCore(ctxNew); err != nil {
			return err
		}
	}

	io.FeedEventLog(&io.Reporter{Message: "Gitrepo synchronizes the latest data", Logtype: "event"})

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
	defer func() {
		cancel()
	}()
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

func gitPlainClone(clonePath, gitURL string, authMethod transport.AuthMethod) error {
	ctxNew, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
	}()
	if _, err := git.PlainCloneContext(ctxNew, clonePath, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:            authMethod,
		URL:             gitURL,
		InsecureSkipTLS: true,
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

				iputs, err := config.ReloadCheckInputCfg()
				if err != nil {
					l.Errorf("ReloadCheckInputCfg failed: %v", err)
					return round, err
				}

				l.Info("before ReloadCheckPipelineCfg")

				opt, err := config.ReloadCheckPipelineCfg(iputs)
				if err != nil {
					if opt != nil {
						l.Errorf("ReloadCheckPipelineCfg failed: %v => Source: %s, Service: %s, Pipeline: %s",
							err, opt.Source, opt.Service, opt.Pipeline)
					} else {
						l.Errorf("ReloadCheckPipelineCfg failed: %v", err)
					}
					return round, err
				}

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

				allGitReposPipelines := config.GetGitReposAllPipelinePath()

				worker.LoadAllDotPScriptForWkr([]string{}, allGitReposPipelines)

			case 4:
				l.Info("before RunInputs")

				if err := inputs.RunInputs(true); err != nil {
					l.Errorf("RunInputs failed: %v", err)
					return round, err
				}

			case 5:
				l.Info("before ReloadTheNormalServer")

				httpd.ReloadTheNormalServer()
			}
		}

		round++
		if round > 5 {
			return round, nil // round + 1
		}
	}
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

func getUserNamePasswordFromGitURL(gitURL string) (gitUserName, gitPassword string, err error) {
	// http only could use username & password auth
	// ssh only could use private key auth
	uGitURL, err := giturls.Parse(gitURL)
	if err != nil {
		l.Errorf("url.Parse failed: %v, url = %s", err, gitURL)
		return
	}

	bIsUseAuthUserNamePassword, err := isUserNamePasswordAuth(gitURL)
	if err != nil {
		l.Errorf("isUserNamePasswordAuth failed: %v, url = %s", err, gitURL)
		return
	}

	if bIsUseAuthUserNamePassword {
		gitUserName = uGitURL.User.Username()
		if password, ok := uGitURL.User.Password(); !ok {
			tip := "invalid git password"
			l.Error(tip)
			err = fmt.Errorf(tip)
			return
		} else {
			gitPassword = password
		}

		if gitUserName == "" || gitPassword == "" {
			tip := "http need username password"
			l.Error(tip)
			err = fmt.Errorf(tip)
			return
		}
	}
	return gitUserName, gitPassword, nil
}

func getAuthMethod(gitUserName, gitPassword string, c *config.GitRepository) (transport.AuthMethod, error) {
	var authMethod transport.AuthMethod
	if gitUserName != "" {
		// use username & password to auth
		authMethod = &http.BasicAuth{
			Username: gitUserName,
			Password: gitPassword,
		}
	} else {
		// use ssh to auth
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
	}
	return authMethod, nil
}
