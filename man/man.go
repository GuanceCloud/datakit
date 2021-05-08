package man

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	ManualBox = packr.New("manuals", "./manuals")
	OtherDocs = map[string]interface{}{
		// value not used, just document the markdown relative path
		// all manuals under man/manuals/
		"pipeline":             "man/manuals/pipeline.md",
		"telegraf":             "man/manuals/telegraf.md",
		"changelog":            "man/manuals/changelog.md",
		"datatypes":            "man/manuals/datatypes.md",
		"apis":                 "man/manuals/apis.md",
		"sec-checker":          "man/manuals/sec-checker.md",
		"datakit-how-to":       "man/manuals/datakit-how-to.md",
		"datakit-arch":         "man/manuals/datakit-arch.md",
		"proxy":                "man/manuals/proxy.md",
		"dataway":              "man/manuals/dataway.md",
		"datakit-batch-deploy": "man/manuals/datakit-batch-deploy.md",
		"prometheus":           "man/manuals/prometheus.md",
		"datakit-on-public":    "man/manuals/datakit-on-public.md",
		"election":             "man/manuals/election.md",
	}

	l = logger.DefaultSLogger("man")
)

type Params struct {
	InputName      string
	Catalog        string
	InputSample    string
	Version        string
	ReleaseDate    string
	Measurements   []*inputs.MeasurementInfo
	CSS            string
	AvailableArchs string
}

func Get(name string) (string, error) {
	return ManualBox.FindString(name + ".md")
}

type Option struct {
	WithCSS                       bool
	IgnoreMissing                 bool
	DisableMonofontOnTagFieldName bool
}

func BuildMarkdownManual(name string, opt *Option) ([]byte, error) {

	var p *Params

	css := MarkdownCSS

	if !opt.WithCSS {
		css = ""
	}

	if opt.DisableMonofontOnTagFieldName {
		inputs.MonofontOnTagFieldName = false
	}

	if _, ok := OtherDocs[name]; ok {
		p = &Params{
			Version:     git.Version,
			ReleaseDate: git.BuildAt,
			CSS:         css,
		}
	} else {
		c, ok := inputs.Inputs[name]
		if !ok {
			return nil, fmt.Errorf("intput %s not found", name)
		}

		input := c()
		switch i := input.(type) {
		case inputs.InputV2:
			sampleMeasurements := i.SampleMeasurement()
			p = &Params{
				InputName:      name,
				InputSample:    i.SampleConfig(),
				Catalog:        i.Catalog(),
				Version:        git.Version,
				ReleaseDate:    git.BuildAt,
				CSS:            css,
				AvailableArchs: strings.Join(i.AvailableArchs(), ","),
			}
			for _, m := range sampleMeasurements {
				p.Measurements = append(p.Measurements, m.Info())
			}

		default:
			return nil, nil
		}
	}

	md, err := Get(name)
	if err != nil {
		if !opt.IgnoreMissing {
			return nil, err
		} else {
			l.Warn(err)
			return nil, nil
		}
	}
	temp, err := template.New(name).Parse(md)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
