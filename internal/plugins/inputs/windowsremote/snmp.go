// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package windowsremote

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/gosnmp/gosnmp"
)

type snmp struct {
	cfg *SnmpConfig
}

func newSnmp(cfg *SnmpConfig) *snmp {
	return &snmp{cfg: cfg}
}

func (s *snmp) Name() string { return "snmp" }

func (s *snmp) CollectMetric(ip string, timestamp int64) (pts []*point.Point) {
	for _, port := range s.cfg.Ports {
		client, err := newSnmpClient(s.cfg, ip, port)
		if err != nil {
			l.Warn(err)
			return nil
		}
		defer client.Close()

		stats, errors := client.getAllStats(false)
		for idx, err := range errors {
			l.Warnf("errros[%d]: %s", idx, err)
		}

		pts = append(pts, s.systemMetric(stats))
		pts = append(pts, s.cpuMetric(stats))
		pts = append(pts, s.memMetric(stats))
		pts = append(pts, s.netMetric(stats)...)
		pts = append(pts, s.diskMetric(stats)...)
	}
	return
}

func (s *snmp) CollectObject(ip string) (pts []*point.Point) {
	for _, port := range s.cfg.Ports {
		client, err := newSnmpClient(s.cfg, ip, port)
		if err != nil {
			l.Warn(err)
			return nil
		}
		defer client.Close()

		stats, errors := client.getAllStats(true)
		for idx, err := range errors {
			l.Warnf("errros[%d]: %s", idx, err)
		}

		pts = append(pts, s.hostobject(stats))
		pts = append(pts, s.hostprocessesObject(stats)...)
	}
	return
}

func (s *snmp) CollectLogging(ip string) (pts []*point.Point) {
	return
}

func (s *snmp) systemMetric(stats *snmpStats) *point.Point {
	var kvs point.KVs
	kvs = kvs.AddTag("host", stats.hostname)
	kvs = kvs.Add("uptime", stats.uptime)

	kvs = kvs.Add("cpu_total_usage", stats.cpuUsagePercent)
	kvs = kvs.Add("n_cpus", stats.cpuCores)
	// kvs = kvs.Add("process_count", len(stats.processes))

	memUsedPercent := 0.0
	if stats.memSize != 0 && stats.memUsed != 0 {
		memUsedPercent = float64(stats.memUsed) / float64(stats.memSize) * 100 // unit percent
	}
	kvs = kvs.Add("memory_usage", memUsedPercent)

	kvs = append(kvs, point.NewTags(s.cfg.extraTags)...)
	return point.NewPoint(systemMeasurement, kvs, point.DefaultMetricOptions()...)
}

func (s *snmp) cpuMetric(stats *snmpStats) *point.Point {
	var kvs point.KVs
	kvs = kvs.AddTag("host", stats.hostname)
	kvs = kvs.Add("usage_total", stats.cpuUsagePercent)

	kvs = append(kvs, point.NewTags(s.cfg.extraTags)...)
	return point.NewPoint(cpuMeasurement, kvs, point.DefaultMetricOptions()...)
}

func (s *snmp) memMetric(stats *snmpStats) *point.Point {
	var kvs point.KVs
	kvs = kvs.AddTag("host", stats.hostname)
	kvs = kvs.Add("total", stats.memSize)
	kvs = kvs.Add("used", stats.memUsed)
	kvs = kvs.Add("available", stats.memSize-stats.memUsed)

	if stats.memSize != 0 && stats.memUsed != 0 {
		usedPercent := float64(stats.memUsed) / float64(stats.memSize) * 100                    // unit percent
		availablePercent := float64(stats.memSize-stats.memUsed) / float64(stats.memSize) * 100 // unit percent

		kvs = kvs.Add("used_percent", usedPercent)
		kvs = kvs.Add("available_percent", availablePercent)
	}

	kvs = append(kvs, point.NewTags(s.cfg.extraTags)...)
	return point.NewPoint(memoryMeasurement, kvs, point.DefaultMetricOptions()...)
}

