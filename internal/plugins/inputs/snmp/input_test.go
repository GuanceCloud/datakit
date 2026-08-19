// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmp

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils"
	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmpmeasurement"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmprefiles"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/traps"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func TestCollectTopologyConfig(t *testing.T) {
	ipt := &Input{}
	assert.False(t, ipt.CollectTopology)

	_, err := toml.Decode("collect_topology = true", ipt)
	require.NoError(t, err)
	assert.True(t, ipt.CollectTopology)
	assert.Contains(t, ipt.SampleConfig(), "collect_topology = false")
}

func TestObjectCollectionReportsEmptyTopologyLinks(t *testing.T) {
	session := snmputil.CreateFakeSession()
	session.SetTime("1.3.6.1.2.1.1.3.0", 100)

	ipt := &Input{
		CollectTopology:    true,
		DeviceNamespace:    "prod",
		OIDBatchSize:       5,
		BulkMaxRepetitions: 10,
		Tagger:             testutils.NewTaggerHost(),
		Metrics:            []snmputil.MetricsConfig{snmputil.UptimeMetricConfig},
	}
	ipt.Metadata = snmputil.UpdateMetadataDefinitionWithLegacyFallback(nil)
	ipt.OidConfig.AddScalarOids(snmputil.ParseScalarOids(ipt.Metrics, nil, ipt.Metadata, true))
	ipt.OidConfig.AddColumnOids(snmputil.ParseColumnOids(ipt.Metrics, ipt.Metadata, true))
	device := NewDeviceInfo(ipt, "192.0.2.12", "prod", "", session)

	points := ipt.CollectingMeasurements("192.0.2.12", device, true)
	require.Len(t, points, 1)
	assert.Equal(t, "[]", points[0].Get("links"))
	assert.Equal(t, "prod", points[0].GetTag(deviceNamespaceTagKey))
}

func TestCustomTagsCannotOverrideDeviceNamespace(t *testing.T) {
	session := snmputil.CreateFakeSession()
	session.SetTime("1.3.6.1.2.1.1.3.0", 100)

	ipt := &Input{
		DeviceNamespace:    "prod",
		OIDBatchSize:       5,
		BulkMaxRepetitions: 10,
		Tagger:             testutils.NewTaggerHost(),
		Metrics:            []snmputil.MetricsConfig{snmputil.UptimeMetricConfig},
		Tags:               map[string]string{deviceNamespaceTagKey: "custom"},
		ptsTime:            time.Now(),
	}
	ipt.Metadata = snmputil.UpdateMetadataDefinitionWithLegacyFallback(nil)
	ipt.OidConfig.AddScalarOids(snmputil.ParseScalarOids(ipt.Metrics, nil, ipt.Metadata, true))
	ipt.OidConfig.AddColumnOids(snmputil.ParseColumnOids(ipt.Metrics, ipt.Metadata, true))
	device := NewDeviceInfo(ipt, "192.0.2.12", "prod", "", session)

	points := ipt.CollectingMeasurements("192.0.2.12", device, false)
	require.Len(t, points, 1)
	assert.Equal(t, "prod", points[0].GetTag(deviceNamespaceTagKey))
}

