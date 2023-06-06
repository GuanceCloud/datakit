// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func formatValue(value ResultValue, format string) (ResultValue, error) {
	switch value.Value.(type) { //nolint:gocritic
	case []byte:
		val := value.Value.([]byte)
		switch format {
		case "mac_address":
			// Format mac address from OctetString to IEEE 802.1a canonical format e.g. `82:a5:6e:a5:c8:01`
			value.Value = formatColonSepBytes(val)
		default:
			return ResultValue{}, fmt.Errorf("unknown format `%s` (value type `%T`)", format, value.Value)
		}
	default:
		return ResultValue{}, fmt.Errorf("value type `%T` not supported (format `%s`)", value.Value, format)
	}
	return value, nil
}

func formatColonSepBytes(val []byte) string {
	octetsList := make([]string, 0, 11)
	for _, b := range val {
		octetsList = append(octetsList, hex.EncodeToString([]byte{b}))
	}
	return strings.Join(octetsList, ":")
}
