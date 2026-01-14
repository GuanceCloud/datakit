// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"strings"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// fakeSessionWrapper wraps FakeSession to implement the Session interface used by Fetch
type fakeSessionWrapper struct {
	session *snmputil.FakeSession
}

func (f *fakeSessionWrapper) Connect() error {
	return f.session.Connect()
}

func (f *fakeSessionWrapper) Close() error {
	return f.session.Close()
}

func (f *fakeSessionWrapper) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	return f.session.Get(oids)
}

func (f *fakeSessionWrapper) GetBulk(oids []string, bulkMaxRepetitions uint32) (*gosnmp.SnmpPacket, error) {
	return f.session.GetBulk(oids, bulkMaxRepetitions)
}

func (f *fakeSessionWrapper) GetNext(oids []string) (*gosnmp.SnmpPacket, error) {
	return f.session.GetNext(oids)
}

func (f *fakeSessionWrapper) GetWalkAll(rootOid string) ([]gosnmp.SnmpPDU, error) {
	return f.session.GetWalkAll(rootOid)
}

func (f *fakeSessionWrapper) GetVersion() gosnmp.SnmpVersion {
	return f.session.GetVersion()
}

func (f *fakeSessionWrapper) BulkWalk(rootOid string, walkFn gosnmp.WalkFunc) error {
	return f.session.BulkWalk(rootOid, walkFn)
}

