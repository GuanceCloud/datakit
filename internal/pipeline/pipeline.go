// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pipeline implement datakit's logging pipeline.
package pipeline

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	// it will use this embedded information in time/tzdata.
	_ "time/tzdata"

	"github.com/GuanceCloud/cliutils/logger"
	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/point"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"

	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var l = logger.DefaultSLogger("pipeline")

func InitPipeline(cfg *plmanager.PipelineCfg, upFn plmap.UploadFunc, gTags map[string]string,
	installDir string,
) error {
	l = logger.SLogger("pipeline")
	return plval.InitPlVal(cfg, upFn, gTags, installDir)
}

func NewPlScriptSampleFromFile(category point.Category, path string, buks ...*plmap.AggBuckets) (*plmanager.PlScript, error) {
	name, script, err := plmanager.ReadPlScriptFromFile(path)
	if err != nil {
		return nil, err
	}

	return NewPlScriptSimple(category, name, script, buks...)
}

func NewPlScriptSimple(category point.Category, name, script string, buks ...*plmap.AggBuckets) (*plmanager.PlScript, error) {
	scs, errs := plmanager.NewScripts(map[string]string{name: script}, map[string]string{}, "", category, buks...)

	if v, ok := errs[name]; ok {
		return nil, v
	}

	if sc, ok := scs[name]; ok {
		return sc, nil
	}

	return nil, fmt.Errorf("unknown error")
}

func NewPipelineMulti(category point.Category, scripts map[string]string,
	scriptPath map[string]string, buks *plmap.AggBuckets,
) (map[string]*plmanager.PlScript, map[string]error) {
	return plmanager.NewScripts(scripts, scriptPath, "", category, buks)
}

type Pipeline struct {
	Script *plmanager.PlScript
}

func (p *Pipeline) Run(cat point.Category, pt *point.Point, plOpt *plmanager.Option,
	signal plruntime.Signal, buks ...*plmap.AggBuckets,
) (ptinput.PlInputPt, error) {
	if p.Script == nil || p.Script.Engine() == nil {
		return nil, fmt.Errorf("pipeline engine not initialized")
	}

	if pt == nil {
		return nil, fmt.Errorf("no data")
	}

	plpt := ptinput.WrapPoint(cat, pt)

	if len(buks) > 0 {
		p.Script.SetAggBuks(buks[0])
	}

	if err := p.Script.Run(plpt, signal, plOpt); err != nil {
		return nil, err
	} else {
		return plpt, nil
	}
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
	d, e := io.ReadAll(reader)
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
