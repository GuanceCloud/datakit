// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat heartbeat
package cat

import (
	"encoding/xml"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

/*
示例：
指标集名称 cat
	指标 os.available-processors 16
	指标 os.system-load-average 	0.46
	指标 os.process-time 		9330000000
	等等。。。

	see: readme.md
*/

const (
	metricName = "cat"
)

type Status struct {
	XMLName   xml.Name     `xml:"status"`
	Timestamp string       `xml:"timestamp,attr"`
	Runtime   *Runtime     `xml:"runtime"`
	OS        *OS          `xml:"os"`
	Disk      *Disk        `xml:"disk"`
	Memory    *Memory      `xml:"memory"`
	Thread    *Thread      `xml:"thread"`
	Message   *MessageXML  `xml:"message"`
	Extension []*Extension `xml:"extension"`
}

func (s *Status) toPoint(domain, hostName string) []*point.Point {
	pts := make([]*point.Point, 0)
	log.Debugf("status_time %s", s.Timestamp)
	t, err := time.ParseInLocation("2006-01-02 15:04:05.000", s.Timestamp, time.Local)
	if err != nil {
		t = time.Now()
	}

	if s.Runtime != nil {
		pts = append(pts, s.Runtime.toPoint(domain, hostName, t))
	}
	if s.OS != nil {
		pts = append(pts, s.OS.toPoint(domain, hostName, t))
	}
	if s.Disk != nil {
		pts = append(pts, s.Disk.toPoint(domain, hostName, t)...)
	}
	if s.Memory != nil {
		pts = append(pts, s.Memory.toPoint(domain, hostName, t)...)
	}

	if s.Thread != nil {
		pts = append(pts, s.Thread.toPoint(domain, hostName, t))
	}

	log.Infof("points len = %d", len(pts))
	return pts
}

type Runtime struct {
	StartTime     int64  `xml:"start-time,attr"`
	Uptime        int64  `xml:"up-time,attr"`
	JavaVersion   string `xml:"java-version,attr"`
	UserName      string `xml:"user-name,attr"`
	UserDir       string `xml:"user-dir"`
	JavaClasspath string `xml:"java-classpath"`
}

func (r *Runtime) toPoint(domain, hostName string, t time.Time) *point.Point {
	var kvs point.KVs
	kvs = make([]*point.Field, 0)
	kvs = kvs.AddTag("runtime_java-version", r.JavaVersion)
	kvs = kvs.AddTag("runtime_user-name", r.UserName)
	kvs = kvs.AddTag("runtime_user-dir", r.UserDir)
	kvs = kvs.AddTag("domain", domain)
	kvs = kvs.AddTag("hostName", hostName)

	kvs = kvs.Add("runtime_up-time", r.Uptime, false, true)
	kvs = kvs.Add("runtime_start-time", r.StartTime, false, true)

	opts := point.DefaultMetricOptions()

	opts = append(opts, point.WithTime(t))
	return point.NewPointV2(metricName, kvs, opts...)
}

type OS struct {
	Name                   string  `xml:"name,attr"`
	Arch                   string  `xml:"arch,attr"`
	Version                string  `xml:"version,attr"`
	AvailableProcessors    int     `xml:"available-processors,attr"`
	SystemLoadAverage      float64 `xml:"system-load-average,attr"`
	ProcessTime            int64   `xml:"process-time,attr"`
	TotalPhysicalMemory    int64   `xml:"total-physical-memory,attr"`
	FreePhysicalMemory     int64   `xml:"free-physical-memory,attr"`
	CommittedVirtualMemory int64   `xml:"committed-virtual-memory,attr"`
	TotalSwapSpace         int64   `xml:"total-swap-space,attr"`
	FreeSwapSpace          int64   `xml:"free-swap-space,attr"`
}

func (o *OS) toPoint(domain, hostName string, t time.Time) *point.Point {
	tags := make(map[string]string)
	tags["os_name"] = o.Name
	tags["os_arch"] = o.Arch
	tags["os_version"] = o.Version
	tags["domain"] = domain
	tags["hostName"] = hostName

	fieds := make(map[string]interface{})
	fieds["os_available-processors"] = o.AvailableProcessors
	fieds["os_system-load-average"] = o.SystemLoadAverage
	fieds["os_total-physical-memory"] = o.TotalPhysicalMemory
	fieds["os_free-physical-memory"] = o.FreePhysicalMemory
	fieds["os_committed-virtual-memory"] = o.CommittedVirtualMemory
	fieds["os_total-swap-space"] = o.TotalSwapSpace
	fieds["os_free-swap-space"] = o.FreeSwapSpace

	opts := point.DefaultMetricOptions()

	opts = append(opts, point.WithTime(t))
	return point.NewPointV2(metricName, append(point.NewTags(tags), point.NewKVs(fieds)...), opts...)
}

type Disk struct {
	DiskVolumes []DiskVolume `xml:"disk-volume"`
}

type DiskVolume struct {
	ID     string `xml:"id,attr"`
	Total  int64  `xml:"total,attr"`
	Free   int64  `xml:"free,attr"`
	Usable int64  `xml:"usable,attr"`
}

func (d *Disk) toPoint(domain, hostName string, t time.Time) []*point.Point {
	pts := make([]*point.Point, 0)

	for _, dv := range d.DiskVolumes {
		tags := make(map[string]string)
		tags["os_name"] = dv.ID
		tags["domain"] = domain
		tags["hostName"] = hostName

		fieds := make(map[string]interface{})
		fieds["disk_total"] = dv.Total
		fieds["disk_free"] = dv.Free
		fieds["disk_usable"] = dv.Usable

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(t))
		p := point.NewPointV2(metricName, append(point.NewTags(tags), point.NewKVs(fieds)...), opts...)
		pts = append(pts, p)
	}

	return pts
}

