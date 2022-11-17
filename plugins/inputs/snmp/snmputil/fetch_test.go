// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
)

func Test_fetchColumnOids(t *testing.T) {
	sess := CreateMockSession()

	bulkPacket := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1",
				Type:  gosnmp.TimeTicks,
				Value: 11,
			},
			{
				Name:  "1.1.2.1",
				Type:  gosnmp.TimeTicks,
				Value: 21,
			},
			{
				Name:  "1.1.1.2",
				Type:  gosnmp.TimeTicks,
				Value: 12,
			},
			{
				Name:  "1.1.2.2",
				Type:  gosnmp.TimeTicks,
				Value: 22,
			},
			{
				Name:  "1.1.1.3",
				Type:  gosnmp.TimeTicks,
				Value: 13,
			},
			{
				Name:  "1.1.3.1",
				Type:  gosnmp.TimeTicks,
				Value: 31,
			},
		},
	}
	bulkPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.4",
				Type:  gosnmp.TimeTicks,
				Value: 14,
			},
			{
				Name:  "1.1.1.5",
				Type:  gosnmp.TimeTicks,
				Value: 15,
			},
		},
	}
	bulkPacket3 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.3.1",
				Type:  gosnmp.TimeTicks,
				Value: 34,
			},
		},
	}
	sess.On("GetBulk", []string{"1.1.1", "1.1.2"}, DefaultBulkMaxRepetitions).Return(&bulkPacket, nil)
	sess.On("GetBulk", []string{"1.1.1.3"}, DefaultBulkMaxRepetitions).Return(&bulkPacket2, nil)
	sess.On("GetBulk", []string{"1.1.1.5"}, DefaultBulkMaxRepetitions).Return(&bulkPacket3, nil)

	oids := map[string]string{"1.1.1": "1.1.1", "1.1.2": "1.1.2"}

	columnValues, err := fetchColumnOidsWithBatching(sess, oids, 100, DefaultBulkMaxRepetitions, useGetBulk)
	assert.Nil(t, err)

	expectedColumnValues := ColumnResultValuesType{
		"1.1.1": {
			"1": ResultValue{Value: float64(11)},
			"2": ResultValue{Value: float64(12)},
			"3": ResultValue{Value: float64(13)},
			"4": ResultValue{Value: float64(14)},
			"5": ResultValue{Value: float64(15)},
		},
		"1.1.2": {
			"1": ResultValue{Value: float64(21)},
			"2": ResultValue{Value: float64(22)},
		},
	}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchColumnOidsBatch_usingGetBulk(t *testing.T) {
	sess := CreateMockSession()

	bulkPacket := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1",
				Type:  gosnmp.TimeTicks,
				Value: 11,
			},
			{
				Name:  "1.1.2.1",
				Type:  gosnmp.TimeTicks,
				Value: 21,
			},
			{
				Name:  "1.1.1.2",
				Type:  gosnmp.TimeTicks,
				Value: 12,
			},
			{
				Name:  "1.1.2.2",
				Type:  gosnmp.TimeTicks,
				Value: 22,
			},
			{
				Name:  "1.1.1.3",
				Type:  gosnmp.TimeTicks,
				Value: 13,
			},
			{
				Name:  "1.1.9.1",
				Type:  gosnmp.TimeTicks,
				Value: 31,
			},
		},
	}

	bulkPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.4",
				Type:  gosnmp.TimeTicks,
				Value: 14,
			},
			{
				Name:  "1.1.1.5",
				Type:  gosnmp.TimeTicks,
				Value: 15,
			},
		},
	}
	bulkPacket3 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.3.1",
				Type:  gosnmp.TimeTicks,
				Value: 34,
			},
		},
	}
	// First bulk iteration with two batches with batch size 2
	sess.On("GetBulk", []string{"1.1.1", "1.1.2"}, DefaultBulkMaxRepetitions).Return(&bulkPacket, nil)

	// Second bulk iteration
	sess.On("GetBulk", []string{"1.1.1.3"}, DefaultBulkMaxRepetitions).Return(&bulkPacket2, nil)

	// Third bulk iteration
	sess.On("GetBulk", []string{"1.1.1.5"}, DefaultBulkMaxRepetitions).Return(&bulkPacket3, nil)

	oids := map[string]string{"1.1.1": "1.1.1", "1.1.2": "1.1.2"}

	columnValues, err := fetchColumnOidsWithBatching(sess, oids, 2, 10, useGetBulk)
	assert.Nil(t, err)

	expectedColumnValues := ColumnResultValuesType{
		"1.1.1": {
			"1": ResultValue{Value: float64(11)},
			"2": ResultValue{Value: float64(12)},
			"3": ResultValue{Value: float64(13)},
			"4": ResultValue{Value: float64(14)},
			"5": ResultValue{Value: float64(15)},
		},
		"1.1.2": {
			"1": ResultValue{Value: float64(21)},
			"2": ResultValue{Value: float64(22)},
		},
	}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchColumnOidsBatch_usingGetNext(t *testing.T) {
	sess := CreateMockSession()
	sess.Version = gosnmp.Version1

	bulkPacket := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1",
				Type:  gosnmp.TimeTicks,
				Value: 11,
			},
			{
				Name:  "1.1.2.1",
				Type:  gosnmp.TimeTicks,
				Value: 21,
			},
		},
	}

	bulkPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.2",
				Type:  gosnmp.TimeTicks,
				Value: 12,
			},
			{
				Name:  "1.1.9.1",
				Type:  gosnmp.TimeTicks,
				Value: 91,
			},
		},
	}
	bulkPacket3 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.9.2",
				Type:  gosnmp.TimeTicks,
				Value: 91,
			},
		},
	}

	secondBatchPacket1 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.3.1",
				Type:  gosnmp.TimeTicks,
				Value: 31,
			},
		},
	}

	secondBatchPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.9.1",
				Type:  gosnmp.TimeTicks,
				Value: 91,
			},
		},
	}

	// First bulk iteration with two batches with batch size 2
	sess.On("GetNext", []string{"1.1.1", "1.1.2"}).Return(&bulkPacket, nil)

	// Second bulk iteration
	sess.On("GetNext", []string{"1.1.1.1", "1.1.2.1"}).Return(&bulkPacket2, nil)

	// Third bulk iteration
	sess.On("GetNext", []string{"1.1.1.2"}).Return(&bulkPacket3, nil)

	// Second batch
	sess.On("GetNext", []string{"1.1.3"}).Return(&secondBatchPacket1, nil)
	sess.On("GetNext", []string{"1.1.3.1"}).Return(&secondBatchPacket2, nil)

	oids := map[string]string{"1.1.1": "1.1.1", "1.1.2": "1.1.2", "1.1.3": "1.1.3"}

	columnValues, err := fetchColumnOidsWithBatching(sess, oids, 2, 10, useGetBulk)
	assert.Nil(t, err)

	expectedColumnValues := ColumnResultValuesType{
		"1.1.1": {
			"1": ResultValue{Value: float64(11)},
			"2": ResultValue{Value: float64(12)},
		},
		"1.1.2": {
			"1": ResultValue{Value: float64(21)},
		},
		"1.1.3": {
			"1": ResultValue{Value: float64(31)},
		},
	}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchColumnOidsBatch_usingGetBulkAndGetNextFallback(t *testing.T) {
	sess := CreateMockSession()
	// When using snmp v2+, we will try GetBulk first and fallback using GetNext
	sess.Version = gosnmp.Version2c

	bulkPacket := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1",
				Type:  gosnmp.TimeTicks,
				Value: 11,
			},
			{
				Name:  "1.1.2.1",
				Type:  gosnmp.TimeTicks,
				Value: 21,
			},
		},
	}

	bulkPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.2",
				Type:  gosnmp.TimeTicks,
				Value: 12,
			},
			{
				Name:  "1.1.9.1",
				Type:  gosnmp.TimeTicks,
				Value: 91,
			},
		},
	}
	bulkPacket3 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.9.2",
				Type:  gosnmp.TimeTicks,
				Value: 91,
			},
		},
	}

	secondBatchPacket1 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.3.1",
				Type:  gosnmp.TimeTicks,
				Value: 31,
			},
		},
	}

	secondBatchPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.9.1",
				Type:  gosnmp.TimeTicks,
				Value: 91,
			},
		},
	}

	sess.On("GetBulk", []string{"1.1.1", "1.1.2"}, DefaultBulkMaxRepetitions).Return(&gosnmp.SnmpPacket{}, fmt.Errorf("bulk error"))

	// First batch
	sess.On("GetNext", []string{"1.1.1", "1.1.2"}).Return(&bulkPacket, nil)
	sess.On("GetNext", []string{"1.1.1.1", "1.1.2.1"}).Return(&bulkPacket2, nil)
	sess.On("GetNext", []string{"1.1.1.2"}).Return(&bulkPacket3, nil)

	// Second batch
	sess.On("GetNext", []string{"1.1.3"}).Return(&secondBatchPacket1, nil)
	sess.On("GetNext", []string{"1.1.3.1"}).Return(&secondBatchPacket2, nil)

	config := &FetchOpts{
		BulkMaxRepetitions: DefaultBulkMaxRepetitions,
		OidBatchSize:       2,
		OidConfig: OidConfig{
			ColumnOids: []string{"1.1.1", "1.1.2", "1.1.3"},
		},
	}
	columnValues, err := Fetch(sess, config)
	assert.Nil(t, err)

	expectedColumnValues := &ResultValueStore{
		ScalarValues: ScalarResultValuesType{},
		ColumnValues: ColumnResultValuesType{
			"1.1.1": {
				"1": ResultValue{Value: float64(11)},
				"2": ResultValue{Value: float64(12)},
			},
			"1.1.2": {
				"1": ResultValue{Value: float64(21)},
			},
			"1.1.3": {
				"1": ResultValue{Value: float64(31)},
			},
		},
	}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchOidBatchSize(t *testing.T) {
	session := CreateMockSession()

	getPacket1 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1.0",
				Type:  gosnmp.Gauge32,
				Value: 10,
			},
			{
				Name:  "1.1.1.2.0",
				Type:  gosnmp.Gauge32,
				Value: 20,
			},
		},
	}

	getPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.3.0",
				Type:  gosnmp.Gauge32,
				Value: 30,
			},
			{
				Name:  "1.1.1.4.0",
				Type:  gosnmp.Gauge32,
				Value: 40,
			},
		},
	}

	getPacket3 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.5.0",
				Type:  gosnmp.Gauge32,
				Value: 50,
			},
			{
				Name:  "1.1.1.6.0",
				Type:  gosnmp.Gauge32,
				Value: 60,
			},
		},
	}

	session.On("Get", []string{"1.1.1.1.0", "1.1.1.2.0"}).Return(&getPacket1, nil)
	session.On("Get", []string{"1.1.1.3.0", "1.1.1.4.0"}).Return(&getPacket2, nil)
	session.On("Get", []string{"1.1.1.5.0", "1.1.1.6.0"}).Return(&getPacket3, nil)

	oids := []string{"1.1.1.1.0", "1.1.1.2.0", "1.1.1.3.0", "1.1.1.4.0", "1.1.1.5.0", "1.1.1.6.0"}

	columnValues, err := fetchScalarOidsWithBatching(session, oids, 2)
	assert.Nil(t, err)

	expectedColumnValues := ScalarResultValuesType{
		"1.1.1.1.0": {Value: float64(10)},
		"1.1.1.2.0": {Value: float64(20)},
		"1.1.1.3.0": {Value: float64(30)},
		"1.1.1.4.0": {Value: float64(40)},
		"1.1.1.5.0": {Value: float64(50)},
		"1.1.1.6.0": {Value: float64(60)},
	}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchOidBatchSize_zeroSizeError(t *testing.T) {
	sess := CreateMockSession()

	oids := []string{"1.1.1.1.0", "1.1.1.2.0", "1.1.1.3.0", "1.1.1.4.0", "1.1.1.5.0", "1.1.1.6.0"}
	columnValues, err := fetchScalarOidsWithBatching(sess, oids, 0)

	assert.EqualError(t, err, "failed to create oid batches: batch size must be positive. invalid size: 0")
	assert.Nil(t, columnValues)
}