func (s *snmp) netMetric(stats *snmpStats) []*point.Point {
	var pts []*point.Point
	for _, net := range stats.networks {
		var kvs point.KVs
		kvs = kvs.AddTag("host", stats.hostname)
		kvs = kvs.AddTag("interface", net.name)

		kvs = kvs.Add("bytes_recv", net.inBytes)
		kvs = kvs.Add("bytes_sent", net.outBytes)
		kvs = kvs.Add("drop_in", net.inDiscards)
		kvs = kvs.Add("drop_out", net.outDiscards)
		kvs = kvs.Add("err_in", net.inErrors)
		kvs = kvs.Add("err_out", net.outErrors)

		kvs = append(kvs, point.NewTags(s.cfg.extraTags)...)
		pts = append(pts, point.NewPoint(netMeasurement, kvs, point.DefaultMetricOptions()...))
	}
	return pts
}

func (s *snmp) diskMetric(stats *snmpStats) []*point.Point {
	var pts []*point.Point
	for _, storage := range stats.storages {
		if storage.name == "Physical Memory" || storage.name == "Virtual Memory" {
			continue
		}

		var kvs point.KVs
		kvs = kvs.AddTag("host", stats.hostname)
		kvs = kvs.AddTag("device", storage.name)

		kvs = kvs.Add("total", storage.size)
		kvs = kvs.Add("used", storage.used)
		kvs = kvs.Add("free", storage.size-storage.used)

		if storage.size != 0 && storage.used != 0 {
			usedPercent := float64(storage.used) / float64(storage.size) * 100 // unit percent
			kvs = kvs.Add("used_percent", usedPercent)
		}

		kvs = append(kvs, point.NewTags(s.cfg.extraTags)...)
		pts = append(pts, point.NewPoint(diskMeasurement, kvs, point.DefaultMetricOptions()...))
	}
	return pts
}

func (s *snmp) hostprocessesObject(stats *snmpStats) []*point.Point {
	var pts []*point.Point
	for _, process := range stats.processes {
		var kvs point.KVs

		kvs = kvs.AddTag("host", stats.hostname)
		kvs = kvs.AddTag("process_name", process.name)
		kvs = kvs.Add("pid", process.id)

		name := fmt.Sprintf("%s_%d", stats.hostname, process.id)
		kvs = kvs.AddTag("name", name)

		cmdline := fmt.Sprintf("%s%s %s", process.path, process.name, process.parameters)
		kvs = kvs.Add("cmdline", cmdline)

		msg := &HostprocessesObjectMessage{
			Pid:         process.id,
			Name:        name,
			ProcessName: process.name,
			Cmdline:     cmdline,
		}

		message, err := json.Marshal(msg)
		if err != nil {
			l.Warnf("message marshal err: %s", err)
		}
		kvs = kvs.Add("message", string(message))

		kvs = append(kvs, point.NewTags(s.cfg.extraTags)...)
		pts = append(pts, point.NewPoint(hostprocessesObjectMeasurement, kvs, point.DefaultObjectOptions()...))
	}
	return pts
}

