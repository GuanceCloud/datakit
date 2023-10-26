// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package solr

import (
	"testing"
)

func TestUrl(t *testing.T) {}

func TestInstanceName(t *testing.T) {
	serverWResultExpect := map[string]string{
		"http://0.0.0.0:123456":    "0.0.0.0_12345",
		"https://127.0.0.1:8983":   "127.0.0.1_8983",
		"http://localhost:8983/":   "localhost_8983",
		"https://golang.org:12345": "golang.org_12345",
		"https://[::]:12345/":      "[::]_12345",
		"https://1.1":              "1.1", // 视为域名
		"golang.org":               "",
		"http://[a:b":              "",
	}
	for k, v := range serverWResultExpect {
		if m, err := instanceName(k); err != nil {
			t.Error(err)
		} else if m != v {
			t.Errorf("expect: %s  actual: %s", v, m)
		}
	}
}