func Test_fetchOidBatchSize_fetchError(t *testing.T) {
	sess := CreateMockSession()

	sess.On("Get", []string{"1.1.1.1.0", "1.1.1.2.0"}).Return(&gosnmp.SnmpPacket{}, fmt.Errorf("my error"))

	oids := []string{"1.1.1.1.0", "1.1.1.2.0", "1.1.1.3.0", "1.1.1.4.0", "1.1.1.5.0", "1.1.1.6.0"}
	columnValues, err := fetchScalarOidsWithBatching(sess, oids, 2)

	assert.EqualError(t, err, "failed to fetch scalar oids: fetch scalar: error getting oids `[1.1.1.1.0 1.1.1.2.0]`: my error")
	assert.Nil(t, columnValues)
}

func Test_fetchScalarOids_retry(t *testing.T) {
	sess := CreateMockSession()

	getPacket := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1.0",
				Type:  gosnmp.Gauge32,
				Value: 10,
			},
			{
				Name:  "1.1.1.2",
				Type:  gosnmp.NoSuchInstance,
				Value: 20,
			},
			{
				Name:  ".1.1.1.3", // test `.` prefix is trimmed
				Type:  gosnmp.NoSuchObject,
				Value: 30,
			},
			{
				Name:  "1.1.1.4.0",
				Type:  gosnmp.NoSuchInstance,
				Value: 40,
			},
		},
	}
	retryGetPacket := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.2.0",
				Type:  gosnmp.Gauge32,
				Value: 20,
			},
			{
				Name:  "1.1.1.3.0",
				Type:  gosnmp.Gauge32,
				Value: 30,
			},
		},
	}

	sess.On("Get", []string{"1.1.1.1.0", "1.1.1.2", "1.1.1.3", "1.1.1.4.0"}).Return(&getPacket, nil)
	sess.On("Get", []string{"1.1.1.2.0", "1.1.1.3.0"}).Return(&retryGetPacket, nil)

	oids := []string{"1.1.1.1.0", "1.1.1.2", "1.1.1.3", "1.1.1.4.0"}

	columnValues, err := fetchScalarOids(sess, oids)
	assert.Nil(t, err)

	expectedColumnValues := ScalarResultValuesType{
		"1.1.1.1.0": {Value: float64(10)},
		"1.1.1.2":   {Value: float64(20)},
		"1.1.1.3":   {Value: float64(30)},
	}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchScalarOids_v1NoSuchName(t *testing.T) {
	sess := CreateMockSession()
	sess.Version = gosnmp.Version1

	getPacket := gosnmp.SnmpPacket{
		Error:      gosnmp.NoSuchName,
		ErrorIndex: 2,
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.1.1.1.0",
				Type: gosnmp.Null,
			},
			{
				Name: "1.1.1.2.0",
				Type: gosnmp.Null,
			},
			{
				Name: "1.1.1.3.0",
				Type: gosnmp.Null,
			},
			{
				Name: "1.1.1.4.0",
				Type: gosnmp.Null,
			},
		},
	}

	getPacket2 := gosnmp.SnmpPacket{
		Error:      gosnmp.NoSuchName,
		ErrorIndex: 3,
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.1.1.1.0",
				Type: gosnmp.Null,
			},
			{
				Name: "1.1.1.3.0",
				Type: gosnmp.Null,
			},
			{
				Name: "1.1.1.4.0",
				Type: gosnmp.Null,
			},
		},
	}

	getPacket3 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1.0",
				Type:  gosnmp.Gauge32,
				Value: 10,
			},
			{
				Name:  "1.1.1.3.0",
				Type:  gosnmp.Gauge32,
				Value: 30,
			},
		},
	}

	sess.On("Get", []string{"1.1.1.1.0", "1.1.1.2", "1.1.1.3", "1.1.1.4.0"}).Return(&getPacket, nil)
	sess.On("Get", []string{"1.1.1.1.0", "1.1.1.3", "1.1.1.4.0"}).Return(&getPacket2, nil)
	sess.On("Get", []string{"1.1.1.1.0", "1.1.1.3"}).Return(&getPacket3, nil)

	oids := []string{"1.1.1.1.0", "1.1.1.2", "1.1.1.3", "1.1.1.4.0"}

	columnValues, err := fetchScalarOids(sess, oids)
	assert.Nil(t, err)

	expectedColumnValues := ScalarResultValuesType{
		"1.1.1.1.0": {Value: float64(10)},
		"1.1.1.3.0": {Value: float64(30)},
	}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchScalarOids_v1NoSuchName_noValidOidsLeft(t *testing.T) {
	sess := CreateMockSession()
	sess.Version = gosnmp.Version1

	getPacket := gosnmp.SnmpPacket{
		Error:      gosnmp.NoSuchName,
		ErrorIndex: 1,
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.1.1.1.0",
				Type: gosnmp.Null,
			},
		},
	}

	sess.On("Get", []string{"1.1.1.1.0"}).Return(&getPacket, nil)

	oids := []string{"1.1.1.1.0"}

	columnValues, err := fetchScalarOids(sess, oids)
	assert.Nil(t, err)

	expectedColumnValues := ScalarResultValuesType{}
	assert.Equal(t, expectedColumnValues, columnValues)
}

