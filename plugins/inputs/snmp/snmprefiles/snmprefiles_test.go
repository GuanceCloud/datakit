// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmprefiles

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

// Check whether on development machine, exit immediately when it it not.
// Development machine has environment variable LOCAL_UNIT_TEST.
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

// go test -v -timeout 30s -run ^TestReleaseFiles$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmprefiles
func TestReleaseFiles(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name  string
		confd string
		err   error
	}{
		{
			name:  "normal",
			confd: "/tmp/gotest/confd",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			saved := datakit.ConfdDir

			datakit.ConfdDir = tc.confd
			err := ReleaseFiles()
			assert.Equal(t, tc.err, err)

			datakit.ConfdDir = saved
		})
	}
}
