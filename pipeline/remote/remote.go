// Package remote contains pipeline remote pulling source code
package remote

import (
	"context"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

const (
	pipelineRemoteName       = "PipelineRemote"
	pipelineRemoteConfigFile = ".config"
	pipelineWarning          = `
	#------------------------------------   警告   -------------------------------------
	# 不要修改或删除本文件
	#-----------------------------------------------------------------------------------
	`
)

var (
	l                 = logger.DefaultSLogger(pipelineRemoteName)
	runPipelineRemote sync.Once
)

type pipelineRemoteConfig struct {
	SiteURL    string
	UpdateTime int64
}

func StartPipelineRemote(urls []string) {
	runPipelineRemote.Do(func() {
		l = logger.SLogger(pipelineRemoteName)
		g := datakit.G(pipelineRemoteName)

		pls, err := config.GetNamespacePipelineFiles(datakit.StrPipelineRemote)
		if err != nil {
			l.Errorf("GetNamespacePipelineFiles failed: %v", err)
		} else {
			worker.ReloadAllRemoteDotPScript2Store(pls)
		}

		g.Go(func(ctx context.Context) error {
			return pullMain(urls, &pipelineRemoteImpl{})
		})
	})
}

//------------------------------------------------------------------------------

type IPipelineRemote interface {
	FileExist(filename string) bool
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm fs.FileMode) error
	ReadDir(dirname string) ([]fs.FileInfo, error)
	PullPipeline(ts int64) (mFiles map[string]string, updateTime int64, err error)
	GetTickerDurationAndBreak() (time.Duration, bool)
}

type pipelineRemoteImpl struct{}

func (*pipelineRemoteImpl) FileExist(filename string) bool {
	return datakit.FileExist(filename)
}

func (*pipelineRemoteImpl) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (*pipelineRemoteImpl) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (*pipelineRemoteImpl) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename) //nolint:gosec
}

func (*pipelineRemoteImpl) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (*pipelineRemoteImpl) ReadDir(dirname string) ([]fs.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (*pipelineRemoteImpl) PullPipeline(ts int64) (mFiles map[string]string, updateTime int64, err error) {
	return io.PullPipeline(ts)
}

func (*pipelineRemoteImpl) GetTickerDurationAndBreak() (time.Duration, bool) {
	dr, err := time.ParseDuration(config.Cfg.Pipeline.RemotePullInterval)
	if err != nil {
		l.Errorf("time.ParseDuration failed: err = (%v), str = (%s)", err, config.Cfg.Pipeline.RemotePullInterval)
		dr = time.Minute
	}
	return dr, false
}

//------------------------------------------------------------------------------

func pullMain(urls []string, ipr IPipelineRemote) error {
	l.Info("start")

	if len(urls) == 0 {
		return nil
	}

	pathConfig := filepath.Join(datakit.PipelineRemoteDir, pipelineRemoteConfigFile)

	td, isReturn := ipr.GetTickerDurationAndBreak()
	l.Infof("duration: %s", td.String())

	tick := time.NewTicker(td)
	defer tick.Stop()

	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return nil

		case <-tick.C:
			l.Debug("triggered")

			err = doPull(pathConfig, urls[0], ipr)
			if err != nil {
				io.FeedLastError(datakit.DatakitInputName, err.Error())
			}

			if isReturn {
				return nil
			}
		} // select
	} // for
}

func doPull(pathConfig, siteURL string, ipr IPipelineRemote) error {
	localTS, err := getPipelineRemoteConfig(pathConfig, siteURL, ipr)
	if err != nil {
		l.Errorf("getPipelineRemoteConfig failed: %s", err.Error())
		return err
	}

	mFiles, updateTime, err := ipr.PullPipeline(localTS)
	if err != nil {
		l.Errorf("PullPipeline failed: %s", err.Error())
		return err
	}

	if localTS == updateTime || updateTime <= 0 {
		l.Debugf("already up to date: %d", updateTime)
	} else {
		l.Infof("localTS = %d, updateTime = %d, so update", localTS, updateTime)

		files, err := dumpFiles(mFiles, ipr)
		if err != nil {
			l.Errorf("dumpFiles failed: %s", err.Error())
			return err
		}

		worker.ReloadAllRemoteDotPScript2Store(files)

		err = updatePipelineRemoteConfig(pathConfig, siteURL, updateTime, ipr)
		if err != nil {
			l.Errorf("updatePipelineRemoteConfig failed: %s", err.Error())
			return err
		}

		l.Debug("update completed.")
	}

	return nil
}

func dumpFiles(mFiles map[string]string, ipr IPipelineRemote) ([]string, error) {
	l.Debugf("dumpFiles: %#v", mFiles)

	// remove lcoal files
	files, err := ipr.ReadDir(datakit.PipelineRemoteDir)
	if err != nil {
		return nil, err
	}
	for _, fi := range files {
		if !fi.IsDir() {
			localName := strings.ToLower(fi.Name())
			if localName != pipelineRemoteConfigFile {
				fullPath := filepath.Join(datakit.PipelineRemoteDir, localName)
				if err = os.Remove(fullPath); err != nil {
					l.Errorf("failed to remove pipeline remote %s, err: %s", fi.Name(), err.Error())
					continue
				}
			}
		}
	}

	// dump
	var plPaths []string
	for k, v := range mFiles {
		fullPath := filepath.Join(datakit.PipelineRemoteDir, k)
		if err = ipr.WriteFile(fullPath, []byte(pipelineWarning+v), datakit.ConfPerm); err != nil {
			l.Errorf("failed to write pipeline remote %s, err: %s", k, err.Error())
			continue
		}
		plPaths = append(plPaths, fullPath)
	}
	l.Debugf("plPaths: %#v", plPaths)
	return plPaths, nil
}

func getPipelineRemoteConfig(pathConfig, siteURL string, ipr IPipelineRemote) (int64, error) {
	if !ipr.FileExist(pathConfig) {
		return 0, nil //nolint:nilerr
	}
	bys, err := ipr.ReadFile(pathConfig) //nolint:gosec
	if err != nil {
		return 0, err
	}
	var cf pipelineRemoteConfig
	if err := ipr.Unmarshal(bys, &cf); err != nil {
		return 0, err
	}
	if cf.SiteURL != siteURL {
		return 0, nil // need update when token has changed
	}
	return cf.UpdateTime, nil
}

func updatePipelineRemoteConfig(pathConfig, siteURL string, latestTime int64, ipr IPipelineRemote) error {
	cf := pipelineRemoteConfig{
		SiteURL:    siteURL,
		UpdateTime: latestTime,
	}
	bys, err := ipr.Marshal(cf)
	if err != nil {
		return err
	}
	if err := ipr.WriteFile(pathConfig, bys, os.ModePerm); err != nil {
		return err
	}
	return nil
}
