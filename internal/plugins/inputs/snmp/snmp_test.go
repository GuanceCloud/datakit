// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmp

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^Test_BuildSNMPParams$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_BuildSNMPParams(t *testing.T) {
	cases := []struct {
		name     string
		ipt      *Input
		deviceIP string
		out      *gosnmp.GoSNMP
		err      error
	}{
		{
			name:     "empty",
			ipt:      &Input{},
			deviceIP: "192.168.1.220",
			err:      errors.New("no authentication mechanism specified"),
		},
		{
			name:     "snmp_version_zero",
			ipt:      &Input{V2CommunityString: "v2_password"},
			deviceIP: "192.168.1.220",
			err:      fmt.Errorf("SNMP version not supported: 0"),
		},
		{
			name:     "unsupported authentication protocol",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 1, V3AuthProtocol: "unknown"},
			deviceIP: "192.168.1.220",
			err:      fmt.Errorf("unsupported authentication protocol: unknown"),
		},
		{
			name:     "unsupported authentication protocol",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 1, V3PrivProtocol: "unknown"},
			deviceIP: "192.168.1.220",
			err:      fmt.Errorf("unsupported privacy protocol: unknown"),
		},
		{
			name:     "snmp_version_1",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 1, Port: 161},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version1,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.NoPriv,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "snmp_version_2",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 2, Port: 161},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version2c,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.NoPriv,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "snmp_version_3",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.NoPriv,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "AuthProtocol_MD5",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3AuthProtocol: "md5"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.MD5,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.NoPriv,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "AuthProtocol_SHA",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3AuthProtocol: "sha"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.SHA,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.NoPriv,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "PrivProtocol_DES",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3PrivProtocol: "des"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.DES,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "PrivProtocol_AES",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3PrivProtocol: "aes"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.AES,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "PrivProtocol_AES192",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3PrivProtocol: "aes192"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.AES192,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "PrivProtocol_AES192C",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3PrivProtocol: "aes192c"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.AES192C,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "PrivProtocol_AES256",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3PrivProtocol: "aes256"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.AES256,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "PrivProtocol_AES256C",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3PrivProtocol: "aes256c"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.NoAuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.AES256C,
					PrivacyPassphrase:        "",
				},
			},
		},
		{
			name:     "AuthPriv",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3PrivKey: "V3PrivKey"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.AuthPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "",
					PrivacyProtocol:          gosnmp.NoPriv,
					PrivacyPassphrase:        "V3PrivKey",
				},
			},
		},
		{
			name:     "AuthPriv",
			ipt:      &Input{V2CommunityString: "v2_password", SNMPVersion: 3, Port: 161, V3AuthKey: "V3AuthKey"},
			deviceIP: "192.168.1.220",
			out: &gosnmp.GoSNMP{
				Target:          "192.168.1.220",
				Port:            161,
				Community:       "v2_password",
				Transport:       "udp",
				Version:         gosnmp.Version3,
				Timeout:         time.Duration(defaultTimeout) * time.Second,
				Retries:         defaultRetries,
				SecurityModel:   gosnmp.UserSecurityModel,
				MsgFlags:        gosnmp.AuthNoPriv,
				ContextEngineID: "",
				ContextName:     "",
				SecurityParameters: &gosnmp.UsmSecurityParameters{
					UserName:                 "",
					AuthenticationProtocol:   gosnmp.NoAuth,
					AuthenticationPassphrase: "V3AuthKey",
					PrivacyProtocol:          gosnmp.NoPriv,
					PrivacyPassphrase:        "",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := tc.ipt
			out, err := ipt.BuildSNMPParams(tc.deviceIP)
			assert.Equal(t, tc.out, out)
			assert.Equal(t, tc.err, err)
		})
	}
}