func (s *snmp) hostobject(stats *snmpStats) *point.Point {
	var kvs point.KVs

	kvs = kvs.AddTag("name", stats.hostname)
	kvs = kvs.AddTag("host", stats.hostname)
	kvs = kvs.AddTag("arch", stats.arch)
	kvs = kvs.AddTag("os", "windows")
	kvs = kvs.Add("start_time", stats.startTime)

	kvs = kvs.Add("num_cpu", stats.cpuCores)
	kvs = kvs.Add("cpu_usage", float64(stats.cpuUsagePercent))

	kvs = kvs.Add("mem_total", stats.memSize)
	kvs = kvs.Add("mem_used", stats.memUsed)

	memUsedPercent := 0.0
	if stats.memSize != 0 && stats.memUsed != 0 {
		memUsedPercent = float64(stats.memUsed) / float64(stats.memSize) * 100 // unit percent
	}
	kvs = kvs.Add("mem_used_percent", memUsedPercent)

	msg := &HostObjectMessage{
		Host: &HostInfo{
			HostMeta: &HostMetaInfo{
				HostName: stats.hostname,
				BootTime: uint64(stats.startTime / 1000), /* seconds */
				OS:       "windows",
				Arch:     stats.arch,
			},
			CPU: []*CPUInfo{
				{Cores: int32(stats.cpuCores)},
			},
			Mem: &MemInfo{
				MemoryTotal: uint64(stats.memSize),
			},
		},
	}

	for _, net := range stats.networks {
		msg.Host.Net = append(msg.Host.Net, &NetInfo{Name: net.name})
	}

	for _, storage := range stats.storages {
		if storage.name == "Physical Memory" || storage.name == "Virtual Memory" {
			continue
		}
		msg.Host.Disk = append(msg.Host.Disk, &DiskInfo{Device: storage.name, Total: uint64(storage.size)})
	}

	message, err := json.Marshal(msg)
	if err != nil {
		l.Warnf("message marshal err: %s", err)
	}
	kvs = kvs.Add("message", string(message))

	kvs = append(kvs, point.NewTags(s.cfg.extraTags)...)
	return point.NewPoint(hostobjectMeasurement, kvs, point.DefaultObjectOptions()...)
}

type snmpClient struct {
	ip     string
	client *gosnmp.GoSNMP
}

func newSnmpClient(cfg *SnmpConfig, ip string, port int) (*snmpClient, error) {
	client := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      uint16(port),
		Community: cfg.Community,
		Timeout:   time.Second * 3,
		Version:   gosnmp.Version2c,
	}

	if err := client.Connect(); err != nil {
		return nil, err
	}

	return &snmpClient{ip: ip, client: client}, nil
}

func (sc *snmpClient) Close() {
	if err := sc.client.Conn.Close(); err != nil {
		l.Warnf("close conn error: %s", err)
	}
}

func (sc *snmpClient) getAllStats(collectProcesses bool) (*snmpStats, []error) {
	stats := &snmpStats{ip: sc.ip}
	var errors []error

	hostname, arch, uptime, startTime, err := sc.getHostInfo()
	if err != nil {
		errors = append(errors, err)
	} else {
		stats.hostname = hostname
		stats.arch = arch
		stats.uptime = int(uptime)
		stats.startTime = int(startTime)
	}

	cpuCores, cpuUsagePercent, err := sc.getCPUCoresAndUsage()
	if err != nil {
		errors = append(errors, err)
	} else {
		stats.cpuCores = cpuCores
		stats.cpuUsagePercent = cpuUsagePercent
	}

	// 进程采集耗时较多，选择性开启
	if collectProcesses {
		processes, err := sc.getProcesses()
		if err != nil {
			errors = append(errors, err)
		} else {
			stats.processes = processes
		}
	}

	networks, err := sc.getNetworkStats()
	if err != nil {
		errors = append(errors, err)
	} else {
		stats.networks = networks
	}

	storages, err := sc.getStorageStats()
	if err != nil {
		errors = append(errors, err)
	} else {
		stats.storages = storages
	}

	for _, storage := range stats.storages {
		if storage.name == "Physical Memory" {
			stats.memSize = storage.size
			stats.memUsed = storage.used
		}
	}

	return stats, errors
}

