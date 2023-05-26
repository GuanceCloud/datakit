// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"bytes"
	"fmt"
)

const (
	deviceHostnamePrefix = "device:"
)

// GetDeviceHostname returns DeviceID as hostname.
func GetDeviceHostname(deviceID string) (string, error) {
	hostname := deviceHostnamePrefix + deviceID
	normalizedHostname, err := NormalizeHost(hostname)
	if err != nil {
		return "", err
	}
	return normalizedHostname, nil
}

// NormalizeHost applies a liberal policy on host names.
func NormalizeHost(host string) (string, error) {
	var buf bytes.Buffer

	// hosts longer than 253 characters are illegal
	if len(host) > 253 {
		return "", fmt.Errorf("hostname is too long, should contain less than 253 characters")
	}

	for _, r := range host {
		switch r {
		// has null rune just toss the whole thing
		case '\x00':
			return "", fmt.Errorf("hostname cannot contain null character")
		// drop these characters entirely
		case '\n', '\r', '\t':
			continue
		// replace characters that are generally used for xss with '-'
		case '>', '<':
			buf.WriteByte('-')
		default:
			buf.WriteRune(r)
		}
	}

	return buf.String(), nil
}
