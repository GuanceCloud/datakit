// +build linux

package sensors

import (
	"log"
	"strconv"
	"strings"
	"testing"
)

type entry struct {
	tags   map[string]string
	fields map[string]interface{}
}

func TestParse(t *testing.T) {
	output := `coretemp-isa-0000
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
		} else {
			if strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if strings.HasSuffix(line, ":") {
					if len(fields) != 0 {
						entries = append(entries, entry{tags: tags, fields: fields})
						tmp := make(map[string]string)
						for k, v := range tags {
							tmp[k] = v
						}
						tags = tmp
						fields = make(map[string]interface{})
					}
					tags["feature"] = strings.ToLower(strings.Replace(strings.TrimSpace(parts[0]), " ", "_", -1))
				} else if strings.HasPrefix(parts[0], " ") {
					if value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err != nil {
						log.Println(err.Error())
					} else {
						fields[strings.ToLower(strings.TrimSpace(parts[0]))] = value
					}
				} else {
					tags[strings.ToLower(parts[0])] = strings.TrimSpace(parts[1])
				}
			} else {
				tags["chip"] = line
			}
		}
	}
	entries = append(entries, entry{tags: tags, fields: fields})

	for _, v := range entries {
		log.Println(v.tags)
		log.Println(v.fields)
		log.Println("##############")
	}
}