func TestObjectCollectionReportsTopologyLinks(t *testing.T) {
	session := snmputil.CreateFakeSession()
	session.SetTime("1.3.6.1.2.1.1.3.0", 100)

	// IF-MIB metadata for the resolved local interface.
	session.SetStr("1.3.6.1.2.1.31.1.1.1.1.12", "Ethernet1/12")
	session.SetStr("1.3.6.1.2.1.2.2.1.2.12", "Ethernet1/12")
	session.SetStr("1.3.6.1.2.1.31.1.1.1.18.12", "uplink")
	session.SetByte("1.3.6.1.2.1.2.2.1.6.12", []byte{0x82, 0xa5, 0x6e, 0xa5, 0xc9, 0x01})
	session.SetInt("1.3.6.1.2.1.2.2.1.3.12", 6)
	session.SetInt("1.3.6.1.2.1.2.2.1.7.12", 1)
	session.SetInt("1.3.6.1.2.1.2.2.1.8.12", 1)

	// LLDP local port 7 resolves to IF-MIB interface 12 by MAC address.
	session.SetInt("1.0.8802.1.1.2.1.3.7.1.2.7", 3)
	session.SetByte("1.0.8802.1.1.2.1.3.7.1.3.7", []byte{0x82, 0xa5, 0x6e, 0xa5, 0xc9, 0x01})

	// LLDP remote entry index: timeMark.localPortNum.remIndex = 0.7.1.
	session.SetInt("1.0.8802.1.1.2.1.4.1.1.4.0.7.1", 4)
	session.SetByte("1.0.8802.1.1.2.1.4.1.1.5.0.7.1", []byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x02})
	session.SetInt("1.0.8802.1.1.2.1.4.1.1.6.0.7.1", 5)
	session.SetStr("1.0.8802.1.1.2.1.4.1.1.7.0.7.1", "Ethernet1/7")
	session.SetStr("1.0.8802.1.1.2.1.4.1.1.8.0.7.1", "remote uplink")
	session.SetStr("1.0.8802.1.1.2.1.4.1.1.9.0.7.1", "switch-b")
	session.SetStr("1.0.8802.1.1.2.1.4.1.1.10.0.7.1", "remote switch")
	// The management address is encoded in the lldpRemManAddrEntry index.
	session.SetInt("1.0.8802.1.1.2.1.4.2.1.3.0.7.1.1.4.10.250.0.6", 2)
	// CDP is also present, but LLDP must take precedence.
	session.SetStr("1.3.6.1.4.1.9.9.23.1.2.1.1.6.12.1", "ignored-cdp-device")
	session.SetStr("1.3.6.1.4.1.9.9.23.1.2.1.1.7.12.1", "Ethernet1/1")

	ipt := &Input{
		CollectTopology:    true,
		DeviceNamespace:    "prod",
		OIDBatchSize:       5,
		BulkMaxRepetitions: 10,
		Tagger:             testutils.NewTaggerHost(),
		Metrics:            []snmputil.MetricsConfig{snmputil.UptimeMetricConfig},
	}
	ipt.Metadata = snmputil.UpdateMetadataDefinitionWithLegacyFallback(nil)
	ipt.OidConfig.AddScalarOids(snmputil.ParseScalarOids(ipt.Metrics, nil, ipt.Metadata, true))
	ipt.OidConfig.AddColumnOids(snmputil.ParseColumnOids(ipt.Metrics, ipt.Metadata, true))
	device := NewDeviceInfo(ipt, "192.0.2.10", "prod", "", session)

	points := ipt.CollectingMeasurements("192.0.2.10", device, true)
	require.Len(t, points, 1)

	linksValue, ok := points[0].Get("links").(string)
	require.True(t, ok)
	var links []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(linksValue), &links))
	require.Len(t, links, 1)
	require.JSONEq(t, `{
		"id":"prod:192.0.2.10:7.1",
		"source_type":"lldp",
		"integration":"snmp",
		"local":{
			"device":{"resolved_id":"prod:192.0.2.10"},
			"interface":{"resolved_id":"prod:192.0.2.10:12","id":"82:a5:6e:a5:c9:01","id_type":"mac_address"}
		},
		"remote":{
			"device":{"id":"01:00:00:00:01:02","id_type":"mac_address","name":"switch-b","description":"remote switch","ip_address":"10.250.0.6"},
			"interface":{"id":"Ethernet1/7","id_type":"interface_name","description":"remote uplink"}
		}
	}`, mustJSON(t, links[0]))
	deviceMeta, ok := points[0].Get("device_meta").(string)
	require.True(t, ok)
	assert.NotContains(t, deviceMeta, `"links"`)
	assert.Equal(t, gosnmp.Version2c, session.GetVersion())
}

