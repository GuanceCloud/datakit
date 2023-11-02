// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupCDNName(t *testing.T) {
	type testCase struct {
		name     string
		domain   string
		throwErr bool
		cName    string
		cdnName  string
	}

	cases := []*testCase{
		{
			name:     "xxxxx.xxxxxx",
			domain:   "xxxxx.xxxxxx",
			throwErr: true,
			cdnName:  "",
		},
		{
			name:     "www-static.qbox.me",
			domain:   "www-static.qbox.me",
			throwErr: false,
			cdnName:  "白山云",
		},
		{
			name:     "tbcache.com",
			domain:   "tbcache.com",
			throwErr: false,
			cName:    "",
			cdnName:  "阿里云",
		},
		{
			name:     "go.dev",
			domain:   "go.dev",
			throwErr: true,
			cName:    "",
			cdnName:  "",
		},
		{
			name:     "ucc.alicdn.com",
			domain:   "ucc.alicdn.com",
			throwErr: false,
			cName:    "",
			cdnName:  "阿里云",
		},
		{
			name:     "static.guance.com",
			domain:   "static.guance.com",
			throwErr: false,
			cName:    "",
			cdnName:  "阿里云",
		},
		{
			name:     "res.vmallres.com",
			domain:   "res.vmallres.com",
			throwErr: false,
			cName:    "",
			cdnName:  "华为云",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, cdnName, err := lookupCDNName(c.domain)
			if c.throwErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			if c.cdnName != "" {
				assert.True(t, strings.Contains(cdnName, c.cdnName))
			}
		})
	}
}
