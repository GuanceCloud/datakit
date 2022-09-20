// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pipeline implement datakit's logging pipeline.
package pipeline

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	// it will use this embedded information in time/tzdata.
	_ "time/tzdata"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/engine/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb/geoip"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb/iploc"
	plrefertable "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/refertable"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var ipdbInstance ipdb.IPdb // get ip location and isp

var (
	l                               = logger.DefaultSLogger("pipeline")
	pipelineDefaultCfg *PipelineCfg = &PipelineCfg{
		IPdbType: "iploc",
		IPdbAttr: map[string]string{
			"iploc_file": "iploc.bin",
			"isp_file":   "ip2isp.txt",
		},
	}
	pipelineIPDbmap = map[string]ipdb.IPdb{
		"iploc":    &iploc.IPloc{},
		"geolite2": &geoip.Geoip{},
	}
)

func GetIPdb() ipdb.IPdb {
	return ipdbInstance
}

type PipelineCfg struct {
	IPdbAttr               map[string]string `toml:"ipdb_attr"`
	IPdbType               string            `toml:"ipdb_type"`
	RemotePullInterval     string            `toml:"remote_pull_interval"`
	ReferTableURL          string            `toml:"refer_table_url"`
	ReferTablePullInterval string            `toml:"refer_table_pull_interval"`
}

func NewPipelineFromFile(category string, path string) (*Pipeline, error) {
	name, script, err := plscript.ReadPlScriptFromFile(path)
	if err != nil {
		return nil, err
	}

	return NewPipeline(category, name, script)
}

func NewPipeline(category string, name, script string) (*Pipeline, error) {
	scs, errs := plscript.NewScripts(map[string]string{name: script}, map[string]string{}, "", category)

	if v, ok := errs[name]; ok {
		return nil, v
	}

	if sc, ok := scs[name]; ok {
		return &Pipeline{
			Script: sc,
		}, nil
	}

	return nil, fmt.Errorf("unknown error")
}

func NewPipelineMulti(category string, scripts map[string]string, scriptPath map[string]string) (map[string]*Pipeline, map[string]error) {
	ret, retErr := plscript.NewScripts(scripts, scriptPath, "", category)

	retPl := map[string]*Pipeline{}
	for k, v := range ret {
		retPl[k] = &Pipeline{
			Script: v,
		}
	}

	return retPl, retErr
}

type Pipeline struct {
	Script *plscript.PlScript
}

func (p *Pipeline) Run(pt *point.Point, plOpt *plscript.Option, ioPtOpt *point.PointOption) (*point.Point, bool, error) {
	if p.Script == nil || p.Script.Engine() == nil {
		return nil, false, fmt.Errorf("pipeline engine not initialized")
	}
	if pt == nil {
		return nil, false, fmt.Errorf("no data")
	}
	fields, err := pt.Fields()
	if err != nil {
		return nil, false, err
	}

	if measurement, tags, fileds, tn, drop, err := p.Script.Run(pt.Name(), pt.Tags(), fields, ioPtOpt.Time, plOpt); err != nil {
		return nil, drop, err
	} else {
		if !tn.IsZero() {
			ioPtOpt.Time = tn
		}
		if pt, err := point.NewPoint(measurement, tags, fileds, ioPtOpt); err != nil {
			// stats.WriteScriptStats(p.script.Category(), p.script.NS(), p.script.Name(), 0, 0, 1, err)
			return nil, false, err
		} else {
			return pt, drop, nil
		}
	}
}

func Init(pipelineCfg *PipelineCfg) error {
	l = logger.SLogger("pipeline")
	plscript.InitStore()
	funcs.InitLog()
	parser.InitLog()
	plrefertable.InitLog()

	if _, err := InitIPdb(pipelineCfg); err != nil {
		l.Warnf("init ipdb error: %s", err.Error())
	}

	if pipelineCfg.ReferTableURL != "" {
		dur, err := time.ParseDuration(pipelineCfg.ReferTablePullInterval)
		if err != nil {
			l.Warnf("refer table pull interval %s, err: %v", dur, err)
			dur = time.Minute * 5
		}
		if err := plrefertable.InitReferTableRunner(
			pipelineCfg.ReferTableURL, dur); err != nil {
			l.Error("init refer table, error: %v", err)
		}
	}

	if err := loadPatterns(); err != nil {
		return err
	}

	return nil
}

// InitIPdb init ipdb instance.
func InitIPdb(pipelineCfg *PipelineCfg) (ipdb.IPdb, error) {
	if pipelineCfg == nil {
		pipelineCfg = pipelineDefaultCfg
	}
	if instance, ok := pipelineIPDbmap[pipelineCfg.IPdbType]; ok {
		ipdbInstance = instance
		ipdbInstance.Init(datakit.DataDir, pipelineCfg.IPdbAttr)
		funcs.InitIPdb(ipdbInstance)
		if pipelineCfg.IPdbType != "geolite2" {
			ip2isp.InitIPdb(ipdbInstance)
		}
	} else { // invalid ipdb type, then use the default iploc to ignore the error.
		l.Warnf("invalid ipdb_type %s", pipelineCfg.IPdbType)
		return pipelineIPDbmap["iploc"], nil
	}

	return ipdbInstance, nil
}

func loadPatterns() error {
	loadedPatterns, err := grok.LoadPatternsFromPath(datakit.PipelinePatternDir)
	if err != nil {
		return err
	}

	for k, v := range grok.CopyDefalutPatterns() {
		loadedPatterns[k] = v
	}

	denormalizedGlobalPatterns, invalidPatterns := grok.DenormalizePatternsFromMap(loadedPatterns)

	for k, v := range denormalizedGlobalPatterns {
		if _, err := regexp.Compile(v.Denormalized()); err != nil {
			invalidPatterns[k] = err.Error()
		}
	}

	if len(invalidPatterns) != 0 {
		for k, v := range invalidPatterns {
			delete(denormalizedGlobalPatterns, k)
			l.Errorf("load pattern '%s', err: '%s'", k, v)
		}
	}

	runtime.DenormalizedGlobalPatterns = denormalizedGlobalPatterns

	return nil
}

// GbToUtf8 Gb to UTF-8.
// http/api_pipeline.go.
func GbToUtf8(s []byte, encoding string) ([]byte, error) {
	var t transform.Transformer
	switch encoding {
	case "gbk":
		t = simplifiedchinese.GBK.NewDecoder()
	case "gb18030":
		t = simplifiedchinese.GB18030.NewDecoder()
	}
	reader := transform.NewReader(bytes.NewReader(s), t)
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func DecodeContent(content []byte, encode string) (string, error) {
	var err error
	if encode != "" {
		encode = strings.ToLower(encode)
	}
	switch encode {
	case "gbk", "gb18030":
		content, err = GbToUtf8(content, encode)
		if err != nil {
			return "", err
		}
	case "utf8", "utf-8":
	default:
	}
	return string(content), nil
}
