// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package man manages all datakit documents
package man

import (
	"bytes"
	"sort"

	// nolint:typecheck
	"strings"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	plfuncs "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var l = logger.DefaultSLogger("man")

type Params struct {
	InputName      string
	Catalog        string
	InputSample    string
	Version        string
	ReleaseDate    string
	Measurements   []*inputs.MeasurementInfo
	CSS            string
	AvailableArchs string
	PipelineFuncs  string
}

type Option struct {
	WithCSS                       bool
	IgnoreMissing                 bool
	DisableMonofontOnTagFieldName bool
	ManVersion                    string
}

func BuildMarkdownManual(name string, opt *Option) ([]byte, error) {
	var p *Params

	css := MarkdownCSS
	ver := datakit.Version

	if !opt.WithCSS {
		css = ""
	}

	if opt.ManVersion != "" {
		ver = opt.ManVersion
	}

	if opt.DisableMonofontOnTagFieldName {
		inputs.MonofontOnTagFieldName = false
	}

	// check if doc's name is a input name
	if creator, ok := inputs.Inputs[name]; ok {
		l.Debugf("build doc for input %s...", name)

		ipt := creator()
		switch i := ipt.(type) {
		case inputs.InputV2:

			sampleMeasurements := i.SampleMeasurement()
			p = &Params{
				InputName:      name,
				InputSample:    i.SampleConfig(),
				Catalog:        i.Catalog(),
				Version:        ver,
				ReleaseDate:    git.BuildAt,
				CSS:            css,
				AvailableArchs: strings.Join(i.AvailableArchs(), " "),
			}

			for _, m := range sampleMeasurements {
				p.Measurements = append(p.Measurements, m.Info())
			}
		default:
			l.Warnf("incomplete input: %s", name)
			return nil, nil
		}
	} else {
		p = &Params{
			Version:     ver,
			ReleaseDate: git.BuildAt,
			CSS:         css,
		}

		l.Debugf("build doc for %s...", name)

		// NOTE: pipeline.md is not input's doc, we have to put all pipeline functions documents
		// to pipeline.md
		if name == "pipeline" {
			sb := strings.Builder{}
			arr := []string{}
			for k := range plfuncs.PipelineFunctionDocs {
				arr = append(arr, k)
			}

			sort.Strings(arr) // order by name

			for _, elem := range arr {
				sb.WriteString(plfuncs.PipelineFunctionDocs[elem].Doc)
			}
			p.PipelineFuncs = sb.String()
		}
	}

	md, err := docs.ReadFile("manuals/" + name + ".md")
	if err != nil {
		if !opt.IgnoreMissing {
			return nil, err
		} else {
			l.Warn(err)
			return nil, nil
		}
	}

	temp, err := template.New(name).Parse(string(md))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
