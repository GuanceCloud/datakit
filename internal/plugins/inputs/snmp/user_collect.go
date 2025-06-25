// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/gosnmp/gosnmp"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/maputil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmpmeasurement"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/snmputil"
)

// clone ipt.UserProfileStore.ZabbixStores[?].Definition -> ipt.userSpecificDevices[deviceIP].
func (ipt *Input) initUserDefinition() {
	if ipt.UserProfileStore == nil {
		return
	}

	for idx, store := range ipt.UserProfileStore.ZabbixStores {
		for _, ip := range store.IPList {
			ipt.jobs <- snmpJob{
				ID:  USER_DISCOVERY,
				IP:  ip,
				Idx: idx,
			}
		}
	}
}

func (ipt *Input) userAutoDiscovery() {
	if ipt.UserProfileStore == nil {
		return
	}

	if len(ipt.UserProfileStore.ZabbixStores) == 0 || len(ipt.mAutoDiscovery) == 0 {
		return
	}

	g.Go(func(ctx context.Context) error {
		if !atomic.CompareAndSwapUint32(&ipt.DiscoveryRunning, 0, 1) {
			// last loop discovery is running
			return nil
		}

		tn := time.Now().UTC()
		for subnet, discovery := range ipt.mAutoDiscovery {
			ipt.dispatchUserDiscovery(subnet, discovery)

			select {
			case <-datakit.Exit.Wait():
				l.Debugf("subnet %s: Stop scheduling devices, exit", subnet)
				atomic.StoreUint32(&ipt.DiscoveryRunning, 0)
				return nil
			case <-ipt.semStop.Wait():
				l.Debugf("subnet %s: Stop scheduling devices, stop", subnet)
				atomic.StoreUint32(&ipt.DiscoveryRunning, 0)
				return nil
			default:
			}
		}
		discoveryCostVec.WithLabelValues("zabbix").Observe(float64(time.Since(tn)) / float64(time.Second))

		atomic.StoreUint32(&ipt.DiscoveryRunning, 0)
		return nil
	})
}

