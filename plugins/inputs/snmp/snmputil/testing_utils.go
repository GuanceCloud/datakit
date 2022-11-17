// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/mock"
)

// copyProfileDefinition copies a profile, it's used for testing.
func copyProfileDefinition(profileDef ProfileDefinition) ProfileDefinition {
	newDef := ProfileDefinition{}
	newDef.Metrics = append(newDef.Metrics, profileDef.Metrics...)
	newDef.MetricTags = append(newDef.MetricTags, profileDef.MetricTags...)
	newDef.StaticTags = append(newDef.StaticTags, profileDef.StaticTags...)
	newDef.Metadata = make(MetadataConfig)
	newDef.Device = profileDef.Device
	newDef.Extends = append(newDef.Extends, profileDef.Extends...)
	newDef.SysObjectIds = append(newDef.SysObjectIds, profileDef.SysObjectIds...)

	for resName, resource := range profileDef.Metadata {
		resConfig := MetadataResourceConfig{}
		resConfig.Fields = make(map[string]MetadataField)
		for fieldName, field := range resource.Fields {
			resConfig.Fields[fieldName] = field
		}
		resConfig.IDTags = append(resConfig.IDTags, resource.IDTags...)
		newDef.Metadata[resName] = resConfig
	}
	return newDef
}

// MockTimeNow mocks time.Now.
var MockTimeNow = func() time.Time {
	layout := "2006-01-02 15:04:05"
	str := "2000-01-01 00:00:00"
	t, _ := time.Parse(layout, str)
	return t
}

// MockSession mocks a connection session.
type MockSession struct {
	mock.Mock
	ConnectErr error
	CloseErr   error
	Version    gosnmp.SnmpVersion
}

// Connect is used to create a new connection.
func (s *MockSession) Connect() error {
	return s.ConnectErr
}

// Close is used to close the connection.
func (s *MockSession) Close() error {
	return s.CloseErr
}

// Get will send a SNMPGET command.
func (s *MockSession) Get(oids []string) (result *gosnmp.SnmpPacket, err error) {
	args := s.Mock.Called(oids)
	return args.Get(0).(*gosnmp.SnmpPacket), args.Error(1)
}

// GetBulk will send a SNMP BULKGET command.
func (s *MockSession) GetBulk(oids []string, bulkMaxRepetitions uint32) (result *gosnmp.SnmpPacket, err error) {
	args := s.Mock.Called(oids, bulkMaxRepetitions)
	return args.Get(0).(*gosnmp.SnmpPacket), args.Error(1)
}

// GetNext will send a SNMP GETNEXT command.
func (s *MockSession) GetNext(oids []string) (result *gosnmp.SnmpPacket, err error) {
	args := s.Mock.Called(oids)
	return args.Get(0).(*gosnmp.SnmpPacket), args.Error(1)
}

// GetVersion returns the snmp version used.
func (s *MockSession) GetVersion() gosnmp.SnmpVersion {
	return s.Version
}

// CreateMockSession creates a mock session.
func CreateMockSession() *MockSession {
	session := &MockSession{}
	session.Version = gosnmp.Version2c
	return session
}

// NewMockSession creates a mock session.
func NewMockSession() (Session, error) {
	return CreateMockSession(), nil
}
