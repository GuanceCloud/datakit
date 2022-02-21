// Package pipeline implement datakit's logging pipeline.
package pipeline

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	// it will use this embedded information in time/tzdata.
	_ "time/tzdata"

	influxm "github.com/influxdata/influxdb1-client/models"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb/iploc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
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
	IPdbAttr map[string]string `toml:"ipdb_attr"`
	IPdbType string            `toml:"ipdb_type"`
}

type Pipeline struct {
	engine  *parser.Engine
	output  *parser.Output // 这是一个map指针，不需要make初始化
	lastErr error
}

func NewPipelineByScriptPath(scriptFullPath string, debug bool) (*Pipeline, error) {
	data, err := ioutil.ReadFile(filepath.Clean(scriptFullPath))
	if err != nil {
		return nil, err
	}
	return NewPipeline(string(data), debug)
}

func NewPipeline(script string, debug bool) (*Pipeline, error) {
	ng, err := parser.NewEngine(script, funcs.FuncsMap, funcs.FuncsCheckMap, debug)
	if err != nil {
		return nil, err
	}
	p := &Pipeline{
		engine: ng,
	}

	return p, nil
}

func NewPipelineFromFile(filename string, debug bool) (*Pipeline, error) {
	b, err := ioutil.ReadFile(filename) //nolint:gosec
	if err != nil {
		return nil, err
	}
	return NewPipeline(string(b), debug)
}

// RunPoint line protocol point to pipeline JSON.
func (p *Pipeline) RunPoint(point influxm.Point) *Pipeline {
	defer func() {
		r := recover()
		if r != nil {
			p.lastErr = fmt.Errorf("%v", r)
		}
	}()

	m := map[string]interface{}{"measurement": string(point.Name())}

	if tags := point.Tags(); len(tags) > 0 {
		m["tags"] = map[string]string{}
		for _, tag := range tags {
			m["tags"].(map[string]string)[string(tag.Key)] = string(tag.Value)
		}
	}

	fields, err := point.Fields()
	if err != nil {
		p.lastErr = err
		return p
	}

	for k, v := range fields {
		m[k] = v
	}

	m["time"] = point.UnixNano()

	j, err := json.Marshal(m)
	if err != nil {
		p.lastErr = err
		return p
	}

	return p.Run(string(j))
}

func (p *Pipeline) Run(data string) *Pipeline {
	// reset
	p.output = nil
	p.lastErr = nil

	var f rtpanic.RecoverCallback

	f = func(trace []byte, _ error) {
		defer rtpanic.Recover(f, nil)

		if trace != nil {
			l.Error("panic: %s", string(trace))
			p.lastErr = fmt.Errorf("%s", trace)
			return
		}

		if p.engine == nil {
			p.lastErr = fmt.Errorf("pipeline engine not initialized")
			l.Error(p.lastErr)
			return
		}

		if err := p.engine.Run(data); err != nil {
			p.lastErr = fmt.Errorf("pipeline run error: %w", err)
			l.Error(p.lastErr)
			return
		}

		p.output = p.engine.Result()
	}

	f(nil, nil)
	return p
}

func (p *Pipeline) Result() (*parser.Output, error) {
	return p.output, p.lastErr
}

func (p *Pipeline) LastError() error {
	return p.lastErr
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