// TestProfileIntegration_CyberpowerPDU tests the complete flow for cyberpower-pdu profile
// This tests: profile loading -> SNMP fetching -> metric reporting -> metadata reporting -> final tags/fields
func TestProfileIntegration_CyberpowerPDU(t *testing.T) {
	t.Skip("skipping cyberpower-pdu test")

	// Load the cyberpower-pdu profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"cyberpower-pdu": {
			DefinitionFile: "cyberpower-pdu.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "cyberpower-pdu")

	profileDef := profiles["cyberpower-pdu"]

	// Verify profile loaded correctly
	assert.NotEmpty(t, profileDef.Metrics)
	assert.NotEmpty(t, profileDef.Metadata)
	assert.Equal(t, "cyberpower", profileDef.Metadata["device"].Fields["vendor"].Value)
	assert.Contains(t, profileDef.SysObjectIds, "1.3.6.1.4.1.3808.1.1.*")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "test-cyberpower-pdu")    // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "CyberPower PDU Test")    // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.3808.1.1.1") // sysObjectID
	fakeSession.SetStr("1.3.6.1.2.1.1.6.0", "Data Center A")          // sysLocation

	// Set up global metric tags (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "test-cyberpower-pdu") // sysName for metric tags

	// Set up device metadata metric tags (from cyberpower-pdu.yaml)
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.1.1.0", "PDU-001")              // ePDUIdentName
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.1.5.0", "CP1500AVRLCD")         // ePDUIdentModelNumber
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.1.6.0", "SN123456789")          // ePDUIdentSerialNumber
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.4.1.1.0", "Environment-Sensor-1") // envirIdentName
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.4.1.2.0", "Server Room")          // envirIdentLocation

	// Set up table data for ePDULoadStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.1.1", 1)   // ePDULoadStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.2.1", 500) // cyberpower.ePDULoadStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.3.1", 1)   // ePDULoadStatusLoadState (1=load_normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.6.1", 120) // cyberpower.ePDULoadStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.7.1", 60)  // cyberpower.ePDULoadStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.8.1", 700) // cyberpower.ePDULoadStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.9.1", 80)  // cyberpower.ePDULoadStatusPowerFactor
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.1.2", 2)   // ePDULoadStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.2.2", 750) // cyberpower.ePDULoadStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.3.2", 1)   // ePDULoadStatusLoadState
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.6.2", 120) // cyberpower.ePDULoadStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.7.2", 90)  // cyberpower.ePDULoadStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.8.2", 850) // cyberpower.ePDULoadStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.9.2", 85)  // cyberpower.ePDULoadStatusPowerFactor

	// Set up table data for ePDULoadBankConfigTable (constant_value_one test)
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.1.1", 1) // ePDULoadBankConfigIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.5.1", 1) // ePDULoadBankConfigAlarm (1=no_load_alarm)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.1.2", 2) // ePDULoadBankConfigIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.5.2", 1) // ePDULoadBankConfigAlarm

	// Set up ePDUOutletStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.1.1", 1)          // ePDUOutletStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.2.1", "Outlet-1") // ePDUOutletStatusOutletName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.4.1", 1)          // ePDUOutletStatusOutletState (on)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.7.1", 150)        // cyberpower.ePDUOutletStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.8.1", 100)        // cyberpower.ePDUOutletStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.9.1", 1)          // ePDUOutletStatusAlarm (no_load_alarm)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.1.2", 2)          // ePDUOutletStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.2.2", "Outlet-2") // ePDUOutletStatusOutletName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.4.2", 1)          // ePDUOutletStatusOutletState (on)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.7.2", 200)        // cyberpower.ePDUOutletStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.8.2", 150)        // cyberpower.ePDUOutletStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.9.2", 2)          // ePDUOutletStatusAlarm (under_current_alarm)
	// Index: 3
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.1.3", 3)          // ePDUOutletStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.2.3", "Outlet-3") // ePDUOutletStatusOutletName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.4.3", 2)          // ePDUOutletStatusOutletState (off)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.7.3", 0)          // cyberpower.ePDUOutletStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.8.3", 0)          // cyberpower.ePDUOutletStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.9.3", 1)          // ePDUOutletStatusAlarm (no_load_alarm)

	// Set up ePDUStatusBankTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.1.1", 1) // ePDUStatusBankIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.2.1", 1) // ePDUStatusBankNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.3.1", 1) // ePDUStatusBankState (bank_load_normal)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.1.2", 2) // ePDUStatusBankIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.2.2", 2) // ePDUStatusBankNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.3.2", 3) // ePDUStatusBankState (bank_load_near_overload)

	// Set up ePDUStatusPhaseTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.1.1", 1) // ePDUStatusPhaseIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.2.1", 1) // ePDUStatusPhaseNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.3.1", 1) // ePDUStatusPhaseState (phase_load_normal)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.1.2", 2) // ePDUStatusPhaseIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.2.2", 2) // ePDUStatusPhaseNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.3.2", 3) // ePDUStatusPhaseState (phase_load_near_overload)

	// Set up ePDU2DeviceStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.1.1", 1)          // ePDU2DeviceStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.3.4.1.3.1", "Device-A") // ePDU2DeviceStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.4.1", 1)          // ePDU2DeviceStatusLoadState (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.5.1", 300)        // cyberpower.ePDU2DeviceStatusCurrentLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.6.1", 350)        // cyberpower.ePDU2DeviceStatusCurrentPeakLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.12.1", 1)         // ePDU2DeviceStatusPowerSupplyAlarm (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.13.1", 1)         // ePDU2DeviceStatusPowerSupply1Status (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.14.1", 1)         // ePDU2DeviceStatusPowerSupply2Status (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.15.1", 400)       // cyberpower.ePDU2DeviceStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.16.1", 80)        // cyberpower.ePDU2DeviceStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.17.1", 1)         // ePDU2DeviceStatusRoleType (standalone)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.1.2", 2)          // ePDU2DeviceStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.3.4.1.3.2", "Device-B") // ePDU2DeviceStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.4.2", 3)          // ePDU2DeviceStatusLoadState (near_overload)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.5.2", 500)        // cyberpower.ePDU2DeviceStatusCurrentLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.6.2", 600)        // cyberpower.ePDU2DeviceStatusCurrentPeakLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.12.2", 2)         // ePDU2DeviceStatusPowerSupplyAlarm (alarm)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.13.2", 2)         // ePDU2DeviceStatusPowerSupply1Status (alarm)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.14.2", 1)         // ePDU2DeviceStatusPowerSupply2Status (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.15.2", 700)       // cyberpower.ePDU2DeviceStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.16.2", 85)        // cyberpower.ePDU2DeviceStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.17.2", 2)         // ePDU2DeviceStatusRoleType (host)

	// Set up ePDU2PhaseStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.1.1", 1)     // ePDU2PhaseStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.4.1", 1)     // ePDU2PhaseStatusLoadState (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.5.1", 1000)  // cyberpower.ePDU2PhaseStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.6.1", 230)   // cyberpower.ePDU2PhaseStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.7.1", 800)   // cyberpower.ePDU2PhaseStatusPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.8.1", 1200)  // cyberpower.ePDU2PhaseStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.9.1", 95)    // cyberpower.ePDU2PhaseStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.10.1", 1100) // cyberpower.ePDU2PhaseStatusPeakLoad
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.1.2", 2)     // ePDU2PhaseStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.4.2", 3)     // ePDU2PhaseStatusLoadState (near_overload)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.5.2", 1500)  // cyberpower.ePDU2PhaseStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.6.2", 240)   // cyberpower.ePDU2PhaseStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.7.2", 1200)  // cyberpower.ePDU2PhaseStatusPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.8.2", 1800)  // cyberpower.ePDU2PhaseStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.9.2", 90)    // cyberpower.ePDU2PhaseStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.10.2", 1700) // cyberpower.ePDU2PhaseStatusPeakLoad

	// Set up ePDU2BankStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.1.1", 1)   // ePDU2BankStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.4.1", 1)   // ePDU2BankStatusLoadState (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.5.1", 500) // cyberpower.ePDU2BankStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.6.1", 600) // cyberpower.ePDU2BankStatusPeakLoad
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.1.2", 2)   // ePDU2BankStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.4.2", 3)   // ePDU2BankStatusLoadState (near_overload)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.5.2", 800) // cyberpower.ePDU2BankStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.6.2", 900) // cyberpower.ePDU2BankStatusPeakLoad

	// Set up ePDU2OutletSwitchedStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.1.1", 1)                  // ePDU2OutletSwitchedStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.4.1", "SwitchedOutlet-1") // ePDU2OutletSwitchedStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.5.1", 1)                  // ePDU2OutletSwitchedStatusState (on)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.1.2", 2)                  // ePDU2OutletSwitchedStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.4.2", "SwitchedOutlet-2") // ePDU2OutletSwitchedStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.5.2", 2)                  // ePDU2OutletSwitchedStatusState (off)

	// Set up scalar metrics
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.7", 120)  // cyberpower.ePDUStatusInputVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.8", 60)   // cyberpower.ePDUStatusInputFrequency
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.4.2.1.0", 25) // cyberpower.envirTemperature
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.4.2.6.0", 25) // cyberpower.envirTemperatureCelsius
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.4.3.1.0", 50) // cyberpower.envirHumidity

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "test"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "192.168.1.100",
		Namespace:             "default",
		CollectDeviceMetadata: true,
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("cyberpower-pdu")
	require.NoError(t, err)
	assert.Equal(t, "cyberpower-pdu", deviceInfo.Profile)

	// Fetch SNMP values using FakeSession
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)
	require.NotNil(t, values)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:192.168.1.100", "snmp_profile:cyberpower-pdu", "device_vendor:cyberpower"}
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Verify constant_value_one metrics
	constantMetricFound := false
	for _, metric := range metricData.Data {
		if metric.Name == "cyberpower.ePDULoadBankConfig" {
			constantMetricFound = true
			assert.Equal(t, 1.0, metric.Value, "constant_value_one metric should have value 1.0")
			hasIndexTag := false
			for _, tag := range metric.Tags {
				if strings.HasPrefix(tag, "e_pdu_load_bank_config_index:") {
					hasIndexTag = true
					break
				}
			}
			assert.True(t, hasIndexTag, "should have index tag")
		}
	}
	assert.True(t, constantMetricFound, "should have found constant_value_one metric")

	// Verify table metrics with mapping
	loadStateMappingFound := false
	for _, metric := range metricData.Data {
		if metric.Name == "cyberpower.ePDULoadStatusLoad" {
			loadStateMappingFound = true
			// Verify mapping tag is applied
			hasMappingTag := false
			for _, tag := range metric.Tags {
				if tag == "e_pdu_load_status_load_state:load_normal" {
					hasMappingTag = true
					break
				}
			}
			assert.True(t, hasMappingTag, "should have mapped load_state tag")
		}
	}
	assert.True(t, loadStateMappingFound, "should have found metrics with mapping")

	// Test metadata reporting
	var metaData deviceMetaData
	metaData.collectMeta = true
	deviceInfo.ReportNetworkDeviceMetadata(values, tags, deviceInfo.Metadata, time.Now(), snmputil.DeviceStatusReachable, &metaData)

	// Verify metadata was collected
	assert.NotEmpty(t, metaData.data, "should have collected metadata")

	// Verify device type is set
	// The metadata should contain "type": "PDU" from cyberpower-pdu.yaml
	metadataJSON := metaData.data[0]
	assert.Contains(t, metadataJSON, `"type":"pdu"`, "metadata should contain device type")

	// Test final tag/field generation
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	// Verify object data (with device_meta)
	objectFound := false
	for _, data := range fts.Data {
		if deviceMeta, ok := data.Fields["device_meta"]; ok {
			objectFound = true
			assert.NotEmpty(t, deviceMeta, "device_meta should not be empty")
			// Verify device_meta contains type
			deviceMetaStr, ok := deviceMeta.(string)
			assert.True(t, ok, "device_meta should be string")
			assert.Contains(t, deviceMetaStr, `"type":"pdu"`, "device_meta should contain type field")
		}
	}
	assert.True(t, objectFound, "should have object data with device_meta")

	// Verify metric data has correct tags
	metricFound := false
	for _, data := range fts.Data {
		if allJSON, ok := data.Fields["all"]; ok {
			allJSONStr, ok := allJSON.(string)
			assert.True(t, ok, "all field should be string")
			if strings.Contains(allJSONStr, "cyberpower.ePDULoadStatusLoad") {
				metricFound = true
				// Verify tags
				assert.Contains(t, allJSONStr, `"e_pdu_load_status_index"`, "should have index tag")
				assert.Contains(t, allJSONStr, `"e_pdu_load_status_load_state":"load_normal"`, "should have mapped tag")

				assert.Contains(t, data.Tags, "snmp_profile", "should have snmp_profile tag")
				assert.Equal(t, "cyberpower-pdu", data.Tags["snmp_profile"])
				assert.Contains(t, data.Tags, "device_vendor", "should have device_vendor tag")
				assert.Equal(t, "cyberpower", data.Tags["device_vendor"])
			}
		}
	}
	assert.True(t, metricFound, "should have metric data with correct tags")

	pts := ipt.CollectingMeasurements("192.168.1.100", deviceInfo, true)
	t.Logf("Final points count (Object mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Object): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_CyberpowerPDU_MetricMode tests the complete flow for cyberpower-pdu profile in Metric mode
// This tests: profile loading -> SNMP fetching -> metric reporting -> final tags/fields (Metric mode, no device_meta)
func TestProfileIntegration_CyberpowerPDU_MetricMode(t *testing.T) {
	t.Skip("skipping cyberpower-pdu metric mode test")
	// Load the cyberpower-pdu profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"cyberpower-pdu": {
			DefinitionFile: "cyberpower-pdu.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "cyberpower-pdu")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "test-cyberpower-pdu")    // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "CyberPower PDU Test")    // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.3808.1.1.1") // sysObjectID
	fakeSession.SetStr("1.3.6.1.2.1.1.6.0", "Data Center A")          // sysLocation

	// Set up global metric tags (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "test-cyberpower-pdu") // sysName for metric tags

	// Set up device metadata metric tags (from cyberpower-pdu.yaml)
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.1.1.0", "PDU-001")              // ePDUIdentName
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.1.5.0", "CP1500AVRLCD")         // ePDUIdentModelNumber
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.1.6.0", "SN123456789")          // ePDUIdentSerialNumber
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.4.1.1.0", "Environment-Sensor-1") // envirIdentName
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.4.1.2.0", "Server Room")          // envirIdentLocation

	// Set up table data for ePDULoadStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.1.1", 1)   // ePDULoadStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.2.1", 500) // cyberpower.ePDULoadStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.3.1", 1)   // ePDULoadStatusLoadState (1=load_normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.6.1", 120) // cyberpower.ePDULoadStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.7.1", 60)  // cyberpower.ePDULoadStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.8.1", 700) // cyberpower.ePDULoadStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.9.1", 80)  // cyberpower.ePDULoadStatusPowerFactor
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.1.2", 2)   // ePDULoadStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.2.2", 750) // cyberpower.ePDULoadStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.3.2", 1)   // ePDULoadStatusLoadState
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.6.2", 120) // cyberpower.ePDULoadStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.7.2", 90)  // cyberpower.ePDULoadStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.8.2", 850) // cyberpower.ePDULoadStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.3.1.1.9.2", 85)  // cyberpower.ePDULoadStatusPowerFactor

	// Set up table data for ePDULoadBankConfigTable (constant_value_one test)
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.1.1", 1) // ePDULoadBankConfigIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.5.1", 1) // ePDULoadBankConfigAlarm (1=no_load_alarm)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.1.2", 2) // ePDULoadBankConfigIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.2.4.1.1.5.2", 1) // ePDULoadBankConfigAlarm

	// Set up ePDUOutletStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.1.1", 1)          // ePDUOutletStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.2.1", "Outlet-1") // ePDUOutletStatusOutletName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.4.1", 1)          // ePDUOutletStatusOutletState (on)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.7.1", 150)        // cyberpower.ePDUOutletStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.8.1", 100)        // cyberpower.ePDUOutletStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.9.1", 1)          // ePDUOutletStatusAlarm (no_load_alarm)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.1.2", 2)          // ePDUOutletStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.2.2", "Outlet-2") // ePDUOutletStatusOutletName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.4.2", 1)          // ePDUOutletStatusOutletState (on)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.7.2", 200)        // cyberpower.ePDUOutletStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.8.2", 150)        // cyberpower.ePDUOutletStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.9.2", 2)          // ePDUOutletStatusAlarm (under_current_alarm)
	// Index: 3
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.1.3", 3)          // ePDUOutletStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.2.3", "Outlet-3") // ePDUOutletStatusOutletName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.4.3", 2)          // ePDUOutletStatusOutletState (off)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.7.3", 0)          // cyberpower.ePDUOutletStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.8.3", 0)          // cyberpower.ePDUOutletStatusActivePower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.3.5.1.1.9.3", 1)          // ePDUOutletStatusAlarm (no_load_alarm)

	// Set up ePDUStatusBankTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.1.1", 1) // ePDUStatusBankIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.2.1", 1) // ePDUStatusBankNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.3.1", 1) // ePDUStatusBankState (bank_load_normal)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.1.2", 2) // ePDUStatusBankIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.2.2", 2) // ePDUStatusBankNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.2.1.3.2", 3) // ePDUStatusBankState (bank_load_near_overload)

	// Set up ePDUStatusPhaseTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.1.1", 1) // ePDUStatusPhaseIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.2.1", 1) // ePDUStatusPhaseNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.3.1", 1) // ePDUStatusPhaseState (phase_load_normal)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.1.2", 2) // ePDUStatusPhaseIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.2.2", 2) // ePDUStatusPhaseNumber
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.4.1.3.2", 3) // ePDUStatusPhaseState (phase_load_near_overload)

	// Set up ePDU2DeviceStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.1.1", 1)          // ePDU2DeviceStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.3.4.1.3.1", "Device-A") // ePDU2DeviceStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.4.1", 1)          // ePDU2DeviceStatusLoadState (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.5.1", 300)        // cyberpower.ePDU2DeviceStatusCurrentLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.6.1", 350)        // cyberpower.ePDU2DeviceStatusCurrentPeakLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.12.1", 1)         // ePDU2DeviceStatusPowerSupplyAlarm (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.13.1", 1)         // ePDU2DeviceStatusPowerSupply1Status (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.14.1", 1)         // ePDU2DeviceStatusPowerSupply2Status (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.15.1", 400)       // cyberpower.ePDU2DeviceStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.16.1", 80)        // cyberpower.ePDU2DeviceStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.17.1", 1)         // ePDU2DeviceStatusRoleType (standalone)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.1.2", 2)          // ePDU2DeviceStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.3.4.1.3.2", "Device-B") // ePDU2DeviceStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.4.2", 3)          // ePDU2DeviceStatusLoadState (near_overload)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.5.2", 500)        // cyberpower.ePDU2DeviceStatusCurrentLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.6.2", 600)        // cyberpower.ePDU2DeviceStatusCurrentPeakLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.12.2", 2)         // ePDU2DeviceStatusPowerSupplyAlarm (alarm)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.13.2", 2)         // ePDU2DeviceStatusPowerSupply1Status (alarm)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.14.2", 1)         // ePDU2DeviceStatusPowerSupply2Status (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.15.2", 700)       // cyberpower.ePDU2DeviceStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.16.2", 85)        // cyberpower.ePDU2DeviceStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.3.4.1.17.2", 2)         // ePDU2DeviceStatusRoleType (host)

	// Set up ePDU2PhaseStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.1.1", 1)     // ePDU2PhaseStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.4.1", 1)     // ePDU2PhaseStatusLoadState (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.5.1", 1000)  // cyberpower.ePDU2PhaseStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.6.1", 230)   // cyberpower.ePDU2PhaseStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.7.1", 800)   // cyberpower.ePDU2PhaseStatusPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.8.1", 1200)  // cyberpower.ePDU2PhaseStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.9.1", 95)    // cyberpower.ePDU2PhaseStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.10.1", 1100) // cyberpower.ePDU2PhaseStatusPeakLoad
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.1.2", 2)     // ePDU2PhaseStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.4.2", 3)     // ePDU2PhaseStatusLoadState (near_overload)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.5.2", 1500)  // cyberpower.ePDU2PhaseStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.6.2", 240)   // cyberpower.ePDU2PhaseStatusVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.7.2", 1200)  // cyberpower.ePDU2PhaseStatusPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.8.2", 1800)  // cyberpower.ePDU2PhaseStatusApparentPower
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.9.2", 90)    // cyberpower.ePDU2PhaseStatusPowerFactor
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.4.4.1.10.2", 1700) // cyberpower.ePDU2PhaseStatusPeakLoad

	// Set up ePDU2BankStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.1.1", 1)   // ePDU2BankStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.4.1", 1)   // ePDU2BankStatusLoadState (normal)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.5.1", 500) // cyberpower.ePDU2BankStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.6.1", 600) // cyberpower.ePDU2BankStatusPeakLoad
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.1.2", 2)   // ePDU2BankStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.4.2", 3)   // ePDU2BankStatusLoadState (near_overload)
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.5.2", 800) // cyberpower.ePDU2BankStatusLoad
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.5.4.1.6.2", 900) // cyberpower.ePDU2BankStatusPeakLoad

	// Set up ePDU2OutletSwitchedStatusTable
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.1.1", 1)                  // ePDU2OutletSwitchedStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.4.1", "SwitchedOutlet-1") // ePDU2OutletSwitchedStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.5.1", 1)                  // ePDU2OutletSwitchedStatusState (on)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.1.2", 2)                  // ePDU2OutletSwitchedStatusIndex
	fakeSession.SetStr("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.4.2", "SwitchedOutlet-2") // ePDU2OutletSwitchedStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.6.6.1.4.1.5.2", 2)                  // ePDU2OutletSwitchedStatusState (off)

	// Set up scalar metrics
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.7", 120)  // cyberpower.ePDUStatusInputVoltage
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.3.5.8", 60)   // cyberpower.ePDUStatusInputFrequency
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.4.2.1.0", 25) // cyberpower.envirTemperature
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.4.2.6.0", 25) // cyberpower.envirTemperatureCelsius
	fakeSession.SetInt("1.3.6.1.4.1.3808.1.1.4.3.1.0", 50) // cyberpower.envirHumidity

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "test"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "192.168.1.100",
		Namespace:             "default",
		CollectDeviceMetadata: false, // Metric mode doesn't collect metadata
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("cyberpower-pdu")
	require.NoError(t, err)
	assert.Equal(t, "cyberpower-pdu", deviceInfo.Profile)

	// Fetch SNMP values using FakeSession
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)
	require.NotNil(t, values)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:192.168.1.100", "snmp_profile:cyberpower-pdu", "device_vendor:cyberpower", "snmp_host:test-cyberpower-pdu"}
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Test final tag/field generation (Metric mode: collectMeta = false)
	var metaData deviceMetaData
	metaData.collectMeta = false // Metric mode
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	// In Metric mode, data should be directly in fts.Data fields, not in "all" field
	metricFound := false
	for _, data := range fts.Data {
		// Check normalized field name (CamelCase conversion)
		if _, ok := data.Fields["cyberpowerEPDULoadStatusLoad"]; ok {
			metricFound = true
			// Verify tags are directly in data.Tags (not in JSON)
			assert.Contains(t, data.Tags, "snmp_profile", "should have snmp_profile tag")
			assert.Equal(t, "cyberpower-pdu", data.Tags["snmp_profile"])
			assert.Contains(t, data.Tags, "device_vendor", "should have device_vendor tag")
			assert.Equal(t, "cyberpower", data.Tags["device_vendor"])
			assert.Contains(t, data.Tags, "snmp_host", "should have snmp_host tag")
			assert.Equal(t, "test-cyberpower-pdu", data.Tags["snmp_host"])
			// Verify metric-specific tags
			assert.Contains(t, data.Tags, "e_pdu_load_status_index", "should have index tag")
			assert.Contains(t, data.Tags, "e_pdu_load_status_load_state", "should have mapped tag")
			assert.Equal(t, "load_normal", data.Tags["e_pdu_load_status_load_state"])
			break
		}
	}
	assert.True(t, metricFound, "should have metric data with correct tags in Metric mode")

	// Verify no device_meta field in Metric mode
	for _, data := range fts.Data {
		_, hasDeviceMeta := data.Fields["device_meta"]
		assert.False(t, hasDeviceMeta, "Metric mode should not have device_meta field")
		_, hasAllField := data.Fields["all"]
		assert.False(t, hasAllField, "Metric mode should not have 'all' field")
	}

	pts := ipt.CollectingMeasurements("192.168.1.100", deviceInfo, false)
	t.Logf("Final points count (Metric mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Metric): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_F5BigIP tests the complete flow for f5-big-ip profile
// This tests: profile loading -> metric_type -> metadata type field
func TestProfileIntegration_F5BigIP(t *testing.T) {
	t.Skip("skipping f5-big-ip test")

	// Load the f5-big-ip profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"f5-big-ip": {
			DefinitionFile: "f5-big-ip.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "f5-big-ip")

	profileDef := profiles["f5-big-ip"]

	// Verify profile loaded correctly
	assert.NotEmpty(t, profileDef.Metrics)
	assert.Equal(t, "f5", profileDef.Device.Vendor)
	assert.Contains(t, profileDef.SysObjectIds, "1.3.6.1.4.1.3375.2.1.3.4.*")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "f5-bigip-01")                // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "BIG-IP Virtual Edition")     // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.3375.2.1.3.4.1") // sysObjectID

	// Set up F5-specific metadata
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.3.3.3.0", "SN123456789")                 // sysGeneralChassisSerialNum
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.4.2.0", "15.0.1")                        // sysProductVersion
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.4.1.0", "BIG-IP")                        // sysProductName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.3.3.1.0", "Z100")                        // sysGeneralHwName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.6.1.0", "Linux")                         // sysSystemName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.6.3.0", "3.10.0-862.14.4.el7.ve.x86_64") // sysSystemRelease

	// Set up F5 metrics (gauge type)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.1.44.0", 8589934592) // sysStatMemoryTotal (8GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.1.45.0", 4294967296) // sysStatMemoryUsed (4GB)

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "production"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "10.0.0.1",
		Namespace:             "default",
		CollectDeviceMetadata: true,
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("f5-big-ip")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting with metric_type
	var metricData snmputil.MetricDatas
	tags := []string{"ip:10.0.0.1", "snmp_profile:f5-big-ip", "device_vendor:f5"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify gauge metrics were collected
	gaugeMetricFound := false
	for _, metric := range metricData.Data {
		if metric.Name == "sysStatMemoryTotal" || metric.Name == "sysStatMemoryUsed" {
			gaugeMetricFound = true
			// Verify metric has a value
			assert.Greater(t, metric.Value, 0.0, "metric should have a value")
		}
	}
	assert.True(t, gaugeMetricFound, "should have found gauge metrics")

	// Test metadata reporting
	var metaData deviceMetaData
	metaData.collectMeta = true
	deviceInfo.ReportNetworkDeviceMetadata(values, tags, deviceInfo.Metadata, time.Now(), snmputil.DeviceStatusReachable, &metaData)

	// Verify metadata contains type field
	metadataJSON := metaData.data[0]
	assert.Contains(t, metadataJSON, `"type":"load_balancer"`, "metadata should contain device type from profile")
	assert.Contains(t, metadataJSON, `"vendor":"f5"`, "metadata should contain vendor")

	// Test final tag/field generation
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	pts := ipt.CollectingMeasurements("10.0.0.1", deviceInfo, true)
	t.Logf("Final points count (Object mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Object): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_F5BigIP_MetricMode tests the complete flow for f5-big-ip profile in Metric mode
func TestProfileIntegration_F5BigIP_MetricMode(t *testing.T) {
	t.Skip("skipping f5-big-ip metric mode test")

	// Load the f5-big-ip profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"f5-big-ip": {
			DefinitionFile: "f5-big-ip.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "f5-big-ip")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "f5-bigip-01")                // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "BIG-IP Virtual Edition")     // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.3375.2.1.3.4.1") // sysObjectID
	fakeSession.SetStr("1.3.6.1.2.1.1.6.0", "Data Center F5")             // sysLocation

	// Set up F5-specific metadata
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.3.3.3.0", "SN123456789")                 // sysGeneralChassisSerialNum
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.4.2.0", "15.0.1")                        // sysProductVersion
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.4.1.0", "BIG-IP")                        // sysProductName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.3.3.1.0", "Z100")                        // sysGeneralHwName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.6.1.0", "Linux")                         // sysSystemName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.6.3.0", "3.10.0-862.14.4.el7.ve.x86_64") // sysSystemRelease

	// Set up Memory Stats
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.1.44.0", 8589934592)   // sysStatMemoryTotal (8GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.1.45.0", 4294967296)   // sysStatMemoryUsed (4GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.21.28.0", 16106127360) // sysGlobalTmmStatMemoryTotal (15GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.21.29.0", 8053063680)  // sysGlobalTmmStatMemoryUsed (7.5GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.20.44.0", 2147483648)  // sysGlobalHostOtherMemoryTotal (2GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.20.45.0", 1073741824)  // sysGlobalHostOtherMemoryUsed (1GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.20.46.0", 4294967296)  // sysGlobalHostSwapTotal (4GB)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.20.47.0", 2147483648)  // sysGlobalHostSwapUsed (2GB)

	// Set up CPU Stats (sysMultiHostCpuTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.1.7.5.2.1.3.1", "0") // sysMultiHostCpuId
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.11.1", 50) // sysMultiHostCpuUsageRatio
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.27.1", 50) // cpu.usage (sysMultiHostCpuUsageRatio1m)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.4.1", 30)  // sysMultiHostCpuUser
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.5.1", 5)   // sysMultiHostCpuNice
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.6.1", 10)  // sysMultiHostCpuSystem
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.7.1", 50)  // sysMultiHostCpuIdle
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.8.1", 2)   // sysMultiHostCpuIrq
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.9.1", 1)   // sysMultiHostCpuSoftirq
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.7.5.2.1.10.1", 2)  // sysMultiHostCpuIowait

	// Set up TCP Stats
	fakeSession.SetInt("1.3.6.1.2.1.6.18.0", 100)               // tcpHCInExtSegs
	fakeSession.SetInt("1.3.6.1.2.1.6.5.0", 12345)              // tcpActiveOpens
	fakeSession.SetInt("1.3.6.1.2.1.6.6.0", 67890)              // tcpPassiveOpens
	fakeSession.SetInt("1.3.6.1.2.1.6.7.0", 100)                // tcpAttemptFails
	fakeSession.SetInt("1.3.6.1.2.1.6.8.0", 50)                 // tcpEstabResets
	fakeSession.SetInt("1.3.6.1.2.1.6.9.0", 200)                // tcpCurrEstab
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.2.0", 500)  // sysTcpStatOpen
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.3.0", 20)   // sysTcpStatCloseWait
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.4.0", 10)   // sysTcpStatFinWait
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.5.0", 5)    // sysTcpStatTimeWait
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.6.0", 1000) // sysTcpStatAccepts
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.7.0", 50)   // sysTcpStatAcceptfails
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.8.0", 800)  // sysTcpStatConnects
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.12.9.0", 30)   // sysTcpStatConnfails

	// Set up UDP Stats
	fakeSession.SetInt("1.3.6.1.2.1.7.2.0", 50000)              // udpInDatagrams
	fakeSession.SetInt("1.3.6.1.2.1.7.3.0", 100)                // udpNoPorts
	fakeSession.SetInt("1.3.6.1.2.1.7.8.0", 40000)              // udpOutDatagrams
	fakeSession.SetInt("1.3.6.1.2.1.7.9.0", 5)                  // udpRcvbufErrors
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.13.2.0", 200)  // sysUdpStatOpen
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.13.3.0", 1500) // sysUdpStatAccepts
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.13.4.0", 10)   // sysUdpStatAcceptfails
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.13.5.0", 1200) // sysUdpStatConnects
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.13.6.0", 5)    // sysUdpStatConnfails

	// Set up IP Stats
	fakeSession.SetInt("1.3.6.1.2.1.4.10.0", 10000) // ipInReceives
	fakeSession.SetInt("1.3.6.1.2.1.4.3.0", 50)     // ipInHdrErrors

	// Set up Client SSL Stats
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.9.2.0", 100)    // sysClientsslStatCurConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.9.10.0", 10240) // sysClientsslStatEncryptedBytesIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.9.11.0", 5120)  // sysClientsslStatEncryptedBytesOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.9.12.0", 15360) // sysClientsslStatDecryptedBytesIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.9.13.0", 7680)  // sysClientsslStatDecryptedBytesOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.1.1.2.9.29.0", 5)     // sysClientsslStatHandshakeFailures

	// Set up Virtual Servers (Scalar)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.1.1.0", 5) // ltmVirtualServNumber

	// Set up Virtual Servers (ltmVirtualServTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.10.1.2.1.1.1", "virtual_server_a") // ltmVirtualServName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.1.2.1.9.1", 1)                  // ltmVirtualServEnabled (1=enabled)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.1.2.1.10.1", 1000)              // ltmVirtualServConnLimit

	// Set up Virtual Server Status (ltmVsStatusTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.10.13.2.1.1.1", "virtual_server_a") // ltmVsStatusName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.13.2.1.2.1", 1)                  // ltmVsStatusAvailState (1=green)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.13.2.1.3.1", 1)                  // ltmVsStatusEnabledState (1=enabled)

	// Set up Virtual Server Stats (ltmVirtualServStatTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.10.2.3.1.1.1", "virtual_server_a") // ltmVirtualServStatName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.6.1", 10000)              // ltmVirtualServStatClientPktsIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.7.1", 102400)             // ltmVirtualServStatClientBytesIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.8.1", 8000)               // ltmVirtualServStatClientPktsOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.9.1", 81920)              // ltmVirtualServStatClientBytesOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.12.1", 500)               // ltmVirtualServStatClientCurConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.31.1", 60)                // ltmVirtualServStatVsUsageRatio5s
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.32.1", 50)                // ltmVirtualServStatVsUsageRatio1m
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.33.1", 40)                // ltmVirtualServStatVsUsageRatio5m
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.34.1", 100)               // ltmVirtualServStatCurrentConnsPerSec
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.35.1", 5)                 // ltmVirtualServStatDurationRateExceeded
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.5.1", 2)                  // ltmVirtualServStatNoNodesErrors
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.11.1", 15000)             // ltmVirtualServStatClientTotConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.27.1", 2000)              // ltmVirtualServStatTotRequests
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.40.1", 10)                // ltmVirtualServStatClientEvictedConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.10.2.3.1.41.1", 5)                 // ltmVirtualServStatClientSlowKilled

	// Set up Nodes (Scalar)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.1.1.0", 3) // ltmNodeAddrNumber

	// Set up Nodes (ltmNodeAddrTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.4.1.2.1.17.1", "node_server_a") // ltmNodeAddrName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.1.2.1.3.1", 2000)             // ltmNodeAddrConnLimit
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.1.2.1.4.1", 10)               // ltmNodeAddrRatio
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.1.2.1.5.1", 5)                // ltmNodeAddrDynamicRatio
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.1.2.1.6.1", 4)                // ltmNodeAddrMonitorState (4=up)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.1.2.1.7.1", 4)                // ltmNodeAddrMonitorStatus (4=up)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.1.2.1.10.1", 1)               // ltmNodeAddrSessionStatus (1=enabled)

	// Set up Node Stats (ltmNodeAddrStatTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.4.2.3.1.20.1", "node_server_a") // ltmNodeAddrStatNodeName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.3.1", 5000)             // ltmNodeAddrStatServerPktsIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.4.1", 51200)            // ltmNodeAddrStatServerBytesIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.5.1", 3000)             // ltmNodeAddrStatServerPktsOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.6.1", 30720)            // ltmNodeAddrStatServerBytesOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.9.1", 100)              // ltmNodeAddrStatServerCurConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.21.1", 200)             // ltmNodeAddrStatCurSessions
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.22.1", 50)              // ltmNodeAddrStatCurrentConnsPerSec
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.23.1", 2)               // ltmNodeAddrStatDurationRateExceeded
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.8.1", 8000)             // ltmNodeAddrStatServerTotConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.4.2.3.1.17.1", 1500)            // ltmNodeAddrStatTotRequests

	// Set up Pools (Scalar)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.1.1.0", 2) // ltmPoolNumber

	// Set up Pools (ltmPoolTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.5.1.2.1.1.1", "pool_web_servers") // ltmPoolName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.1.2.1.8.1", 2)                  // ltmPoolActiveMemberCnt
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.1.2.1.16.1", 10)                // ltmPoolDynamicRatioSum
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.1.2.1.23.1", 3)                 // ltmPoolMemberCnt

	// Set up Pool Stats (ltmPoolStatTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.5.2.3.1.1.1", "pool_web_servers") // ltmPoolStatName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.2.1", 6000)               // ltmPoolStatServerPktsIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.3.1", 61440)              // ltmPoolStatServerBytesIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.4.1", 4500)               // ltmPoolStatServerPktsOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.5.1", 46080)              // ltmPoolStatServerBytesOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.8.1", 80)                 // ltmPoolStatServerCurConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.18.1", 10)                // ltmPoolStatConnqDepth
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.19.1", 500)               // ltmPoolStatConnqAgeHead
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.31.1", 150)               // ltmPoolStatCurSessions
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.7.1", 9000)               // ltmPoolStatServerTotConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.23.1", 20)                // ltmPoolStatConnqServiced
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.2.3.1.30.1", 1800)              // ltmPoolStatTotRequests

	// Set up Pool Members (Scalar)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.3.1.0", 6) // ltmPoolMemberNumber

	// Set up Pool Members (ltmPoolMemberTable) - Index 1 (for pool_web_servers, node_server_a)
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.5.3.2.1.1.1", "pool_web_servers") // ltmPoolMemberPoolName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.5.3.2.1.19.1", "node_server_a")   // ltmPoolMemberNodeName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.3.2.1.5.1", 500)                // ltmPoolMemberConnLimit
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.3.2.1.6.1", 5)                  // ltmPoolMemberRatio
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.3.2.1.9.1", 2)                  // ltmPoolMemberDynamicRatio
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.3.2.1.10.1", 4)                 // ltmPoolMemberMonitorState (4=up)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.3.2.1.11.1", 4)                 // ltmPoolMemberMonitorStatus (4=up)
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.3.2.1.13.1", 1)                 // ltmPoolMemberSessionStatus (1=enabled)

	// Set up Pool Member Stats (ltmPoolMemberStatTable) - Index 1
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.5.4.3.1.1.1", "pool_web_servers") // ltmPoolMemberStatPoolName
	fakeSession.SetStr("1.3.6.1.4.1.3375.2.2.5.4.3.1.28.1", "node_server_a")   // ltmPoolMemberStatNodeName
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.5.1", 1000)               // ltmPoolMemberStatServerPktsIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.6.1", 10240)              // ltmPoolMemberStatServerBytesIn
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.7.1", 700)                // ltmPoolMemberStatServerPktsOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.8.1", 7168)               // ltmPoolMemberStatServerBytesOut
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.11.1", 20)                // ltmPoolMemberStatServerCurConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.22.1", 5)                 // ltmPoolMemberStatConnqDepth
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.23.1", 100)               // ltmPoolMemberStatConnqAgeHead
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.29.1", 30)                // ltmPoolMemberStatCurSessions
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.30.1", 10)                // ltmPoolMemberStatCurrentConnsPerSec
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.31.1", 1)                 // ltmPoolMemberStatDurationRateExceeded
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.10.1", 1200)              // ltmPoolMemberStatServerTotConns
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.19.1", 250)               // ltmPoolMemberStatTotRequests
	fakeSession.SetInt("1.3.6.1.4.1.3375.2.2.5.4.3.1.27.1", 5)                 // ltmPoolMemberStatConnqServiced

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "production"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "10.0.0.1",
		Namespace:             "default",
		CollectDeviceMetadata: false, // Metric mode
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("f5-big-ip")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:10.0.0.1", "snmp_profile:f5-big-ip", "device_vendor:f5"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Test final tag/field generation (Metric mode: collectMeta = false)
	var metaData deviceMetaData
	metaData.collectMeta = false // Metric mode
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	// Verify no device_meta field in Metric mode
	for _, data := range fts.Data {
		_, hasDeviceMeta := data.Fields["device_meta"]
		assert.False(t, hasDeviceMeta, "Metric mode should not have device_meta field")
		_, hasAllField := data.Fields["all"]
		assert.False(t, hasAllField, "Metric mode should not have 'all' field")
	}

	pts := ipt.CollectingMeasurements("10.0.0.1", deviceInfo, false)
	t.Logf("Final points count (Metric mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Metric): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_CiscoNexus tests the complete flow for cisco-nexus profile
// This tests: extends, mapping, type metadata (switch)
func TestProfileIntegration_CiscoNexus(t *testing.T) {
	t.Skip("skipping cisco-nexus test")

	// Load the cisco-nexus profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"cisco-nexus": {
			DefinitionFile: "cisco-nexus.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "cisco-nexus")

	profileDef := profiles["cisco-nexus"]

	// Verify profile loaded correctly with extends
	assert.NotEmpty(t, profileDef.Metrics)
	assert.NotEmpty(t, profileDef.Metadata)
	assert.Equal(t, "cisco", profileDef.Device.Vendor)
	assert.Contains(t, profileDef.SysObjectIds, "1.3.6.1.4.1.9.1.1216")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "cisco-nexus-01")               // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "Cisco Nexus Operating System") // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.9.1.1216")         // sysObjectID

	// Set up sensor table data (from cisco-nexus.yaml)
	// Index: 1 (sensor type: 8=celsius)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.91.1.1.1.1.1.1", 8)   // entSensorType
	fakeSession.SetInt("1.3.6.1.4.1.9.9.91.1.1.1.1.4.1", 350) // entSensorValue (35.0 degrees)
	// Index: 2 (sensor type: 4=volts_dc)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.91.1.1.1.1.1.2", 4)   // entSensorType
	fakeSession.SetInt("1.3.6.1.4.1.9.9.91.1.1.1.1.4.2", 120) // entSensorValue (12.0 volts)

	// Set up memory table data
	fakeSession.SetInt("1.3.6.1.4.1.9.9.221.1.1.1.1.7.1", 4294967296) // memory.used (4GB)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.221.1.1.1.1.8.1", 8589934592) // memory.free (8GB)

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "production"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "10.0.0.10",
		Namespace:             "default",
		CollectDeviceMetadata: true,
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("cisco-nexus")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:10.0.0.10", "snmp_profile:cisco-nexus", "device_vendor:cisco"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Verify mapping tag is applied (sensor_type should be mapped)
	sensorMetricFound := false
	for _, metric := range metricData.Data {
		if metric.Name == "entSensorValue" {
			sensorMetricFound = true
			// Verify mapping tag is applied
			hasMappingTag := false
			for _, tag := range metric.Tags {
				if tag == "sensor_type:celsius" || tag == "sensor_type:volts_dc" {
					hasMappingTag = true
					break
				}
			}
			assert.True(t, hasMappingTag, "should have mapped sensor_type tag")
		}
	}
	assert.True(t, sensorMetricFound, "should have found sensor metrics with mapping")

	// Test metadata reporting
	var metaData deviceMetaData
	metaData.collectMeta = true
	deviceInfo.ReportNetworkDeviceMetadata(values, tags, deviceInfo.Metadata, time.Now(), snmputil.DeviceStatusReachable, &metaData)

	// Verify metadata contains type field (switch)
	metadataJSON := metaData.data[0]
	assert.Contains(t, metadataJSON, `"type":"switch"`, "metadata should contain device type from profile")
	assert.Contains(t, metadataJSON, `"vendor":"cisco"`, "metadata should contain vendor")

	// Test final tag/field generation
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	pts := ipt.CollectingMeasurements("10.0.0.10", deviceInfo, true)
	t.Logf("Final points count (Object mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Object): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_CiscoNexus_MetricMode tests the complete flow for cisco-nexus profile in Metric mode
func TestProfileIntegration_CiscoNexus_MetricMode(t *testing.T) {
	t.Skip("skipping cisco-nexus metric mode test")

	// Load the cisco-nexus profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"cisco-nexus": {
			DefinitionFile: "cisco-nexus.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "cisco-nexus")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "cisco-nexus-01")                                                                                                                                                      // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "Cisco NX-OS(tm) Nexus9000 C9364C, Software (NXOS 32-bit), Version 9.3(9), RELEASE SOFTWARE Copyright (c) 2002-2022 by Cisco Systems, Inc. Compiled 2/4/2022 7:00:00") // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.9.1.1216")                                                                                                                                                // sysObjectID
	fakeSession.SetStr("1.3.6.1.2.1.1.6.0", "Data Center C")                                                                                                                                                       // sysLocation

	// Set up global metric tags (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "cisco-nexus-01") // sysName for metric tags

	// Set up metadata fields (from _cisco-metadata.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "Cisco NX-OS(tm) Nexus9000 C9364C, Software (NXOS 32-bit), Version 9.3(9), RELEASE SOFTWARE Copyright (c) 2002-2022 by Cisco Systems, Inc. Compiled 2/4/2022 7:00:00") // sysDescr for version, model, os_name extraction

	// Set up interface table data (from _generic-if.yaml)
	// Interface 1
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.1.1", 1)                   // ifIndex
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.2.1", "eth0")              // ifDescr
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.6.1", "00:0c:29:4a:d6:94") // ifPhysAddress
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.7.1", 1)                   // ifAdminStatus (1=up)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.8.1", 1)                   // ifOperStatus (1=up)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.5.1", 1000000000)          // ifSpeed (1Gbps)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.14.1", 10)                 // ifInErrors
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.13.1", 0)                  // ifInDiscards
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.20.1", 5)                  // ifOutErrors
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.19.1", 0)                  // ifOutDiscards
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.1.1", "eth0")           // ifName
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.18.1", "Ethernet0/0")   // ifAlias
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.6.1", 10000000)         // ifHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.10.1", 20000000)        // ifHCOutOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.7.1", 10000)            // ifHCInUcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.8.1", 500)              // ifHCInMulticastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.9.1", 100)              // ifHCInBroadcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.11.1", 20000)           // ifHCOutUcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.12.1", 600)             // ifHCOutMulticastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.13.1", 150)             // ifHCOutBroadcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.15.1", 1000)            // ifHighSpeed (Mbps)

	// Interface 2
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.1.2", 2)                   // ifIndex
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.2.2", "eth1")              // ifDescr
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.6.2", "00:0c:29:4a:d6:95") // ifPhysAddress
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.7.2", 2)                   // ifAdminStatus (2=down)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.8.2", 2)                   // ifOperStatus (2=down)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.5.2", 100000000)           // ifSpeed (100Mbps)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.14.2", 2)                  // ifInErrors
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.13.2", 1)                  // ifInDiscards
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.20.2", 1)                  // ifOutErrors
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.19.2", 0)                  // ifOutDiscards
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.1.2", "eth1")           // ifName
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.18.2", "Ethernet0/1")   // ifAlias
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.6.2", 500000)           // ifHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.10.2", 1000000)         // ifHCOutOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.7.2", 500)              // ifHCInUcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.8.2", 50)               // ifHCInMulticastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.9.2", 10)               // ifHCInBroadcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.11.2", 1000)            // ifHCOutUcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.12.2", 60)              // ifHCOutMulticastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.13.2", 15)              // ifHCOutBroadcastPkts
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.15.2", 100)             // ifHighSpeed (Mbps)

	// Set up TCP Stats (from _generic-tcp.yaml)
	fakeSession.SetInt("1.3.6.1.2.1.6.5.0", 123456)   // tcpActiveOpens
	fakeSession.SetInt("1.3.6.1.2.1.6.6.0", 67890)    // tcpPassiveOpens
	fakeSession.SetInt("1.3.6.1.2.1.6.7.0", 100)      // tcpAttemptFails
	fakeSession.SetInt("1.3.6.1.2.1.6.8.0", 50)       // tcpEstabResets
	fakeSession.SetInt("1.3.6.1.2.1.6.9.0", 200)      // tcpCurrEstab
	fakeSession.SetInt("1.3.6.1.2.1.6.17.0", 1000000) // tcpHCInSegs
	fakeSession.SetInt("1.3.6.1.2.1.6.18.0", 2000000) // tcpHCOutSegs
	fakeSession.SetInt("1.3.6.1.2.1.6.12.0", 20)      // tcpRetransSegs
	fakeSession.SetInt("1.3.6.1.2.1.6.14.0", 5)       // tcpInErrs
	fakeSession.SetInt("1.3.6.1.2.1.6.15.0", 2)       // tcpOutRsts

	// Set up UDP Stats (from _generic-udp.yaml)
	fakeSession.SetInt("1.3.6.1.2.1.7.8.0", 500000) // udpHCInDatagrams
	fakeSession.SetInt("1.3.6.1.2.1.7.2.0", 100)    // udpNoPorts
	fakeSession.SetInt("1.3.6.1.2.1.7.3.0", 50)     // udpInErrors
	fakeSession.SetInt("1.3.6.1.2.1.7.9.0", 250000) // udpHCOutDatagrams

	// Set up OSPF NBR Table (from _generic-ospf.yaml)
	// Index 1 (ospfNbrIpAddr: 10.0.0.101, ospfNbrRtrId: 10.0.0.101)
	fakeSession.SetInt("1.3.6.1.2.1.14.10.1.6.10.0.0.101.1", 8)            // ospfNbrState (full)
	fakeSession.SetInt("1.3.6.1.2.1.14.10.1.7.10.0.0.101.1", 10)           // ospfNbrEvents
	fakeSession.SetInt("1.3.6.1.2.1.14.10.1.8.10.0.0.101.1", 0)            // ospfNbrLsRetransQLen
	fakeSession.SetStr("1.3.6.1.2.1.14.10.1.3.10.0.0.101.1", "10.0.0.101") // ospfNbrRtrId
	fakeSession.SetStr("1.3.6.1.2.1.14.10.1.1.10.0.0.101.1", "10.0.0.101") // ospfNbrIpAddr

	// Index 2 (ospfNbrIpAddr: 10.0.0.102, ospfNbrRtrId: 10.0.0.102)
	fakeSession.SetInt("1.3.6.1.2.1.14.10.1.6.10.0.0.102.1", 4)            // ospfNbrState (two_way)
	fakeSession.SetInt("1.3.6.1.2.1.14.10.1.7.10.0.0.102.1", 5)            // ospfNbrEvents
	fakeSession.SetInt("1.3.6.1.2.1.14.10.1.8.10.0.0.102.1", 1)            // ospfNbrLsRetransQLen
	fakeSession.SetStr("1.3.6.1.2.1.14.10.1.3.10.0.0.102.1", "10.0.0.102") // ospfNbrRtrId
	fakeSession.SetStr("1.3.6.1.2.1.14.10.1.1.10.0.0.102.1", "10.0.0.102") // ospfNbrIpAddr

	// Set up BGP Peer Table (from _generic-bgp4.yaml)
	// Index 1 (bgpPeerRemoteAddr: 172.16.0.2)
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.3.172.16.0.2", 2)            // bgpPeerAdminStatus (start)
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.4.172.16.0.2", 4)            // bgpPeerNegotiatedVersion
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.9.172.16.0.2", 65001)        // bgpPeerRemoteAs
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.2.172.16.0.2", 6)            // bgpPeerState (established)
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.10.172.16.0.2", 100)         // bgpPeerInUpdates
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.11.172.16.0.2", 50)          // bgpPeerOutUpdates
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.12.172.16.0.2", 1000)        // bgpPeerInTotalMessages
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.13.172.16.0.2", 800)         // bgpPeerOutTotalMessages
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.15.172.16.0.2", 5)           // bgpPeerFsmEstablishedTransitions
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.16.172.16.0.2", 3600)        // bgpPeerFsmEstablishedTime
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.17.172.16.0.2", 60)          // bgpPeerConnectRetryInterval
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.18.172.16.0.2", 180)         // bgpPeerHoldTime
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.19.172.16.0.2", 60)          // bgpPeerKeepAlive
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.20.172.16.0.2", 180)         // bgpPeerHoldTimeConfigured
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.21.172.16.0.2", 60)          // bgpPeerKeepAliveConfigured
	fakeSession.SetInt("1.3.6.1.2.1.15.3.1.22.172.16.0.2", 30)          // bgpPeerMinASOriginationInterval
	fakeSession.SetStr("1.3.6.1.2.1.15.3.1.7.172.16.0.2", "172.16.0.2") // bgpPeerRemoteAddr

	// Set up IP System Stats (from _generic-ip.yaml)
	// Index: 1 (ipv4)
	fakeSession.SetInt("1.3.6.1.2.1.4.31.1.1.4.1", 100000)   // ipSystemStatsHCInReceives
	fakeSession.SetInt("1.3.6.1.2.1.4.31.1.1.6.1", 5000000)  // ipSystemStatsHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.4.31.1.1.7.1", 10)       // ipSystemStatsInHdrErrors
	fakeSession.SetInt("1.3.6.1.2.1.4.31.1.1.8.1", 5)        // ipSystemStatsInNoRoutes
	fakeSession.SetInt("1.3.6.1.2.1.4.31.1.1.19.1", 90000)   // ipSystemStatsHCInDelivers
	fakeSession.SetInt("1.3.6.1.2.1.4.31.1.1.21.1", 80000)   // ipSystemStatsHCOutRequests
	fakeSession.SetInt("1.3.6.1.2.1.4.31.1.1.33.1", 4000000) // ipSystemStatsHCOutOctets

	// Set up IP If Stats (from _generic-ip.yaml)
	// Index: 1 (ipv4, ifIndex 1)
	fakeSession.SetInt("1.3.6.1.2.1.4.31.3.1.6.1.1", 2000000)  // ipIfStatsHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.4.31.3.1.7.1.1", 2)        // ipIfStatsInHdrErrors
	fakeSession.SetInt("1.3.6.1.2.1.4.31.3.1.19.1.1", 18000)   // ipIfStatsHCInDelivers
	fakeSession.SetInt("1.3.6.1.2.1.4.31.3.1.33.1.1", 1500000) // ipIfStatsHCOutOctets

	// Set up CPU Stats (from _cisco-cpu-memory.yaml and _cisco-generic.yaml)
	// Index: 1 (cpu: 1)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.109.1.1.1.1.7.1", 30)       // cpu.usage (cpmCPUTotal1minRev)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.109.1.1.1.1.10.1", 60)      // cpmCPUTotalMonIntervalValue
	fakeSession.SetInt("1.3.6.1.4.1.9.9.109.1.1.1.1.12.1", 1000000) // cpmCPUMemoryUsed
	fakeSession.SetInt("1.3.6.1.4.1.9.9.109.1.1.1.1.13.1", 2000000) // cpmCPUMemoryFree

	// Set up Memory Pool Stats (from _cisco-cpu-memory.yaml and _cisco-generic.yaml)
	// Index: 1 (mem_pool_name: Processor)
	fakeSession.SetStr("1.3.6.1.4.1.9.9.48.1.1.1.2.1", "Processor") // ciscoMemoryPoolName
	fakeSession.SetInt("1.3.6.1.4.1.9.9.48.1.1.1.5.1", 2048576)     // memory.used (ciscoMemoryPoolUsed)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.48.1.1.1.6.1", 4096000)     // memory.free (ciscoMemoryPoolFree)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.48.1.1.1.7.1", 1024000)     // ciscoMemoryPoolLargestFree

	// Set up FRU Power Status Table (from _cisco-generic.yaml)
	// Index: 1 (fru: PS1)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.1.2.1.1.1", 2)   // cefcFRUPowerAdminStatus (off)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.1.2.1.2.1", 2)   // cefcFRUPowerOperStatus (on)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.1.2.1.3.1", 500) // cefcFRUCurrent
	// constant_value_one metric
	// fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.1.2.1.X.1", 1) // cefcFRUPowerStatus, actual value is derived

	// Set up Fan Tray Status Table (from _cisco-generic.yaml)
	// Index: 1 (fru: Fan1)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.4.1.1.3.1", 2) // cefcFanTrayState (up)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.4.1.1.1.1", 1) // cefcFanTrayStatusIndex
	fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.4.1.1.2.1", 2) // cefcFanTrayDirection (front_to_back)
	// constant_value_one metric
	// fakeSession.SetInt("1.3.6.1.4.1.9.9.117.1.4.1.1.X.1", 1) // cefcFanTrayStatus, actual value is derived

	// Set up Cisco IF Extension MIB (from _cisco-generic.yaml)
	// Index: 1 (interface: eth0)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.276.1.1.2.1.1.1", 5) // cieIfResetCount

	// Set up Cisco EnvMon Temperature Status Table (from _cisco-generic.yaml)
	// Index: 1 (temp_index: 1)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.13.1.3.1.3.1", 40) // ciscoEnvMonTemperatureStatusValue
	fakeSession.SetInt("1.3.6.1.4.1.9.9.13.1.3.1.6.1", 1)  // ciscoEnvMonTemperatureState (normal)

	// Set up Cisco EnvMon Supply Status Table (from _cisco-generic.yaml)
	// Index: 1 (power_status_descr: PowerSupply1)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.13.1.5.1.3.1", 1)              // ciscoEnvMonSupplyState (normal)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.13.1.5.1.4.1", 2)              // ciscoEnvMonSupplySource (ac)
	fakeSession.SetStr("1.3.6.1.4.1.9.9.13.1.5.1.2.1", "PowerSupply1") // ciscoEnvMonSupplyStatusDescr
	// constant_value_one metric
	// fakeSession.SetInt("1.3.6.1.4.1.9.9.13.1.5.1.X.1", 1) // ciscoEnvMonSupplyStatus

	// Set up Cisco Stackwise MIB (cswStackPortInfoTable)
	// Index: 1 (interface: eth0)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.500.1.2.2.1.1.1", 1) // cswStackPortOperStatus (up)

	// Set up Cisco Stackwise MIB (cswSwitchInfoTable)
	// Index: 1 (mac_addr: 00:0c:29:4a:d6:94, entity_name: Chassis)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.500.1.2.1.1.6.1", 4)                   // cswSwitchState (ready)
	fakeSession.SetStr("1.3.6.1.4.1.9.9.500.1.2.1.1.7.1", "00:0c:29:4a:d6:94") // cswSwitchMacAddress
	fakeSession.SetStr("1.3.6.1.2.1.47.1.1.1.1.7.1", "Chassis")                // entPhysicalName
	// constant_value_one metric
	// fakeSession.SetInt("1.3.6.1.4.1.9.9.500.1.2.1.X.1", 1) // cswSwitchInfo

	// Set up Cisco Firewall MIB (cfwConnectionStatTable)
	// Index: 1 (connection_type: total_open)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.147.1.2.2.2.1.4.1", 1000) // cfwConnectionStatCount
	fakeSession.SetInt("1.3.6.1.4.1.9.9.147.1.2.2.2.1.2.1", 2)    // cfwConnectionStatType (total_open)

	// Set up Cisco Firewall MIB (cfwHardwareStatusTable)
	// Index: 1 (hardware_type: 1, hardware_desc: CPU)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.147.1.2.1.1.1.3.1", 1)     // cfwHardwareStatusValue (normal)
	fakeSession.SetStr("1.3.6.1.4.1.9.9.147.1.2.1.1.1.2.1", "CPU") // cfwHardwareInformation

	// Set up Cisco Virtual Switch MIB (cvsChassisTable)
	// Index: 1 (chassis_switch_id: 1)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.388.1.2.2.1.3.1", 86400) // cvsChassisUpTime (1 day)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.388.1.2.2.1.1.1", 1)     // cvsChassisSwitchID

	// Set up RTT Mon Latest RTT Oper Table (from _cisco-generic.yaml)
	// Index: 1 (rtt_index: 1, rtt_state: active, rtt_type: echo, rtt_sense: ok)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.42.1.2.10.1.1.1", 100)        // rttMonLatestRttOperCompletionTime
	fakeSession.SetInt("1.3.6.1.4.1.9.9.42.1.2.10.1.2.1", 1)          // rttMonLatestRttOperSense (ok)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.42.1.2.9.1.10.1", 6)          // rttMonCtrlOperState (active)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.42.1.2.1.1.4.1", 1)           // rttMonCtrlAdminRttType (echo)
	fakeSession.SetStr("1.3.6.1.4.1.9.9.42.1.2.2.1.6.1", "10.0.0.10") // rttMonEchoAdminSourceAddress
	fakeSession.SetStr("1.3.6.1.4.1.9.9.42.1.2.2.1.2.1", "10.0.0.20") // rttMonEchoAdminTargetAddress

	// Set up RTT Mon Ctrl Oper Table (from _cisco-generic.yaml)
	// Index: 1 (rtt_index: 1, rtt_state: active, rtt_type: echo, rtt_timeout: false)
	fakeSession.SetInt("1.3.6.1.4.1.9.9.42.1.2.9.1.6.1", 2) // rttMonCtrlOperTimeoutOccurred (false)

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "production"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "10.0.0.10",
		Namespace:             "default",
		CollectDeviceMetadata: false, // Metric mode
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("cisco-nexus")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:10.0.0.10", "snmp_profile:cisco-nexus", "device_vendor:cisco", "snmp_host:cisco-nexus-01"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Test final tag/field generation (Metric mode: collectMeta = false)
	var metaData deviceMetaData
	metaData.collectMeta = false // Metric mode
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	// Verify no device_meta field in Metric mode
	for _, data := range fts.Data {
		_, hasDeviceMeta := data.Fields["device_meta"]
		assert.False(t, hasDeviceMeta, "Metric mode should not have device_meta field")
		_, hasAllField := data.Fields["all"]
		assert.False(t, hasAllField, "Metric mode should not have 'all' field")
	}

	pts := ipt.CollectingMeasurements("10.0.0.10", deviceInfo, false)
	t.Logf("Final points count (Metric mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Metric): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_DellPowerEdge tests the complete flow for dell-poweredge profile
// This tests: extends, constant_value_one, mapping, type metadata (server)
func TestProfileIntegration_DellPowerEdge(t *testing.T) {
	t.Skip("skipping dell-poweredge test")

	// Load the dell-poweredge profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"dell-poweredge": {
			DefinitionFile: "dell-poweredge.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "dell-poweredge")

	profileDef := profiles["dell-poweredge"]

	// Verify profile loaded correctly
	assert.NotEmpty(t, profileDef.Metrics)
	assert.NotEmpty(t, profileDef.Metadata)
	assert.Equal(t, "dell", profileDef.Device.Vendor)
	assert.Contains(t, profileDef.SysObjectIds, "1.3.6.1.4.1.674.10892.1")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "dell-poweredge-r730")         // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "Dell PowerEdge R730")         // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.674.10892.1.1.1") // sysObjectID

	// Set up systemStateTable data
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.1.1", 1)  // systemStatechassisIndex
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.4.1", 3)  // systemStateChassisStatus (3=ok)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.9.1", 3)  // systemStatePowerSupplyStatusCombined (3=ok)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.24.1", 3) // systemStateTemperatureStatusCombined (3=ok)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.1.2", 2)  // systemStatechassisIndex
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.4.2", 3)  // systemStateChassisStatus
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.9.2", 4)  // systemStatePowerSupplyStatusCombined (4=non_critical)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.24.2", 3) // systemStateTemperatureStatusCombined

	// Set up operatingSystemMemoryTable data
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.400.20.1.6.1", 8589934592)  // operatingSystemMemoryAvailablePhysicalSize (8GB)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.400.20.1.7.1", 17179869184) // operatingSystemMemoryTotalPageFileSize (16GB)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.400.20.1.8.1", 4294967296)  // operatingSystemMemoryAvailablePageFileSize (4GB)

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "datacenter"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "192.168.1.50",
		Namespace:             "default",
		CollectDeviceMetadata: true,
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("dell-poweredge")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:192.168.1.50", "snmp_profile:dell-poweredge", "device_vendor:dell"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Verify constant_value_one metric
	constantMetricFound := false
	for _, metric := range metricData.Data {
		if metric.Name == "systemState" {
			constantMetricFound = true
			// Verify value is 1.0
			assert.Equal(t, 1.0, metric.Value, "constant_value_one metric should have value 1.0")
			// Verify tags include mapping
			hasMappingTag := false
			for _, tag := range metric.Tags {
				if tag == "system_state_power_supply_status_combined:ok" ||
					tag == "system_state_power_supply_status_combined:non_critical" {
					hasMappingTag = true
					break
				}
			}
			assert.True(t, hasMappingTag, "should have mapped power supply status tag")
		}
	}
	assert.True(t, constantMetricFound, "should have found constant_value_one metric")

	// Test metadata reporting
	var metaData deviceMetaData
	metaData.collectMeta = true
	deviceInfo.ReportNetworkDeviceMetadata(values, tags, deviceInfo.Metadata, time.Now(), snmputil.DeviceStatusReachable, &metaData)

	// Verify metadata contains type field (server)
	metadataJSON := metaData.data[0]
	assert.Contains(t, metadataJSON, `"type":"server"`, "metadata should contain device type from profile")
	assert.Contains(t, metadataJSON, `"vendor":"dell"`, "metadata should contain vendor")

	// Test final tag/field generation
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	pts := ipt.CollectingMeasurements("192.168.1.50", deviceInfo, true)
	t.Logf("Final points count (Object mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Object): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_DellPowerEdge_MetricMode tests the complete flow for dell-poweredge profile in Metric mode
func TestProfileIntegration_DellPowerEdge_MetricMode(t *testing.T) {
	t.Skip("skipping dell-poweredge metric mode test")

	// Load the dell-poweredge profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"dell-poweredge": {
			DefinitionFile: "dell-poweredge.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "dell-poweredge")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "dell-poweredge-r730")         // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "Dell PowerEdge R730")         // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.4.1.674.10892.1.1.1") // sysObjectID

	// Set up systemStateTable data
	// Index: 1
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.1.1", 1)  // systemStatechassisIndex
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.4.1", 3)  // systemStateChassisStatus (3=ok)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.9.1", 3)  // systemStatePowerSupplyStatusCombined (3=ok)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.24.1", 3) // systemStateTemperatureStatusCombined (3=ok)
	// Index: 2
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.1.2", 2)  // systemStatechassisIndex
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.4.2", 3)  // systemStateChassisStatus
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.9.2", 4)  // systemStatePowerSupplyStatusCombined (4=non_critical)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.200.10.1.24.2", 3) // systemStateTemperatureStatusCombined

	// Set up operatingSystemMemoryTable data
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.400.20.1.6.1", 8589934592)  // operatingSystemMemoryAvailablePhysicalSize (8GB)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.400.20.1.7.1", 17179869184) // operatingSystemMemoryTotalPageFileSize (16GB)
	fakeSession.SetInt("1.3.6.1.4.1.674.10892.1.400.20.1.8.1", 4294967296)  // operatingSystemMemoryAvailablePageFileSize (4GB)

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "datacenter"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "192.168.1.50",
		Namespace:             "default",
		CollectDeviceMetadata: false, // Metric mode
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("dell-poweredge")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:192.168.1.50", "snmp_profile:dell-poweredge", "device_vendor:dell"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Test final tag/field generation (Metric mode: collectMeta = false)
	var metaData deviceMetaData
	metaData.collectMeta = false // Metric mode
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	// Verify no device_meta field in Metric mode
	for _, data := range fts.Data {
		_, hasDeviceMeta := data.Fields["device_meta"]
		assert.False(t, hasDeviceMeta, "Metric mode should not have device_meta field")
		_, hasAllField := data.Fields["all"]
		assert.False(t, hasAllField, "Metric mode should not have 'all' field")
	}

	pts := ipt.CollectingMeasurements("192.168.1.50", deviceInfo, false)
	t.Logf("Final points count (Metric mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Metric): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_GenericRouter tests the complete flow for generic-router profile
// This tests: extends (generic-device), basic router functionality
func TestProfileIntegration_GenericRouter(t *testing.T) {
	t.Skip("skipping generic-router test")

	// Load the generic-router profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"generic-router": {
			DefinitionFile: "generic-router.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "generic-router")

	profileDef := profiles["generic-router"]

	// Verify profile loaded correctly with extends
	assert.NotEmpty(t, profileDef.Metrics, "should have metrics from generic-device")
	assert.NotEmpty(t, profileDef.Metadata, "should have metadata from generic-device")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml via generic-device)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "generic-router-01")     // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "Generic Router Device") // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.2.1.1.0")       // sysObjectID

	// Set up interface table data (from _generic-if.yaml via generic-device)
	// Interface 1
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.1.1", 1)          // ifIndex
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.2.1", "eth0")     // ifDescr
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.7.1", 1)          // ifAdminStatus (1=up)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.8.1", 1)          // ifOperStatus (1=up)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.5.1", 1000000000) // ifSpeed (1Gbps)
	// Add ifName for interface tag (from ifXTable)
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.1.1", "eth0") // ifName
	// Add ifHCInOctets and ifHCOutOctets (from ifXTable, these are in the profile)
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.6.1", 1000000)  // ifHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.10.1", 2000000) // ifHCOutOctets
	// Interface 2
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.1.2", 2)          // ifIndex
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.2.2", "eth1")     // ifDescr
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.7.2", 2)          // ifAdminStatus (2=down)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.8.2", 2)          // ifOperStatus (2=down)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.5.2", 1000000000) // ifSpeed
	// Add ifName for interface tag (from ifXTable)
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.1.2", "eth1") // ifName
	// Add ifHCInOctets and ifHCOutOctets (from ifXTable)
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.6.2", 500000)   // ifHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.10.2", 1000000) // ifHCOutOctets

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "test"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "172.16.0.1",
		Namespace:             "default",
		CollectDeviceMetadata: true,
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("generic-router")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:172.16.0.1", "snmp_profile:generic-router"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Verify interface metrics were collected
	// Note: _generic-if.yaml defines ifHCInOctets and ifHCOutOctets (not ifInOctets/ifOutOctets)
	interfaceMetricFound := false
	for _, metric := range metricData.Data {
		if metric.Name == "ifHCInOctets" || metric.Name == "ifHCOutOctets" {
			interfaceMetricFound = true
			// Verify metric has tags
			assert.Greater(t, len(metric.Tags), 0, "interface metric should have tags")
			// Verify interface tag is present
			hasInterfaceTag := false
			for _, tag := range metric.Tags {
				if tag == "interface:eth0" || tag == "interface:eth1" {
					hasInterfaceTag = true
					break
				}
			}
			assert.True(t, hasInterfaceTag, "should have interface tag")
		}
	}
	assert.True(t, interfaceMetricFound, "should have found interface metrics")

	// Test metadata reporting
	var metaData deviceMetaData
	metaData.collectMeta = true
	deviceInfo.ReportNetworkDeviceMetadata(values, tags, deviceInfo.Metadata, time.Now(), snmputil.DeviceStatusReachable, &metaData)

	// Verify metadata was collected
	assert.NotEmpty(t, metaData.data, "should have collected metadata")

	// Test final tag/field generation
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	pts := ipt.CollectingMeasurements("172.16.0.1", deviceInfo, true)
	t.Logf("Final points count (Object mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Object): %s", i, pt.MustLPPoint().String())
	}
}

