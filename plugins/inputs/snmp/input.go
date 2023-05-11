// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package snmp contains snmp collector implement.
package snmp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmpmeasurement"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmprefiles"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmputil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/traps"
)

const (
	// nolint:lll
	sampleCfg = `
[[inputs.snmp]]
  ## Filling in specific device IP address, like ["10.200.10.240", "10.200.10.241"].
  ## And you can use auto_discovery and specific_devices at the same time.
  ## If you don't want to specific device, you don't need provide this.
  #
  # specific_devices = ["***"] # SNMP Device IP.

  ## Filling in autodiscovery CIDR subnet, like ["10.200.10.0/24", "10.200.20.0/24"].
  ## If you don't want to enable autodiscovery feature, you don't need provide this.
  #
  # auto_discovery = ["***"] # Used in autodiscovery mode only, ignore this in other cases.

  ## SNMP protocol version the devices using, fill in 2 or 3.
  ## If you using the version 1, just fill in 2. Version 2 supported version 1.
  ## This is must be provided.
  #
  snmp_version = 2

  ## SNMP port in the devices. Default is 161. In most cases, you don't need change this.
  ## This is optional.
  #
  # port = 161

  ## Password in SNMP v2, enclose with single quote. Only worked in SNMP v2.
  ## If you are using SNMP v2, this is must be provided.
  ## If you are using SNMP v3, you don't need provide this.
  #
  # v2_community_string = "***"

  ## Authentication stuff in SNMP v3.
  ## If you are using SNMP v2, you don't need provide this.
  ## If you are using SNMP v3, this is must be provided.
  #
  # v3_user = "***"
  # v3_auth_protocol = "***"
  # v3_auth_key = "***"
  # v3_priv_protocol = "***"
  # v3_priv_key = "***"
  # v3_context_engine_id = "***"
  # v3_context_name = "***"

  ## Number of workers used to collect and discovery devices concurrently. Default is 100.
  ## Modifying it based on device's number and network scale.
  ## This is optional.
  #
  # workers = 100

  ## Interval between each autodiscovery in seconds. Default is "1h".
  ## Only worked in autodiscovery feature.
  ## This is optional.
  #
  # discovery_interval = "1h"

  ## Filling in excluded device IP address, like ["10.200.10.220", "10.200.10.221"].
  ## Only worked in autodiscovery feature.
  ## This is optional.
  #
  # discovery_ignored_ip = []

  ## Set true to enable election
  #
  # election = true

  ## Device Namespace. Default is "default".
  #
  # device_namespace = "default"

  ## Picking the metric data only contains the field's names below.
  #
  # enable_picking_data = true # Default is "false", which means collecting all data.
  # status = ["sysUpTimeInstance", "tcpCurrEstab", "ifAdminStatus", "ifOperStatus", "cswSwitchState"]
  # speed = ["ifHCInOctets", "ifHCInOctetsRate", "ifHCOutOctets", "ifHCOutOctetsRate", "ifHighSpeed", "ifSpeed", "ifBandwidthInUsageRate", "ifBandwidthOutUsageRate"]
  # cpu = ["cpuUsage"]
  # mem = ["memoryUsed", "memoryUsage", "memoryFree"]
  # extra = []

  [inputs.snmp.tags]
  # tag1 = "val1"
  # tag2 = "val2"

  [inputs.snmp.traps]
  # enable = true
  # bind_host = "0.0.0.0"
  # port = 9162
  # stop_timeout = 3    # stop timeout in seconds.
`  // sampleCfg

	defaultPort              = uint16(161)
	defaultWorkers           = 100
	defaultDiscoveryInterval = time.Hour
	defaultObjectInterval    = 5 * time.Minute
	defaultMetricInterval    = 10 * time.Second

	// Using high oid batch size might lead to snmp calls timing out.
	// For some devices, the default oid_batch_size of 5 might be high (leads to timeouts),
	// and require manual setting oid_batch_size to a lower value.
	defaultOidBatchSize = 5

	// DefaultBulkMaxRepetitions is the default max rep
	// Using too high max repetitions might lead to tooBig SNMP error messages.
	// - Java SNMP and gosnmp (gosnmp.defaultMaxRepetitions) uses 50
	// - snmp-net uses 10.
	defaultBulkMaxRepetitions = uint32(10)

	defaultDeviceNamespace = "default"
	deviceNamespaceTagKey  = "device_namespace"
	deviceIPTagKey         = "snmp_device"
	subnetTagKey           = "autodiscovery_subnet"
	agentHostKey           = "agent_host"
	agentVersionKey        = "agent_version"
)

var (
	// Make sure Input implements the inputs.InputV2 interface.
	_                   inputs.InputV2 = &Input{}
	l                                  = logger.DefaultSLogger(snmpmeasurement.InputName)
	g                                  = datakit.G("inputs_snmp_")
	onceReleasePrefiles sync.Once
	onceSetLog          sync.Once
)