func (sc *snmpClient) getHostInfo() (hostname, arch string, uptime, startTime int64, err error) {
	oids := []string{
		".1.3.6.1.2.1.1.5.0", // string, hostname
		".1.3.6.1.2.1.1.1.0", // string, arch
		".1.3.6.1.2.1.1.3.0", // time-ticks, uptime, 0.01s
	}

	res, err := sc.client.Get(oids)
	if err != nil {
		err = fmt.Errorf("hostinfo get error: %w", err)
		return
	}

	if len(res.Variables) != len(oids) {
		err = fmt.Errorf("unexpected hostinfo response: expect len(%d), actual len(%d)", len(oids), len(res.Variables))
		return
	}

	if res.Variables[0].Type != gosnmp.OctetString {
		err = fmt.Errorf("unexpected hostname resp: expect OctetString, actual %s", res.Variables[0].Type)
	} else {
		s, ok := res.Variables[0].Value.([]byte)
		if !ok {
			err = fmt.Errorf("converting to string failed, %v", res.Variables[0].Value)
			return
		}
		hostname = string(s)
	}

	if res.Variables[1].Type != gosnmp.OctetString {
		err = fmt.Errorf("unexpected arch resp: expect OctetString, actual %s", res.Variables[1].Type)
	} else {
		s, ok := res.Variables[1].Value.([]byte)
		if !ok {
			err = fmt.Errorf("converting to string failed, %v", res.Variables[1].Value)
			return
		}
		arch = string(s)
	}

	if res.Variables[2].Type != gosnmp.TimeTicks {
		err = fmt.Errorf("unexpected uptime resp: expect TimeTicks, actual %s", res.Variables[2].Type)
	} else {
		n := gosnmp.ToBigInt(res.Variables[2].Value).Int64()

		uptime = n * 10 // converting to millisecond
		duration := time.Duration(uptime) * time.Millisecond
		startTime = time.Now().Add(-duration).UnixMilli()
	}

	return // nolint:nakedret
}

func (sc *snmpClient) getCPUCoresAndUsage() (cores, usagePercent int, err error) {
	oid := ".1.3.6.1.2.1.25.3.3.1.2" // int, hrProcessorLoad

	values, err := sc.client.BulkWalkAll(oid)
	if err != nil {
		err = fmt.Errorf("cpu walk error: %w", err)
		return
	}

	for _, value := range values {
		if value.Type == gosnmp.Integer {
			n := gosnmp.ToBigInt(value.Value).Int64()
			cores += 1
			usagePercent += int(n)
		}
	}
	return
}

func (sc *snmpClient) getProcesses() ([]*snmpProcess, error) {
	oids := []string{
		".1.3.6.1.2.1.25.4.2.1.1", // int, hrSWRunIndex
		".1.3.6.1.2.1.25.4.2.1.2", // string, hrSWRunName
		".1.3.6.1.2.1.25.4.2.1.4", // string, hrSWRunPath
		".1.3.6.1.2.1.25.4.2.1.5", // string, hrSWRunParameters
	}

	indexes, err := sc.client.BulkWalkAll(oids[0])
	if err != nil {
		return nil, fmt.Errorf("processes walk err: %w", err)
	}

	var (
		nameOids      []string
		pathOids      []string
		parameterOids []string

		names      []string
		paths      []string
		parameters []string

		res       []*snmpProcess
		batchSize = 20
	)

	for _, value := range indexes {
		if value.Type != gosnmp.Integer {
			return nil, fmt.Errorf("unexpected hrSWRunIndex resp: expect Integer, actual %s", value.Type)
		}

		id := gosnmp.ToBigInt(value.Value).Int64()
		res = append(res, &snmpProcess{id: id})

		index := strconv.Itoa(int(id))

		nameOids = append(nameOids, oids[1]+"."+index)
		pathOids = append(pathOids, oids[2]+"."+index)
		parameterOids = append(parameterOids, oids[3]+"."+index)
	}

	{
		for i := 0; i < len(nameOids); i += batchSize {
			end := i + batchSize
			if end > len(nameOids) {
				end = len(nameOids)
			}
			batch := nameOids[i:end]

			res, err := sc.client.Get(batch)
			if err != nil {
				return nil, fmt.Errorf("hrSWRunName get error: %w", err)
			}

			for _, value := range res.Variables {
				if value.Type != gosnmp.OctetString {
					return nil, fmt.Errorf("unexpected hrSWRunName resp: expect OctetString, actual %s", value.Type)
				}
				s, ok := value.Value.([]byte)
				if !ok {
					return nil, fmt.Errorf("converting to string failed, %v", value.Value)
				}
				names = append(names, string(s))
			}
		}
	}

	{
		for i := 0; i < len(pathOids); i += batchSize {
			end := i + batchSize
			if end > len(pathOids) {
				end = len(pathOids)
			}
			batch := pathOids[i:end]

			res, err := sc.client.Get(batch)
			if err != nil {
				return nil, fmt.Errorf("hrSWRunPath get error: %w", err)
			}

			for _, value := range res.Variables {
				if value.Type != gosnmp.OctetString {
					return nil, fmt.Errorf("unexpected hrSWRunPath resp: expect OctetString, actual %s", value.Type)
				}
				s, ok := value.Value.([]byte)
				if !ok {
					return nil, fmt.Errorf("converting to string failed, %v", value.Value)
				}
				paths = append(paths, string(s))
			}
		}
	}

	{
		for i := 0; i < len(parameterOids); i += batchSize {
			end := i + batchSize
			if end > len(parameterOids) {
				end = len(parameterOids)
			}
			batch := parameterOids[i:end]

			res, err := sc.client.Get(batch)
			if err != nil {
				return nil, fmt.Errorf("hrSWRunParameters get error: %w", err)
			}

			for _, value := range res.Variables {
				if value.Type != gosnmp.OctetString {
					return nil, fmt.Errorf("unexpected hrSWRunParameters resp: expect OctetString, actual %s", value.Type)
				}
				s, ok := value.Value.([]byte)
				if !ok {
					return nil, fmt.Errorf("converting to string failed, %v", value.Value)
				}
				parameters = append(parameters, string(s))
			}
		}
	}

	if !(len(res) == len(names) && len(names) == len(paths) && len(paths) == len(parameters)) {
		return nil, fmt.Errorf("unexpected processes response: expect len(%d)", len(res))
	}

	for idx, name := range names {
		res[idx].name = name
	}

	for idx, path := range paths {
		res[idx].path = path
	}

	for idx, parameters := range parameters {
		res[idx].parameters = parameters
	}

	return res, nil
}

