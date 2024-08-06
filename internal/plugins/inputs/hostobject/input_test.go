// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// go test -v -timeout 30s -run ^TestCollect$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/hostobject
func TestCollect(t *testing.T) {
	ipt := defaultInput()

	// ipt.OnlyPhysicalDevice = true

	datakit.IsTestMode = true
	if err := ipt.collect(); err != nil {
		t.Error(err)
	}

	var pts []*point.Point
	pt := ipt.collectCache[0]
	pts = append(pts, pt)

	mpts := make(map[string][]*point.Point)
	mpts[datakit.Object] = pts

	if len(mpts) == 0 {
		t.Error("collect empty!")
		return
	}

	for category := range mpts {
		if category != "/v1/write/object" {
			t.Errorf("category not object: %s", category)
			return
		}
	}

	t.Log("TestCollect succeeded!")
}

func TestInput_setup(t *testing.T) {
	tests := []struct {
		name         string
		conf         string
		wantElection bool
		wantHost     bool
	}{
		{
			name:         "nil is true",
			conf:         ``,
			wantElection: true,
			wantHost:     true,
		},
		{
			name: "real = false and deprecated = nil",
			conf: `
			enable_cloud_host_tags_as_global_election_tags = false
            enable_cloud_host_tags_as_global_host_tags = false
			`,
			wantElection: false,
			wantHost:     false,
		},
		{
			name: "real = nil and deprecated = false",
			conf: `
			enable_cloud_host_tags_global_election = false
            enable_cloud_host_tags_global_host = false
			`,
			wantElection: false,
			wantHost:     false,
		},
		{
			name: "real = false and deprecated = false",
			conf: `
			enable_cloud_host_tags_as_global_election_tags = false
            enable_cloud_host_tags_as_global_host_tags = false
			enable_cloud_host_tags_global_election = false
            enable_cloud_host_tags_global_host = false
			`,
			wantElection: false,
			wantHost:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := defaultInput()

			_, err := toml.Decode(tt.conf, ipt)
			assert.NoError(t, err)

			ipt.setup()

			assert.Equal(t, tt.wantElection, ipt.EnableCloudHostTagsGlobalElection)
			assert.Equal(t, tt.wantHost, ipt.EnableCloudHostTagsGlobalHost)
		})
	}
}
