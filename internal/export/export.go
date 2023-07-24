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

	"github.com/GuanceCloud/cliutils/logger"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	plfuncs "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var l = logger.DefaultSLogger("man")

func SetLog() {
	l = logger.SLogger("man")
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
	langs         []inputs.I18n
	exclude       string
	topDir        string
	version       string
	ignoreMissing bool
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

// buildInputDoc render inputs docs based on input document template.
func buildInputDoc(inputName string, md []byte, opt *exportOptions) ([]byte, error) {
	c := inputs.Inputs[inputName]

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
		Version:        opt.version,
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

// buildNonInputDocs render non-inputs docs.
func buildNonInputDocs(md []byte, opt *exportOptions) ([]byte, error) {
	p := &Params{
		Version:     opt.version,
		ReleaseDate: git.BuildAt,

		DatakitConfSample: DatakitConfSample,
	}

	if buf, err := renderBuf(md, p); err != nil {
		return nil, err
	} else {
		return buf, nil
	}
}

// buildPipelineDocs render pipeline function docs.
func buildPipelineDocs(
	md []byte,
	fndocs map[string]*plfuncs.PLDoc,
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
		Version:     opt.version,
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