type Input struct {
	AutoDiscovery       []string          `toml:"auto_discovery"`
	SpecificDevices     []string          `toml:"specific_devices"`
	SNMPVersion         uint8             `toml:"snmp_version"`
	Port                uint16            `toml:"port"`
	V2CommunityString   string            `toml:"v2_community_string"`
	V3User              string            `toml:"v3_user"`
	V3AuthProtocol      string            `toml:"v3_auth_protocol"`
	V3AuthKey           string            `toml:"v3_auth_key"`
	V3PrivProtocol      string            `toml:"v3_priv_protocol"`
	V3PrivKey           string            `toml:"v3_priv_key"`
	V3ContextEngineID   string            `toml:"v3_context_engine_id"`
	V3ContextName       string            `toml:"v3_context_name"`
	Workers             int               `toml:"workers"`
	DiscoveryInterval   time.Duration     `toml:"discovery_interval"`
	DiscoveryIgnoredIPs []string          `toml:"discovery_ignored_ip"`
	Tags                map[string]string `toml:"tags"`
	Traps               TrapsConfig       `toml:"traps"`
	Election            bool              `toml:"election"`
	DeviceNamespace     string            `toml:"device_namespace"`
	EnablePickingData   bool              `toml:"enable_picking_data"`
	PickingStatus       []string          `toml:"status"`
	PickingSpeed        []string          `toml:"speed"`
	PickingCPU          []string          `toml:"cpu"`
	PickingMem          []string          `toml:"mem"`
	PickingExtra        []string          `toml:"extra"`
	ObjectInterval      time.Duration     `toml:"object_interval,omitempty"`
	MetricInterval      time.Duration     `toml:"metric_interval,omitempty"`

	Profiles       snmputil.ProfileDefinitionMap
	CustomProfiles snmputil.ProfileConfigMap `toml:"custom_profiles,omitempty"`

	// Those need pass to devices, because they could be changed inside deviceInfo.
	ProfileTags []string
	OidConfig   snmputil.OidConfig
	Profile     string `toml:"profile,omitempty"`
	ProfileDef  *snmputil.ProfileDefinition
	Metadata    snmputil.MetadataConfig
	Metrics     []snmputil.MetricsConfig   `toml:"metrics,omitempty"`     // SNMP metrics definition
	MetricTags  []snmputil.MetricTagConfig `toml:"metric_tags,omitempty"` // SNMP metric tags definition

	semStop              *cliutils.Sem // start stop signal
	mAutoDiscovery       map[string]*discoveryInfo
	mDiscoveryIgnoredIPs map[string]struct{}
	mSpecificDevices     map[string]*deviceInfo
	mDynamicDevices      sync.Map
	mFieldNameSpecified  map[string]struct{}
	jobs                 chan Job
	autodetectProfile    bool
	feeder               io.Feeder
}

type TrapsConfig struct {
	Enable      bool   `toml:"enable"`
	BindHost    string `toml:"bind_host"`
	Port        uint16 `toml:"port"`
	StopTimeout int    `toml:"stop_timeout"`
}

func (*Input) Catalog() string { return snmpmeasurement.InputName }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&snmpmeasurement.SNMPObject{}, &snmpmeasurement.SNMPMetric{}}
}

