// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
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

func (s *MockSession) GetWalkAll(oid string) (result []gosnmp.SnmpPDU, err error) {
	args := s.Mock.Called(oid)
	return args.Get(0).([]gosnmp.SnmpPDU), args.Error(1)
}

// GetVersion returns the snmp version used.
func (s *MockSession) GetVersion() gosnmp.SnmpVersion {
	return s.Version
}

// BulkWalk will send a SNMP BulkWalk command (mock implementation).
func (s *MockSession) BulkWalk(rootOid string, walkFn gosnmp.WalkFunc) error {
	args := s.Mock.Called(rootOid, walkFn)
	return args.Error(0)
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

// FakeSession implements Session wrapping around a fixed set of PDUs.
// It's similar to Datadog's FakeSession but adapted for Datakit.
// Caveats:
//   - Fetching an object that isn't there will always return NoSuchObject,
//     never NoSuchInstance.
type FakeSession struct {
	data             map[string]gosnmp.SnmpPDU
	oids             [][]int // sorted slice of all OIDs in data, stored as []ints
	dirty            bool
	snmpGetCount     *atomic.Uint32
	snmpGetBulkCount *atomic.Uint32
	snmpGetNextCount *atomic.Uint32
	Version          gosnmp.SnmpVersion
}

// oidToNumbers parses an OID into a list of numbers.
func oidToNumbers(oid string) ([]int, error) {
	oid = strings.TrimLeft(oid, ".")
	strNumbers := strings.Split(oid, ".")
	var numbers []int
	for _, strNumber := range strNumbers {
		num, err := strconv.Atoi(strNumber)
		if err != nil {
			return nil, fmt.Errorf("error converting digit %s (oid=%s)", strNumber, oid)
		}
		numbers = append(numbers, num)
	}
	return numbers, nil
}

// numbersToOID converts a list of numbers back to an OID.
func numbersToOID(nums []int) string {
	segments := make([]string, len(nums))
	for i, k := range nums {
		segments[i] = strconv.Itoa(k)
	}
	return strings.Join(segments, ".")
}

// cmpOIDs return -1 if a < b, 1 if a > b, and 0 otherwise.
func cmpOIDs(a, b []int) int {
	for i := range a {
		if i >= len(b) {
			return 1
		}
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}
	if len(b) > len(a) {
		return -1
	}
	return 0
}

// CreateFakeSession creates a new FakeSession with an empty set of data.
func CreateFakeSession() *FakeSession {
	return &FakeSession{
		data:             make(map[string]gosnmp.SnmpPDU),
		snmpGetCount:     &atomic.Uint32{},
		snmpGetBulkCount: &atomic.Uint32{},
		snmpGetNextCount: &atomic.Uint32{},
		Version:          gosnmp.Version2c,
	}
}

// Set creates a new PDU with the given attributes, replacing any in the session with the same OID.
func (fs *FakeSession) Set(oid string, typ gosnmp.Asn1BER, val interface{}) *FakeSession {
	if _, ok := fs.data[oid]; !ok {
		fs.dirty = true
	}
	fs.data[oid] = gosnmp.SnmpPDU{
		Name:  oid,
		Type:  typ,
		Value: val,
	}
	return fs
}

// SetMany adds many PDUs to the session at once.
func (fs *FakeSession) SetMany(pdus ...gosnmp.SnmpPDU) {
	fs.dirty = true
	for _, pdu := range pdus {
		fs.data[pdu.Name] = pdu
	}
}

// getOIDs returns a sorted list of all OIDs in fs.data.
func (fs *FakeSession) getOIDs() [][]int {
	if fs.dirty {
		fs.oids = [][]int{}
		for oid := range fs.data {
			nums, err := oidToNumbers(oid)
			if err != nil {
				continue
			}
			fs.oids = append(fs.oids, nums)
		}
		sort.Slice(fs.oids, func(i, j int) bool {
			return cmpOIDs(fs.oids[i], fs.oids[j]) < 0
		})
		fs.dirty = false
	}
	return fs.oids
}

// Connect is a no-op.
func (fs *FakeSession) Connect() error {
	return nil
}

// Close is a no-op.
func (fs *FakeSession) Close() error {
	return nil
}

// GetVersion returns the SNMP version.
func (fs *FakeSession) GetVersion() gosnmp.SnmpVersion {
	return fs.Version
}

// Get gets the values for the given OIDs. OIDs not in the session will return PDUs of type NoSuchObject.
func (fs *FakeSession) Get(oids []string) (result *gosnmp.SnmpPacket, err error) {
	fs.snmpGetCount.Add(1)
	vars := make([]gosnmp.SnmpPDU, len(oids))
	for i, oid := range oids {
		v, ok := fs.data[oid]
		if !ok {
			v = gosnmp.SnmpPDU{
				Name:  oid,
				Type:  gosnmp.NoSuchObject,
				Value: nil,
			}
		}
		vars[i] = v
	}
	return &gosnmp.SnmpPacket{
		Variables: vars,
	}, nil
}

// nextIndex finds the index of the first OID greater than the input.
func (fs *FakeSession) nextIndex(oid []int) int {
	oids := fs.getOIDs()
	return sort.Search(len(oids), func(i int) bool {
		return cmpOIDs(oid, oids[i]) < 0
	})
}

// getNexts returns the items expected by a GetBulk request.
func (fs *FakeSession) getNexts(oids []string, count int) ([]gosnmp.SnmpPDU, error) {
	knownOIDs := fs.getOIDs()
	vars := make([]gosnmp.SnmpPDU, len(oids)*count)
	for i, oid := range oids {
		index := len(knownOIDs)
		nums, err := oidToNumbers(oid)
		if err == nil {
			index = fs.nextIndex(nums)
		}
		for offset := 0; offset < count; offset++ {
			var v gosnmp.SnmpPDU
			if index+offset >= len(knownOIDs) {
				v = gosnmp.SnmpPDU{
					Name:  oid,
					Type:  gosnmp.EndOfMibView,
					Value: nil,
				}
			} else {
				v = fs.data[numbersToOID(knownOIDs[index+offset])]
			}
			vars[offset*len(oids)+i] = v
		}
	}
	return vars, nil
}

// GetBulk returns the `count` next PDUs after each of the given oids.
func (fs *FakeSession) GetBulk(oids []string, count uint32) (*gosnmp.SnmpPacket, error) {
	fs.snmpGetBulkCount.Add(1)
	vars, err := fs.getNexts(oids, int(count))
	if err != nil {
		return nil, err
	}
	return &gosnmp.SnmpPacket{
		Variables: vars,
	}, nil
}

// GetNext returns the first PDU after each of the given OIDs.
func (fs *FakeSession) GetNext(oids []string) (*gosnmp.SnmpPacket, error) {
	fs.snmpGetNextCount.Add(1)
	vars, err := fs.getNexts(oids, 1)
	if err != nil {
		return nil, err
	}
	return &gosnmp.SnmpPacket{
		Variables: vars,
	}, nil
}

// GetWalkAll returns all PDUs for the given root OID.
func (fs *FakeSession) GetWalkAll(rootOid string) ([]gosnmp.SnmpPDU, error) {
	rootNums, err := oidToNumbers(rootOid)
	if err != nil {
		return nil, err
	}
	knownOIDs := fs.getOIDs()
	var result []gosnmp.SnmpPDU
	for _, oidNums := range knownOIDs {
		// Check if this OID is under the root
		if len(oidNums) >= len(rootNums) {
			match := true
			for i := 0; i < len(rootNums); i++ {
				if oidNums[i] != rootNums[i] {
					match = false
					break
				}
			}
			if match {
				result = append(result, fs.data[numbersToOID(oidNums)])
			}
		}
	}
	return result, nil
}

// BulkWalk performs a bulk walk operation.
func (fs *FakeSession) BulkWalk(rootOid string, walkFn gosnmp.WalkFunc) error {
	pdus, err := fs.GetWalkAll(rootOid)
	if err != nil {
		return err
	}
	for _, pdu := range pdus {
		err := walkFn(pdu)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSnmpGetCount returns the number of SNMP GET requests.
func (fs *FakeSession) GetSnmpGetCount() uint32 {
	return fs.snmpGetCount.Load()
}

// GetSnmpGetBulkCount returns the number of SNMP BULKGET requests.
func (fs *FakeSession) GetSnmpGetBulkCount() uint32 {
	return fs.snmpGetBulkCount.Load()
}

// GetSnmpGetNextCount returns the number of SNMP GETNEXT requests.
func (fs *FakeSession) GetSnmpGetNextCount() uint32 {
	return fs.snmpGetNextCount.Load()
}

// SetByte adds an OctetString PDU with the given OID and value.
func (fs *FakeSession) SetByte(oid string, value []byte) *FakeSession {
	return fs.Set(oid, gosnmp.OctetString, value)
}

// SetStr adds an OctetString PDU with the given OID and value.
func (fs *FakeSession) SetStr(oid string, value string) *FakeSession {
	return fs.SetByte(oid, []byte(value))
}

// SetObj adds an ObjectIdentifier PDU with the given OID and value.
func (fs *FakeSession) SetObj(oid string, value string) *FakeSession {
	return fs.Set(oid, gosnmp.ObjectIdentifier, value)
}

// SetTime adds a TimeTicks PDU with the given OID and value.
func (fs *FakeSession) SetTime(oid string, ticks uint32) *FakeSession {
	return fs.Set(oid, gosnmp.TimeTicks, ticks)
}

// SetInt adds an Integer PDU with the given OID and value.
func (fs *FakeSession) SetInt(oid string, value int) *FakeSession {
	return fs.Set(oid, gosnmp.Integer, value)
}

// SetCounter64 adds a Counter64 PDU with the given OID and value.
func (fs *FakeSession) SetCounter64(oid string, value uint64) *FakeSession {
	return fs.Set(oid, gosnmp.Counter64, value)
}

// SetCounter32 adds a Counter32 PDU with the given OID and value.
func (fs *FakeSession) SetCounter32(oid string, value uint32) *FakeSession {
	return fs.Set(oid, gosnmp.Counter32, value)
}

// SetIP adds an IPAddress PDU with the given OID and value.
func (fs *FakeSession) SetIP(oid string, value string) *FakeSession {
	return fs.Set(oid, gosnmp.IPAddress, value)
}
