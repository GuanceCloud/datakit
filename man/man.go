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
	ManualBox = packr.New("manulas", "./manuals")
	OtherDocs = map[string]interface{}{
		// value not used, just document the markdown relative path
		"pipeline": "man/manuals/pipeline.md",
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

func BuildMarkdownManual(name string) ([]byte, error) {

	var p *Params

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
		case inputs.InputV2: // pass
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
		return nil, err
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
