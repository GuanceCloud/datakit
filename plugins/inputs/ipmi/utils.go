// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

package ipmi

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// NewFieldInfoB new ibyte data type.
func NewFieldInfoB(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

// NewFieldInfoP new precent filed.
func NewFieldInfoP(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Float,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}

// NewFieldInfoC new count field.
func NewFieldInfoC(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func Strings2StringSlice(str string) (strs []string, err error) {
	// trim blank
	str = strings.Trim(str, " ")
	// trim "[" "]"
	str = strings.Trim(str, "[")
	str = strings.Trim(str, "]")
	str = strings.Trim(str, " ")
	if len(str) < 1 {
		return nil, fmt.Errorf("length of data is 0")
	}
	// Split by ","
	strs = strings.Split(str, ",")
	// trim `"`
	for i := 0; i < len(strs); i++ {
		// trim `"`
		strs[i] = strings.Trim(strs[i], "\"")
		strs[i] = strings.Trim(strs[i], " ")
	}

	return
}

func Ints2IntSlice(str string) (ints []int, err error) {
	// trim blank
	str = strings.Trim(str, " ")
	// trim "[" "]"
	str = strings.Trim(str, "[")
	str = strings.Trim(str, "]")
	str = strings.Trim(str, " ")
	if len(str) < 1 {
		return nil, fmt.Errorf("length of data is 0")
	}
	// Split by ","
	strs := strings.Split(str, ",")

	for i := 0; i < len(strs); i++ {
		// trim `"`
		strs[i] = strings.Trim(strs[i], " ")
		// Atoi
		v, err := strconv.Atoi(strs[i])
		if err != nil {
			return ints, err
		}
		ints = append(ints, v)
	}

	return
}