func TestObjectCollectionReportsCDPTopologyLinksWhenLLDPIsUnavailable(t *testing.T) {
	session := snmputil.CreateFakeSession()
	session.SetTime("1.3.6.1.2.1.1.3.0", 100)

	// CDP cache entry index: cdpCacheIfIndex.cdpCacheDeviceIndex = 12.3.
	session.SetStr("1.3.6.1.4.1.9.9.23.1.2.1.1.5.12.3", "Cisco IOS")
	session.SetStr("1.3.6.1.4.1.9.9.23.1.2.1.1.6.12.3", "switch-c-device-id")
	session.SetStr("1.3.6.1.4.1.9.9.23.1.2.1.1.7.12.3", "Ethernet1/1")
	session.SetStr("1.3.6.1.4.1.9.9.23.1.2.1.1.17.12.3", "switch-c")
	session.SetInt("1.3.6.1.4.1.9.9.23.1.2.1.1.19.12.3", 1)
	session.SetByte("1.3.6.1.4.1.9.9.23.1.2.1.1.20.12.3", []byte{10, 251, 0, 7})

	ipt := &Input{
		CollectTopology:    true,
		DeviceNamespace:    "prod",
		OIDBatchSize:       5,
		BulkMaxRepetitions: 10,
		Tagger:             testutils.NewTaggerHost(),
		Metrics:            []snmputil.MetricsConfig{snmputil.UptimeMetricConfig},
	}
	ipt.Metadata = snmputil.UpdateMetadataDefinitionWithLegacyFallback(nil)
	ipt.OidConfig.AddScalarOids(snmputil.ParseScalarOids(ipt.Metrics, nil, ipt.Metadata, true))
	ipt.OidConfig.AddColumnOids(snmputil.ParseColumnOids(ipt.Metrics, ipt.Metadata, true))
	device := NewDeviceInfo(ipt, "192.0.2.11", "prod", "", session)

	points := ipt.CollectingMeasurements("192.0.2.11", device, true)
	require.Len(t, points, 1)

	linksValue, ok := points[0].Get("links").(string)
	require.True(t, ok)
	var links []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(linksValue), &links))
	require.Len(t, links, 1)
	require.JSONEq(t, `{
		"id":"prod:192.0.2.11:12.3",
		"source_type":"cdp",
		"integration":"snmp",
		"local":{
			"device":{"resolved_id":"prod:192.0.2.11"},
			"interface":{"resolved_id":"prod:192.0.2.11:12","id":""}
		},
		"remote":{
			"device":{"id":"switch-c-device-id","name":"switch-c","description":"Cisco IOS","ip_address":"10.251.0.7"},
			"interface":{"id":"Ethernet1/1","id_type":"interface_name"}
		}
	}`, mustJSON(t, links[0]))
}

func mustJSON(t *testing.T, value interface{}) string {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return string(data)
}

func TestMergeInterfacesByInterfacePreservesIndex(t *testing.T) {
	interfaces := []*interfaceAttribute{
		{
			Interface: "Ethernet1/12",
			Fields:    map[string]interface{}{"ifHCInOctets": float64(100)},
		},
		{
			Interface:      "Ethernet1/12",
			InterfaceIndex: "12",
			Fields:         map[string]interface{}{"ifHCOutOctets": float64(200)},
		},
	}

	merged := mergeInterfacesByInterface(interfaces)
	require.Len(t, merged, 1)
	assert.Equal(t, "12", merged[0].InterfaceIndex)
	assert.Equal(t, float64(100), merged[0].Fields["ifHCInOctets"])
	assert.Equal(t, float64(200), merged[0].Fields["ifHCOutOctets"])
}

// go test -v -timeout 30s -run ^Test_AvailableArchs$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_AvailableArchs(t *testing.T) {
	ipt := &Input{}
	out := ipt.AvailableArchs()
	assert.Equal(t, datakit.AllOS, out)
}

// go test -v -timeout 30s -run ^Test_SampleMeasurement$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_SampleMeasurement(t *testing.T) {
	ipt := &Input{}
	out := ipt.SampleMeasurement()
	assert.Equal(t, []inputs.Measurement{
		&snmpmeasurement.SNMPObject{},
		&snmpmeasurement.SNMPMetric{},
		&snmpmeasurement.SNMPLLDP{},
	}, out)
}

func setupTrapServerTest(t *testing.T) {
	t.Helper()

	oldConfdDir := datakit.ConfdDir
	datakit.ConfdDir = t.TempDir()
	traps.StopServer()
	t.Cleanup(func() {
		traps.StopServer()
		datakit.ConfdDir = oldConfdDir
	})
	require.NoError(t, snmprefiles.ReleaseFiles())
}

