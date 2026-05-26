//go:build linux
// +build linux

package dnsflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFlowAgg_QueryMeta(t *testing.T) {
	var agg FlowAgg

	err := agg.Append(DNSQAKey{
		IsUDP:      true,
		IsV4:       true,
		ClientPort: 43210,
		ServerPort: 53,
		ClientIP:   [4]uint32{0, 0, 0, 0x0a00000a},
		ServerIP:   [4]uint32{0, 0, 0, 0x08080808},
	}, DNSStats{
		RCODE:       0,
		RespTime:    5 * time.Millisecond,
		QueryDomain: "example.com",
		QueryType:   "A",
	})
	require.NoError(t, err)
	require.Len(t, agg.data, 1)

	for key := range agg.data {
		require.Equal(t, "example.com", key.queryDomain)
		require.Equal(t, "A", key.queryType)
	}

	pts := agg.ToPoint(nil, nil)
	require.Len(t, pts, 1)

	fields := pts[0].InfluxFields()
	require.Equal(t, "example.com", fields["query_domain"])
	require.Equal(t, "A", pts[0].GetTag("query_type"))
}
