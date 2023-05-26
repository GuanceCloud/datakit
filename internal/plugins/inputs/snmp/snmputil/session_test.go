// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"io/ioutil"
	stdlog "log"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_snmpSession_Configure(t *testing.T) {
	tests := []struct {
		name                       string
		config                     SessionOpts
		expectedError              error
		expectedVersion            gosnmp.SnmpVersion
		expectedTimeout            time.Duration
		expectedRetries            int
		expectedCommunity          string
		expectedMsgFlags           gosnmp.SnmpV3MsgFlags
		expectedContextName        string
		expectedSecurityParameters gosnmp.SnmpV3SecurityParameters
	}{
		{
			name: "no auth method",
			config: SessionOpts{
				IPAddress: "1.2.3.4",
				Port:      uint16(1234),
			},
			expectedError: fmt.Errorf("an authentication method needs to be provided"),
		},
		{
			name: "valid v1 config",
			config: SessionOpts{
				IPAddress:       "1.2.3.4",
				Port:            uint16(1234),
				SnmpVersion:     uint8(1),
				CommunityString: "abc",
			},
			expectedVersion:   gosnmp.Version1,
			expectedError:     nil,
			expectedTimeout:   time.Duration(4) * time.Second,
			expectedRetries:   3,
			expectedCommunity: "abc",
			expectedMsgFlags:  gosnmp.NoAuthNoPriv,
		},
		{
			name: "valid default v2 config",
			config: SessionOpts{
				IPAddress:       "1.2.3.4",
				Port:            uint16(1234),
				CommunityString: "abc",
			},
			expectedVersion:   gosnmp.Version2c,
			expectedError:     nil,
			expectedTimeout:   time.Duration(4) * time.Second,
			expectedRetries:   3,
			expectedCommunity: "abc",
			expectedMsgFlags:  gosnmp.NoAuthNoPriv,
		},
		{
			name: "valid v2 config",
			config: SessionOpts{
				IPAddress:       "1.2.3.4",
				Port:            uint16(1234),
				CommunityString: "abc",
			},
			expectedVersion:   gosnmp.Version2c,
			expectedError:     nil,
			expectedTimeout:   time.Duration(4) * time.Second,
			expectedRetries:   3,
			expectedCommunity: "abc",
			expectedMsgFlags:  gosnmp.NoAuthNoPriv,
		},
		{
			name: "valid v2c config",
			config: SessionOpts{
				IPAddress:       "1.2.3.4",
				Port:            uint16(1234),
				CommunityString: "abc",
			},
			expectedVersion:   gosnmp.Version2c,
			expectedError:     nil,
			expectedTimeout:   time.Duration(4) * time.Second,
			expectedRetries:   3,
			expectedCommunity: "abc",
			expectedMsgFlags:  gosnmp.NoAuthNoPriv,
		},
		{
			name: "valid v3 AuthPriv config",
			config: SessionOpts{
				IPAddress:    "1.2.3.4",
				Port:         uint16(1234),
				ContextName:  "myContext",
				User:         "myUser",
				AuthKey:      "myAuthKey",
				AuthProtocol: "md5",
				PrivKey:      "myPrivKey",
				PrivProtocol: "aes",
			},
			expectedVersion:     gosnmp.Version3,
			expectedError:       nil,
			expectedTimeout:     time.Duration(4) * time.Second,
			expectedRetries:     3,
			expectedCommunity:   "",
			expectedMsgFlags:    gosnmp.AuthPriv,
			expectedContextName: "myContext",
			expectedSecurityParameters: &gosnmp.UsmSecurityParameters{
				UserName:                 "myUser",
				AuthenticationProtocol:   gosnmp.MD5,
				AuthenticationPassphrase: "myAuthKey",
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        "myPrivKey",
			},
		},
		{
			name: "valid v3 AuthNoPriv config",
			config: SessionOpts{
				IPAddress:    "1.2.3.4",
				Port:         uint16(1234),
				User:         "myUser",
				AuthKey:      "myAuthKey",
				AuthProtocol: "md5",
			},
			expectedVersion:   gosnmp.Version3,
			expectedError:     nil,
			expectedTimeout:   time.Duration(4) * time.Second,
			expectedRetries:   3,
			expectedCommunity: "",
			expectedMsgFlags:  gosnmp.AuthNoPriv,
			expectedSecurityParameters: &gosnmp.UsmSecurityParameters{
				UserName:                 "myUser",
				AuthenticationProtocol:   gosnmp.MD5,
				AuthenticationPassphrase: "myAuthKey",
				PrivacyProtocol:          gosnmp.NoPriv,
				PrivacyPassphrase:        "",
			},
		},
		{
			name: "invalid v3 authProtocol",
			config: SessionOpts{
				IPAddress:    "1.2.3.4",
				Port:         uint16(1234),
				User:         "myUser",
				AuthKey:      "myAuthKey",
				AuthProtocol: "invalid",
			},
			expectedVersion:            gosnmp.Version1, // default, not configured
			expectedError:              fmt.Errorf("unsupported authentication protocol: invalid"),
			expectedSecurityParameters: nil, // default, not configured
		},
		{
			name: "invalid v3 privProtocol",
			config: SessionOpts{
				IPAddress:    "1.2.3.4",
				Port:         uint16(1234),
				User:         "myUser",
				AuthKey:      "myAuthKey",
				AuthProtocol: "md5",
				PrivKey:      "myPrivKey",
				PrivProtocol: "invalid",
			},
			expectedVersion:            gosnmp.Version1, // default, not configured
			expectedError:              fmt.Errorf("unsupported privacy protocol: invalid"),
			expectedSecurityParameters: nil, // default, not configured
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewGosnmpSession(&tt.config)
			assert.Equal(t, tt.expectedError, err)
			if tt.expectedError == nil {
				gosnmpSess := s.(*GosnmpSession)
				assert.Equal(t, tt.expectedVersion, gosnmpSess.gosnmpInst.Version)
				assert.Equal(t, tt.expectedRetries, gosnmpSess.gosnmpInst.Retries)
				assert.Equal(t, tt.expectedTimeout, gosnmpSess.gosnmpInst.Timeout)
				assert.Equal(t, tt.expectedCommunity, gosnmpSess.gosnmpInst.Community)
				assert.Equal(t, tt.expectedContextName, gosnmpSess.gosnmpInst.ContextName)
				assert.Equal(t, tt.expectedMsgFlags, gosnmpSess.gosnmpInst.MsgFlags)
				assert.Equal(t, tt.expectedSecurityParameters, gosnmpSess.gosnmpInst.SecurityParameters)
			}
		})
	}
}

