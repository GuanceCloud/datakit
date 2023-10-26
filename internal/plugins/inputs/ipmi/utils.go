// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ipmi collects host ipmi metrics.
package ipmi

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type metricType struct {
	dataName string
	dataType string
}

var metricTypes = []metricType{
	{"status", "int"},
	{"current", "float"},
	{"voltage", "float"},
	{"power", "float"},
	{"temp", "float"},
	{"fan_speed", "int"},
	{"usage", "float"},
	{"count", "int"},
}

// The bytes got from ipmi server.
type dataStruct struct {
	server        string
	metricVersion int
	data          []byte
}

func (ipt *Input) getBytes(index int) (*dataStruct, error) {
	if err := ipt.checkConfig(); err != nil {
		return nil, err
	}

	// Get exec parameters.
	opts, metricVersion, err := ipt.getParameters(index)
	if err != nil {
		return nil, fmt.Errorf("getParameters :%w", err)
	}

	l.Debugf("before get bytes, server: ", ipt.IpmiServers[index])
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout)
	defer cancel()
	//nolint:gosec
	c := exec.CommandContext(ctx, ipt.BinPath, opts...)

	// exec.Command ENV
	if len(ipt.Envs) != 0 {
		// In windows here will broken old PATH.
		c.Env = ipt.Envs
	}

	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("c.Start(): %w, %v", err, b.String())
	}
	err = c.Wait()
	if err != nil {
		return nil, fmt.Errorf("c.Wait(): %s, %w, %v", ipt.IpmiServers[index], err, b.String())
	}

	bytes := b.Bytes()
	l.Debugf("get bytes len: %v. consuming: %v", len(bytes), time.Since(start))

	return &dataStruct{
		server:        ipt.IpmiServers[index],
		metricVersion: metricVersion,
		data:          bytes,
	}, err
}

func (ipt *Input) getPoints(data []byte, metricVersion int, server string) ([]*point.Point, error) {
	// data just like
	// V1
	// Temp             | 45 degrees C      | ok
	// NDC PG           | 0x00              | ok

	// V2
	// Inlet Temp       | 05h | ok  |  7.1 | 23 degrees C
	// NDC PG           | 08h | ok  |  7.1 | State Deasserted

	pts := make([]*point.Point, 0)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.start))

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

		var kvs point.KVs

		kvs = kvs.Add("host", server, true, true)
		kvs = kvs.Add("unit", strs[0], true, true)
		kvs = kvs.Add(metricTypes[metricIndex].dataName, value, false, true)

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		pts = append(pts, point.NewPointV2(inputName, kvs, opts...))
	}

	return pts, nil
}

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
