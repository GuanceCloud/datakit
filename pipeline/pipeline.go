// Package pipeline implement datakit's logging pipeline.
package pipeline

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	// it will use this embedded information in time/tzdata.
	_ "time/tzdata"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
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

type Pipeline struct {
	scriptInfo *scriptstore.ScriptInfo
}

func (p *Pipeline) Run(data string) (*Result, error) {
	if p.scriptInfo.Engine() == nil {
		return nil, fmt.Errorf("pipeline engine not initialized")
	}

	if result, err := RunPlStr(data, "", 0, p.scriptInfo.Engine()); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (p *Pipeline) UpdateScriptInfo() error {
	var err error = nil
	p.scriptInfo, err = scriptstore.QueryScript(p.scriptInfo.Name(), p.scriptInfo)
	return err
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
	loadedPatterns, err := readPatternsFromDir(datakit.PipelinePatternDir)
	if err != nil {
		return err
	}

	for k, v := range loadedPatterns {
		if _, ok := parser.GlobalPatterns[k]; !ok {
			parser.GlobalPatterns[k] = v
		} else {
			l.Warnf("can not overwrite internal pattern `%s', skipped `%s'", k, k)
		}
	}
	return nil
}

func readPatternsFromDir(path string) (map[string]string, error) {
	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path += "/*"
		}
	} else {
		return nil, fmt.Errorf("invalid path : %s", path)
	}

	files, _ := filepath.Glob(path)

	patterns := make(map[string]string)
	for _, fileName := range files {
		file, err := os.Open(filepath.Clean(fileName))
		if err != nil {
			return patterns, err
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 && l[0] != '#' {
				names := strings.SplitN(l, " ", 2)
				patterns[names[0]] = names[1]
			}
		}

		if err := file.Close(); err != nil {
			l.Warnf("Close: %s, ignored", err.Error())
		}
	}

	return patterns, nil
}

func RunPlStr(cntStr, source string, maxMessageLen int, ng *parser.Engine) (*Result, error) {
	result := &Result{
		Output: nil,
	}
	if ng != nil {
		if err := ng.Run(cntStr); err != nil {
			l.Debug(err)
			result.Err = err.Error()
		}
		result.Output = ng.Result()
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
