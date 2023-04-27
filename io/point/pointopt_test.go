// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"testing"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestPointOptions(t *testing.T) {
	cases := []struct {
		name string

		electionable bool
		election     bool

		envTags, hostTags map[string]string

		t map[string]string
		f map[string]interface{}
	}{
		{
			name: "no-electionable-pts",
			f:    map[string]interface{}{"f1": 1},

			envTags: map[string]string{
				"env_t1": "env_val1",
			},

			hostTags: map[string]string{
				"host_t1": "host_val1",
			},
		},

		{
			name: "no-electionable-pts-with-election-on",
			f:    map[string]interface{}{"f1": 1},

			election: true,
			envTags: map[string]string{
				"env_t1": "env_val1",
			},

			hostTags: map[string]string{
				"host_t1": "host_val1",
			},
		},

		{
			name: "env-pts-with-election-off",
			f:    map[string]interface{}{"f1": 1},

			election:     false,
			electionable: true,

			envTags: map[string]string{
				"env_t1": "env_val1",
			},

			hostTags: map[string]string{
				"host_t1": "host_val1",
			},
		},

		{
			name: "env-pts-with-election-on",
			f:    map[string]interface{}{"f1": 1},

			election:     true,
			electionable: true,

			envTags: map[string]string{
				"env_t1": "env_val1",
			},

			hostTags: map[string]string{
				"host_t1": "host_val1",
			},
		},
	}

	for _, tc := range cases {
		EnableElection = tc.election
		ClearGlobalTags()

		for _, cat := range []string{
			datakit.MetricDeprecated,
			datakit.Metric,
			datakit.Network,
			datakit.Object,
			datakit.Logging,
			datakit.Profiling,
		} {
			t.Run(tc.name+cat, func(t *testing.T) {
				for k, v := range tc.envTags {
					SetGlobalElectionTags(k, v)
				}

				for k, v := range tc.hostTags {
					SetGlobalHostTags(k, v)
				}

				var pt *Point
				var err error
				var opt *PointOption

				switch cat {
				case datakit.MetricDeprecated:
					if tc.electionable {
						opt = MOptElection()
					} else {
						opt = MOpt()
					}
				case datakit.Metric:
					if tc.electionable {
						opt = MOptElection()
					} else {
						opt = MOpt()
					}
				case datakit.Network:
					if tc.electionable {
						opt = NOptElection()
					} else {
						opt = NOpt()
					}
				case datakit.Object:
					if tc.electionable {
						opt = OOptElection()
					} else {
						opt = OOpt()
					}
				case datakit.Logging:
					if tc.electionable {
						opt = LOptElection()
					} else {
						opt = LOpt()
					}
				case datakit.Profiling:
					if tc.electionable {
						opt = POptElection()
					} else {
						opt = POpt()
					}
				}

				pt, err = NewPoint("test", tc.t, tc.f, opt)

				if err != nil {
					t.Error(err)
					return
				}

				jp, err := pt.ToJSON()
				if err != nil {
					t.Error(err)
					return
				}

				if tc.electionable {
					if tc.election { // election on: only global-env-tags
						for k, v := range tc.envTags {
							if _, ok := jp.Tags[k]; !ok {
								tu.Assert(t, ok, "expect tag %s=%s", k, v)
							}
						}
					} else { // election off: only global-host-tags
						for k, v := range tc.hostTags {
							if _, ok := jp.Tags[k]; !ok {
								tu.Assert(t, ok, "expect tag %s=%s", k, v)
							}
						}
					}
				} else { // not electionable: only global-host-tags
					for k, v := range tc.hostTags {
						if _, ok := jp.Tags[k]; !ok {
							tu.Assert(t, ok, "expect tag %s=%s", k, v)
						}
					}
				}

				t.Log(pt.String())
			})
		}
	}
}
