// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

/*
Test this file with docker:
docker run --rm -it -w /root/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/sensors -v $GOPATH:/root/go ubuntu.golang:latest go test -v .
*/
package sensors

import (
	"log"
	"strconv"
	"strings"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/command"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var output = `coretemp-isa-0000
Adapter: ISA adapter
Package id 0:
  temp1_input: 32.000
  temp1_max: 80.000
  temp1_crit: 100.000
  temp1_crit_alarm: 0.000
Core 0:
  temp2_input: 28.000
  temp2_max: 80.000
  temp2_crit: 100.000
  temp2_crit_alarm: 0.000
Core 1:
  temp3_input: 30.000
  temp3_max: 80.000
  temp3_crit: 100.000
  temp3_crit_alarm: 0.000
Core 2:
  temp4_input: 29.000
  temp4_max: 80.000
  temp4_crit: 100.000
  temp4_crit_alarm: 0.000
Core 3:
  temp5_input: 27.000
  temp5_max: 80.000
  temp5_crit: 100.000
  temp5_crit_alarm: 0.000

acpitz-acpi-0
Adapter: ACPI interface
temp1:
  temp1_input: 27.800
  temp1_crit: 119.000
temp2:
  temp2_input: 29.800
  temp2_crit: 119.000

nouveau-pci-0100
Adapter: PCI adapter
fan1:
  fan1_input: 869.000
temp1:
  temp1_input: 36.000
  temp1_max: 95.000
  temp1_max_hyst: 3.000
  temp1_crit: 105.000
  temp1_crit_hyst: 5.000
  temp1_emergency: 135.000
  temp1_emergency_hyst: 5.000
`

type entry struct {
	tags   map[string]string
	fields map[string]interface{}
}

func TestParseLogic(t *T.T) {
	var (
		lines   = strings.Split(strings.TrimSpace(output), "\n")
		tags    = make(map[string]string)
		fields  = make(map[string]interface{})
		entries []entry
	)
	for _, line := range lines {
		if line == "" {
			entries = append(entries, entry{tags: tags, fields: fields})
			tags = make(map[string]string)
			fields = make(map[string]interface{})
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.Split(line, ":")
			switch {
			case strings.HasSuffix(line, ":"):
				if len(fields) != 0 {
					entries = append(entries, entry{tags: tags, fields: fields})
					tmp := make(map[string]string)
					for k, v := range tags {
						tmp[k] = v
					}
					tags = tmp
					fields = make(map[string]interface{})
				}
				tags["feature"] = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(parts[0]), " ", "_"))
			case strings.HasPrefix(parts[0], " "):
				if value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err != nil {
					t.Error(err.Error())
				} else {
					fields[strings.ToLower(strings.TrimSpace(parts[0]))] = value
				}
			default:
				tags[strings.ToLower(parts[0])] = strings.TrimSpace(parts[1])
			}
		} else {
			tags["chip"] = line
		}
	}
	entries = append(entries, entry{tags: tags, fields: fields})

	for _, v := range entries {
		t.Logf("%v", v.tags)
		t.Logf("%v", v.fields)
		t.Logf("##############")
	}
}

func TestParseOutput(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		input := &Input{
			Path:     "/usr/bin/sensors",
			Interval: datakit.Duration{Duration: 10 * time.Second},
			Timeout:  datakit.Duration{Duration: 3 * time.Second},
			Tags:     map[string]string{"key1": "tag1", "key2": "tag2"},
		}

		cmdout := `
coretemp-isa-0000
Adapter: ISA adapter
Package id 0:
  temp1_input: 32.000
  temp1_max: 80.000
  temp1_crit: 100.000
  temp1_crit_alarm: 0.000
Core 0:
  temp2_input: 28.000
  temp2_max: 80.000
  temp2_crit: 100.000
  temp2_crit_alarm: 0.000
Core 1:
  temp3_input: 30.000
  temp3_max: 80.000
  temp3_crit: 100.000
  temp3_crit_alarm: 0.000
Core 2:
  temp4_input: 29.000
  temp4_max: 80.000
  temp4_crit: 100.000
  temp4_crit_alarm: 0.000
Core 3:
  temp5_input: 27.000
  temp5_max: 80.000
  temp5_crit: 100.000
  temp5_crit_alarm: 0.000
		`

		pts, err := input.parse(cmdout)
		assert.NoError(t, err)

		for idx, pt := range pts {
			assert.Equalf(t, "tag1", pt.Get("key1"), "%d got %s", idx, pt.Pretty())
			assert.Equalf(t, "tag2", pt.Get("key2"), "%d got %s", idx, pt.Pretty())
			assert.Equalf(t, "coretemp-isa-0000", pt.Get("chip"), "%d got %s", idx, pt.Pretty())
			assert.Equalf(t, "ISA adapter", pt.Get("adapter"), "%d got %s", idx, pt.Pretty())
		}

		pt := pts[0]
		assert.Equalf(t, "package_id_0", pt.Get("feature"), "got %s", pt.Pretty())
		assert.Equalf(t, 32.0, pt.Get("temp1_input"), "got %s", pt.Pretty())
		assert.Equalf(t, 80.0, pt.Get("temp1_max"), "got %s", pt.Pretty())
		assert.Equalf(t, 100.0, pt.Get("temp1_crit"), "got %s", pt.Pretty())
		assert.Equalf(t, 0.0, pt.Get("temp1_crit_alarm"), "got %s", pt.Pretty())

		pt = pts[1]
		assert.Equalf(t, "core_0", pt.Get("feature"), "got %s", pt.Pretty())
		assert.Equalf(t, 28.0, pt.Get("temp2_input"), "got %s", pt.Pretty())
		assert.Equalf(t, 80.0, pt.Get("temp2_max"), "got %s", pt.Pretty())
		assert.Equalf(t, 100.0, pt.Get("temp2_crit"), "got %s", pt.Pretty())
		assert.Equalf(t, 0.0, pt.Get("temp2_crit_alarm"), "got %s", pt.Pretty())

		pt = pts[2]
		assert.Equalf(t, "core_1", pt.Get("feature"), "got %s", pt.Pretty())
		assert.Equalf(t, 30.0, pt.Get("temp3_input"), "got %s", pt.Pretty())
		assert.Equalf(t, 80.0, pt.Get("temp3_max"), "got %s", pt.Pretty())
		assert.Equalf(t, 100.0, pt.Get("temp3_crit"), "got %s", pt.Pretty())
		assert.Equalf(t, 0.0, pt.Get("temp3_crit_alarm"), "got %s", pt.Pretty())

		pt = pts[3]
		assert.Equalf(t, "core_2", pt.Get("feature"), "got %s", pt.Pretty())
		assert.Equalf(t, 29.0, pt.Get("temp4_input"), "got %s", pt.Pretty())
		assert.Equalf(t, 80.0, pt.Get("temp4_max"), "got %s", pt.Pretty())
		assert.Equalf(t, 100.0, pt.Get("temp4_crit"), "got %s", pt.Pretty())
		assert.Equalf(t, 0.0, pt.Get("temp4_crit_alarm"), "got %s", pt.Pretty())

		pt = pts[4]
		assert.Equalf(t, "core_3", pt.Get("feature"), "got %s", pt.Pretty())
		assert.Equalf(t, 27.0, pt.Get("temp5_input"), "got %s", pt.Pretty())
		assert.Equalf(t, 80.0, pt.Get("temp5_max"), "got %s", pt.Pretty())
		assert.Equalf(t, 100.0, pt.Get("temp5_crit"), "got %s", pt.Pretty())
		assert.Equalf(t, 0.0, pt.Get("temp5_crit_alarm"), "got %s", pt.Pretty())
	})
}

func TestRunCommand(t *T.T) {
	output, err := command.RunWithTimeout(3*time.Second, false, defPath, "-u")
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(string(output))
}
