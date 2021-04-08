package man

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/gobuffalo/packr/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	ManualBox = packr.New("manulas", "./manuals")

	l = logger.DefaultSLogger("man")
)

type Input struct {
	InputName    string
	Catalog      string
	InputSample  string
	Version      string
	ReleaseDate  string
	Measurements []*inputs.MeasurementInfo
	CSS          string
}

func Get(name string) (string, error) {
	return ManualBox.FindString(name + ".md")
}

func BuildMarkdownManual(name string) ([]byte, error) {

	x, ok := inputs.Inputs[name]
	if !ok {
		return nil, fmt.Errorf("intput %s not found", name)
	}

	input := x()
	switch i := input.(type) {
	case inputs.ManualInput: // pass
		sampleMeasurements := i.SampleMeasurement()
		x := Input{
			InputName:   name,
			InputSample: i.SampleConfig(),
			Catalog:     i.Catalog(),
			Version:     git.Version,
			ReleaseDate: git.BuildAt,
			CSS:         css,
		}
		for _, m := range sampleMeasurements {
			x.Measurements = append(x.Measurements, m.Info())
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
		if err := temp.Execute(&buf, x); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		return nil, nil
	}
}
