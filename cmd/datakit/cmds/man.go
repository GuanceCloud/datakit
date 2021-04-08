package cmds

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/c-bata/go-prompt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func Man() {

	// load input-names
	for k, _ := range inputs.Inputs {
		suggestions = append(suggestions, prompt.Suggest{Text: k, Description: ""})
	}

	// TODO: add suggestions for pipeline

	c, _ := newCompleter()

	p := prompt.New(
		runMan,
		c.Complete,
		prompt.OptionTitle("man: DataKit manual query"),
		prompt.OptionPrefix("man > "),
	)

	p.Run()
}

func ExportMan(to string) error {
	if err := os.MkdirAll(to, os.ModePerm); err != nil {
		return err
	}

	for k, _ := range inputs.Inputs {
		data, err := GetMan(k)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			continue
		}

		if err := ioutil.WriteFile(filepath.Join(to, k+".md"), []byte(data), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func runMan(txt string) {
	s := strings.Join(strings.Fields(strings.TrimSpace(txt)), " ")
	if s == "" {
		return
	}

	switch s {
	case "Q", "q", "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		x, err := GetMan(s)
		if err != nil {
			fmt.Printf("[E] %s\n", err.Error())
		} else {
			if len(x) == 0 {
				fmt.Printf("[E] intput %s got no manual", s)
			} else {
				result := markdown.Render(x, 80, 6)
				fmt.Println(string(result))
			}
		}
	}
}

func GetMan(inputName string) (string, error) {
	c, ok := inputs.Inputs[inputName]
	if !ok {
		return "", fmt.Errorf("intput %s not found: %s", inputName)
	}

	input := c() // construct input

	var sampleMeasurements []inputs.Measurement
	switch i := input.(type) {
	case inputs.ManualInput:
		sampleMeasurements = i.SampleMeasurement()
	default:
		return "", nil
	}

	md, err := man.Get(inputName)
	if err != nil {
		return "", err
	}

	temp, err := template.New(inputName).Parse(md)
	if err != nil {
		return "", err
	}

	x := man.Input{
		InputName:   inputName,
		InputSample: input.SampleConfig(),
		Version:     git.Version,
		ReleaseDate: git.BuildAt,
	}

	for _, m := range sampleMeasurements {
		x.Measurements = append(x.Measurements, m.Info())
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, x); err != nil {
		return "", err
	}
	return buf.String(), nil
}
