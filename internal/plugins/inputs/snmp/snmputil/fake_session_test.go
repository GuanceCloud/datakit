// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFakeSessionImplementation(t *testing.T) {
	sess := CreateFakeSession()

	// Set up test data - simulating a network device
	sess.SetStr("1.3.6.1.2.1.1.1.0", "test-description")
	sess.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.3375.2.1.3.4.1") // F5 sysObjectID
	sess.SetStr("1.3.6.1.2.1.1.5.0", "test-device")
	sess.SetInt("1.3.6.1.2.1.2.2.1.10.1", 1234567890) // ifInOctets.1
	sess.SetInt("1.3.6.1.2.1.2.2.1.10.2", 2345678901) // ifInOctets.2
	sess.SetStr("1.3.6.1.2.1.2.2.1.2.1", "eth0")      // ifDescr.1
	sess.SetStr("1.3.6.1.2.1.2.2.1.2.2", "eth1")      // ifDescr.2

	t.Run("Connect and Close", func(t *testing.T) {
		err := sess.Connect()
		assert.NoError(t, err)
		err = sess.Close()
		assert.NoError(t, err)
	})

	t.Run("GetVersion", func(t *testing.T) {
		assert.Equal(t, gosnmp.Version2c, sess.GetVersion())
	})

	t.Run("Get single OID", func(t *testing.T) {
		result, err := sess.Get([]string{"1.3.6.1.2.1.1.5.0"})
		require.NoError(t, err)
		require.Len(t, result.Variables, 1)
		assert.Equal(t, "1.3.6.1.2.1.1.5.0", result.Variables[0].Name)
		assert.Equal(t, gosnmp.OctetString, result.Variables[0].Type)
		assert.Equal(t, "test-device", string(result.Variables[0].Value.([]byte)))
	})

	t.Run("Get multiple OIDs", func(t *testing.T) {
		result, err := sess.Get([]string{"1.3.6.1.2.1.1.1.0", "1.3.6.1.2.1.1.2.0"})
		require.NoError(t, err)
		require.Len(t, result.Variables, 2)
		assert.Equal(t, "1.3.6.1.2.1.1.1.0", result.Variables[0].Name)
		assert.Equal(t, "1.3.6.1.2.1.1.2.0", result.Variables[1].Name)
		assert.Equal(t, "test-description", string(result.Variables[0].Value.([]byte)))
		assert.Equal(t, "1.3.6.1.4.1.3375.2.1.3.4.1", result.Variables[1].Value.(string))
	})

	t.Run("Get non-existent OID", func(t *testing.T) {
		result, err := sess.Get([]string{"1.3.6.1.2.1.1.999.0"})
		require.NoError(t, err)
		require.Len(t, result.Variables, 1)
		assert.Equal(t, gosnmp.NoSuchObject, result.Variables[0].Type)
		assert.Equal(t, "1.3.6.1.2.1.1.999.0", result.Variables[0].Name)
	})

	t.Run("GetNext", func(t *testing.T) {
		result, err := sess.GetNext([]string{"1.3.6.1.2.1.1.0"})
		require.NoError(t, err)
		require.Len(t, result.Variables, 1)
		// Should return the next OID after 1.3.6.1.2.1.1.0 (which is 1.3.6.1.2.1.1.1.0)
		assert.NotEqual(t, gosnmp.EndOfMibView, result.Variables[0].Type)
		assert.Equal(t, "1.3.6.1.2.1.1.1.0", result.Variables[0].Name)
	})

	t.Run("GetBulk", func(t *testing.T) {
		result, err := sess.GetBulk([]string{"1.3.6.1.2.1.1.0"}, 5)
		require.NoError(t, err)
		require.Len(t, result.Variables, 5)
		// Should return 5 next OIDs
		for i := 0; i < 5; i++ {
			assert.NotEqual(t, gosnmp.EndOfMibView, result.Variables[i].Type)
		}
	})

	t.Run("GetWalkAll", func(t *testing.T) {
		result, err := sess.GetWalkAll("1.3.6.1.2.1.2.2.1.10")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2) // Should have at least 2 interfaces
		// Verify all returned OIDs are under the root
		for _, pdu := range result {
			assert.Contains(t, pdu.Name, "1.3.6.1.2.1.2.2.1.10")
		}
	})

	t.Run("BulkWalk", func(t *testing.T) {
		var walkedPDUs []gosnmp.SnmpPDU
		err := sess.BulkWalk("1.3.6.1.2.1.2.2.1.10", func(pdu gosnmp.SnmpPDU) error {
			walkedPDUs = append(walkedPDUs, pdu)
			return nil
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(walkedPDUs), 2)
	})

	t.Run("Request counts", func(t *testing.T) {
		initialGetCount := sess.GetSnmpGetCount()
		initialBulkCount := sess.GetSnmpGetBulkCount()
		initialNextCount := sess.GetSnmpGetNextCount()

		_, _ = sess.Get([]string{"1.3.6.1.2.1.1.5.0"})
		_, _ = sess.GetBulk([]string{"1.3.6.1.2.1.1.0"}, 2)
		_, _ = sess.GetNext([]string{"1.3.6.1.2.1.1.0"})

		assert.Equal(t, uint32(1), sess.GetSnmpGetCount()-initialGetCount)
		assert.Equal(t, uint32(1), sess.GetSnmpGetBulkCount()-initialBulkCount)
		assert.Equal(t, uint32(1), sess.GetSnmpGetNextCount()-initialNextCount)
	})

	t.Run("Set helper methods", func(t *testing.T) {
		sess2 := CreateFakeSession()
		sess2.SetStr("1.2.3.4.5", "string-value")
		sess2.SetInt("1.2.3.4.6", 123)
		sess2.SetObj("1.2.3.4.7", "1.2.3.4.8")
		sess2.SetTime("1.2.3.4.8", 1000)
		sess2.SetCounter32("1.2.3.4.9", 2000)
		sess2.SetCounter64("1.2.3.4.10", 3000)

		result, err := sess2.Get([]string{"1.2.3.4.5", "1.2.3.4.6", "1.2.3.4.7", "1.2.3.4.8", "1.2.3.4.9", "1.2.3.4.10"})
		require.NoError(t, err)
		require.Len(t, result.Variables, 6)
		assert.Equal(t, "string-value", string(result.Variables[0].Value.([]byte)))
		assert.Equal(t, 123, result.Variables[1].Value.(int))
		assert.Equal(t, "1.2.3.4.8", result.Variables[2].Value.(string))
		assert.Equal(t, uint32(1000), result.Variables[3].Value.(uint32))
		assert.Equal(t, uint32(2000), result.Variables[4].Value.(uint32))
		assert.Equal(t, uint64(3000), result.Variables[5].Value.(uint64))
	})

	t.Run("SetMany", func(t *testing.T) {
		sess3 := CreateFakeSession()
		pdus := []gosnmp.SnmpPDU{
			{Name: "1.1.1.1", Type: gosnmp.OctetString, Value: []byte("value1")},
			{Name: "1.1.1.2", Type: gosnmp.Integer, Value: 42},
		}
		sess3.SetMany(pdus...)

		result, err := sess3.Get([]string{"1.1.1.1", "1.1.1.2"})
		require.NoError(t, err)
		require.Len(t, result.Variables, 2)
		assert.Equal(t, "value1", string(result.Variables[0].Value.([]byte)))
		assert.Equal(t, 42, result.Variables[1].Value.(int))
	})
}