func (ipt *Input) dispatchUserDiscovery(subnet string, discovery *discoveryInfo) {
	l.Debugf("subnet %s: Run discovery", subnet)
	for currentIP := cloneIP(discovery.StartingIP); discovery.Network.Contains(currentIP); incrementIP(currentIP) {
		deviceIP := currentIP.String()

		if ignored := ipt.isIPIgnored(deviceIP); ignored {
			continue
		}

		ipt.jobs <- snmpJob{
			ID:  USER_DISCOVERY,
			IP:  deviceIP,
			Idx: -1,
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

// if from Zabbix, idx must > -1
// if from consul discover, know ip and deviceType
// if from auto discovery, will try ipt.UserProfileStore with oid 1.3.6.1.2.1.1.2.0.
func (ipt *Input) doAfterDiscovery(idx int, deviceIP, deviceType string, tags map[string]string) {
	if ipt.UserProfileStore == nil {
		return
	}

	// ip from consul discover, must have ip and deviceType
	if deviceType != "" {
		for i, store := range ipt.UserProfileStore.ZabbixStores {
			if deviceType == store.Definition.DeviceType {
				ipt.addDevice(i, deviceIP, tags)
				return
			}
		}
		l.Warnf("stores have not this device type: %s", deviceType)
		return
	}

	l.Debugf("discovery idx: %d, ip: %s, deviceType: %s", idx, deviceIP, deviceType)
	// ip already in collect array
	if _, ok := ipt.userSpecificDevices.Load(deviceIP); ok {
		l.Debugf("device ip already in collect array: %s", deviceIP)
		return
	}

	// 0 stores
	if len(ipt.UserProfileStore.ZabbixStores) == 0 {
		l.Error("have no profile")
		return
	}

	// ip from snmp.conf file,
	if idx > -1 {
		ipt.addDevice(idx, deviceIP, tags)
		return
	}

	// from auto discover, need try all stores from oid 1.3.6.1.2.1.1.2.0

	params, err := ipt.BuildSNMPParams(deviceIP)
	if err != nil {
		l.Errorf("Error building params for device %s: %v", deviceIP, err)
		return
	}
	if err := params.Connect(); err != nil {
		l.Debugf("SNMP connect to %s error: %v", deviceIP, err)
		return
	}
	defer params.Conn.Close() //nolint:errcheck

	// Since `params<GoSNMP>.ContextEngineID` is empty
	// `params.GetNext` might lead to multiple SNMP GET calls when using SNMP v3
	value, err := params.GetNext([]string{snmputil.DeviceReachableGetNextOid})
	if err != nil { //nolint:gocritic
		l.Debugf("SNMP get to %s error: %v", deviceIP, err)
		return
	} else if len(value.Variables) < 1 || value.Variables[0].Value == nil {
		l.Debugf("SNMP get to %s no data", deviceIP)
		return
	}
	l.Debugf("SNMP get to %s success: %v", deviceIP, value.Variables[0].Value)

	tryIdx, err := ipt.tryDevice(deviceIP, params)
	if err != nil {
		l.Debugf("ip : %s compare stores fail : %w", deviceIP, err)
		return
	}

	l.Debugf("add device ip : %s, store index : %d", deviceIP, tryIdx)
	ipt.addDevice(tryIdx, deviceIP, tags)
}

// add device to ipt.userSpecificDevices.
func (ipt *Input) addDevice(idx int, deviceIP string, tags map[string]string) {
	if len(ipt.UserProfileStore.ZabbixStores) <= idx {
		return
	}

	v2CommunityString := ipt.V2CommunityString
	if ipt.UserProfileStore.ZabbixStores[idx].Definition.Community != "" {
		// dome times prom.yml have community
		v2CommunityString = ipt.UserProfileStore.ZabbixStores[idx].Definition.Community
	}

	session, err := snmputil.NewGosnmpSession(&snmputil.SessionOpts{
		IPAddress:       deviceIP,
		Port:            ipt.Port,
		SnmpVersion:     ipt.SNMPVersion,
		CommunityString: v2CommunityString,
		User:            ipt.V3User,
		AuthProtocol:    ipt.V3AuthProtocol,
		AuthKey:         ipt.V3AuthKey,
		PrivProtocol:    ipt.V3PrivProtocol,
		PrivKey:         ipt.V3PrivKey,
		ContextName:     ipt.V3ContextName,
	})
	if err != nil {
		l.Errorf("NewGosnmpSession failed: err = (%v), ip = (%s)", err, deviceIP)
		return
	}

	di := NewDeviceInfo(ipt, deviceIP, ipt.DeviceNamespace, "", session)
	if err := di.initialize(); err != nil {
		l.Errorf("Input initialize failed: err = (%v), ip = (%s)", err, deviceIP)
		return
	}
	di.UserProfileDefinition = ipt.cloneZabbixDefinition(ipt.UserProfileStore.ZabbixStores[idx])
	di.UserProfileDefinition.InputTags = maputil.MergeMapString(di.UserProfileDefinition.InputTags, tags)
	di.preProcess()

	ipt.userSpecificDevices.Store(deviceIP, di)
}

func (ipt *Input) tryDevice(deviceIP string, params *gosnmp.GoSNMP) (int, error) {
	value, err := params.Get([]string{sysObjectOID})
	if err != nil || value == nil { //nolint:gocritic
		l.Debugf("SNMP get to %s error: %v", deviceIP, err)
		return -1, fmt.Errorf("SNMP get to %s error: %w", deviceIP, err)
	}

	nextOID := ""
	for _, variable := range value.Variables {
		// example: .1.3.6.1.4.1.2011.2.240.12
		if nextOID, err = assertString(variable.Value); err != nil {
			l.Debugf("assert snmpPDU.Value: %v to string : failed: %w", variable.Value, err)
			return -1, fmt.Errorf("assert snmpPDU.Value: %v to string : failed: %w", variable.Value, err)
		}
	}

	value, err = params.GetNext([]string{nextOID})
	if err != nil || value == nil { //nolint:gocritic
		l.Debugf("SNMP get to %s error: %v", deviceIP, err)
		return -1, fmt.Errorf("SNMP get to %s error: %w", deviceIP, err)
	}

	compareOID := ""
	for _, variable := range value.Variables {
		compareOID = variable.Name // example: .1.3.6.1.4.1.2011.5.2.1.1.1.1.6.114.97.100.105.117.115
	}

	return ipt.compareStores(compareOID)
}

func (ipt *Input) compareStores(compareOID string) (int, error) {
	compareOID = FormatOID(compareOID)

	long := 0
	idx := -1
	for i, store := range ipt.UserProfileStore.ZabbixStores {
		l := prefixLong(compareOID, store)
		if long < l && l >= 7 {
			// example compareOID 1.3.6.1.4.1.674.10892.5.1.1.1.0
			// if l < 7, can not compare the manufacturer info
			long = l
			idx = i
		}
	}

	if idx == -1 {
		l.Debugf("stores can not compared oid : %s", compareOID)
		return -1, fmt.Errorf("stores can not compared oid : %s", compareOID)
	}

	return idx, nil
}

func prefixLong(compareOID string, store *snmputil.ProfileStore) int {
	long := 0
	for _, host := range store.Definition.ZabbixExport.Hosts {
		for _, item := range host.Items {
			l := samePrefixLong(compareOID, item.SnmpOID)
			if long < l {
				long = l
			}
		}
	}

	for _, template := range store.Definition.ZabbixExport.Templates {
		for _, item := range template.Items {
			l := samePrefixLong(compareOID, item.SnmpOID)
			if long < l {
				long = l
			}
		}

		for _, rule := range template.DiscoveryRules {
			for _, item := range rule.ItemPrototypes {
				l := samePrefixLong(compareOID, item.SnmpOID)
				if long < l {
					long = l
				}
			}
		}
	}

	return long
}

func samePrefixLong(x, y string) int {
	arrX := strings.Split(x, ".")
	arrY := strings.Split(y, ".")

	l := len(arrX)
	if l > len(arrY) {
		l = len(arrY)
	}

	for i := 0; i < l; i++ {
		if arrX[i] != arrY[i] {
			return i
		}
	}

	return l
}

func (ipt *Input) collectUserObject() {
	ipt.userSpecificDevices.Range(func(deviceIP, v interface{}) bool {
		device, ok := v.(*deviceInfo)
		if !ok {
			l.Warn("assert deviceInfo fail, ip:", deviceIP)
			return true
		}

		ip, ok := deviceIP.(string)
		if !ok {
			l.Warn("assert deviceIP fail, ip:", deviceIP)
			return true
		}

		ipt.jobs <- snmpJob{
			ID:     COLLECT_USER_OBJECT,
			IP:     ip,
			Device: device,
		}
		return true
	})
}

func (ipt *Input) doCollectUserObject(deviceIP string, device *deviceInfo) {
	if ipt.UserProfileStore == nil {
		return
	}

	tn := time.Now().UTC()
	points, _ := device.getUserMeasurements(deviceIP, tn, true)

	if len(points) == 0 {
		return
	}

	if err := ipt.feeder.Feed(point.CustomObject, points,
		dkio.WithCollectCost(time.Since(tn)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(snmpmeasurement.SNMPObjectName),
	); err != nil {
		l.Errorf("FeedMeasurement metric err: %v", err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(snmpmeasurement.InputName),
			metrics.WithLastErrorCategory(point.CustomObject),
		)
	}
}

func (ipt *Input) collectUserMetrics() {
	tn := time.Now().UTC()
	deviceNumbers := make(map[string]int)

	ipt.userSpecificDevices.Range(func(deviceIP, v interface{}) bool {
		device, ok := v.(*deviceInfo)
		if !ok {
			l.Warn("assert deviceInfo fail, ip:", deviceIP)
			return true
		}

		ip, ok := deviceIP.(string)
		if !ok {
			l.Warn("assert deviceIP fail, ip:", deviceIP)
			return true
		}

		class := device.ClassName()
		deviceNumbers[class]++

		ipt.jobs <- snmpJob{
			ID:     COLLECT_USER_METRICS,
			IP:     ip,
			Device: device,
		}
		return true
	})

	deviceTotal := 0
	for class, i := range deviceNumbers {
		aliveDevicesVec.WithLabelValues(class).Set(float64(i))
		deviceTotal += i
	}
	aliveDevicesVec.WithLabelValues("total").Set(float64(deviceTotal))
	collectCostVec.WithLabelValues().Observe(float64(time.Since(tn)) / float64(time.Second))
}

func (ipt *Input) doCollectUserMetrics(deviceIP string, device *deviceInfo) {
	if ipt.UserProfileStore == nil {
		return
	}

	tn := time.Now().UTC()
	points, _ := device.getUserMeasurements(deviceIP, tn, false)
	l.Debugf("collect points: %d, ip: %s, profile name: %s", len(points), deviceIP, device.UserProfileDefinition.ProfileName)
	deviceCollectCostVec.WithLabelValues(device.ClassName()).Observe(float64(time.Since(tn)) / float64(time.Second))

	if len(points) == 0 {
		return
	}

	if err := ipt.feeder.Feed(point.Metric, points,
		dkio.WithCollectCost(time.Since(tn)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(snmpmeasurement.SNMPMetricName),
	); err != nil {
		l.Errorf("FeedMeasurement metric err: %v", err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(snmpmeasurement.InputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
	}
}

// AddTags add tag into kvs after ignore filter.
func (ipt *Input) AddTags(kvs point.KVs, k string, v string) point.KVs {
	for _, s := range ipt.TagsIgnore {
		if s == k {
			return kvs
		}
	}

	for _, r := range ipt.TagsIgnoreRule {
		if r.MatchString(k) {
			return kvs
		}
	}

	return kvs.AddTag(k, v)
}

func (di *deviceInfo) getUserMeasurements(ip string, tn time.Time, collectObject bool) ([]*point.Point, error) {
	var z snmputil.Session
	if di.Session == z {
		l.Error("di.Session is nil")
		return nil, fmt.Errorf("di.Session is nil")
	}

	err := di.Session.Connect()
	if err != nil {
		l.Warnf("Connect failed: err = (%v), ip = (%s)", err, di.IP)
		return nil, err
	}
	defer func() {
		err := di.Session.Close()
		if err != nil {
			l.Warnf("failed to close session: err = (%v), ip = (%s)", err, di.IP)
		}
	}()

	// Check if the device is reachable
	_, err = di.Session.GetNext([]string{snmputil.DeviceReachableGetNextOid})
	if err != nil {
		l.Errorf("check %s device reachable: failed: %v", di.IP, err)
		return nil, err
	}

	if collectObject {
		return di.getUserObjectPoint(di.UserProfileDefinition.Items, di.UserProfileDefinition.StringItems, ip, tn), nil
	}

	var points []*point.Point

	points = append(points, di.getUserPoints(di.UserProfileDefinition.Items, ip, tn)...)

	for _, item := range di.UserProfileDefinition.DiscoveryItems {
		points = append(points, di.getUserDiscoveryPoints(item, ip, tn)...)
	}

	return points, nil
}

func (di *deviceInfo) getUserObjectPoint(items, stringItems []*snmputil.Item, ip string, tn time.Time) []*point.Point {
	var kvs point.KVs

	// must have obj points
	if len(items) == 0 {
		items = append(items, &snmputil.Item{
			SnmpOID: "1.3.6.1.2.1.1.3.0",
			Key:     "netUptime",
		}, &snmputil.Item{
			SnmpOID: "1.3.6.1.2.1.25.1.1.0",
			Key:     "uptime",
		})
	}

	for _, item := range items {
		if item.SnmpOID == "" {
			continue
		}

		fieldName := item.Key
		if fieldName == "" {
			continue
		}

		snmpPacket, err := di.Session.Get([]string{item.SnmpOID})
		if err != nil {
			l.Debugf("get snmp: failed: %v", err)
			continue
		}

		f, _, isFloat, err := assertSNMPPacket(snmpPacket)
		if err != nil && !isFloat {
			continue
		}

		kvs = kvs.Add(fieldName, f, false, false)
	}

	kvs = di.Ipt.AddTags(kvs, "ip", ip)
	kvs = di.Ipt.AddTags(kvs, "name", di.UserProfileDefinition.Name)
	kvs = di.Ipt.AddTags(kvs, "sys_name", di.UserProfileDefinition.SysName)
	kvs = di.Ipt.AddTags(kvs, "sys_object_id", di.UserProfileDefinition.SysObjectID)
	kvs = di.Ipt.AddTags(kvs, "device_type", di.UserProfileDefinition.DeviceType)
	for k, v := range di.UserProfileDefinition.InputTags {
		kvs = di.Ipt.AddTags(kvs, k, v)
	}

	for _, item := range stringItems {
		if item.SnmpOID == "" {
			continue
		}

		fieldName := item.Key
		if fieldName == "" {
			continue
		}

		snmpPacket, err := di.Session.Get([]string{item.SnmpOID})
		if err != nil {
			l.Debugf("get snmp: failed: %v", err)
			continue
		}

		v, err := assertSNMPPacketString(snmpPacket)
		if err != nil {
			continue
		}

		kvs = kvs.Add(fieldName, v, false, false)
	}

	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(tn))
	return []*point.Point{point.NewPointV2(di.ClassName(), kvs, opts...)}
}

func (di *deviceInfo) getUserPoints(items []*snmputil.Item, ip string, tn time.Time) []*point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(tn))
	var points []*point.Point

	for _, item := range items {
		if item.SnmpOID == "" {
			continue
		}

		fieldName := item.Key
		if fieldName == "" {
			continue
		}

		snmpPacket, err := di.Session.Get([]string{item.SnmpOID})
		if err != nil {
			l.Debugf("get snmp: failed: %v", err)
			continue
		}

		f, _, isFloat, err := assertSNMPPacket(snmpPacket)
		if err != nil && !isFloat {
			continue
		}
		// example: x = doFunc(x, process.Parameter) -> x * 0.1
		for _, process := range item.Preprocessing.Steps {
			if doFunc, ok := wantPreprocessing[process.Type]; ok {
				f = doFunc(f, process.Parameter)
			}
		}

		var kvs point.KVs
		kvs = kvs.Add(fieldName, f, false, false)

		kvs = di.Ipt.AddTags(kvs, "ip", ip)
		kvs = di.Ipt.AddTags(kvs, "oid", item.SnmpOID)
		kvs = di.Ipt.AddTags(kvs, "name", di.UserProfileDefinition.Name)
		kvs = di.Ipt.AddTags(kvs, "sys_name", di.UserProfileDefinition.SysName)
		kvs = di.Ipt.AddTags(kvs, "sys_object_id", di.UserProfileDefinition.SysObjectID)
		kvs = di.Ipt.AddTags(kvs, "device_type", di.UserProfileDefinition.DeviceType)

		for k, v := range getItemTags(item.Tags) {
			kvs = di.Ipt.AddTags(kvs, k, v)
		}
		for k, v := range di.UserProfileDefinition.InputTags {
			kvs = di.Ipt.AddTags(kvs, k, v)
		}

		points = append(points, point.NewPointV2(di.ClassName(), kvs, opts...))
	}

	return points
}

func (di *deviceInfo) getUserDiscoveryPoints(item *snmputil.Item, ip string, tn time.Time) []*point.Point {
	if len(item.OIDs) == 0 {
		return nil
	}

	var points []*point.Point

	// In github.com/gosnmp/gosnmp, have MaxOids (60).
	// So item.OIDs[i:end] will get max gosnmp.MaxOids (60) strings.
	for i := 0; i < len(item.OIDs); i += gosnmp.MaxOids {
		end := i + gosnmp.MaxOids
		if end > len(item.OIDs) {
			end = len(item.OIDs)
		}

		snmpPacket, err := di.Session.Get(item.OIDs[i:end])
		if err != nil {
			l.Errorf("SNMP session.Get: %v", err)
			return nil
		}

		for _, snmpPDU := range snmpPacket.Variables {
			snmpIndex := getSnmpIndex(snmpPDU.Name)
			pts := di.makeDiscoveryPoint(snmpPDU, ip, tn, item, snmpIndex)
			points = append(points, pts...)
		}
	}

	return points
}

func (di *deviceInfo) makeDiscoveryPoint(
	snmpPDU gosnmp.SnmpPDU,
	ip string,
	tn time.Time,
	item *snmputil.Item,
	snmpIndex string,
) []*point.Point {
	fieldName := item.Key
	if fieldName == "" {
		return nil
	}

	value, err := snmputil.GetValueFromPDU(snmpPDU)
	if err != nil {
		l.Debugf("assertSNMPPacket fail, oid:%s, type:%s", snmpPDU.Name, snmpPDU.Type.String())
		return nil
	}
	f, ok := value.(float64)
	if !ok {
		return nil
	}
	// example: x = doFunc(x, process.Parameter) -> x * 0.1
	for _, process := range item.Preprocessing.Steps {
		if doFunc, ok := wantPreprocessing[process.Type]; ok {
			f = doFunc(f, process.Parameter)
		}
	}

	var kvs point.KVs
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(tn))

	kvs = kvs.Add(fieldName, f, false, false)
	kvs = di.Ipt.AddTags(kvs, "ip", ip)

	// add macro tags, example: DISK_NAME=raid1-sys
	if snmpIndex != "" {
		kvs = di.Ipt.AddTags(kvs, "oid", item.SnmpOID+"."+snmpIndex)
		for k, macro := range item.Macros {
			if v, ok := macro[snmpIndex]; ok {
				kvs = di.Ipt.AddTags(kvs, di.filterTagName(k), v)
			}
		}
	} else {
		kvs = di.Ipt.AddTags(kvs, "oid", item.SnmpOID)
	}

	kvs = di.Ipt.AddTags(kvs, "name", di.UserProfileDefinition.Name)
	kvs = di.Ipt.AddTags(kvs, "sys_name", di.UserProfileDefinition.SysName)
	kvs = di.Ipt.AddTags(kvs, "sys_object_id", di.UserProfileDefinition.SysObjectID)
	kvs = di.Ipt.AddTags(kvs, "device_type", di.UserProfileDefinition.DeviceType)

	for k, v := range getItemTags(item.Tags) {
		kvs = di.Ipt.AddTags(kvs, k, v)
	}
	for k, v := range di.UserProfileDefinition.InputTags {
		kvs = di.Ipt.AddTags(kvs, k, v)
	}

	return []*point.Point{point.NewPointV2(di.ClassName(), kvs, opts...)}
}

func (di *deviceInfo) filterFieldName(name, oid string) string {
	if newName, ok := di.Ipt.OIDKeys[oid]; ok {
		return newName
	}
	if newName, ok := di.Ipt.KeyMapping[name]; ok {
		return newName
	}
	return name
}

func (di *deviceInfo) filterTagName(name string) string {
	if newName, ok := di.Ipt.KeyMapping[name]; ok {
		return newName
	}
	return name
}

func (di *deviceInfo) ClassName() string {
	if di.UserProfileDefinition != nil {
		if di.UserProfileDefinition.Class != "" {
			return "snmp_" + di.UserProfileDefinition.Class
		}
	}

	return "snmp_unknown"
}

// example:
// outOID 1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5.9 -> 9
// outOID .1.3.6.1.4.1.674.10892.5.4.700.20.1.6.1.2. -> 2.
func getSnmpIndex(outOID string) string {
	outOID = FormatOID(outOID)
	strs := strings.Split(outOID, ".")
	str := strs[len(strs)-1]
	if _, err := strconv.Atoi(str); err != nil {
		l.Debugf("get snmpIndex fail : %s", outOID)
		return ""
	}

	return str
}

func getItemTags(t []*snmputil.Tag) map[string]string {
	tags := make(map[string]string)
	for _, tag := range t {
		if v, ok := tags[tag.Tag]; ok {
			// example: "component" -> "diskarray|storage"
			tags[tag.Tag] = v + "|" + tag.Value
		} else {
			// example: "component" -> "storage"
			tags[tag.Tag] = tag.Value
		}
	}

	return tags
}
