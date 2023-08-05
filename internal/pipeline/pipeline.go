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

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/grok"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ipdb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ipdb/geoip"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ipdb/iploc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plmap"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
	plrefertable "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/refertable"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/relation"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/script"
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
	IPdbAttr               map[string]string      `toml:"ipdb_attr"`
	IPdbType               string                 `toml:"ipdb_type"`
	RemotePullInterval     string                 `toml:"remote_pull_interval"`
	ReferTableURL          string                 `toml:"refer_table_url"`
	ReferTablePullInterval string                 `toml:"refer_table_pull_interval"`
	UseSQLite              bool                   `toml:"use_sqlite"`
	SQLiteMemMode          bool                   `toml:"sqlite_mem_mode"`
	Offload                *offload.OffloadConfig `toml:"offload"`
}

func NewPipelineFromFile(category point.Category, path string) (*Pipeline, error) {
	name, script, err := plscript.ReadPlScriptFromFile(path)
	if err != nil {
		return nil, err
	}

	return NewPipeline(category, name, script)
}

func NewPipeline(category point.Category, name, script string) (*Pipeline, error) {
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

func NewPipelineMulti(category point.Category, scripts map[string]string, scriptPath map[string]string) (map[string]*Pipeline, map[string]error) {
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

func (p *Pipeline) Run(cat point.Category, pt *dkpt.Point, plOpt *plscript.Option, ioPtOpt *dkpt.PointOption,
	signal plruntime.Signal, buks ...*plmap.AggBuckets,
) (ptinput.PlInputPt, error) {
	if p.Script == nil || p.Script.Engine() == nil {
		return nil, fmt.Errorf("pipeline engine not initialized")
	}

	if pt == nil {
		return nil, fmt.Errorf("no data")
	}

	plpt, err := ptinput.WrapDeprecatedPoint(cat, pt)
	if err != nil {
		return nil, err
	}

	if len(buks) > 0 {
		p.Script.SetAggBuks(buks[0])
	}

	if err := p.Script.Run(plpt, signal, plOpt); err != nil {
		return nil, err
	} else {
		if !plpt.PtTime().IsZero() {
			ioPtOpt.Time = plpt.PtTime()
		}
		return plpt, nil
	}
}

func Init(pipelineCfg *PipelineCfg) error {
	l = logger.SLogger("pipeline")
	plscript.InitStore()
	funcs.InitLog()
	plrefertable.InitLog()
	relation.InitRelationLog()

	if pipelineCfg == nil {
		pipelineCfg = pipelineDefaultCfg
	}

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
			pipelineCfg.ReferTableURL, dur, pipelineCfg.UseSQLite, pipelineCfg.SQLiteMemMode); err != nil {
			l.Error("init refer table, error: %v", err)
		}
	}

	if pipelineCfg.Offload != nil && pipelineCfg.Offload.Receiver != "" &&
		len(pipelineCfg.Offload.Addresses) != 0 {
		if err := offload.InitOffloaWorker(pipelineCfg.Offload); err != nil {
			l.Errorf("init offload worker, error: %v", err)
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
			ip2isp.InitIPDB(ipdbInstance)
		}
	} else { // invalid ipdb type, then use the default iploc to ignore the error.
		l.Warnf("invalid ipdb_type %s", pipelineCfg.IPdbType)
		return pipelineIPDbmap["iploc"], nil
	}

	return ipdbInstance, nil
}

func loadPatterns() error {
	// 从文件加载 pattern
	loadedPatterns, err := grok.LoadPatternsFromPath(datakit.PipelinePatternDir)
	if err != nil {
		return err
	}

	// 使用内置的 pattern，可能覆盖文件中的 pattern
	for k, v := range CopyDefalutPatterns() {
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

	// 替换 ppl runtime 中的 patterns
	plruntime.DenormalizedGlobalPatterns = denormalizedGlobalPatterns

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
