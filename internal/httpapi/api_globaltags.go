// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var muDatakitConf sync.Mutex

type handler interface {
	dumpMainCfgTOML(c *config.Config)
	getDuplicateCfg() (*config.Config, bool)
}

func getHostTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"host-tags": fmt.Sprint(datakit.GlobalHostTags()),
	})
}

func postHostTags(c *gin.Context) {
	q := getQueryMap(c.Request.URL.Query())

	if !datakit.Docker {
		handleUpdateDatakitConf(q, true, newHandle())
	}

	c.JSON(http.StatusOK, UpdateHostTags(q, "/v1/global/host/tags"))
}

func deleteHostTags(c *gin.Context) {
	q := getQueryArray(c.Request.URL.Query())

	if !datakit.Docker {
		handleDeleteDatakitConf(q, true, newHandle())
	}

	if tags, ok := deleteMerge(q, datakit.GlobalHostTags()); ok {
		datakit.ClearGlobalHostTags()
		datakit.SetGlobalHostTagsByMap(tags)
		if !config.Cfg.Election.Enable {
			datakit.ClearGlobalElectionTags()
			datakit.SetGlobalElectionTagsByMap(tags)
		}
		apiGlobalTagsUpdatedVec.WithLabelValues(c.Request.URL.Path, c.Request.Method, "OK").Set(float64(time.Now().Unix()))
	}

	if tags, ok := deleteMerge(q, config.Cfg.Dataway.GlobalTags()); ok {
		config.Cfg.Dataway.UpdateGlobalTags(tags)
	}

	c.JSON(http.StatusOK, gin.H{
		"dataway-tags":  config.Cfg.Dataway.GlobalTags(),
		"election-tags": datakit.GlobalElectionTags(),
		"host-tags":     datakit.GlobalHostTags(),
	})
}

func getElectionTags(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"election-tags": fmt.Sprint(datakit.GlobalElectionTags()),
	})
}

func postElectionTags(c *gin.Context) {
	if !config.Cfg.Election.Enable {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Can't use this command when global-election is false.",
		})
		return
	}

	q := getQueryMap(c.Request.URL.Query())

	if !datakit.Docker {
		handleUpdateDatakitConf(q, false, newHandle())
	}

	c.JSON(http.StatusOK, UpdateElectionTags(q, "/v1/global/election/tags"))
}

func deleteElectionTags(c *gin.Context) {
	if !config.Cfg.Election.Enable {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Can't use this command when global-election is false.",
		})
		return
	}

	q := getQueryArray(c.Request.URL.Query())

	if !datakit.Docker {
		handleDeleteDatakitConf(q, false, newHandle())
	}

	if tags, ok := deleteMerge(q, datakit.GlobalElectionTags()); ok {
		datakit.ClearGlobalElectionTags()
		datakit.SetGlobalElectionTagsByMap(tags)
		apiGlobalTagsUpdatedVec.WithLabelValues(c.Request.URL.Path, c.Request.Method, "OK").Set(float64(time.Now().Unix()))
	}

	if tags, ok := deleteMerge(q, config.Cfg.Dataway.GlobalTags()); ok {
		config.Cfg.Dataway.UpdateGlobalTags(tags)
	}

	c.JSON(http.StatusOK, gin.H{
		"dataway-tags":  config.Cfg.Dataway.GlobalTags(),
		"election-tags": datakit.GlobalElectionTags(),
		"host-tags":     datakit.GlobalHostTags(),
	})
}

func updateMerge(q map[string]string, tags map[string]string) (map[string]string, bool) {
	m := internal.CopyMapString(tags)
	done := false
	for k, v := range q {
		if m[k] != v {
			m[k] = v
			done = true
		}
	}
	return m, done
}

func deleteMerge(q []string, tags map[string]string) (map[string]string, bool) {
	m := internal.CopyMapString(tags)
	done := false
	for _, k := range q {
		if _, ok := m[k]; ok {
			delete(m, k)
			done = true
		}
	}
	return m, done
}

func getQueryMap(uv url.Values) map[string]string {
	q := make(map[string]string)
	for k, v := range uv {
		if len(v) == 0 || k == "host" {
			continue
		}
		q[k] = v[0]
	}
	return q
}

func getQueryArray(uv url.Values) []string {
	q := make([]string, 0)
	for k, v := range uv {
		if k != "tags" {
			continue
		}
		if len(v) == 0 {
			continue
		}
		for _, v := range strings.Split(v[0], ",") {
			if v != "host" {
				q = append(q, v)
			}
		}
	}
	return q
}