func Test_fetchScalarOids_v1NoSuchName_errorIndexTooHigh(t *testing.T) {
	sess := CreateMockSession()
	sess.Version = gosnmp.Version1

	getPacket := gosnmp.SnmpPacket{
		Error:      gosnmp.NoSuchName,
		ErrorIndex: 3,
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.1.1.1.0",
				Type: gosnmp.Null,
			},
			{
				Name: "1.1.1.2.0",
				Type: gosnmp.Null,
			},
		},
	}

	sess.On("Get", []string{"1.1.1.1.0", "1.1.1.2"}).Return(&getPacket, nil)

	oids := []string{"1.1.1.1.0", "1.1.1.2"}

	columnValues, err := fetchScalarOids(sess, oids)
	assert.EqualError(t, err, "invalid ErrorIndex `3` when fetching oids `[1.1.1.1.0 1.1.1.2]`")
	assert.Nil(t, columnValues)
}

func Test_fetchScalarOids_v1NoSuchName_errorIndexTooLow(t *testing.T) {
	sess := CreateMockSession()
	sess.Version = gosnmp.Version1

	getPacket := gosnmp.SnmpPacket{
		Error:      gosnmp.NoSuchName,
		ErrorIndex: 0,
		Variables: []gosnmp.SnmpPDU{
			{
				Name: "1.1.1.1.0",
				Type: gosnmp.Null,
			},
			{
				Name: "1.1.1.2.0",
				Type: gosnmp.Null,
			},
		},
	}

	sess.On("Get", []string{"1.1.1.1.0", "1.1.1.2"}).Return(&getPacket, nil)

	oids := []string{"1.1.1.1.0", "1.1.1.2"}

	columnValues, err := fetchScalarOids(sess, oids)
	assert.EqualError(t, err, "invalid ErrorIndex `0` when fetching oids `[1.1.1.1.0 1.1.1.2]`")
	assert.Nil(t, columnValues)
}

