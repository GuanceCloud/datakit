// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"errors"

	"github.com/gosnmp/gosnmp"
)

func validatePacket(p *gosnmp.SnmpPacket, opt *TrapsServerOpt) error {
	if p.Version == gosnmp.Version3 {
		// v3 Packets are already decrypted and validated by gosnmp
		return nil
	}

	// At least one of the known community strings must match.
	for _, community := range opt.CommunityStrings {
		if community == p.Community {
			return nil
		}
	}

	return errors.New("unknown community string")
}