func getFreeUDPPorts(t *testing.T, count int) []uint16 {
	t.Helper()

	connections := make([]*net.UDPConn, 0, count)
	ports := make([]uint16, 0, count)
	for i := 0; i < count; i++ {
		conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
		require.NoError(t, err)
		connections = append(connections, conn)
		ports = append(ports, uint16(conn.LocalAddr().(*net.UDPAddr).Port))
	}

	for _, conn := range connections {
		require.NoError(t, conn.Close())
	}
	return ports
}

func startTestTrapServer(t *testing.T, port uint16) *traps.TrapServer {
	t.Helper()

	server, err := traps.StartServer(&traps.TrapsServerOpt{
		Enabled:          true,
		BindHost:         "127.0.0.1",
		Port:             port,
		CommunityStrings: []string{"public"},
		StopTimeout:      1,
	})
	require.NoError(t, err)
	require.NotNil(t, server)
	return server
}

func assertUDPPortAvailable(t *testing.T, port uint16) {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: int(port),
	})
	require.NoError(t, err)
	require.NoError(t, conn.Close())
}

func assertUDPPortInUse(t *testing.T, port uint16) {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: int(port),
	})
	if conn != nil {
		require.NoError(t, conn.Close())
	}
	require.ErrorIs(t, err, syscall.EADDRINUSE)
}

func TestRunDoesNotStartTrapWhenInitializationFails(t *testing.T) {
	setupTrapServerTest(t)
	port := getFreeUDPPorts(t, 1)[0]

	ipt := defaultInput()
	ipt.SNMPVersion = 2
	ipt.ZabbixProfiles = []*snmputil.ZabbixProfile{}
	ipt.Traps = TrapsConfig{
		Enable:      true,
		BindHost:    "127.0.0.1",
		Port:        port,
		StopTimeout: 1,
	}
	ipt.Run()

	assertUDPPortAvailable(t, port)
}

func TestTerminateBeforeStartDoesNotStartTrapServer(t *testing.T) {
	setupTrapServerTest(t)
	port := getFreeUDPPorts(t, 1)[0]

	ipt := defaultInput()
	ipt.Traps = TrapsConfig{
		Enable:      true,
		BindHost:    "127.0.0.1",
		Port:        port,
		StopTimeout: 1,
	}

	ipt.Terminate()
	ipt.startTrapServer()

	require.True(t, ipt.trapTerminated)
	require.Nil(t, ipt.trapServer)
	assertUDPPortAvailable(t, port)
}

func TestInputExitStopsOnlyOwnedTrapServer(t *testing.T) {
	setupTrapServerTest(t)
	ports := getFreeUDPPorts(t, 2)

	firstInput := defaultInput()
	firstInput.trapServer = startTestTrapServer(t, ports[0])
	secondInput := defaultInput()
	secondInput.trapServer = startTestTrapServer(t, ports[1])

	firstInput.exit()
	require.Nil(t, firstInput.trapServer)
	assertUDPPortAvailable(t, ports[0])
	assertUDPPortInUse(t, ports[1])

	secondInput.exit()
	require.Nil(t, secondInput.trapServer)
	assertUDPPortAvailable(t, ports[1])
}

// go test -v -timeout 30s -run ^Test_calcTagsHash$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_calcTagsHash(t *testing.T) {
	cases := []struct {
		name string
		in   *snmputil.MetricDatas
		out  *snmputil.MetricDatas
	}{
		{
			name: "normal",
			in: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "",
					},
				},
			},
			out: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "e80b5017098950fc58aad83c8c14978e",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "e80b5017098950fc58aad83c8c14978e",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "972d3055287ad2a9f007321eaa601c54",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "972d3055287ad2a9f007321eaa601c54",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "972d3055287ad2a9f007321eaa601c54",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calcTagsHash(tc.in)
			assert.Equal(t, tc.out, tc.in)
		})
	}
}