func Test_fetchValues_errors(t *testing.T) {
	tests := []struct {
		name          string
		config        FetchOpts
		bulkPacket    gosnmp.SnmpPacket
		expectedError error
	}{
		{
			name: "invalid batch size",
			config: FetchOpts{
				BulkMaxRepetitions: DefaultBulkMaxRepetitions,
				OidConfig: OidConfig{
					ScalarOids: []string{"1.1", "1.2"},
				},
			},
			expectedError: fmt.Errorf("failed to fetch scalar oids with batching: failed to create oid batches: batch size must be positive. invalid size: 0"),
		},
		{
			name: "get fetch error",
			config: FetchOpts{
				BulkMaxRepetitions: DefaultBulkMaxRepetitions,
				OidBatchSize:       10,
				OidConfig: OidConfig{
					ScalarOids: []string{"1.1", "2.2"},
				},
			},
			expectedError: fmt.Errorf("failed to fetch scalar oids with batching: failed to fetch scalar oids: fetch scalar: error getting oids `[1.1 2.2]`: get error"),
		},
		{
			name: "bulk fetch error",
			config: FetchOpts{
				BulkMaxRepetitions: DefaultBulkMaxRepetitions,
				OidBatchSize:       10,
				OidConfig: OidConfig{
					ScalarOids: []string{},
					ColumnOids: []string{"1.1", "2.2"},
				},
			},
			expectedError: fmt.Errorf("failed to fetch oids with GetNext batching: failed to fetch column oids: fetch column: failed getting oids `[1.1 2.2]` using GetNext: getnext error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := CreateMockSession()
			sess.On("Get", []string{"1.1", "2.2"}).Return(&gosnmp.SnmpPacket{}, fmt.Errorf("get error"))
			sess.On("GetBulk", []string{"1.1", "2.2"}, DefaultBulkMaxRepetitions).Return(&gosnmp.SnmpPacket{}, fmt.Errorf("bulk error"))
			sess.On("GetNext", []string{"1.1", "2.2"}).Return(&gosnmp.SnmpPacket{}, fmt.Errorf("getnext error"))

			_, err := Fetch(sess, &tt.config)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func Test_fetchColumnOids_alreadyProcessed(t *testing.T) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	sess := CreateMockSession()

	bulkPacket := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.1",
				Type:  gosnmp.TimeTicks,
				Value: 11,
			},
			{
				Name:  "1.1.2.1",
				Type:  gosnmp.TimeTicks,
				Value: 21,
			},
			{
				Name:  "1.1.1.2",
				Type:  gosnmp.TimeTicks,
				Value: 12,
			},
			{
				Name:  "1.1.2.2",
				Type:  gosnmp.TimeTicks,
				Value: 22,
			},
			{
				Name:  "1.1.1.3",
				Type:  gosnmp.TimeTicks,
				Value: 13,
			},
			{
				Name:  "1.1.2.3",
				Type:  gosnmp.TimeTicks,
				Value: 23,
			},
		},
	}
	bulkPacket2 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  "1.1.1.4",
				Type:  gosnmp.TimeTicks,
				Value: 14,
			},
			{
				Name:  "1.1.2.4",
				Type:  gosnmp.TimeTicks,
				Value: 24,
			},
			{
				Name:  "1.1.1.5",
				Type:  gosnmp.TimeTicks,
				Value: 15,
			},
			{
				Name:  "1.1.2.5",
				Type:  gosnmp.TimeTicks,
				Value: 25,
			},
		},
	}
	bulkPacket3 := gosnmp.SnmpPacket{
		Variables: []gosnmp.SnmpPDU{
			{
				// this OID is already process, we won't try to fetch it again
				Name:  "1.1.1.4",
				Type:  gosnmp.TimeTicks,
				Value: 14,
			},
			{
				// not processed yet
				Name:  "1.1.2.6",
				Type:  gosnmp.TimeTicks,
				Value: 26,
			},
			{
				// this OID is already process, we won't try to fetch it again
				Name:  "1.1.1.5",
				Type:  gosnmp.TimeTicks,
				Value: 15,
			},
			{
				// this OID is already process, we won't try to fetch it again
				Name:  "1.1.2.5",
				Type:  gosnmp.TimeTicks,
				Value: 25,
			},
		},
	}
	sess.On("GetBulk", []string{"1.1.1", "1.1.2"}, DefaultBulkMaxRepetitions).Return(&bulkPacket, nil)
	sess.On("GetBulk", []string{"1.1.1.3", "1.1.2.3"}, DefaultBulkMaxRepetitions).Return(&bulkPacket2, nil)
	sess.On("GetBulk", []string{"1.1.1.5", "1.1.2.5"}, DefaultBulkMaxRepetitions).Return(&bulkPacket3, nil)

	oids := map[string]string{"1.1.1": "1.1.1", "1.1.2": "1.1.2"}

	columnValues, err := fetchColumnOidsWithBatching(sess, oids, 100, DefaultBulkMaxRepetitions, useGetBulk)
	assert.Nil(t, err)

	expectedColumnValues := ColumnResultValuesType{
		"1.1.1": {
			"1": ResultValue{Value: float64(11)},
			"2": ResultValue{Value: float64(12)},
			"3": ResultValue{Value: float64(13)},
			"4": ResultValue{Value: float64(14)},
			"5": ResultValue{Value: float64(15)},
		},
		"1.1.2": {
			"1": ResultValue{Value: float64(21)},
			"2": ResultValue{Value: float64(22)},
			"3": ResultValue{Value: float64(23)},
			"4": ResultValue{Value: float64(24)},
			"5": ResultValue{Value: float64(25)},
			"6": ResultValue{Value: float64(26)},
		},
	}
	assert.Equal(t, expectedColumnValues, columnValues)

	w.Flush()
	// logs := b.String()
	logs := "[DEBUG] fetchColumnOids: fetch column: request oids (maxRep:10,fetchStrategy:useGetBulk): map[1.1.1:1.1.1 1.1.2:1.1.2][DEBUG] getResults: fetch column: GetBulk results: error=NoError(code:0, idx:0), values=[{\"oid\":\"1.1.1.1\",\"type\":\"TimeTicks\",\"value\":\"11\"},{\"oid\":\"1.1.2.1\",\"type\":\"TimeTicks\",\"value\":\"21\"},{\"oid\":\"1.1.1.2\",\"type\":\"TimeTicks\",\"value\":\"12\"},{\"oid\":\"1.1.2.2\",\"type\":\"TimeTicks\",\"value\":\"22\"},{\"oid\":\"1.1.1.3\",\"type\":\"TimeTicks\",\"value\":\"13\"},{\"oid\":\"1.1.2.3\",\"type\":\"TimeTicks\",\"value\":\"23\"}][DEBUG] fetchColumnOids: fetch column: request oids (maxRep:10,fetchStrategy:useGetBulk): map[1.1.1:1.1.1.3 1.1.2:1.1.2.3][DEBUG] getResults: fetch column: GetBulk results: error=NoError(code:0, idx:0), values=[{\"oid\":\"1.1.1.4\",\"type\":\"TimeTicks\",\"value\":\"14\"},{\"oid\":\"1.1.2.4\",\"type\":\"TimeTicks\",\"value\":\"24\"},{\"oid\":\"1.1.1.5\",\"type\":\"TimeTicks\",\"value\":\"15\"},{\"oid\":\"1.1.2.5\",\"type\":\"TimeTicks\",\"value\":\"25\"}][DEBUG] fetchColumnOids: fetch column: request oids (maxRep:10,fetchStrategy:useGetBulk): map[1.1.1:1.1.1.5 1.1.2:1.1.2.5][DEBUG] getResults: fetch column: GetBulk results: error=NoError(code:0, idx:0), values=[{\"oid\":\"1.1.1.4\",\"type\":\"TimeTicks\",\"value\":\"14\"},{\"oid\":\"1.1.2.6\",\"type\":\"TimeTicks\",\"value\":\"26\"},{\"oid\":\"1.1.1.5\",\"type\":\"TimeTicks\",\"value\":\"15\"},{\"oid\":\"1.1.2.5\",\"type\":\"TimeTicks\",\"value\":\"25\"}][DEBUG] fetchColumnOids: fetch column: request oids (maxRep:10,fetchStrategy:useGetBulk): map[1.1.1:1.1.1.5 1.1.2:1.1.2.5][DEBUG] fetchColumnOids: fetch column: OID already processed: 1.1.1.5[DEBUG] fetchColumnOids: fetch column: OID already processed: 1.1.2.5"
	assert.Nil(t, err)

	assert.Equal(t, 1, strings.Count(logs, "[DEBUG] fetchColumnOids: fetch column: OID already processed: 1.1.1.5"), logs)
	assert.Equal(t, 1, strings.Count(logs, "[DEBUG] fetchColumnOids: fetch column: OID already processed: 1.1.2.5"), logs)
}