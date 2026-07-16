// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type staticFormatter struct {
	payload []byte
}

func (f staticFormatter) FormatPacket(*SnmpPacket) ([]byte, error) {
	return f.payload, nil
}

func TestTrapForwarderUsesConfiguredSource(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name: "default source",
			want: defaultTrapSource,
		},
		{
			name:   "custom source",
			source: "snmp_traps_env_a",
			want:   "snmp_traps_env_a",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feeder := dkio.NewMockedFeeder()
			forwarder, err := NewTrapForwarder(
				staticFormatter{payload: []byte(`{"trap":{"snmpTrapName":"linkDown"}}`)},
				make(PacketsChannel, 1),
				&TrapsServerOpt{
					Source: tc.source,
					Feeder: feeder,
					Tagger: datakit.DefaultGlobalTagger(),
				},
			)
			require.NoError(t, err)

			forwarder.sendTrap(&SnmpPacket{
				Addr: &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 12345},
			})

			pts, err := feeder.AnyPoints(time.Second)
			require.NoError(t, err)
			require.Len(t, pts, 1)
			assert.Equal(t, tc.want, pts[0].Name())
		})
	}
}