func (ipt *Input) Run() {
	setLog()
	l.Info("Run entry")

	onceReleasePrefiles.Do(func() {
		if err := snmprefiles.ReleaseFiles(); err != nil {
			l.Errorf("snmp release prefiles failed: %v", err)
		}
	})

	// starting traps server
	if ipt.Traps.Enable {
		var communityStrings []string
		if len(ipt.V2CommunityString) > 0 {
			communityStrings = []string{ipt.V2CommunityString}
		}
		var v3 []traps.UserV3
		if len(ipt.V3User) > 0 {
			v3 = []traps.UserV3{
				{
					Username:     ipt.V3User,
					AuthKey:      ipt.V3AuthKey,
					AuthProtocol: ipt.V3AuthProtocol,
					PrivKey:      ipt.V3PrivKey,
					PrivProtocol: ipt.V3PrivProtocol,
				},
			}
		}
		if err := traps.StartServer(&traps.TrapsServerOpt{
			Enabled:          ipt.Traps.Enable,
			BindHost:         ipt.Traps.BindHost,
			Port:             ipt.Traps.Port,
			Namespace:        ipt.DeviceNamespace,
			CommunityStrings: communityStrings,
			Users:            v3,
			StopTimeout:      ipt.Traps.StopTimeout,
			Election:         ipt.Election,
		}); err != nil {
			l.Errorf("traps.StartServer failed: %v, port = %d", err, ipt.Traps.Port)
			return
		}
	}

	// starting snmp collecting
	ipt.jobs = make(chan Job)

	if err := ipt.ValidateConfig(); err != nil {
		l.Errorf("validateConfig failed: %v", err)
		return
	}

	if err := ipt.Initialize(); err != nil {
		l.Errorf("initialize failed: %v", err)
		return
	}

	tickerObject := time.NewTicker(ipt.ObjectInterval)
	tickerMetric := time.NewTicker(ipt.MetricInterval)
	tickerDiscovery := time.NewTicker(ipt.DiscoveryInterval)

	workerFunc := func() {
		g.Go(func(ctx context.Context) error {
			for {
				select {
				case job := <-ipt.jobs:
					ipt.doJob(job)

				case <-datakit.Exit.Wait():
					l.Info(snmpmeasurement.InputName + " exit")
					return nil

				case <-ipt.semStop.Wait():
					l.Infof(snmpmeasurement.InputName + " return")
					return nil
				}
			}
		})
	}

	length := 0
	if len(ipt.mAutoDiscovery) > 0 {
		length = ipt.Workers
	} else {
		length = len(ipt.mSpecificDevices)
	}

	for w := 0; w < length; w++ {
		workerFunc()
	}

	ipt.autoDiscovery()
	ipt.collectObject()
	ipt.collectMetrics()

	for {
		select {
		case <-tickerObject.C:
			ipt.collectObject()

		case <-tickerMetric.C:
			ipt.collectMetrics()

		case <-tickerDiscovery.C:
			ipt.autoDiscovery()

		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info(snmpmeasurement.InputName + " exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof(snmpmeasurement.InputName + " return")
			return
		}
	}
}

func (ipt *Input) collectObject() {
	// send specific devices
	for deviceIP, device := range ipt.mSpecificDevices {
		ipt.jobs <- Job{
			ID:     COLLECT_OBJECT,
			IP:     deviceIP,
			Device: device,
		}
	}

	// send dynamic devices
	ipt.mDynamicDevices.Range(func(k, v interface{}) bool {
		deviceIP, ok := k.(string)
		if !ok {
			l.Errorf("should not be here")
			return true
		}
		device, ok := v.(*deviceInfo)
		if !ok {
			l.Errorf("should not be here")
			return true
		}

		ipt.jobs <- Job{
			ID:     COLLECT_OBJECT,
			IP:     deviceIP,
			Device: device,
		}
		return true
	})
}

func (ipt *Input) collectMetrics() {
	// send specific devices
	for deviceIP, device := range ipt.mSpecificDevices {
		ipt.jobs <- Job{
			ID:     COLLECT_METRICS,
			IP:     deviceIP,
			Device: device,
		}
	}

	// send dynamic devices
	ipt.mDynamicDevices.Range(func(k, v interface{}) bool {
		deviceIP, ok := k.(string)
		if !ok {
			l.Errorf("should not be here")
			return true
		}
		device, ok := v.(*deviceInfo)
		if !ok {
			l.Errorf("should not be here")
			return true
		}

		ipt.jobs <- Job{
			ID:     COLLECT_METRICS,
			IP:     deviceIP,
			Device: device,
		}
		return true
	})
}

func (ipt *Input) autoDiscovery() {
	if len(ipt.mAutoDiscovery) == 0 {
		return
	}

	mSpecificDevices := make(map[string]struct{}, len(ipt.SpecificDevices))
	for deviceIP := range ipt.mSpecificDevices {
		mSpecificDevices[deviceIP] = struct{}{}
	}

	g.Go(func(ctx context.Context) error {
		for subnet, discovery := range ipt.mAutoDiscovery {
			ipt.dispatchDiscovery(subnet, discovery, mSpecificDevices)

			select {
			case <-datakit.Exit.Wait():
				l.Debugf("subnet %s: Stop scheduling devices, exit", subnet)
				return nil
			case <-ipt.semStop.Wait():
				l.Debugf("subnet %s: Stop scheduling devices, stop", subnet)
				return nil
			default:
			}
		}
		return nil
	})
}

func (ipt *Input) dispatchDiscovery(subnet string, discovery *discoveryInfo, mSpecificDevices map[string]struct{}) {
	l.Debugf("subnet %s: Run discovery", subnet)
	for currentIP := discovery.StartingIP; discovery.Network.Contains(currentIP); incrementIP(currentIP) {
		deviceIP := currentIP.String()

		if ignored := ipt.isIPIgnored(deviceIP); ignored {
			continue
		}
		if _, ok := mSpecificDevices[deviceIP]; ok {
			continue
		}

		ipt.jobs <- Job{
			ID:     DISCOVERY,
			IP:     deviceIP,
			Subnet: subnet,
		}

		select {
		case <-datakit.Exit.Wait():
			l.Debugf("subnet %s: Stop scheduling devices, exit", subnet)
			return
		case <-ipt.semStop.Wait():
			l.Debugf("subnet %s: Stop scheduling devices, stop", subnet)
			return
		default:
		}
	}
}

func (ipt *Input) doJob(job Job) {
	ipt.checkIPWorking(job.IP)
	defer checkIPDone(job.IP)

	l.Debugf("doJob entry: %#v", job)
	switch job.ID {
	case COLLECT_OBJECT:
		ipt.doCollectObject(job.IP, job.Device)
	case COLLECT_METRICS:
		ipt.doCollectMetrics(job.IP, job.Device)
	case DISCOVERY:
		ipt.doAutoDiscovery(job.IP, job.Subnet)
	}
}

var mWorkingIP sync.Map

// If the IP is working, then waiting.
func (ipt *Input) checkIPWorking(deviceIP string) {
	for {
		if _, ok := mWorkingIP.Load(deviceIP); !ok {
			mWorkingIP.Store(deviceIP, struct{}{})
			return
		}

		l.Debugf("IP working: %s", deviceIP)

		tk := time.NewTicker(time.Second)

		select {
		case <-tk.C:

		case <-datakit.Exit.Wait():
			l.Info(snmpmeasurement.InputName + " exit")
			return

		case <-ipt.semStop.Wait():
			l.Infof(snmpmeasurement.InputName + " return")
			return
		}
	}
}

// If the IP is done, remove it from map.
func checkIPDone(deviceIP string) {
	mWorkingIP.Delete(deviceIP)
}

func (ipt *Input) doCollectObject(deviceIP string, device *deviceInfo) {
	tn := time.Now().UTC()
	points := ipt.CollectingMeasurements(deviceIP, device, tn, true)
	if len(points) == 0 {
		return
	}

	if err := ipt.feeder.Feed(snmpmeasurement.SNMPObjectName, point.Object, points, &io.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("FeedMeasurement object err: %v", err)
		ipt.feeder.FeedLastError(snmpmeasurement.SNMPObjectName, err.Error())
	}
}

func (ipt *Input) doCollectMetrics(deviceIP string, device *deviceInfo) {
	tn := time.Now().UTC()
	points := ipt.CollectingMeasurements(deviceIP, device, tn, false)
	if len(points) == 0 {
		return
	}

	if err := ipt.feeder.Feed(snmpmeasurement.SNMPMetricName, point.Metric, points, &io.Option{CollectCost: time.Since(tn)}); err != nil {
		l.Errorf("FeedMeasurement metric err: %v", err)
		ipt.feeder.FeedLastError(snmpmeasurement.SNMPMetricName, err.Error())
	}
}

func (ipt *Input) CollectingMeasurements(deviceIP string, device *deviceInfo, tn time.Time, isObject bool) []*point.Point {
	var pts []*point.Point

	var fts tagFields

	if isObject {
		ipt.doCollectCore(deviceIP, device, tn, &fts, true) // object need collect meta

		for _, data := range fts.Data {
			sobj := &snmpmeasurement.SNMPObject{
				Name:     snmpmeasurement.SNMPObjectName,
				Tags:     data.Tags,
				Fields:   data.Fields,
				TS:       tn,
				Election: ipt.Election,
			}
			pts = append(pts, sobj.Point())
		}
	} else {
		ipt.doCollectCore(deviceIP, device, tn, &fts, false) // metric not collect meta

		for _, data := range fts.Data {
			smtc := &snmpmeasurement.SNMPMetric{
				Name:     snmpmeasurement.SNMPMetricName,
				Tags:     data.Tags,
				Fields:   data.Fields,
				TS:       tn,
				Election: ipt.Election,
			}
			pts = append(pts, smtc.Point())
		}
	}

	return pts
}

func (ipt *Input) doAutoDiscovery(deviceIP, subnet string) {
	params, err := ipt.BuildSNMPParams(deviceIP)
	if err != nil {
		l.Errorf("Error building params for device %s: %v", deviceIP, err)
		return
	}
	if err := params.Connect(); err != nil {
		l.Debugf("SNMP connect to %s error: %v", deviceIP, err)
		ipt.removeDynamicDevice(deviceIP)
	} else {
		defer params.Conn.Close() //nolint:errcheck

		// Since `params<GoSNMP>.ContextEngineID` is empty
		// `params.GetNext` might lead to multiple SNMP GET calls when using SNMP v3
		value, err := params.GetNext([]string{snmputil.DeviceReachableGetNextOid})
		if err != nil { //nolint:gocritic
			l.Debugf("SNMP get to %s error: %v", deviceIP, err)
			ipt.removeDynamicDevice(deviceIP)
		} else if len(value.Variables) < 1 || value.Variables[0].Value == nil {
			l.Debugf("SNMP get to %s no data", deviceIP)
			ipt.removeDynamicDevice(deviceIP)
		} else {
			l.Debugf("SNMP get to %s success: %v", deviceIP, value.Variables[0].Value)
			ipt.addDynamicDevice(deviceIP, subnet)
		}
	}
}

//------------------------------------------------------------------------------

func (ipt *Input) doCollectCore(ip string, device *deviceInfo, tn time.Time, fts *tagFields, collectObject bool) {
	deviceReachable, tags, values, checkErr, isErrClosed := device.getValuesAndTags()
	if checkErr != nil {
		if isErrClosed && len(device.Subnet) > 0 {
			// used for ignore closed devices failed report
			if _, ok := ipt.mDynamicDevices.Load(ip); !ok {
				// not found means already deleted it in autodiscovery mode.
				return
			}
		}
		l.Warnf("getValuesAndTags failed: %v", checkErr)
	}
	for k, v := range ipt.Tags {
		tags = append(tags, k+":"+v)
	}
	tags = append(tags, "ip:"+ip)
	tags = append(tags, agentHostKey+":"+datakit.DatakitHostName)
	tags = append(tags, agentVersionKey+":"+git.Version)
	if len(device.Subnet) > 0 {
		tags = append(tags, subnetTagKey+":"+device.Subnet)
	}

	var metricData snmputil.MetricDatas
	if values != nil {
		snmputil.ReportMetrics(device.Metrics, values, tags, &metricData)
	}

	var deviceStatus snmputil.DeviceStatus
	if deviceReachable {
		deviceStatus = snmputil.DeviceStatusReachable
	} else {
		deviceStatus = snmputil.DeviceStatusUnreachable
	}

	var metaData deviceMetaData
	if collectObject { // collect object needs to collect meta, so we use "collectMeta" represents collect object.
		metaData.collectMeta = true
		device.ReportNetworkDeviceMetadata(values, tags, device.Metadata, tn, deviceStatus, &metaData)
	}

	aggregateDeviceData(&metricData, fts, &metaData, tags, ipt)
}

type tagFields struct {
	Data []*tagField
}

func (fts *tagFields) Add(ft *tagField) {
	normalizeFieldTags(ft)
	fts.Data = append(fts.Data, ft)
}

type tagField struct {
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
}

func normalizeFieldTags(ft *tagField) {
	for k, v := range ft.Tags {
		tmp := replaceMetricsName(k)
		if len(tmp) > 0 {
			ft.Tags[tmp] = v
			delete(ft.Tags, k)
		}
	}
	for k, v := range ft.Fields {
		tmp := replaceMetricsName(k)
		if len(tmp) > 0 {
			ft.Fields[tmp] = v
			delete(ft.Fields, k)
		}
	}
}

// If underline, replace point to underline
// If without underline, I.e CamelCase, remove point and make the letter behind upper.
// return new when changed, return empty if not fit.
func replaceMetricsName(in string) string {
	if strings.Contains(in, "_") {
		// found _, undeline
		if strings.Contains(in, ".") {
			// found ., replace
			return strings.ReplaceAll(in, ".", "_") // replace
		}
	} else {
		// not found _, CamelCase
		changed := false
		for {
			nIdx := strings.Index(in, ".")
			if nIdx != -1 {
				if !changed {
					changed = true
				}

				newLeft := in[:nIdx] // get left value
				var newRight string
				if len(in) > nIdx+1 {
					right := in[nIdx+1:]
					if len(right) > 0 {
						newRight = strings.ToUpper(string(right[0]))
						newRight += right[1:]
					}
				}
				in = (newLeft + newRight)
			} else {
				break
			}
		}
		if changed {
			return in
		}
	}
	return "" // not replace
}

func aggregateDeviceData(metricData *snmputil.MetricDatas, fts *tagFields, metaData *deviceMetaData, origTags []string, ipt *Input) {
	calcTagsHash(metricData)
	mHash := make(map[string]map[string]interface{}) // map[hash]map[value_key]value_value
	aggregateHash(metricData, mHash)
	getFieldTagArr(metricData, mHash, fts, metaData, origTags, ipt)
}

func calcTagsHash(metricData *snmputil.MetricDatas) {
	// calculate tags hash
	for _, v := range metricData.Data {
		var tagsLine string
		for _, tag := range v.Tags {
			tagsLine += tag
		}
		v.TagsHash = dkstring.MD5Sum(tagsLine)
	}
}

func aggregateHash(metricData *snmputil.MetricDatas, mHash map[string]map[string]interface{}) {
	// aggregate
	for _, v := range metricData.Data {
		if val, ok := mHash[v.TagsHash]; ok { // map[string]interface{}
			if valVal, ok := val[v.Name]; ok { // interface{}
				// If larger then replace, otherwise not.
				if valValFloat64, ok := valVal.(float64); ok {
					if v.Value > valValFloat64 {
						val[v.Name] = v.Value
					}
				} else {
					val[v.Name] = v.Value
				} // float64
			} else {
				val[v.Name] = v.Value
			}
		} else {
			mHash[v.TagsHash] = make(map[string]interface{})
			mHash[v.TagsHash][v.Name] = v.Value
		}
	}
}

// interfaces.
type interfaceAttribute struct {
	Interface      string                 `json:"interface"`
	InterfaceAlias string                 `json:"interface_alias"`
	Fields         map[string]interface{} `json:"fields"`
}

// sensors.
type sensorAttribute struct {
	SensorID   string                 `json:"sensor_id"`
	SensorType string                 `json:"sensor_type"`
	Fields     map[string]interface{} `json:"fields"`
}

// mems.
type memAttribute struct {
	Mem    string                 `json:"mem"`
	Fields map[string]interface{} `json:"fields"`
}

// mem_pool_names.
type memPoolNameAttribute struct {
	MemPoolName string                 `json:"mem_pool_name"`
	Fields      map[string]interface{} `json:"fields"`
}

// cpus.
type cpuAttribute struct {
	CPU    string                 `json:"cpu"`
	Fields map[string]interface{} `json:"fields"`
}

var reservedKeys = []string{
	"device_vendor",
	"host",
	"ip",
	"name",
	"snmp_host",
	"snmp_profile",
}

func isReservedKeys(checkName string, customTags map[string]string) bool {
	// custom tags should be reserved.
	if _, ok := customTags[checkName]; ok {
		return true
	}

	for _, v := range reservedKeys {
		if v == checkName {
			return true
		}
	}

	return false
}

func getFieldTagArr(metricData *snmputil.MetricDatas,
	mHash map[string]map[string]interface{},
	fts *tagFields,
	metaData *deviceMetaData,
	origTags []string,
	ipt *Input,
) {
	if len(mHash) == 0 {
		return
	}

	// for object only.
	objectTags := make(map[string]string)
	objectFields := make(map[string]interface{})
	var objectFieldInterfaces []*interfaceAttribute     // interfaces.
	var objectFieldSensors []*sensorAttribute           // sensors.
	var objectFieldmems []*memAttribute                 // mems
	var objectFieldMemPoolNames []*memPoolNameAttribute // mem_pool_names
	var objectFieldcCPUs []*cpuAttribute                // cpus
	var objectFieldAll []*tagField                      // all

	for hash, fields := range mHash {
		tags := make(map[string]string)

		for _, v := range metricData.Data {
			if v.TagsHash == hash {
				getDatakitStyleTags(v.Tags, tags)
				break
			}
		} // for data

		tags["host"] = tags["ip"] // replace host as ip.
		tags["name"] = tags["ip"] // replace name as ip.

		if metaData.collectMeta {
			// collect object.

			isCreated := false // whether already created data set.

			for tagK, tagV := range tags {
				if isReservedKeys(tagK, ipt.Tags) {
					// reserved, only assignment once.
					if _, ok := objectTags[tagK]; ok {
						continue
					} else {
						objectTags[tagK] = tagV
					}
				} else {
					if !isCreated {
						isCreated = true

						// gathering specific.
						switch tagK {
						case "interface":
							objectFieldInterfaces = append(objectFieldInterfaces, &interfaceAttribute{
								Interface:      tagV,
								InterfaceAlias: tags["interface_alias"],
								Fields:         fields,
							})
						case "sensor_id":
							objectFieldSensors = append(objectFieldSensors, &sensorAttribute{
								SensorID:   tagV,
								SensorType: tags["sensor_type"],
								Fields:     fields,
							})
						case "mem":
							objectFieldmems = append(objectFieldmems, &memAttribute{
								Mem:    tagV,
								Fields: fields,
							})
						case "mem_pool_name":
							objectFieldMemPoolNames = append(objectFieldMemPoolNames, &memPoolNameAttribute{
								MemPoolName: tagV,
								Fields:      fields,
							})
						case "cpu":
							objectFieldcCPUs = append(objectFieldcCPUs, &cpuAttribute{
								CPU:    tagV,
								Fields: fields,
							})
						} // switch tagK

						// gathering all.
						unknownTags := make(map[string]string)
						for tagKK, tagVV := range tags {
							if !isReservedKeys(tagKK, ipt.Tags) {
								unknownTags[tagKK] = tagVV
							}
						}
						objectFieldAll = append(objectFieldAll, &tagField{
							Tags:   unknownTags,
							Fields: fields,
						})
					} // if !isCreated
				}
			}
		} else {
			// collect metrics.

			if ipt.EnablePickingData {
				// collect picking data.

				found := false
				for k := range fields {
					tmp := replaceMetricsName(k)
					if len(tmp) == 0 {
						tmp = k
					}
					if _, ok := ipt.mFieldNameSpecified[tmp]; ok {
						found = true
						break
					}
				}
				if found {
					fts.Add(&tagField{
						Tags:   tags,
						Fields: fields,
					})
				}
			} else {
				// collect every interface data.

				fts.Add(&tagField{
					Tags:   tags,
					Fields: fields,
				})
			}
		}
	}

	if metaData.collectMeta {
		// collect object.

		objectFields["interfaces"] = beJSON(objectFieldInterfaces)
		objectFields["sensors"] = beJSON(objectFieldSensors)
		objectFields["mems"] = beJSON(objectFieldmems)
		objectFields["mem_pool_names"] = beJSON(objectFieldMemPoolNames)
		objectFields["cpus"] = beJSON(objectFieldcCPUs)
		objectFields["all"] = beJSON(objectFieldAll)

		tags := make(map[string]string)
		getDatakitStyleTags(origTags, tags)

		metaAll := strings.Join(metaData.data, ", ")
		objectFields["device_meta"] = metaAll

		fts.Add(&tagField{
			Tags:   objectTags,
			Fields: objectFields,
		})
	}
}

func beJSON(in interface{}) interface{} {
	bys, err := json.Marshal(in)
	if err != nil {
		l.Errorf("json.Marshal failed: %v", err)
		return nil
	}
	return string(bys)
}

func getDatakitStyleTags(tags []string, outTags map[string]string) {
	for _, tag := range tags {
		arr := strings.Split(tag, ":")
		if len(arr) == 2 {
			// ignore specific rules for GuanceCloud
			switch arr[0] {
			case agentHostKey, agentVersionKey: // drop
			default:
				outTags[arr[0]] = arr[1]
			}
		}
	}
}

func (ipt *Input) assignFieldNameSpecified(arr []string) {
	for _, v := range arr {
		ipt.mFieldNameSpecified[v] = struct{}{}
	}
}

func (ipt *Input) ValidateConfig() error {
	ipt.mAutoDiscovery = make(map[string]*discoveryInfo)
	ipt.mSpecificDevices = make(map[string]*deviceInfo)
	ipt.mDiscoveryIgnoredIPs = make(map[string]struct{})

	if ipt.EnablePickingData {
		ipt.mFieldNameSpecified = make(map[string]struct{})
		ipt.assignFieldNameSpecified(ipt.PickingStatus)
		ipt.assignFieldNameSpecified(ipt.PickingSpeed)
		ipt.assignFieldNameSpecified(ipt.PickingCPU)
		ipt.assignFieldNameSpecified(ipt.PickingMem)
		ipt.assignFieldNameSpecified(ipt.PickingExtra)
	}

	// default check zone
	if ipt.Port <= 0 || ipt.Port > 65535 {
		ipt.Port = defaultPort
	}
	if ipt.ObjectInterval == 0 {
		ipt.ObjectInterval = defaultObjectInterval
	}
	if ipt.MetricInterval == 0 {
		ipt.MetricInterval = defaultMetricInterval
	}
	if ipt.Workers == 0 {
		ipt.Workers = defaultWorkers
	}
	if ipt.DiscoveryInterval == 0 {
		ipt.DiscoveryInterval = defaultDiscoveryInterval
	}
	if len(ipt.DeviceNamespace) == 0 {
		ipt.DeviceNamespace = defaultDeviceNamespace
	}

	l.Info(ipt.Port, ipt.ObjectInterval, ipt.MetricInterval, ipt.Workers, ipt.DiscoveryInterval, ipt.DeviceNamespace)

	if err := ipt.validateNetAddress(); err != nil {
		return err
	}

	switch ipt.SNMPVersion {
	case 1, 2, 3:
	default:
		return fmt.Errorf("`snmp_version` must be 1 or 2 or 3")
	}

	return nil
}

func (ipt *Input) validateNetAddress() error {
	for _, v := range ipt.AutoDiscovery {
		if len(v) == 0 {
			continue
		}
		if _, _, err := net.ParseCIDR(v); err != nil {
			return err
		}
		ipt.mAutoDiscovery[v] = &discoveryInfo{}
	}
	for _, v := range ipt.DiscoveryIgnoredIPs {
		if len(v) == 0 {
			continue
		}
		ipt.mDiscoveryIgnoredIPs[v] = struct{}{}
	}
	for _, v := range ipt.SpecificDevices {
		if len(v) == 0 {
			continue
		}
		if ip := net.ParseIP(v); ip == nil {
			return fmt.Errorf("invalid IP address")
		}
		ipt.mSpecificDevices[v] = &deviceInfo{}
	}
	return nil
}

func (ipt *Input) Initialize() error {
	if err := ipt.initializeDiscovery(); err != nil {
		return err
	}
	if err := ipt.initializeSpecificDevices(); err != nil {
		return err
	}
	return nil
}

func (ipt *Input) initializeSpecificDevices() error {
	if len(ipt.Profile) > 0 || len(ipt.Metrics) > 0 {
		ipt.autodetectProfile = false
	} else {
		ipt.autodetectProfile = true
	}

	snmputil.NormalizeMetrics(ipt.Metrics)
	ipt.Metrics = append(ipt.Metrics, snmputil.UptimeMetricConfig) // addUptimeMetric
	ipt.Metadata = snmputil.UpdateMetadataDefinitionWithLegacyFallback(nil)
	ipt.OidConfig.AddScalarOids(snmputil.ParseScalarOids(ipt.Metrics, ipt.MetricTags, ipt.Metadata, true))
	ipt.OidConfig.AddColumnOids(snmputil.ParseColumnOids(ipt.Metrics, ipt.Metadata, true))

	// Profile Configs
	var profiles snmputil.ProfileDefinitionMap
	if len(ipt.CustomProfiles) > 0 {
		// TODO: [PERFORMANCE] Load init config custom profiles once for all integrations
		//   There are possibly multiple init configs
		//
		customProfiles, err := snmputil.LoadProfiles(ipt.CustomProfiles)
		if err != nil {
			return fmt.Errorf("failed to load custom profiles: %w", err)
		}
		profiles = customProfiles
	} else {
		defaultProfiles, err := snmputil.LoadDefaultProfiles()
		if err != nil {
			return fmt.Errorf("failed to load default profiles: %w", err)
		}
		profiles = defaultProfiles
	}

	for _, profileDef := range profiles {
		snmputil.NormalizeMetrics(profileDef.Metrics)
	}

	ipt.Profiles = profiles

	errors := snmputil.ValidateEnrichMetrics(ipt.Metrics)
	errors = append(errors, snmputil.ValidateEnrichMetricTags(ipt.MetricTags)...)
	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "\n"))
	}

	// init session
	for deviceIP := range ipt.mSpecificDevices {
		di, err := ipt.initializeDevice(deviceIP, "")
		if err != nil {
			l.Errorf("initializeDevice failed: err = (%v), ip = (%s)", err, deviceIP)
			return err
		}
		ipt.mSpecificDevices[deviceIP] = di
	}

	return nil
}

