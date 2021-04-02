package cmds

import (
	"bytes"
	"fmt"
	"os"
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
		getMan(s)
	}
}

func getMan(inputName string) {
	c, ok := inputs.Inputs[inputName]
	if !ok {
		fmt.Printf("[E] intput %s not found: %s\n", inputName)
		return
	}

	input := c() // construct input

	var sampleMeasurements []inputs.Measurement
	switch i := input.(type) {
	case inputs.ManualInput:
		sampleMeasurements = i.SampleMeasurement()
	default:
		fmt.Printf("[E] intput %s not implement SampleMeasurement()\n", inputName)
		return
	}

	md, err := man.Get(inputName)
	if err != nil {
		fmt.Printf("[E] get manual failed: %s\n", err.Error())
		return
	}

	temp, err := template.New(inputName).Parse(md)
	if err != nil {
		fmt.Printf("[E] invalid markdown template in %s: %s\n", inputName, err.Error())
		return
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
	temp.Execute(&buf, x)

	result := markdown.Render(buf.String(), 80, 6)
	fmt.Println(string(result))
}