type Memory struct {
	Max          int64 `xml:"max,attr"`
	Total        int64 `xml:"total,attr"`
	Free         int64 `xml:"free,attr"`
	HeapUsage    int64 `xml:"heap-usage,attr"`
	NonHeapUsage int64 `xml:"non-heap-usage,attr"`
	GC           []GC  `xml:"gc"`
}

type GC struct {
	Name  string `xml:"name,attr"`
	Count int    `xml:"count,attr"`
	Time  int    `xml:"time,attr"`
}

func (m *Memory) toPoint(domain, hostName string, t time.Time) []*point.Point {
	pts := make([]*point.Point, 0)

	tags := make(map[string]string)
	tags["domain"] = domain
	tags["hostName"] = hostName

	fieds := make(map[string]interface{})
	fieds["memory_max"] = m.Max
	fieds["memory_total"] = m.Total
	fieds["memory_free"] = m.Free
	fieds["memory_heap-usage"] = m.HeapUsage
	fieds["memory_non-heap-usage"] = m.NonHeapUsage

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(t))

	p := point.NewPointV2(metricName, append(point.NewTags(tags), point.NewKVs(fieds)...), opts...)
	pts = append(pts, p)

	for _, gc := range m.GC {
		gctags := make(map[string]string)
		gctags["domain"] = domain
		gctags["hostName"] = hostName
		gctags["memory_gc_name"] = gc.Name

		gcfieds := make(map[string]interface{})
		gcfieds["memory_gc_count"] = gc.Count
		gcfieds["memory_gc_time"] = gc.Time

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTime(t))

		p := point.NewPointV2(metricName, append(point.NewTags(gctags), point.NewKVs(gcfieds)...), opts...)
		pts = append(pts, p)
	}

	return pts
}

type Thread struct {
	Count             int `xml:"count,attr"`
	DaemonCount       int `xml:"daemon-count,attr"`
	PeekCount         int `xml:"peek-count,attr"`
	TotalStartedCount int `xml:"total-started-count,attr"`
	CatThreadCount    int `xml:"cat-thread-count,attr"`
	PigeonThreadCount int `xml:"pigeon-thread-count,attr"`
	HTTPThreadCount   int `xml:"http-thread-count,attr"`
}

func (th *Thread) toPoint(domain, hostName string, t time.Time) *point.Point {
	tags := make(map[string]string)
	tags["domain"] = domain
	tags["hostName"] = hostName

	fieds := make(map[string]interface{})
	fieds["thread_count"] = th.Count
	fieds["thread_daemon_count"] = th.DaemonCount
	fieds["thread_peek_count"] = th.PeekCount
	fieds["thread_total_started_count"] = th.TotalStartedCount
	fieds["thread_cat_thread_count"] = th.CatThreadCount
	fieds["thread_pigeon_thread_count"] = th.PigeonThreadCount
	fieds["thread_http_thread_count"] = th.HTTPThreadCount

	opts := point.DefaultMetricOptions()

	opts = append(opts, point.WithTime(t))
	return point.NewPointV2(metricName, append(point.NewTags(tags), point.NewKVs(fieds)...), opts...)
}

// MessageXML MessageXML: 没有指标意义。
type MessageXML struct {
	Produced   int64 `xml:"produced,attr"`
	Overflowed int64 `xml:"overflowed,attr"`
	Bytes      int64 `xml:"bytes,attr"`
}

// Extension 重复的指标可以忽略。
type Extension struct {
	ID               string            `xml:"id,attr"`
	ExtensionDetails []ExtensionDetail `xml:"extensionDetail"`
}

type ExtensionDetail struct {
	ID    string  `xml:"id,attr"`
	Value float64 `xml:"value,attr"`
}
