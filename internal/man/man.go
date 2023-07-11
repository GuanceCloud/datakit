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
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
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

	// Dashboard and Monitor used to mapping en/zh contents into dashboard/monitor JSON
	Dashboard, Monitor map[string]string

	ic     *installCmd
	delims [2]string
}

// A ExportOption defined various doc export options.
type ExportOption struct {
	Skips,
	Path,
	ManVersion string

	Lang          inputs.I18n
	IgnoreMissing bool
}

func BuildMonitor(inputName string, lang inputs.I18n) (map[string][]byte, error) {
	c, ok := inputs.Inputs[inputName]
	if !ok {
		return nil, fmt.Errorf("unknown input %s", inputName)
	}

	var (
		ipt = c()

		monitor map[string]string

		templateMap = map[string][]byte{}
		resMap      = map[string][]byte{}

		p *Params
	)

	// load default monitor
	if x, err := monitorTryLoad(inputName, lang); len(x) > 0 && err == nil {
		templateMap[inputName] = x
	}

	switch i := ipt.(type) {
	case inputs.Monitor:
		monitor = i.Monitor(lang)
		ml := []string{inputName}
		if arr := i.MonitorList(); len(arr) > 0 {
			ml = arr
		}

		l.Infof("input %s got %d monitor rendering", inputName, len(ml))

		for _, elem := range i.MonitorList() {
			l.Infof("rendering monitor %q ...", elem)
			if x, err := monitorTryLoad(elem, lang); len(x) > 0 && err == nil {
				templateMap[elem] = x
			}
		}

	default:
		cp.Warnf("[W] input %s not implement Monitor interfaces, ignored\n", inputName)
	}

	p = &Params{
		Monitor: monitor,

		// For monitor json, we have to escape jinja2 template(also the format {{ xx }}),
		// so we use customer delimeters for Go template.
		delims: [2]string{"<<", ">>"},
	}

	l.Infof("build %d monitors for %s...", len(templateMap), inputName)
	for k, t := range templateMap {
		l.Infof("render monitor %s...", k)
		buf, err := renderBuf(t, p)
		if err != nil {
			l.Errorf("renderBuf: render monitor on input %q[%q]: %s, ignored", inputName, k, err)
			return nil, err
		}

		// check if JSON ok
		if !json.Valid(buf) {
			return nil, fmt.Errorf("invalid monitor JSON on input %q[%q]", inputName, k)
		}

		resMap[k] = buf
	}

	return resMap, nil
}

// BuildDashboard render all dashboard JSON of input inputName.
func BuildDashboard(inputName string, lang inputs.I18n) (map[string][]byte, error) {
	c, ok := inputs.Inputs[inputName]
	if !ok {
		return nil, fmt.Errorf("unknown input %s", inputName)
	}

	var (
		ipt = c()

		dashboard map[string]string

		templateMap = map[string][]byte{}
		resMap      = map[string][]byte{}

		p *Params
	)

	// load default dashboard
	if x, err := dashboardTryLoad(inputName, lang); err == nil && len(x) > 0 {
		templateMap[inputName] = x
	}

	switch i := ipt.(type) {
	case inputs.Dashboard:
		dashboard = i.Dashboard(lang)
		dl := []string{inputName}

		if arr := i.DashboardList(); len(arr) > 0 {
			dl = arr
		}

		l.Infof("input %s got %d dashboard rendering", inputName, len(dl))

		for _, elem := range dl {
			l.Infof("rendering dashboard %q ...", elem)
			if x, err := dashboardTryLoad(elem, lang); len(x) > 0 && err == nil {
				templateMap[elem] = x
			} else {
				cp.Warnf("[W] dashboardTryLoad %s/%s failed: %s\n", elem, lang.String(), err)
			}
		}

	default:
		cp.Warnf("[W] input %s not implement Dashboard interfaces, ignored", inputName)
	}

	p = &Params{
		Dashboard: dashboard,
	}

	l.Infof("build %d dashboards for %s...", len(templateMap), inputName)

	for k, t := range templateMap {
		l.Infof("render dashboard %s...", k)
		buf, err := renderBuf(t, p)
		if err != nil {
			return nil, fmt.Errorf("renderBuf: render dashboard on input %q[%q]: %w",
				inputName, k, err)
		} else {
			// check if JSON ok
			if !json.Valid(buf) {
				return nil, fmt.Errorf("invalid dashboard JSON on input %q[%q]", inputName, k)
			}

			resMap[k] = buf
		}
	}

	return resMap, nil
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
		cp.Warnf("[W] input %s not implement InputV2 interfaces, ignored\n", inputName)
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

	var (
		temp *template.Template
		err  error
	)

	if len(p.delims) == 2 {
		temp, err = template.New("").
			Delims(p.delims[0], p.delims[1]). // use customer delimeter
			Funcs(map[string]interface{}{
				"CodeBlock": codeBlock,
				"InstallCmd": func(indent int, opts ...InstallOpt) string {
					p.ic = InstallCommand(opts...)
					return codeBlock(p.ic.String(), indent)
				},
			}).Parse(string(md))
		if err != nil {
			return nil, err
		}
	} else {
		temp, err = template.New("").Funcs(map[string]interface{}{
			"CodeBlock": codeBlock,
			"InstallCmd": func(indent int, opts ...InstallOpt) string {
				p.ic = InstallCommand(opts...)
				return codeBlock(p.ic.String(), indent)
			},
		}).Parse(string(md))
		if err != nil {
			return nil, err
		}
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
