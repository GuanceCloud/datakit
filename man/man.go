// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package man manages all datakit documents
package man

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"

	// nolint:typecheck
	"strings"
	"text/template"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	plfuncs "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var l = logger.DefaultSLogger("man")

type Params struct {
	InputName       string
	Catalog         string
	InputSample     string
	Version         string
	ReleaseDate     string
	Measurements    []*inputs.MeasurementInfo
	CSS             string
	AvailableArchs  string
	PipelineFuncs   string
	PipelineFuncsEN string
}

type i18n int

const (
	I18nZH = iota
	I18nEN = iota
)

func (x i18n) String() string {
	switch x {
	case I18nZH:
		return "zh"
	case I18nEN:
		return "en"
	default:
		return ""
	}
}

type Option struct {
	WithCSS                       bool
	IgnoreMissing                 bool
	Skips                         string
	DisableMonofontOnTagFieldName bool
	ManVersion                    string
	Path                          string
}

func BuildMarkdownManual(name string, opt *Option) (map[i18n][]byte, error) {
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
			{ // zh
				sb := strings.Builder{}
				arr := []string{}
				for k := range plfuncs.PipelineFunctionDocs {
					arr = append(arr, k)
				}

				sort.Strings(arr) // order by name

				for _, elem := range arr {
					sb.WriteString(plfuncs.PipelineFunctionDocs[elem].Doc + "\n\n")
				}
				p.PipelineFuncs = sb.String()
			}

			{ // en
				sb := strings.Builder{}
				arr := []string{}
				for k := range plfuncs.PipelineFunctionDocsEN {
					arr = append(arr, k)
				}

				sort.Strings(arr) // order by name

				for _, elem := range arr {
					sb.WriteString(plfuncs.PipelineFunctionDocsEN[elem].Doc + "\n\n")
				}
				p.PipelineFuncsEN = sb.String()
			}
		}
	}

	res := map[i18n][]byte{}
	for _, x := range []i18n{I18nZH, I18nEN} {
		// read raw markdown from embed repository
		md, err := docs.ReadFile(filepath.Join("docs", x.String(), name+".md"))
		if err != nil {
			if !opt.IgnoreMissing {
				return nil, err
			} else {
				l.Warn(err)
				continue
			}
		}

		// render raw markdown
		temp, err := template.New(name).Funcs(map[string]interface{}{
			"CodeBlock": func(code string, indent int) string {
				arr := []string{}
				for _, line := range strings.Split(code, "\n") {
					arr = append(arr, strings.Repeat(" ", indent)+line)
				}
				return strings.Join(arr, "\n")
			},
		}).Parse(string(md))
		if err != nil {
			return nil, fmt.Errorf("[%s] template.New(%s): %w", x, name, err)
		}

		var buf bytes.Buffer
		if err := temp.Execute(&buf, p); err != nil {
			return nil, err
		}

		res[x] = buf.Bytes()
	}

	return res, nil
}