func Test_snmpSession_traceLog_disabled(t *testing.T) {
	config := SessionOpts{
		IPAddress:       "1.2.3.4",
		CommunityString: "abc",
	}

	s, err := NewGosnmpSession(&config)
	gosnmpSess := s.(*GosnmpSession)
	assert.Nil(t, err)
	assert.Equal(t, gosnmp.Logger{}, gosnmpSess.gosnmpInst.Logger)
}

func Test_snmpSession_Connect_Logger(t *testing.T) {
	config := SessionOpts{
		IPAddress:       "1.2.3.4",
		CommunityString: "abc",
	}
	s, err := NewGosnmpSession(&config)
	gosnmpSess := s.(*GosnmpSession)
	require.NoError(t, err)

	logger := gosnmp.NewLogger(stdlog.New(ioutil.Discard, "abc", 0))
	gosnmpSess.gosnmpInst.Logger = logger
	s.Connect()
	assert.Equal(t, logger, gosnmpSess.gosnmpInst.Logger)

	s.Connect()
	assert.Equal(t, logger, gosnmpSess.gosnmpInst.Logger)

	logger2 := gosnmp.NewLogger(stdlog.New(ioutil.Discard, "123", 0))
	gosnmpSess.gosnmpInst.Logger = logger2
	s.Connect()
	assert.NotEqual(t, logger, gosnmpSess.gosnmpInst.Logger)
	assert.Equal(t, logger2, gosnmpSess.gosnmpInst.Logger)
}