func (sc *snmpClient) getStorageStats() ([]*snmpStorageStats, error) {
	oids := []string{
		".1.3.6.1.2.1.25.2.3.1.3", // string, hrStorageDescr
		".1.3.6.1.2.1.25.2.3.1.4", // int, hrStorageAllocationUnits
		".1.3.6.1.2.1.25.2.3.1.5", // int, hrStorageSize
		".1.3.6.1.2.1.25.2.3.1.6", // int, hrStorageUsed
	}

	var values [][]gosnmp.SnmpPDU

	for _, oid := range oids {
		value, err := sc.client.BulkWalkAll(oid)
		if err != nil {
			return nil, fmt.Errorf("storage walk err: %w", err)
		}
		values = append(values, value)
	}

	if len(values) != len(oids) {
		return nil, fmt.Errorf("unexpected storage response: expect len(%d), actual len(%d)", len(oids), len(values))
	}

	var res []*snmpStorageStats

	for _, value := range values[0] {
		if value.Type != gosnmp.OctetString {
			return nil, fmt.Errorf("unexpected hrStorageDescr resp: expect OctetString, actual %s", value.Type)
		}
		s, ok := value.Value.([]byte)
		if !ok {
			return nil, fmt.Errorf("converting to string failed, %v", value.Value)
		}
		res = append(res, &snmpStorageStats{name: string(s)})
	}

	for idx, value := range values[1] {
		if value.Type != gosnmp.Integer {
			return nil, fmt.Errorf("unexpected hrStorageAllocationUnits resp: expect Integer, actual %s", value.Type)
		}
		res[idx].unit = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[2] {
		if value.Type != gosnmp.Integer {
			return nil, fmt.Errorf("unexpected hrStorageSize resp: expect OctetString, actual %s", value.Type)
		}
		res[idx].size = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[3] {
		if value.Type != gosnmp.Integer {
			return nil, fmt.Errorf("unexpected hrStorageUsed resp: expect Integer, actual %s", value.Type)
		}
		res[idx].used = gosnmp.ToBigInt(value.Value).Int64()
	}

	for _, stats := range res {
		stats.size *= stats.unit
		stats.used *= stats.unit
	}

	return res, nil
}

func (sc *snmpClient) getNetworkStats() ([]*snmpNetworkStats, error) {
	oids := []string{
		".1.3.6.1.2.1.2.2.1.2",  // string, ifDescr, e.g. Ethernet0、Local Area Connection
		".1.3.6.1.2.1.2.2.1.8",  // int, ifOperStatus, up/down/testing/...
		".1.3.6.1.2.1.2.2.1.10", // counter32, ifInOctets
		".1.3.6.1.2.1.2.2.1.13", // counter32, ifInDiscards
		".1.3.6.1.2.1.2.2.1.14", // counter32, ifInErrors
		".1.3.6.1.2.1.2.2.1.16", // counter32, ifOutOctets
		".1.3.6.1.2.1.2.2.1.19", // counter32, ifOutDiscards
		".1.3.6.1.2.1.2.2.1.20", // counter32, ifOutErrors
	}

	var values [][]gosnmp.SnmpPDU

	for _, oid := range oids {
		value, err := sc.client.BulkWalkAll(oid)
		if err != nil {
			return nil, fmt.Errorf("network walk err: %w", err)
		}
		values = append(values, value)
	}

	if len(values) != len(oids) {
		return nil, fmt.Errorf("unexpected network response: expect len(%d), actual len(%d)", len(oids), len(values))
	}

	var res []*snmpNetworkStats

	for _, value := range values[0] {
		if value.Type != gosnmp.OctetString {
			return nil, fmt.Errorf("unexpected ifDescr resp: expect OctetString, actual %s", value.Type)
		}
		s, ok := value.Value.([]byte)
		if !ok {
			return nil, fmt.Errorf("converting to string failed, %v", value.Value)
		}
		res = append(res, &snmpNetworkStats{name: string(s)})
	}

	for idx, value := range values[1] {
		if value.Type != gosnmp.Integer {
			return nil, fmt.Errorf("unexpected ifOperStatus resp: expect Integer, actual %s", value.Type)
		}
		res[idx].up = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[2] {
		if value.Type != gosnmp.Counter32 {
			return nil, fmt.Errorf("unexpected ifInOctets resp: expect Counter32, actual %s", value.Type)
		}
		res[idx].inBytes = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[3] {
		if value.Type != gosnmp.Counter32 {
			return nil, fmt.Errorf("unexpected ifOutOctets resp: expect Counter32, actual %s", value.Type)
		}
		res[idx].inDiscards = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[4] {
		if value.Type != gosnmp.Counter32 {
			return nil, fmt.Errorf("unexpected ifInErrors resp: expect Counter32, actual %s", value.Type)
		}
		res[idx].inErrors = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[5] {
		if value.Type != gosnmp.Counter32 {
			return nil, fmt.Errorf("unexpected ifOutErrors resp: expect Counter32, actual %s", value.Type)
		}
		res[idx].outBytes = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[6] {
		if value.Type != gosnmp.Counter32 {
			return nil, fmt.Errorf("unexpected ifInOctets resp: expect Counter32, actual %s", value.Type)
		}
		res[idx].outDiscards = gosnmp.ToBigInt(value.Value).Int64()
	}

	for idx, value := range values[7] {
		if value.Type != gosnmp.Counter32 {
			return nil, fmt.Errorf("unexpected ifInOctets resp: expect Counter32, actual %s", value.Type)
		}
		res[idx].outErrors = gosnmp.ToBigInt(value.Value).Int64()
	}

	return res, nil
}

type snmpStats struct {
	hostname, ip, arch        string
	uptime, startTime         int // unit milliseconds
	cpuCores, cpuUsagePercent int
	memSize, memUsed          int64
	processes                 []*snmpProcess
	storages                  []*snmpStorageStats
	networks                  []*snmpNetworkStats
}

type snmpProcess struct {
	id                     int64
	name, path, parameters string
}

type snmpNetworkStats struct {
	name                    string
	up                      int64
	inBytes, outBytes       int64
	inErrors, outErrors     int64
	inDiscards, outDiscards int64
}

type snmpStorageStats struct {
	name             string
	unit, size, used int64
}
