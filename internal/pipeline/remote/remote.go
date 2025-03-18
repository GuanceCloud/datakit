// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package remote contains pipeline remote pulling source code
package remote

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/targzutil"
)

const (
	pipelineRemoteName              = "PipelineRemote"
	pipelineRemoteConfigFile        = "pull_config.json"
	pipelineRemoteContentFile       = "scripts.tar.gz"
	pipelineRemoteRelationDumpFile  = "relation.json"
	pipelineRemoteDefaultScriptFile = "default.json"
	deleteAll                       = 1
)

var (
	l                 = logger.DefaultSLogger(pipelineRemoteName)
	runPipelineRemote sync.Once
	isFirst           = true

	pathContent string
)

type pipelineRemoteConfig struct {
	SiteURL    string
	UpdateTime int64
}

func StartPipelineRemote(urls []string) {
	runPipelineRemote.Do(func() {
		l = logger.SLogger(pipelineRemoteName)
		if runner, err := pullMain(urls, &pipelineRemoteImpl{}); err != nil {
			l.Error(err)
			return
		} else if runner != nil {
			g := datakit.G("pipeline_remote")
			g.Go(func(ctx context.Context) error {
				return runner()
			})
		}
	})
}

func GetConentFileName() string {
	return pipelineRemoteContentFile
}

//------------------------------------------------------------------------------

type IPipelineRemote interface {
	FileExist(filename string) bool
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm fs.FileMode) error
	ReadDir(dirname string) ([]fs.FileInfo, error)
	PullPipeline(int64, int64) (mFiles map[point.Category]map[string]string, plRelation map[point.Category]map[string]string,
		defaultPl map[point.Category]string, updateTime int64, relationTS int64, err error)
	GetTickerDurationAndBreak() (time.Duration, bool)
	Remove(name string) error
	FeedLastError(inputName string, err string)
	ReadTarToMap(srcFile string) (map[string]string, error)
	WriteTarFromMap(data map[string]string, dest string) error
}

// Make sure pipelineRemoteImpl implements the IPipelineRemote interface.
var _ IPipelineRemote = (*pipelineRemoteImpl)(nil)

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
	return os.ReadFile(filename) //nolint:gosec
}

