// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package enrichment contains netflow enrich feature.
package enrichment

// RemapDirection remaps direction from 0 or 1 to respectively ingress or egress.
func RemapDirection(direction uint32) string {
	if direction == 1 {
		return "egress"
	}
	return "ingress"
}
