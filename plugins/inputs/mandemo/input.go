package mandemo

import (
	//"github.com/gobuffalo/packr/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	//manBox        = packr.New("man", "./man")
	//manBoxAnother = packr.New("man", "./man") // same-name seems ok
	inputName = "mandemo"
)

type Demo struct{}

func (i *Demo) Man() (string, error) {
	md, err := man.Get(inputName + ".md")
	if err != nil {
		return "", err
	}

	return md, nil
}

func (i *Demo) Run()                              { return }
func (i *Demo) Catalog() string                   { return "testing" }
func (i *Demo) SampleConfig() string              { return "[inputs.mandemo]" }
func (a *Demo) Test() (*inputs.TestResult, error) { return nil, nil }

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Demo{}
	})
}
