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
	"path/filepath"
	"regexp"
	"strings"

	// it will use this embedded information in time/tzdata.
	_ "time/tzdata"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb/iploc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/scriptstore"
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
		"iploc": &iploc.IPloc{},
	}
)

func GetIPdb() ipdb.IPdb {
	return ipdbInstance
}

type PipelineCfg struct {
	IPdbAttr           map[string]string `toml:"ipdb_attr"`
	IPdbType           string            `toml:"ipdb_type"`
	RemotePullInterval string            `toml:"remote_pull_interval"`
}

func NewPipeline(srciptname string) (*Pipeline, error) {
	script, err := scriptstore.QueryScript(srciptname, nil)
	if err != nil {
		return nil, err
	}
	return &Pipeline{
		scriptInfo: script,
	}, nil
}

func NewPipelineFromFile(path string) (*Pipeline, error) {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	sc, err := scriptstore.NewScriptInfo("", string(data), "")
	if err != nil {
		return nil, err
	}
	return &Pipeline{
		DisableUpdate: true,
		scriptInfo:    sc,
	}, nil
}

type Pipeline struct {
	DisableUpdate bool
	scriptInfo    *scriptstore.ScriptInfo
}

func (p *Pipeline) Run(data string, source string) (*Result, error) {
	if p.scriptInfo.Engine() == nil {
		return nil, fmt.Errorf("pipeline engine not initialized")
	}

	if result, err := RunPlStr(data, source, 0, p.scriptInfo.Engine()); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (p *Pipeline) RunByte(data []byte, encode string, source string) (*Result, error) {
	if p.scriptInfo.Engine() == nil {
		return nil, fmt.Errorf("pipeline engine not initialized")
	}

	if result, err := RunPlByte(data, encode, source, 0, p.scriptInfo.Engine()); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (p *Pipeline) UpdateScriptInfo() error {
	var err error
	if p.DisableUpdate {
		return fmt.Errorf("current pipeline update disabled")
	}
	if p.scriptInfo, err = scriptstore.QueryScript(p.scriptInfo.Name(), p.scriptInfo); err != nil {
		return err
	}
	return nil
}

func Init(pipelineCfg *PipelineCfg) error {
	l = logger.SLogger("pipeline")
	funcs.InitLog()
	parser.InitLog()

	if _, err := InitIPdb(pipelineCfg); err != nil {
		l.Warnf("init ipdb error: %s", err.Error())
	}

	if err := loadPatterns(); err != nil {
		return err
	}

	scriptstore.InitStore()

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
		ip2isp.InitIPdb(ipdbInstance)
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

	parser.DenormalizedGlobalPatterns = denormalizedGlobalPatterns

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

func RunPlStr(cntStr, source string, maxMessageLen int, ng *parser.Engine) (*Result, error) {
	result := &Result{
		Output: nil,
	}
	if ng != nil {
		var err error
		if result.Output, err = ng.Run(cntStr); err != nil {
			l.Debug(err)
			result.Err = err.Error()
		}
	} else {
		result.Output = &parser.Output{
			Cost: map[string]string{},
			Tags: map[string]string{},
			Fields: map[string]interface{}{
				PipelineMessageField: cntStr,
			},
		}
	}
	result.preprocessing(source, maxMessageLen)
	return result, nil
}

func RunPlByte(cntByte []byte, encode string, source string, maxMessageLen int, ng *parser.Engine) (*Result, error) {
	cntStr, err := DecodeContent(cntByte, encode)
	if err != nil {
		return nil, err
	}
	return RunPlStr(cntStr, source, maxMessageLen, ng)
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