// go test -v -timeout 30s -run ^Test_aggregateDeviceDataWithTagsIgnore$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_aggregateDeviceDataWithTagsIgnore(t *testing.T) {
	ipt := &Input{
		TagsIgnore: []string{"interface_index"},
		TagsIgnoreRule: []*regexp.Regexp{
			regexp.MustCompile(`^volatile_`),
		},
	}

	metricData := &snmputil.MetricDatas{
		Data: []*snmputil.MetricData{
			{
				Name:  "ifOperStatus",
				Value: 1.0,
				Tags: []string{
					"ip:192.168.1.1",
					"interface:eth0",
					"interface_index:100",
					"volatile_tag:first",
				},
			},
			{
				Name:  "ifOperStatus",
				Value: 2.0,
				Tags: []string{
					"ip:192.168.1.1",
					"interface:eth0",
					"interface_index:200",
					"volatile_tag:second",
				},
			},
		},
	}

	fts := &tagFields{}
	aggregateDeviceData(metricData, fts, &deviceMetaData{}, nil, ipt)

	require.Len(t, fts.Data, 1)
	assert.NotContains(t, fts.Data[0].Tags, "interface_index")
	assert.NotContains(t, fts.Data[0].Tags, "volatile_tag")
	assert.Equal(t, "192.168.1.1", fts.Data[0].Tags["ip"])
	assert.Equal(t, "eth0", fts.Data[0].Tags["interface"])
	assert.Equal(t, 2.0, fts.Data[0].Fields["ifOperStatus"])
}

func Test_filterMetricDataTagsWithSharedTags(t *testing.T) {
	ipt := &Input{
		TagsIgnore: []string{"interface_index"},
		TagsIgnoreRule: []*regexp.Regexp{
			regexp.MustCompile(`^volatile_`),
		},
	}

	sharedTags := []string{
		"ip:192.168.1.1",
		"volatile_tag:temporary",
		"interface:eth0",
		"interface_index:100",
	}
	metricData := &snmputil.MetricDatas{
		Data: []*snmputil.MetricData{
			{Name: "ifOperStatus", Value: 1.0, Tags: sharedTags},
			{Name: "ifAdminStatus", Value: 1.0, Tags: sharedTags},
		},
	}

	ipt.filterMetricDataTags(metricData)

	wantTags := []string{"ip:192.168.1.1", "interface:eth0"}
	require.Equal(t, wantTags, metricData.Data[0].Tags)
	require.Equal(t, wantTags, metricData.Data[1].Tags)
}

// go test -v -timeout 30s -run ^Test_aggregateHash$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_aggregateHash(t *testing.T) {
	cases := []struct {
		name       string
		metricData *snmputil.MetricDatas
		in         map[string]map[string]interface{}
		out        map[string]map[string]interface{}
	}{
		{
			name: "normal",
			metricData: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
				},
			},
			in: make(map[string]map[string]interface{}),
			out: map[string]map[string]interface{}{
				"fake_hash_1": {
					"key1": float64(1.0),
					"key2": float64(2.0),
				},
				"fake_hash_2": {
					"key3": float64(3.0),
					"key4": float64(4.0),
					"key5": float64(5.0),
				},
			},
		},
		{
			name: "larger",
			metricData: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc", "def"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    5.0,
						Tags:     []string{"abc", "def", "apple"},
						TagsHash: "fake_hash_2",
					},
				},
			},
			in: make(map[string]map[string]interface{}),
			out: map[string]map[string]interface{}{
				"fake_hash_1": {
					"key1": float64(1.0),
					"key2": float64(2.0),
				},
				"fake_hash_2": {
					"key3": float64(3.0),
					"key4": float64(5.0),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			aggregateHash(tc.metricData, tc.in)
			assert.Equal(t, tc.out, tc.in)
		})
	}
}

