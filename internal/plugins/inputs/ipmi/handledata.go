// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ipmi collects host ipmi metrics.
package ipmi

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

// handleData handle ipmitools Data.
func (ipt *Input) handleData(dataCache []dataStruct) error {
	for _, dataStru := range dataCache {
		// Got data, delay warning.  (server, isActive, isOnlyAdd)
		ipt.serverOnlineInfo(dataStru.server, true, false)
		// Convert data.
		ipt.convert(dataStru.data, dataStru.metricVersion, dataStru.server)
	}
	return nil
}

// covert to metric.
func (ipt *Input) convert(data []byte, metricVersion int, server string) {
	// data just like
	// V1
	// Temp             | 45 degrees C      | ok
	// NDC PG           | 0x00              | ok

	// V2
	// Inlet Temp       | 05h | ok  |  7.1 | 23 degrees C
	// NDC PG           | 08h | ok  |  7.1 | State Deasserted

	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		// Format the data.
		strs := strings.Split(scanner.Text(), "|")

		if metricVersion == 2 && len(strs) != 5 {
			continue
		}
		if metricVersion != 2 && len(strs) != 3 {
			continue
		}

		// The data in strs[1]
		if metricVersion == 2 {
			strs[1] = strs[4]
		}

		strs[0] = strings.Trim(strs[0], " ")
		strs[0] = strings.ToLower(strs[0])
		strs[0] = strings.ReplaceAll(strs[0], " ", "_")

		strs[1] = strings.Trim(strs[1], " ")
		strs[1] = strings.Split(strs[1], " ")[0]

		strs[2] = strings.Trim(strs[2], " ")
		strs[2] = strings.ToLower(strs[2])

		metricIndex, value, err := ipt.getMetric(strs)
		if err != nil {
			l.Errorf(" error: %v", err)
			continue
		}

		if metricIndex == -1 {
			// Not error, only give up this data.
			continue
		}

		tags := map[string]string{
			"host": server,
			"unit": strs[0],
		}

		fields := map[string]interface{}{
			metricTypes[metricIndex].dataName: value,
		}

		ipt.collectCache = append(ipt.collectCache, ipt.newPoint(tags, fields))
	}
}

// newPoint Create a new point.
func (ipt *Input) newPoint(tags map[string]string, fields map[string]interface{}) *point.Point {
	opts := point.DefaultMetricOptions()

	if ipt != nil && ipt.Election {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalElectionTags()))
	}

	return point.NewPointV2([]byte(inputName),
		append(point.NewTags(tags), point.NewKVs(fields)...),
		opts...)
}

// get metric.
func (ipt *Input) getMetric(strs []string) (int, interface{}, error) {
	// Step 1, Try convert as value metric, like 1234 or 1234.56
	metricIndex, value, err := ipt.getValueMetric(strs)
	if err != nil {
		return -1, "", err
	}
	if metricIndex > -1 {
		// Convert as value metric success.
		return metricIndex, value, nil
	}

	// Step 2, Try convert as status metric.
	return ipt.getStatusMetric(strs)
}

func (ipt *Input) getValueMetric(strs []string) (int, interface{}, error) {
	if strs[2] != "ok" {
		// Not error, only try status metric.
		return -1, "", nil
	}

	var expValue *regexp.Regexp
	var expStr string
	var expStrs []string
	var err error

	// Try if it's a value, only like 1234 or 1234.56
	expValue = regexp.MustCompile(`^\d+(\.?\d+|\d*)$`)
	if !expValue.MatchString(strs[1]) {
		// Not like 1234 or 1234.56, not error, only try status metric.
		return -1, "", nil
	}

	// Traversal try all value metric kind.
	// i form 1, want skip "status".
	for i := 1; i < len(metricTypes); i++ {
		// Get one metric kind regexp list. Like regexp_power = ["pwr","power"]
		expStrs, err = ipt.getExpStrs(metricTypes[i].dataName)
		if err != nil {
			l.Errorf("ipmi getExpStrs : %v", err)
			continue
		}

		// Traversal try all Regexp in this metric kind. Like ["pwr","power"].
		for _, expStr = range expStrs {
			expName, err := regexp.Compile(expStr)
			if err != nil {
				l.Errorf("parsing regexp:: %v", err)
				return 0, "", err
			}

			if expName.MatchString(strs[0]) {
				// Match name
				switch metricTypes[i].dataType {
				case "float":
					expValue = regexp.MustCompile(`^\d+(\.?\d+|\d*)$`)
				case "int":
					expValue = regexp.MustCompile(`^\d+$`)
				default:
					return i, "", fmt.Errorf("ipmi error value type : %s", strs[1])
				}

				if expValue.MatchString(strs[1]) {
					// Matching value succeeded.
					// Get the value.
					switch metricTypes[i].dataType {
					case "float":
						value, err := strconv.ParseFloat(strs[1], 64)
						if err != nil {
							return i, "", fmt.Errorf(" error strconv.ParseFloat : %s", strs[1])
						}
						return i, value, nil
					case "int":
						value, err := strconv.Atoi(strs[1])
						if err != nil {
							return i, "", fmt.Errorf(" error strconv.Atoi : %s", strs[1])
						}
						return i, value, nil
					default:
						return i, "", fmt.Errorf(" error value type: %s", metricTypes[i].dataName)
					}
				} else {
					// Matching value failed
					return i, "", fmt.Errorf("can not matching metric value: %s", strs[0])
				}
			}
		}
	}
	// Not error, only try status metric.
	return -1, "", nil
}

func (ipt *Input) getStatusMetric(strs []string) (int, interface{}, error) {
	// If it's status, only "ok" or "ns".
	// Traversal try only status kind from valueType.
	// metricTypes[0] is "status"
	expStrs, err := ipt.getExpStrs(metricTypes[0].dataName)
	if err != nil {
		l.Errorf(" error: %v", err)
		return 0, "", err
	}

	// Traversal try all Regexp in “status” kind. Like regexp_power = ["pwr","power"]
	for _, expStr := range expStrs {
		expName, err := regexp.Compile(expStr)
		if err != nil {
			l.Errorf("parsing regexp:: %v", err)
			return 0, "", err
		}

		if expName.MatchString(strs[0]) {
			if strs[2] == "ok" {
				return 0, 1, nil
			} else {
				return 0, 0, nil
			}
		}
	}
	// Not error, only give up this data.
	return -1, "", nil
}

// get metric regexp list.
func (ipt *Input) getExpStrs(metricName string) (expStrs []string, err error) {
	switch metricName {
	case "current":
		return ipt.RegexpCurrent, nil
	case "voltage":
		return ipt.RegexpVoltage, nil
	case "power":
		return ipt.RegexpPower, nil
	case "temp":
		return ipt.RegexpTemp, nil
	case "fan_speed":
		return ipt.RegexpFanSpeed, nil
	case "usage":
		return ipt.RegexpUsage, nil
	case "count":
		return ipt.RegexpCount, nil
	case "status":
		return ipt.RegexpStatus, nil
	default:
		return nil, fmt.Errorf("not find metric name: %s", metricName)
	}
}