func (*pipelineRemoteImpl) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (*pipelineRemoteImpl) ReadDir(dirname string) ([]fs.FileInfo, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	infos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (*pipelineRemoteImpl) PullPipeline(ts, relationTS int64) (
	map[point.Category]map[string]string, map[point.Category]map[string]string,
	map[point.Category]string, int64, int64, error,
) {
	return io.PullPipeline(ts, relationTS)
}

func (*pipelineRemoteImpl) GetTickerDurationAndBreak() (time.Duration, bool) {
	dr, err := time.ParseDuration(config.Cfg.Pipeline.RemotePullInterval)
	if err != nil {
		l.Warnf("time.ParseDuration failed: err = (%v), str = (%s)", err, config.Cfg.Pipeline.RemotePullInterval)
		dr = time.Minute
	}
	return dr, false
}

func (*pipelineRemoteImpl) Remove(name string) error {
	return os.Remove(name)
}

func (*pipelineRemoteImpl) FeedLastError(inputName string, err string) {
	metrics.FeedLastError(inputName, err)
}

func (*pipelineRemoteImpl) ReadTarToMap(srcFile string) (map[string]string, error) {
	return targzutil.ReadTarToMap(srcFile)
}

func (*pipelineRemoteImpl) WriteTarFromMap(data map[string]string, dest string) error {
	return targzutil.WriteTarFromMap(data, dest)
}

//------------------------------------------------------------------------------

func pullMain(urls []string, ipr IPipelineRemote) (func() error, error) {
	l.Info("start")

	if len(urls) == 0 {
		return nil, nil
	}

	pathConfig := filepath.Join(datakit.PipelineRemoteDir, pipelineRemoteConfigFile)
	pathRelation := filepath.Join(datakit.PipelineRemoteDir, pipelineRemoteRelationDumpFile)
	pathContent = filepath.Join(datakit.PipelineRemoteDir, pipelineRemoteContentFile)

	td, isReturn := ipr.GetTickerDurationAndBreak()
	l.Infof("duration: %s", td.String())

	// first run
	var err error
	err = doPull(pathConfig, pathRelation, urls[0], ipr)
	if err != nil {
		l.Warnf("doPull: %s, ignored", err.Error())
	}
	if isReturn {
		return nil, nil
	}

	return func() error {
		tick := time.NewTicker(td)
		defer tick.Stop()

		for {
			err = doPull(pathConfig, pathRelation, urls[0], ipr)
			if err != nil {
				l.Warnf("doPull: %s, ignored", err.Error())
			}

			if isReturn {
				return nil
			}

			select {
			case <-datakit.Exit.Wait():
				l.Info("exit")
				return nil

			case <-tick.C:
				l.Debug("triggered")
			}
		}
	}, nil
}

func doPull(pathConfig, pathRelation, siteURL string, ipr IPipelineRemote) error {
	localTS, err := getPipelineRemoteConfig(pathConfig, siteURL, ipr)
	if err != nil {
		l.Errorf("getPipelineRemoteConfig failed: %s", err.Error())
		return err
	}

	managerWkr, ok := plval.GetManager()
	if !ok || managerWkr == nil {
		return nil
	}

	spRelation := managerWkr.GetScriptRelation()
	if spRelation == nil {
		return nil
	}

	relationTS := spRelation.UpdateAt()

	mFiles, pRelation, defaultPl, updateTime, relationUpdateTime, err := ipr.PullPipeline(localTS, relationTS)
	if err != nil {
		l.Errorf("PullPipeline failed: %s", err.Error())
		return err
	}

	if localTS == updateTime || updateTime <= 0 {
		l.Debugf("pipeline already up to date: %d", updateTime)
	} else {
		if updateTime == deleteAll {
			l.Debug("deleteAll")

			// cleanup default pipeline
			managerWkr.UpdateDefaultScript(nil)
			if m, ok := plval.GetManager(); ok && m != nil {
				// cleanup all remote scripts
				m.LoadScripts(constants.NSRemote, nil, nil)
			}

			// remove lcoal files
			if err := removeLocalRemote(ipr); err != nil {
				return err
			}
		} else {
			l.Infof("localTS = %d, updateTime = %d, so update", localTS, updateTime)

			err := dumpFiles(mFiles, defaultPl, ipr)
			if err != nil {
				l.Errorf("dumpFiles failed: %s", err.Error())
				return err
			}

			l.Debug("dumpFiles succeeded")

			loadContentPipeline(mFiles)
			combineLocal := plval.PreferLocalDefaultPipeline(defaultPl)
			managerWkr.UpdateDefaultScript(combineLocal)

			err = updatePipelineRemoteConfig(pathConfig, siteURL, updateTime, ipr)
			if err != nil {
				l.Errorf("updatePipelineRemoteConfig failed: %s", err.Error())
				return err
			}

			l.Debugf("update completed: %d", updateTime)
		}
	}

	if relationUpdateTime != -1 {
		// 通常无正常状态的 pl 关系时，update time 不会更新，
		// 中心不会返回最近（禁用/删除的）的关系的 ts 值，此时返回 ts 默认值 0
		// 这种情况会将存储的 relation_update_at 置为 0

		l.Info("update remote pipeline relation map")
		spRelation.UpdateRelation(relationUpdateTime, pRelation)
		if err := dumpRelation(pathRelation, pRelation); err != nil {
			l.Debug(err)
		}
	}

	return nil
}

func removeLocalRemote(ipr IPipelineRemote) error {
	files, err := ipr.ReadDir(datakit.PipelineRemoteDir)
	if err != nil {
		return err
	}
	for _, fi := range files {
		if !fi.IsDir() {
			localName := strings.ToLower(fi.Name())
			if localName != pipelineRemoteConfigFile && localName != pipelineRemoteRelationDumpFile {
				fullPath := filepath.Join(datakit.PipelineRemoteDir, localName)
				if err = ipr.Remove(fullPath); err != nil {
					l.Errorf("failed to remove pipeline remote %s, err: %s", fi.Name(), err.Error())
					continue
				}
			}
		}
	}
	return nil
}

func dumpFiles(mFiles map[point.Category]map[string]string,
	defaultPl map[point.Category]string, ipr IPipelineRemote,
) error {
	l.Debugf("dumpFiles: %#v", mFiles)
	// remove lcoal files
	if err := removeLocalRemote(ipr); err != nil {
		return err
	}
	// dump
	data := convertThreeMapToContentMap(mFiles, defaultPl)
	if err := ipr.WriteTarFromMap(data, pathContent); err != nil {
		return err
	}
	return nil
}

type relationInfo struct {
	Relation map[string]map[string]string `json:"relation"`
}

func dumpRelation(path string, relation map[point.Category]map[string]string) error {
	rl := map[string]map[string]string{}
	for c, r := range relation {
		rl[c.String()] = r
	}

	if body, err := json.Marshal(&relationInfo{
		Relation: rl,
	}); err != nil {
		return err
	} else {
		if err := os.WriteFile(path, body, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
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
		if err := ipr.Remove(pathConfig); err != nil {
			l.Warnf("diff token, remove config failed: %v", err)
		}
		if err := removeLocalRemote(ipr); err != nil {
			l.Warnf("diff token, removeLocalRemote failed: %v", err)
		}
		return 0, nil // need update when token has changed
	} else if isFirst {
		// init load pipeline remotes
		isFirst = false
		cf.UpdateTime = 0 // 永远不从磁盘文件初始化
	} // isFirst

	return cf.UpdateTime, nil
}

// ConvertContentMapToThreeMap more info see test case.
func ConvertContentMapToThreeMap(in map[string]string) map[string]map[string]string {
	out := make(map[string]map[string]string)
	for categoryAndName, content := range in {
		categoryGet, name := filepath.Split(categoryAndName)
		category := filepath.Clean(categoryGet)
		if len(category) > 0 && len(name) > 0 {
			// normal
			if mVal, ok := out[category]; ok {
				mVal[name] = content
				out[category] = mVal
			} else {
				mPiece := make(map[string]string)
				mPiece[name] = content
				out[category] = mPiece
			}
		}
	}
	return out
}

// more info see test case.
func convertThreeMapToContentMap(in map[point.Category]map[string]string,
	defaultPl map[point.Category]string,
) map[string]string {
	out := make(map[string]string)
	for category, mVal := range in {
		for name, content := range mVal {
			out[filepath.Join(category.String(), name)] = content
		}
	}

	if defaultPl != nil {
		df := map[string]string{}
		for cat, name := range defaultPl {
			df[cat.String()] = name
		}
		if v, err := json.Marshal(df); err != nil {
			l.Error(err)
		} else {
			out[pipelineRemoteDefaultScriptFile] = string(v)
		}
	}

	return out
}

func loadContentPipeline(in map[point.Category]map[string]string) {
	managerWkr, ok := plval.GetManager()
	if !ok || managerWkr == nil {
		return
	}

	inS := map[point.Category]map[string]string{}

	for cat, val := range in {
		inS[cat] = val
	}
	managerWkr.LoadScripts(constants.NSRemote, inS, nil)
}
