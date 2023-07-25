// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

const (
	defaultTimeout = 5
	defaultRetries = 3
)

//------------------------------------------------------------------------------
// pool

type jobType int

const (
	COLLECT_OBJECT  = 0 //nolint:stylecheck
	COLLECT_METRICS = 1 //nolint:stylecheck
	DISCOVERY       = 2 //nolint:stylecheck
)

type Job struct {
	ID     jobType
	IP     string
	Device *deviceInfo
	Subnet string
}

//------------------------------------------------------------------------------

// BuildSNMPParams returns a valid GoSNMP struct to start making queries.
func (ipt *Input) BuildSNMPParams(deviceIP string) (*gosnmp.GoSNMP, error) {
	if ipt.V2CommunityString == "" && ipt.V3User == "" {
		return nil, errors.New("no authentication mechanism specified")
	}

	var version gosnmp.SnmpVersion
	switch ipt.SNMPVersion {
	case 1:
		version = gosnmp.Version1
	case 2:
		version = gosnmp.Version2c
	case 3:
		version = gosnmp.Version3
	default:
		return nil, fmt.Errorf("SNMP version not supported: %d", ipt.SNMPVersion)
	}

	var authProtocol gosnmp.SnmpV3AuthProtocol
	lowerAuthProtocol := strings.ToLower(ipt.V3AuthProtocol)
	switch lowerAuthProtocol {
	case "":
		authProtocol = gosnmp.NoAuth
	case "md5":
		authProtocol = gosnmp.MD5
	case "sha":
		authProtocol = gosnmp.SHA
	default:
		return nil, fmt.Errorf("unsupported authentication protocol: %s", ipt.V3AuthProtocol)
	}

	var privProtocol gosnmp.SnmpV3PrivProtocol
	lowerPrivProtocol := strings.ToLower(ipt.V3PrivProtocol)
	switch lowerPrivProtocol {
	case "":
		privProtocol = gosnmp.NoPriv
	case "des":
		privProtocol = gosnmp.DES
	case "aes", "aes128":
		privProtocol = gosnmp.AES
	case "aes192":
		privProtocol = gosnmp.AES192
	case "aes192c":
		privProtocol = gosnmp.AES192C
	case "aes256":
		privProtocol = gosnmp.AES256
	case "aes256c":
		privProtocol = gosnmp.AES256C
	default:
		return nil, fmt.Errorf("unsupported privacy protocol: %s", ipt.V3PrivProtocol)
	}

	msgFlags := gosnmp.NoAuthNoPriv
	if ipt.V3PrivKey != "" {
		msgFlags = gosnmp.AuthPriv
	} else if ipt.V3AuthKey != "" {
		msgFlags = gosnmp.AuthNoPriv
	}

	return &gosnmp.GoSNMP{
		Target:          deviceIP,
		Port:            ipt.Port,
		Community:       ipt.V2CommunityString,
		Transport:       "udp",
		Version:         version,
		Timeout:         time.Duration(defaultTimeout) * time.Second,
		Retries:         defaultRetries,
		SecurityModel:   gosnmp.UserSecurityModel,
		MsgFlags:        msgFlags,
		ContextEngineID: ipt.V3ContextEngineID,
		ContextName:     ipt.V3ContextName,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName:                 ipt.V3User,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: ipt.V3AuthKey,
			PrivacyProtocol:          privProtocol,
			PrivacyPassphrase:        ipt.V3PrivKey,
		},
	}, nil
}
