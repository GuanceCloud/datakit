// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"github.com/gosnmp/gosnmp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
)

// BuildSNMPParams returns a valid GoSNMP params structure from configuration.
func BuildSNMPParams(c *TrapsServerOpt) (*gosnmp.GoSNMP, error) {
	if len(c.Users) == 0 {
		return &gosnmp.GoSNMP{
			Port:      c.Port,
			Transport: "udp",
			Version:   gosnmp.Version2c, // No user configured, let's use Version2 which is enough and doesn't require setting up fake security data.
			Logger:    gosnmp.NewLogger(&trapLogger{}),
		}, nil
	}
	user := c.Users[0]
	authProtocol, err := snmputil.GetAuthProtocol(user.AuthProtocol)
	if err != nil {
		return nil, err
	}

	privProtocol, err := snmputil.GetPrivProtocol(user.PrivProtocol)
	if err != nil {
		return nil, err
	}

	msgFlags := gosnmp.NoAuthNoPriv
	if user.PrivKey != "" {
		msgFlags = gosnmp.AuthPriv
	} else if user.AuthKey != "" {
		msgFlags = gosnmp.AuthNoPriv
	}

	return &gosnmp.GoSNMP{
		Port:          c.Port,
		Transport:     "udp",
		Version:       gosnmp.Version3, // Always using version3 for traps, only option that works with all SNMP versions simultaneously
		SecurityModel: gosnmp.UserSecurityModel,
		MsgFlags:      msgFlags,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName:                 user.Username,
			AuthoritativeEngineID:    c.authoritativeEngineID,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: user.AuthKey,
			PrivacyProtocol:          privProtocol,
			PrivacyPassphrase:        user.PrivKey,
		},
		Logger: gosnmp.NewLogger(&trapLogger{}),
	}, nil
}
