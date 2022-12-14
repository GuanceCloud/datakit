// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

const (
	defaultPort        = uint16(9162) // Standard UDP port for traps.
	defaultStopTimeout = 5
	packetsChanSize    = 100
	genericTrapOid     = "1.3.6.1.6.3.1.1.5"
)