func (ipt *Input) initializeDevice(deviceIP, subnet string) (*deviceInfo, error) {
	session, err := snmputil.NewGosnmpSession(&snmputil.SessionOpts{
		IPAddress:       deviceIP,
		Port:            ipt.Port,
		SnmpVersion:     ipt.SNMPVersion,
		CommunityString: ipt.V2CommunityString,
		User:            ipt.V3User,
		AuthProtocol:    ipt.V3AuthProtocol,
		AuthKey:         ipt.V3AuthKey,
		PrivProtocol:    ipt.V3PrivProtocol,
		PrivKey:         ipt.V3PrivKey,
		ContextName:     ipt.V3ContextName,
	})
	if err != nil {
		l.Errorf("NewGosnmpSession failed: err = (%v), ip = (%s)", err, deviceIP)
		return nil, err
	}
	di := NewDeviceInfo(ipt, deviceIP, ipt.DeviceNamespace, subnet, session)
	if err := di.initialize(); err != nil {
		l.Errorf("Input initialize failed: err = (%v), ip = (%s)", err, deviceIP)
		return nil, err
	}

	return di, nil
}

// only for command "datakit --test-input".

func (ipt *Input) Collect() (map[point.Category][]*point.Point, error) {
	setLog()

	if err := ipt.CheckTestSNMP(); err != nil {
		return nil, err
	}

	if err := ipt.ValidateConfig(); err != nil {
		return nil, err
	}

	if err := ipt.Initialize(); err != nil {
		return nil, err
	}

	l.Infof("Start collecting snmp...\n")
	mpts := make(map[point.Category][]*point.Point)

	specificDevices := ipt.GetSpecificDevices()
	for deviceIP, deviceInfo := range specificDevices {
		tn := time.Now().UTC()
		points := ipt.CollectingMeasurements(deviceIP, deviceInfo, tn, true)
		mpts[point.Object] = append(mpts[point.Object], points...)

		tn = time.Now().UTC()
		points = ipt.CollectingMeasurements(deviceIP, deviceInfo, tn, false)
		mpts[point.Metric] = append(mpts[point.Metric], points...)
	}

	return mpts, nil
}

