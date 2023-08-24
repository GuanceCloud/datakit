// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package enrichment

import (
	"encoding/binary"
	"net"
)

// FormatMacAddress formats mac address from uint64 to "xx:xx:xx:xx:xx:xx" format.
func FormatMacAddress(fieldValue uint64) string {
	mac := make([]byte, 8)
	binary.BigEndian.PutUint64(mac, fieldValue)
	return net.HardwareAddr(mac[2:]).String()
}
