// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package export manage all exporting sources.
package export

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var l = logger.DefaultSLogger("export")

func SetLog() {
	l = logger.SLogger("export")
}

// Exporter provide exporting resource of Datakit, such as
//   - markdown docs
//   - dashboard json
//   - monitor json
//   - pipeline related resource
//   - all metric tag/field specs
type Exporter interface {
	Export() error
	Check() error
}

type exportOptions struct {
	langs   []inputs.I18n
	exclude string
	topDir  string

	dcaVersion, version string

	datakitImageURL string
	ignoreMissing   bool
	allMeasurements string
}

type option func(*exportOptions)

// WithI18n set exported languages.
func WithI18n(langs []inputs.I18n) option {
	return func(o *exportOptions) {
		if len(langs) > 0 {
			o.langs = langs
		}
	}
}

// WithExclude set excluded name(comma splited).
func WithExclude(list string) option {
	return func(o *exportOptions) {
		o.exclude = list
	}
}

// WithTopDir set exported top dir.
func WithTopDir(dir string) option {
	return func(o *exportOptions) {
		o.topDir = dir
	}
}

// WithVersion set exported version.
func WithVersion(v string) option {
	return func(o *exportOptions) {
		o.version = v
	}
}

// WithDCAVersion set exported DCA version.
func WithDCAVersion(v string) option {
	return func(o *exportOptions) {
		o.dcaVersion = v
	}
}

// WithDatakitImageURL set datakit docker image URL.
//
// We need this URL to export yaml for datakit daemonset.
func WithDatakitImageURL(x string) option {
	return func(o *exportOptions) {
		o.datakitImageURL = x
	}
}

// WithIgnoreMissing to ignore or not when resource not found.
func WithIgnoreMissing(on bool) option {
	return func(o *exportOptions) {
		o.ignoreMissing = on
	}
}

// A Params defined various template parameters to build docs
// and command line output.
type Params struct {
	// Various fields used to render meta info into markdown documents.
	InputName   string
	Catalog     string
	InputSample string

	DCAVersion,
	Version string

	InputENVSample      string
	InputENVSampleZh    string
	NonInputENVSample   map[string]string
	NonInputENVSampleZh map[string]string

	ChangeManifests *changes.Manifests

	ReleaseDate       string
	AvailableArchs    string
	PipelineFuncs     string
	PipelineFuncsEN   string
	DatakitConfSample string
	Year              string // year of current time

	// Measurements used to render metric info into markdown documents.
	Measurements []*inputs.MeasurementInfo

	// AllMeasurements used to show all measurement and it's category and source(input name)
	AllMeasurements string

	// Dashboard and Monitor used to mapping en/zh contents into dashboard/monitor JSON
	Dashboard, Monitor map[string]string

	ic             *installCmd
	templateDelims [2]string
}

// buildInputDoc render inputs docs based on input document template.
func buildInputDoc(inputName string, md []byte, opt *exportOptions) ([]byte, error) {
	c := inputs.AllInputs[inputName]

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
		return nil, fmt.Errorf("input %s not implement InputV2 interfaces", inputName)
	}

	p := &Params{
		InputName:   inputName,
		InputSample: ipt.SampleConfig(),
		Catalog:     ipt.Catalog(),

		Version:    opt.version,
		DCAVersion: opt.dcaVersion,

		ChangeManifests: changes.MustLoadAllManifest(),

		ReleaseDate:    git.BuildAt,
		AvailableArchs: archs,
		Measurements:   measurements,
		Year:           fmt.Sprintf("%d", time.Now().Year()),
	}

	if inp, ok := ipt.(inputs.GetENVDoc); ok {
		p.InputENVSample = inputs.GetENVSample(inp.GetENVDoc(), false)
		p.InputENVSampleZh = inputs.GetENVSample(inp.GetENVDoc(), true)
	}

	if buf, err := p.renderBuf(md); err != nil {
		return nil, fmt.Errorf("renderBuf() on input %q: %w", inputName, err)
	} else {
		return buf, nil
	}
}

// buildNonInputDocs render non-inputs docs.
func buildNonInputDocs(fileName string, md []byte, opt *exportOptions) ([]byte, error) {
	p := &Params{
		Version:             opt.version,
		DCAVersion:          opt.dcaVersion,
		ReleaseDate:         git.BuildAt,
		ChangeManifests:     changes.MustLoadAllManifest(),
		NonInputENVSample:   make(map[string]string),
		NonInputENVSampleZh: make(map[string]string),
		DatakitConfSample:   datakit.MainConfSample(datakit.BrandDomainTemplate),
		AllMeasurements:     opt.allMeasurements,
		Year:                fmt.Sprintf("%d", time.Now().Year()),
	}

	if _, ok := nonInputDocs[fileName]; ok {
		for contentName, info := range nonInputDocs[fileName] {
			p.NonInputENVSample[contentName] = inputs.GetENVSample(info, false)
			p.NonInputENVSampleZh[contentName] = inputs.GetENVSample(info, true)
		}
	}

	if buf, err := p.renderBuf(md); err != nil {
		return nil, fmt.Errorf("renderBuf() on %q: %w", fileName, err)
	} else {
		return buf, nil
	}
}

// buildPipelineDocs render pipeline function docs.
func buildPipelineDocs(
	md []byte,
	fndocs map[string]*funcs.PLDoc,
	opt *exportOptions,
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
		Version:         opt.version,
		DCAVersion:      opt.dcaVersion,
		ReleaseDate:     git.BuildAt,
		ChangeManifests: changes.MustLoadAllManifest(),

		DatakitConfSample: datakit.MainConfSample(datakit.BrandDomainTemplate),
		PipelineFuncs:     sb.String(),
		Year:              fmt.Sprintf("%d", time.Now().Year()),
	}

	if buf, err := p.renderBuf(md); err != nil {
		return nil, err
	} else {
		return buf, nil
	}
}

// renderBuf render parameters into raw markdown.
func (p *Params) renderBuf(md []byte) ([]byte, error) {
	var (
		temp *template.Template
		err  error
	)

	if len(p.templateDelims) == 2 {
		temp, err = template.New("").
			Delims(p.templateDelims[0], p.templateDelims[1]). // use customer delimeter
			Funcs(map[string]interface{}{
				"CodeBlock": codeBlock,
				"UISteps":   uiSteps,
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

func uiSteps(steps, sep string) string {
	arr := []string{}
	for _, x := range strings.Split(steps, sep) {
		arr = append(arr, "**"+strings.TrimSpace(x)+"**") // each step are bold font
	}
	return strings.Join(arr, " ➔ ")
}