func setLog() {
	onceSetLog.Do(func() {
		l = logger.SLogger(snmpmeasurement.InputName)
	})
	snmputil.SetLog()
}

func (ipt *Input) CheckTestSNMP() error {
	if len(ipt.mAutoDiscovery) > 0 {
		return fmt.Errorf("autodiscovery_not_empty")
	}

	return nil
}

func (ipt *Input) GetSpecificDevices() map[string]*deviceInfo {
	return ipt.mSpecificDevices
}

func (ipt *Input) exit() {
	traps.StopServer()

	for deviceIP, device := range ipt.mSpecificDevices {
		if err := device.Session.Close(); err != nil {
			l.Warnf("device.Session.Close failed: err = (%v), deviceIP = (%v)", err, deviceIP)
		}
	}

	ipt.mDynamicDevices.Range(func(k, v interface{}) bool {
		deviceIP, ok := k.(string)
		if !ok {
			l.Errorf("should not be here")
			return true
		}
		device, ok := v.(*deviceInfo)
		if !ok {
			l.Errorf("should not be here")
			return true
		}

		l.Debugf("closing %s", deviceIP)
		if err := device.Session.Close(); err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				l.Warnf("device.Session.Close failed: err = (%v), deviceIP = (%v)", err, deviceIP)
			}
		}
		ipt.mDynamicDevices.Delete(k)
		return true
	})
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		Tags:    make(map[string]string),
		semStop: cliutils.NewSem(),
		feeder:  io.DefaultFeeder(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(snmpmeasurement.InputName, func() inputs.Input {
		return defaultInput()
	})
}