// TestProfileIntegration_GenericRouter_MetricMode tests the complete flow for generic-router profile in Metric mode
func TestProfileIntegration_GenericRouter_MetricMode(t *testing.T) {
	t.Skip("skipping generic-router metric mode test")

	// Load the generic-router profile
	datakit.ConfdDir = "/usr/local/datakit/conf.d/"
	profiles, err := snmputil.LoadProfiles(snmputil.ProfileConfigMap{
		"generic-router": {
			DefinitionFile: "generic-router.yaml",
		},
	})
	require.NoError(t, err)
	require.Contains(t, profiles, "generic-router")

	// Create FakeSession with test data
	fakeSession := snmputil.CreateFakeSession()

	// Set up scalar values (from _base.yaml via generic-device)
	fakeSession.SetStr("1.3.6.1.2.1.1.5.0", "generic-router-01")     // sysName
	fakeSession.SetStr("1.3.6.1.2.1.1.1.0", "Generic Router Device") // sysDescr
	fakeSession.SetObj("1.3.6.1.2.1.1.2.0", "1.3.6.1.2.1.1.0")       // sysObjectID

	// Set up interface table data (from _generic-if.yaml via generic-device)
	// Interface 1
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.1.1", 1)          // ifIndex
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.2.1", "eth0")     // ifDescr
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.7.1", 1)          // ifAdminStatus (1=up)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.8.1", 1)          // ifOperStatus (1=up)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.5.1", 1000000000) // ifSpeed (1Gbps)
	// Add ifName for interface tag (from ifXTable)
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.1.1", "eth0") // ifName
	// Add ifHCInOctets and ifHCOutOctets (from ifXTable, these are in the profile)
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.6.1", 1000000)  // ifHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.10.1", 2000000) // ifHCOutOctets
	// Interface 2
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.1.2", 2)          // ifIndex
	fakeSession.SetStr("1.3.6.1.2.1.2.2.1.2.2", "eth1")     // ifDescr
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.7.2", 2)          // ifAdminStatus (2=down)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.8.2", 2)          // ifOperStatus (2=down)
	fakeSession.SetInt("1.3.6.1.2.1.2.2.1.5.2", 1000000000) // ifSpeed
	// Add ifName for interface tag (from ifXTable)
	fakeSession.SetStr("1.3.6.1.2.1.31.1.1.1.1.2", "eth1") // ifName
	// Add ifHCInOctets and ifHCOutOctets (from ifXTable)
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.6.2", 500000)   // ifHCInOctets
	fakeSession.SetInt("1.3.6.1.2.1.31.1.1.1.10.2", 1000000) // ifHCOutOctets

	// Create deviceInfo and apply profile
	ipt := &Input{
		Profiles: profiles,
		Tags:     map[string]string{"env": "test"},
	}
	ipt.Tagger = testutils.NewTaggerHost()
	deviceInfo := &deviceInfo{
		Ipt:                   ipt,
		IP:                    "172.16.0.1",
		Namespace:             "default",
		CollectDeviceMetadata: false, // Metric mode
	}
	deviceInfo.Session = &fakeSessionWrapper{session: fakeSession}

	err = deviceInfo.refreshWithProfile("generic-router")
	require.NoError(t, err)

	// Fetch SNMP values
	fetchSession := &fakeSessionWrapper{session: fakeSession}
	fetchOpts := &snmputil.FetchOpts{
		OidConfig:          deviceInfo.OidConfig,
		OidBatchSize:       10,
		BulkMaxRepetitions: 10,
	}
	values, err := snmputil.Fetch(fetchSession, fetchOpts)
	require.NoError(t, err)

	// Test metric reporting
	var metricData snmputil.MetricDatas
	tags := []string{"ip:172.16.0.1", "snmp_profile:generic-router"}
	// Add global metric tags (like snmp_host from _base.yaml)
	globalTags := snmputil.GetCheckInstanceMetricTags(deviceInfo.MetricTags, values)
	tags = append(tags, globalTags...)
	snmputil.ReportMetrics(deviceInfo.Metrics, values, tags, &metricData)

	// Verify metrics were collected
	assert.Greater(t, len(metricData.Data), 0, "should have collected metrics")

	// Test final tag/field generation (Metric mode: collectMeta = false)
	var metaData deviceMetaData
	metaData.collectMeta = false // Metric mode
	fts := &tagFields{}
	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)

	// Verify final output
	assert.Greater(t, len(fts.Data), 0, "should have generated tag/field data")

	// Verify no device_meta field in Metric mode
	for _, data := range fts.Data {
		_, hasDeviceMeta := data.Fields["device_meta"]
		assert.False(t, hasDeviceMeta, "Metric mode should not have device_meta field")
		_, hasAllField := data.Fields["all"]
		assert.False(t, hasAllField, "Metric mode should not have 'all' field")
	}

	pts := ipt.CollectingMeasurements("172.16.0.1", deviceInfo, false)
	t.Logf("Final points count (Metric mode): %d", len(pts))
	for i, pt := range pts {
		t.Logf("Point[%d] (Metric): %s", i, pt.MustLPPoint().String())
	}
}
