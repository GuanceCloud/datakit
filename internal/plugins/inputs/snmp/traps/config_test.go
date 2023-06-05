// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"strings"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
)

func TestFullConfig(t *testing.T) {
	config := &TrapsServerOpt{
		Enabled: true,
		Port:    1234,
		Users: []UserV3{
			{
				Username:     "user",
				AuthKey:      "password",
				AuthProtocol: "MD5",
				PrivKey:      "password",
				PrivProtocol: "AES",
			},
		},
		BindHost:         "127.0.0.1",
		CommunityStrings: []string{"public"},
		StopTimeout:      12,
		Namespace:        "foo",
	}
	err := checkDefaultConfig(config, "")
	assert.NoError(t, err)
	assert.Equal(t, uint16(1234), config.Port)
	assert.Equal(t, 12, config.StopTimeout)
	assert.Equal(t, []string{"public"}, config.CommunityStrings)
	assert.Equal(t, "127.0.0.1", config.BindHost)
	assert.Equal(t, "foo", config.Namespace)
	assert.Equal(t, []UserV3{
		{
			Username:     "user",
			AuthKey:      "password",
			AuthProtocol: "MD5",
			PrivKey:      "password",
			PrivProtocol: "AES",
		},
	}, config.Users)

	params, err := BuildSNMPParams(config)
	assert.NoError(t, err)
	assert.Equal(t, uint16(1234), params.Port)
	assert.Equal(t, gosnmp.Version3, params.Version)
	assert.Equal(t, "udp", params.Transport)
	assert.NotNil(t, params.Logger)
	assert.Equal(t, gosnmp.UserSecurityModel, params.SecurityModel)
	assert.Equal(t, &gosnmp.UsmSecurityParameters{
		UserName:                 "user",
		AuthoritativeEngineID:    "\x80\xff\xff\xff\xff\xb8J\x19\x82\xe6\xc1\xe4ip/\xee\xb6ӷ\xd7B",
		AuthenticationProtocol:   gosnmp.MD5,
		AuthenticationPassphrase: "password",
		PrivacyProtocol:          gosnmp.AES,
		PrivacyPassphrase:        "password",
	}, params.SecurityParameters)
}

func TestMinimalConfig(t *testing.T) {
	config := &TrapsServerOpt{Enabled: true}
	err := checkDefaultConfig(config, "")
	assert.NoError(t, err)
	assert.Equal(t, uint16(9162), config.Port)
	assert.Equal(t, 5, config.StopTimeout)
	assert.Equal(t, []string{}, config.CommunityStrings)
	assert.Equal(t, "0.0.0.0", config.BindHost)
	assert.Equal(t, []UserV3{}, config.Users)
	assert.Equal(t, "default", config.Namespace)

	params, err := BuildSNMPParams(config)
	assert.NoError(t, err)
	assert.Equal(t, uint16(9162), params.Port)
	assert.Equal(t, gosnmp.Version2c, params.Version)
	assert.Equal(t, "udp", params.Transport)
	assert.NotNil(t, params.Logger)
	assert.Equal(t, nil, params.SecurityParameters)
}

func TestDefaultUsers(t *testing.T) {
	config := &TrapsServerOpt{
		Enabled:          true,
		CommunityStrings: []string{"public"},
		StopTimeout:      11,
	}
	err := checkDefaultConfig(config, "")
	assert.NoError(t, err)

	assert.Equal(t, 11, config.StopTimeout)
}

func TestBuildAuthoritativeEngineID(t *testing.T) {
	expectedEngineIDs := map[string]string{
		"VeryLongHostnameThatDoesNotFitIntoTheByteArray": "\x80\xff\xff\xff\xff\x67\xb2\x0f\xe4\xdf\x73\x7a\xce\x28\x47\x03\x8f\x57\xe6\x5c\x98",
		"VeryLongHostnameThatIsDifferent":                "\x80\xff\xff\xff\xff\xe7\x21\xcc\xd7\x0b\xe1\x60\xc5\x18\xd7\xde\x17\x86\xb0\x7d\x36",
	}
	for hostname, engineID := range expectedEngineIDs {
		config := &TrapsServerOpt{Enabled: true}
		err := checkDefaultConfig(config, hostname)
		assert.NoError(t, err)
		assert.Equal(t, engineID, config.authoritativeEngineID)
	}
}

func TestNamespaceIsNormalized(t *testing.T) {
	config := &TrapsServerOpt{
		Enabled:   true,
		Namespace: "><\n\r\tfoo",
	}
	err := checkDefaultConfig(config, "")
	assert.NoError(t, err)

	assert.Equal(t, "--foo", config.Namespace)
}

func TestInvalidNamespace(t *testing.T) {
	config := &TrapsServerOpt{Enabled: true, Namespace: strings.Repeat("x", 101)}
	err := checkDefaultConfig(config, "")
	assert.Error(t, err)
}

func TestNamespaceSetGlobally(t *testing.T) {
	config := &TrapsServerOpt{Enabled: true, Namespace: "foo"}
	err := checkDefaultConfig(config, "")
	assert.NoError(t, err)

	assert.Equal(t, "foo", config.Namespace)
}

func TestNamespaceSetBothGloballyAndLocally(t *testing.T) {
	config := &TrapsServerOpt{Enabled: true, Namespace: "bar"}
	err := checkDefaultConfig(config, "")
	assert.NoError(t, err)

	assert.Equal(t, "bar", config.Namespace)
}
