// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package man manages all datakit documents
package man

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	plfuncs "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var l = logger.DefaultSLogger("man")

// A Params defined various template parameters to build docs
// and command line output.
type Params struct {
	// Various fields used to render meta info into markdown documents.
	InputName         string
	Catalog           string
	InputSample       string
	Version           string
	ReleaseDate       string
	AvailableArchs    string
	PipelineFuncs     string
	PipelineFuncsEN   string
	DatakitConfSample string

	// Measurements used to render metric info into markdown documents.
	Measurements []*inputs.MeasurementInfo

	// Dashboard used to mapping en/zh contents into dashboard JSON
	Dashboard map[string]string

	ic *installCmd
}

// A ExportOption defined various doc export options.
type ExportOption struct {
	Skips,
	Path,
	ManVersion string

	Lang          inputs.I18n
	IgnoreMissing bool
}

// BuildDashboard render dashboard JSON.
func BuildDashboard(inputName string, j []byte, lang inputs.I18n) ([]byte, error) {
	c, ok := inputs.Inputs[inputName]
	if !ok {
		return nil, fmt.Errorf("unknown input %s", inputName)
	}

	var (
		ipt       = c()
		dashboard map[string]string
	)

	switch i := ipt.(type) {
	case inputs.Dashboard:
		dashboard = i.Dashboard(lang)
		if len(dashboard) == 0 {
			l.Warnf("input %s got no dashboard rendering, ignored", inputName)
			return nil, nil
		}

	default:
		l.Warnf("input %s not implement Dashboard interfaces, ignored", inputName)
		return nil, nil
	}

	p := &Params{
		Dashboard: dashboard,
	}

	buf, err := renderBuf(j, p)
	if err != nil {
		return nil, err
	}

	// check if JSON ok
	if !json.Valid(buf) {
		return nil, fmt.Errorf("invalid dashboard on input %q", inputName)
	}
	return buf, nil
}

// BuildInputDoc render inputs docs based on input document template.
func BuildInputDoc(inputName string, md []byte, opt *ExportOption) ([]byte, error) {
	c, ok := inputs.Inputs[inputName]
	if !ok {
		return nil, fmt.Errorf("unknown input %s", inputName)
	}

	var (
		ipt          = c()
		measurements []*inputs.MeasurementInfo
		archs        string
	)

	switch i := ipt.(type) {
	case inputs.InputV2:
		for _, m := range i.SampleMeasurement() {
			measurements = append(measurements, m.Info())
		}

		archs = strings.Join(i.AvailableArchs(), " ")
	default:
		l.Warnf("input %s not implement InputV2 interfaces, ignored", inputName)
		return nil, nil
	}

	p := &Params{
		InputName:      inputName,
		InputSample:    ipt.SampleConfig(),
		Catalog:        ipt.Catalog(),
		Version:        opt.ManVersion,
		ReleaseDate:    git.BuildAt,
		AvailableArchs: archs,
		Measurements:   measurements,
	}

	if buf, err := renderBuf(md, p); err != nil {
		return nil, fmt.Errorf("template.New(%s): %w", inputName, err)
	} else {
		return buf, nil
	}
}

// BuildNonInputDocs render non-inputs docs.
func BuildNonInputDocs(md []byte, opt *ExportOption) ([]byte, error) {
	p := &Params{
		Version:     opt.ManVersion,
		ReleaseDate: git.BuildAt,

		DatakitConfSample: DatakitConfSample,
	}

	if buf, err := renderBuf(md, p); err != nil {
		return nil, err
	} else {
		return buf, nil
	}
}

// BuildPipelineDocs render pipeline function docs.
func BuildPipelineDocs(
	md []byte,
	fndocs map[string]*plfuncs.PLDoc,
	opt *ExportOption,
) ([]byte, error) {
	arr := []string{}
	for k := range fndocs {
		arr = append(arr, k)
	}

	// Order by name: make the table-of-contents
	// sorted and easy to find function doc by name.
	sort.Strings(arr)

	sb := strings.Builder{}
	for _, elem := range arr {
		sb.WriteString(fndocs[elem].Doc + "\n\n")
	}

	p := &Params{
		Version:     opt.ManVersion,
		ReleaseDate: git.BuildAt,

		DatakitConfSample: DatakitConfSample,
		PipelineFuncs:     sb.String(),
	}

	if buf, err := renderBuf(md, p); err != nil {
		return nil, err
	} else {
		return buf, nil
	}
}

func renderBuf(md []byte, p *Params) ([]byte, error) {
	// render raw markdown
	temp, err := template.New("").Funcs(map[string]interface{}{
		"CodeBlock": codeBlock,
		"InstallCmd": func(indent int, opts ...InstallOpt) string {
			p.ic = InstallCommand(opts...)
			return codeBlock(p.ic.String(), indent)
		},
	}).Parse(string(md))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, p); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func codeBlock(block string, indent int) string {
	arr := []string{}
	for _, line := range strings.Split(block, "\n") {
		arr = append(arr, strings.Repeat(" ", indent)+line)
	}
	return strings.Join(arr, "\n")
}