func handleUpdateDatakitConf(q map[string]string, isHost bool, h handler) {
	muDatakitConf.Lock()
	defer muDatakitConf.Unlock()

	duplicateCfg, ok := h.getDuplicateCfg()
	if !ok {
		return
	}

	if isHost {
		if tags, ok := updateMerge(q, duplicateCfg.GlobalHostTags); ok {
			duplicateCfg.GlobalHostTags = tags
			h.dumpMainCfgTOML(duplicateCfg)
		}
	} else {
		if tags, ok := updateMerge(q, duplicateCfg.Election.Tags); ok {
			duplicateCfg.Election.Tags = tags
			h.dumpMainCfgTOML(duplicateCfg)
		}
	}
}

func handleDeleteDatakitConf(q []string, isHost bool, h handler) {
	muDatakitConf.Lock()
	defer muDatakitConf.Unlock()

	duplicateCfg, ok := h.getDuplicateCfg()
	if !ok {
		return
	}

	if isHost {
		if tags, ok := deleteMerge(q, duplicateCfg.GlobalHostTags); ok {
			duplicateCfg.GlobalHostTags = tags
			h.dumpMainCfgTOML(duplicateCfg)
		}
	} else {
		if tags, ok := deleteMerge(q, duplicateCfg.Election.Tags); ok {
			duplicateCfg.Election.Tags = tags
			h.dumpMainCfgTOML(duplicateCfg)
		}
	}
}

var newHandle = func() handler {
	return &handle{}
}

type handle struct{}

func (*handle) getDuplicateCfg() (*config.Config, bool) {
	c := config.DefaultConfig()

	cfgdata, err := os.ReadFile(filepath.Clean(datakit.MainConfPath))
	if err != nil {
		l.Errorf("os.ReadFile: %w", err)
		return nil, false
	}

	_, err = bstoml.Decode(string(cfgdata), c)
	if err != nil {
		l.Errorf("bstoml.Decode: %w", err)
		return nil, false
	}

	return c, true
}

func (*handle) dumpMainCfgTOML(c *config.Config) {
	fileName := datakit.MainConfPath
	file, err := os.OpenFile(filepath.Clean(fileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		l.Errorf("Dump main Cfg, create file: %s, err: %s", fileName, err.Error())
		return
	}
	encode := bstoml.NewEncoder(file)
	if err := encode.Encode(c); err != nil {
		l.Errorf("Dump main Cfg, encode: %s, err: %s", fileName, err.Error())
	}
	file.Close() //nolint:errcheck,gosec
}

func UpdateHostTags(q map[string]string, source string) map[string](map[string]string) {
	if tags, ok := updateMerge(q, datakit.GlobalHostTags()); ok {
		datakit.ClearGlobalHostTags()
		datakit.SetGlobalHostTagsByMap(tags)

		method := "unknow"
		if source == "/v1/global/host/tags" {
			method = "POST"
		}
		apiGlobalTagsUpdatedVec.WithLabelValues(source, method, "OK").Set(float64(time.Now().Unix()))
	}

	if !config.Cfg.Election.Enable && source == "/v1/global/host/tags" {
		if tags, ok := updateMerge(q, datakit.GlobalElectionTags()); ok {
			datakit.ClearGlobalElectionTags()
			datakit.SetGlobalElectionTagsByMap(tags)
		}
	}

	if tags, ok := updateMerge(q, config.Cfg.Dataway.GlobalTags()); ok {
		config.Cfg.Dataway.UpdateGlobalTags(tags)
	}

	return map[string](map[string]string){
		"dataway-tags":  config.Cfg.Dataway.GlobalTags(),
		"election-tags": datakit.GlobalElectionTags(),
		"host-tags":     datakit.GlobalHostTags(),
	}
}

func UpdateElectionTags(q map[string]string, source string) map[string](map[string]string) {
	if tags, ok := updateMerge(q, datakit.GlobalElectionTags()); ok {
		datakit.ClearGlobalElectionTags()
		datakit.SetGlobalElectionTagsByMap(tags)

		method := "unknow"
		if source == "/v1/global/election/tags" {
			method = "POST"
		}
		apiGlobalTagsUpdatedVec.WithLabelValues(source, method, "OK").Set(float64(time.Now().Unix()))
	}

	if tags, ok := updateMerge(q, config.Cfg.Dataway.GlobalTags()); ok {
		config.Cfg.Dataway.UpdateGlobalTags(tags)
	}

	return map[string](map[string]string){
		"dataway-tags":  config.Cfg.Dataway.GlobalTags(),
		"election-tags": datakit.GlobalElectionTags(),
		"host-tags":     datakit.GlobalHostTags(),
	}
}
