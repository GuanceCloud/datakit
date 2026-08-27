// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// TestDocTemplatesParse ensures the committed doc templates stay syntactically
// valid Go templates. Full rendering happens via `make md_export` in CI; this
// guard catches broken placeholders or control structures early.
func TestDocTemplatesParse(t *testing.T) {
	paths := []string{
		filepath.Join("..", "..", "..", "..", "internal", "export", "doc", "zh", "inputs", "winnetflow.md"),
		filepath.Join("..", "..", "..", "..", "internal", "export", "doc", "en", "inputs", "winnetflow.md"),
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		tmpl := template.New("doc").Funcs(template.FuncMap{
			"CodeBlock": func(interface{}, int) string { return "" },
		})
		if _, err := tmpl.Parse(string(b)); err != nil {
			t.Fatalf("%s is not a valid template: %v", p, err)
		}
	}
}

// TestDocTemplatesRender renders both committed doc templates with real input
// data (sample config, measurement metadata, env docs) the same way the
// exporter does, and verifies the httpflow measurement table appears.
func TestDocTemplatesRender(t *testing.T) {
	ipt := NewInput()
	creator, ok := inputs.AllInputs[inputName]
	if !ok {
		t.Fatalf("input %s not registered", inputName)
	}
	registered := creator().(inputs.InputV2)

	measurements := []*inputs.MeasurementInfo{}
	for _, m := range registered.SampleMeasurement() {
		measurements = append(measurements, m.Info())
	}

	envDoc := ""
	if inp, ok := registered.(inputs.GetENVDoc); ok {
		envDoc = inputs.GetENVSample(inp.GetENVDoc(), true)
	}

	codeBlock := func(block string, indent int) string {
		arr := []string{}
		for _, line := range strings.Split(block, "\n") {
			arr = append(arr, strings.Repeat(" ", indent)+line)
		}
		return strings.Join(arr, "\n")
	}

	paths := map[string]struct {
		file   string
		envVar string
	}{
		"zh": {file: filepath.Join("..", "..", "..", "..", "internal", "export", "doc", "zh", "inputs", "winnetflow.md"), envVar: envDoc},
		"en": {file: filepath.Join("..", "..", "..", "..", "internal", "export", "doc", "en", "inputs", "winnetflow.md"), envVar: envDoc},
	}

	for _, tc := range paths {
		b, err := os.ReadFile(tc.file)
		if err != nil {
			t.Fatalf("read %s: %v", tc.file, err)
		}
		tmpl := template.New("doc").Funcs(template.FuncMap{
			"CodeBlock": codeBlock,
		})
		if _, err := tmpl.Parse(string(b)); err != nil {
			t.Fatalf("%s is not a valid template: %v", tc.file, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, map[string]interface{}{
			"InputName":        inputName,
			"InputSample":      ipt.SampleConfig(),
			"InputENVSample":   tc.envVar,
			"InputENVSampleZh": tc.envVar,
			"AvailableArchs":   strings.Join(registered.AvailableArchs(), " "),
			"Measurements":     measurements,
		}); err != nil {
			t.Fatalf("render %s: %v", tc.file, err)
		}

		out := buf.String()
		if !strings.Contains(out, "httpflow") {
			t.Fatalf("%s rendered doc does not mention httpflow", tc.file)
		}
		if !strings.Contains(out, "enable_httpflow") {
			t.Fatalf("%s rendered doc does not include enable_httpflow sample config", tc.file)
		}
	}
}