// go test -v -timeout 30s -count=1 -run ^Test_getFieldTagArr$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_getFieldTagArr(t *testing.T) {
	cases := []struct {
		name       string
		metricData *snmputil.MetricDatas
		mHash      map[string]map[string]interface{}
		metaData   *deviceMetaData
		origTags   []string
		out        tagFields
	}{
		{
			name: "empty_hash",
		},
		{
			name: "collect_meta",
			metricData: &snmputil.MetricDatas{
				Data: []*snmputil.MetricData{
					{
						Name:     "key1",
						Value:    1.0,
						Tags:     []string{"abc:value1", "def:value2"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key2",
						Value:    2.0,
						Tags:     []string{"abc:value1", "def:value2"},
						TagsHash: "fake_hash_1",
					},
					{
						Name:     "key3",
						Value:    3.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key4",
						Value:    4.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
					{
						Name:     "key5",
						Value:    5.0,
						Tags:     []string{"abc:value1", "def:value2", "apple:value3"},
						TagsHash: "fake_hash_2",
					},
				},
			},
			mHash: map[string]map[string]interface{}{
				"fake_hash_1": {
					"key1": float64(1.0),
					"key2": float64(2.0),
				},
				"fake_hash_2": {
					"key3": float64(3.0),
					"key4": float64(4.0),
					"key5": float64(5.0),
				},
			},
			metaData: &deviceMetaData{
				collectMeta: true,
				data: []string{
					"fruit1=banana",
					"fruit2=pear",
					"fruit3=tomato",
				},
			},
			origTags: []string{
				"device_namespace:default",
				"snmp_device:192.168.1.100",
			},
			out: tagFields{
				Data: []*tagField{
					{
						Tags: map[string]string{
							"name":          "",
							"host":          "",
							"device_type":   "",
							"device_vendor": "",
						},
						Fields: map[string]interface{}{
							"interfaces":     "null",
							"sensors":        "null",
							"mems":           "null",
							"mem_pool_names": "null",
							"cpus":           "null",
							"all":            `[{"tags":{"abc":"value1","apple":"value3","def":"value2"},"fields":{"key3":3,"key4":4,"key5":5}},{"tags":{"abc":"value1","def":"value2"},"fields":{"key1":1,"key2":2}}]`,
							"device_meta":    "fruit1=banana, fruit2=pear, fruit3=tomato",
							"uptime":         float64(0),
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fts := tagFields{}
			getFieldTagArr(tc.metricData, tc.mHash, &fts, tc.metaData, tc.origTags, &Input{})
			for _, v := range tc.out.Data {
				foundIdx := -1
				for kk, vv := range fts.Data {
					// resF := reflect.DeepEqual(v.Fields, vv.Fields)

					resF1 := reflect.DeepEqual(v.Fields["interfaces"], vv.Fields["interfaces"])
					resF2 := reflect.DeepEqual(v.Fields["sensors"], vv.Fields["sensors"])
					resF3 := reflect.DeepEqual(v.Fields["mems"], vv.Fields["mems"])
					resF4 := reflect.DeepEqual(v.Fields["mem_pool_names"], vv.Fields["mem_pool_names"])
					resF5 := reflect.DeepEqual(v.Fields["cpus"], vv.Fields["cpus"])
					// resF6 := reflect.DeepEqual(v.Fields["all"], vv.Fields["all"])
					resF7 := reflect.DeepEqual(v.Fields["device_meta"], vv.Fields["device_meta"])
					resF8 := reflect.DeepEqual(v.Fields["uptime"], vv.Fields["uptime"])

					resT := reflect.DeepEqual(v.Tags, vv.Tags)
					if resT && resF1 && resF2 && resF3 && resF4 && resF5 && resF7 && resF8 {
						foundIdx = kk
						break
					}
				} // for
				assert.NotEqual(t, -1, foundIdx) // must found, so cannot be -1.
			}
		})
	}
}

// go test -v -timeout 30s -run ^Test_getDatakitStyleTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_getDatakitStyleTags(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		out  map[string]string
	}{
		{
			name: agentHostKey,
			in: []string{
				agentHostKey + ":apple",
			},
			out: map[string]string{},
		},
		{
			name: agentVersionKey,
			in: []string{
				agentVersionKey + ":apple",
			},
			out: map[string]string{},
		},
		{
			name: "snmp_host",
			in: []string{
				"snmp_host" + ":apple",
			},
			out: map[string]string{
				"snmp_host": "apple",
			},
		},
		{
			name: "other",
			in: []string{
				"whatever:apple",
			},
			out: map[string]string{
				"whatever": "apple",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := make(map[string]string)
			getDatakitStyleTags(tc.in, out)
			assert.Equal(t, tc.out, out)
		})
	}
}

// go test -v -timeout 30s -run ^Test_validateConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_validateConfig(t *testing.T) {
	cases := []struct {
		name                string
		version             uint8
		autoDiscovery       []string
		discoveryIgnoredIPs []string
		specificDevices     []string
		err                 error
	}{
		{
			name: "snmp version error",
			err:  fmt.Errorf("`snmp_version` must be 1 or 2 or 3"),
		},
		{
			name:    "version only",
			version: 2,
		},
		{
			name:          "CIDR error",
			version:       2,
			autoDiscovery: []string{"", "192.168.1.100"},
			err:           &net.ParseError{Type: "CIDR address", Text: "192.168.1.100"},
		},
		{
			name:            "invalid IP address",
			version:         2,
			specificDevices: []string{"invalid_ip"},
			err:             fmt.Errorf("invalid IP address"),
		},
		{
			name:                "normal",
			version:             2,
			autoDiscovery:       []string{"", "192.168.1.0/24"},
			discoveryIgnoredIPs: []string{"", "192.168.1.101"},
			specificDevices:     []string{"", "192.168.1.100"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &Input{SNMPVersion: tc.version, AutoDiscovery: tc.autoDiscovery, DiscoveryIgnoredIPs: tc.discoveryIgnoredIPs, SpecificDevices: tc.specificDevices} //nolint:lll
			err := ipt.ValidateConfig()
			assert.Equal(t, tc.err, err)
		})
	}
}

func Test_validateConfig_snmpBatchDefaults(t *testing.T) {
	ipt := &Input{SNMPVersion: 2}
	err := ipt.ValidateConfig()

	assert.NoError(t, err)
	assert.Equal(t, defaultOidBatchSize, ipt.OIDBatchSize)
	assert.Equal(t, defaultBulkMaxRepetitions, ipt.BulkMaxRepetitions)
}

func Test_validateConfig_snmpBatchOverrides(t *testing.T) {
	ipt := &Input{
		SNMPVersion:        2,
		OIDBatchSize:       60,
		BulkMaxRepetitions: 20,
	}
	err := ipt.ValidateConfig()

	assert.NoError(t, err)
	assert.Equal(t, 60, ipt.OIDBatchSize)
	assert.Equal(t, uint32(20), ipt.BulkMaxRepetitions)
}

func TestCheckIPWorking(t *testing.T) {
	deviceIP1 := "1.2.3.4"
	deviceIP2 := "2.3.4.5"
	mWorkingIP.Store(deviceIP1, COLLECT_OBJECT)
	t.Cleanup(func() {
		mWorkingIP.Delete(deviceIP1)
		mWorkingIP.Delete(deviceIP2)
	})

	ipt := &Input{semStop: cliutils.NewSem()}
	assert.True(t, ipt.checkIPWorking(deviceIP2, COLLECT_METRICS))
	assert.False(t, ipt.checkIPWorking(deviceIP2, COLLECT_METRICS))
	checkIPDone(deviceIP2)

	go func() {
		time.Sleep(time.Second)
		checkIPDone(deviceIP1)
	}()

	assert.True(t, ipt.checkIPWorking(deviceIP1, COLLECT_METRICS))
}

// go test -v -timeout 30s -run ^Test_normalizeFieldTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_normalizeFieldTags(t *testing.T) {
	cases := []struct {
		name string
		in   *tagField
		out  *tagField
	}{
		{
			name: "normal",
			in: &tagField{
				Tags: map[string]string{
					"aaa_a.a": "not_used",
				},
				Fields: map[string]interface{}{
					"aaa_a.a": "not_used",
				},
			},
			out: &tagField{
				Tags: map[string]string{
					"aaa_a_a": "not_used",
				},
				Fields: map[string]interface{}{
					"aaa_a_a": "not_used",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			normalizeFieldTags(tc.in)
			assert.Equal(t, tc.out, tc.in)
		})
	}
}

// go test -v -timeout 30s -run ^Test_replaceMetricsName$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp
func Test_replaceMetricsName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "underline_with_point",
			in:   "aaa_a.a",
			out:  "aaa_a_a",
		},
		{
			name: "underline_without_point",
			in:   "aaa_aa",
			out:  "",
		},
		{
			name: "underline_with_point_last",
			in:   "aaa_a.",
			out:  "aaa_a_",
		},
		{
			name: "CamelCase_with_point",
			in:   "Aaa.a",
			out:  "AaaA",
		},
		{
			name: "CamelCase_with_point_repeat",
			in:   "Aaa.aaa.aa.a",
			out:  "AaaAaaAaA",
		},
		{
			name: "CamelCase_without_point",
			in:   "Aaa",
			out:  "",
		},
		{
			name: "CamelCase_with_point_last",
			in:   "Aaa.",
			out:  "Aaa",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := replaceMetricsName(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}
