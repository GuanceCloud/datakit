// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package vsphere collects vsphere metrics.
package vsphere

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/confd/log"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	catalogName = "vmware"
	inputName   = "vsphere"
	l           = logger.DefaultSLogger(inputName)

	_ inputs.InputV2       = (*Input)(nil)
	_ inputs.ElectionInput = &Input{}
)

const (
	maxInterval = 1 * time.Minute
	minInterval = 1 * time.Second
)

type Input struct {
	*dknet.TLSClientConfig

	Interval                datakit.Duration `toml:"interval"`
	Vcenter                 string           `toml:"vcenter"`
	Username                string           `toml:"username"`
	Password                string           `toml:"password"`
	Election                bool             `toml:"election"`
	Timeout                 datakit.Duration `toml:"timeout"`
	DatacenterInstances     bool             `toml:"datacenter_instances"`
	DatacenterMetricInclude []string         `toml:"datacenter_metric_include"`
	DatacenterMetricExclude []string         `toml:"datacenter_metric_exclude"`
	DatacenterInclude       []string         `toml:"datacenter_include"`
	DatacenterExclude       []string         `toml:"datacenter_exclude"`
	ClusterInstances        bool             `toml:"cluster_instances"`
	ClusterMetricInclude    []string         `toml:"cluster_metric_include"`
	ClusterMetricExclude    []string         `toml:"cluster_metric_exclude"`
	ClusterInclude          []string         `toml:"cluster_include"`
	ClusterExclude          []string         `toml:"cluster_exclude"`
	// ResourcePoolInstances     bool             `toml:"resource_pool_instances"`
	// ResourcePoolMetricInclude []string         `toml:"resource_pool_metric_include"`
	// ResourcePoolMetricExclude []string         `toml:"resource_pool_metric_exclude"`
	// ResourcePoolInclude       []string         `toml:"resource_pool_include"`
	// ResourcePoolExclude       []string         `toml:"resource_pool_exclude"`
	HostInstances           bool             `toml:"host_instances"`
	HostMetricInclude       []string         `toml:"host_metric_include"`
	HostMetricExclude       []string         `toml:"host_metric_exclude"`
	HostInclude             []string         `toml:"host_include"`
	HostExclude             []string         `toml:"host_exclude"`
	VMInstances             bool             `toml:"vm_instances"`
	VMMetricInclude         []string         `toml:"vm_metric_include"`
	VMMetricExclude         []string         `toml:"vm_metric_exclude"`
	VMInclude               []string         `toml:"vm_include"`
	VMExclude               []string         `toml:"vm_exclude"`
	DatastoreInstances      bool             `toml:"datastore_instances"`
	DatastoreMetricInclude  []string         `toml:"datastore_metric_include"`
	DatastoreMetricExclude  []string         `toml:"datastore_metric_exclude"`
	DatastoreInclude        []string         `toml:"datastore_include"`
	DatastoreExclude        []string         `toml:"datastore_exclude"`
	MaxQueryObjects         int              `toml:"max_query_objects"`
	MaxQueryMetrics         int              `toml:"max_query_metrics"`
	HistoricalInterval      datakit.Duration `toml:"historical_interval"`
	ObjectDiscoveryInterval datakit.Duration `toml:"object_discovery_interval"`

	Tags           map[string]string `toml:"tags"`
	client         *Client
	tail           *tailer.Tailer
	pause          bool
	pauseCh        chan bool
	semStop        *cliutils.Sem // start stop signal
	feeder         dkio.Feeder
	collectCache   []*point.Point
	collectLogs    []*point.Point
	collectObjects []*point.Point
	g              *goroutine.Group
	ptsTime        time.Time
	mutex          sync.Mutex
	duration       time.Duration
	timeout        time.Duration
	vcenter        *url.URL

	hostLastLogTimes map[string]time.Time
	vmLastLogTimes   map[string]time.Time

	datastoreLastLogTimes map[string]time.Time
	networkLastLogTimes   map[string]time.Time
}

func (*Input) Catalog() string { return catalogName }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&clusterMeasurement{},
		&datastoreMeasurement{},
		&hostMeasurement{},
		&vmMeasurement{},
		&clusterObject{},
		&datastoreObject{},
		&hostObject{},
		&vmObject{},
		&eventMeasurement{},
	}
}

func (*Input) SampleConfig() string { return sampleConfig }

func (ipt *Input) init() error {
	if ipt.client != nil {
		return nil
	}

	u, err := url.Parse(ipt.Vcenter)
	if err != nil {
		l.Errorf("failed to parse vcenter URL %s: %s", ipt.Vcenter, err)
	}
	u.Path = "/sdk"
	ipt.vcenter = u

	if err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		return fmt.Errorf("failed to create vSphere client: %w", err)
	}

	if ipt.client, err = ipt.getClient(); err != nil {
		return fmt.Errorf("failed to create vSphere client: %w", err)
	}

	return ipt.startDiscovery()
}

func (ipt *Input) startDiscovery() error {
	if ipt.ObjectDiscoveryInterval.Duration > 0 {
		g := goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName(inputName)})
		g.Go(func(ctx context.Context) error {
			ticker := time.NewTicker(ipt.ObjectDiscoveryInterval.Duration)
			defer ticker.Stop()
			for {
				select {
				case <-datakit.Exit.Wait():
					return nil
				case <-ipt.semStop.Wait():
					return nil
				case <-ticker.C:
				case ipt.pause = <-ipt.pauseCh:
				}

				if !ipt.pause {
					if err := ipt.client.discover(context.Background()); err != nil {
						l.Errorf("failed to discover: %w", err)
					}
				}
			}
		})
	}
	return nil
}

func (ipt *Input) Collect() error {
	if err := ipt.testClient(context.Background()); err != nil {
		ipt.client, err = ipt.getClient()
		if err != nil {
			return fmt.Errorf("failed to create vsphere client: %w", err)
		}
	}

	if ipt.ObjectDiscoveryInterval.Duration == 0 {
		if err := ipt.client.discover(context.Background()); err != nil {
			return fmt.Errorf("failed to discover: %w", err)
		}
	}

	for k, res := range ipt.client.resourceKinds {
		if res.enabled {
			func(resourceType string) {
				ipt.g.Go(func(ctx context.Context) error {
					ipt.collectResource(resourceType)
					ipt.collectResourceObject(resourceType)
					ipt.collectResourceEvent(resourceType)
					return nil
				})
			}(k)
		}
	}

	return ipt.g.Wait()
}

// collectResourceEvent collects events from vcenter.
func (ipt *Input) collectResourceEvent(resourceType string) {
	client := ipt.client
	res := client.resourceKinds[resourceType]
	eventManager := event.NewManager(client.Client.Client)

	pts := []*point.Point{}
	for _, obj := range res.objects {
		latestTime := ipt.ptsTime
		if obj.lastLogTime == nil {
			obj.lastLogTime = make(map[string]time.Time)
		} else if t, exists := obj.lastLogTime[resourceType]; exists {
			latestTime = t
		}

		filter := types.EventFilterSpec{}
		filter.Time = &types.EventFilterSpecByTime{
			BeginTime: &latestTime,
		}
		filter.Entity = &types.EventFilterSpecByEntity{
			Entity:    obj.ref,
			Recursion: types.EventFilterSpecRecursionOptionAll,
		}

		events, err := eventManager.QueryEvents(context.Background(), filter)
		if err != nil {
			l.Errorf("Error querying events for %s,  %s: %s", resourceType, obj.name, err.Error())
			continue
		}

		if len(events) == 0 {
			continue
		}

		for _, e := range events {
			event := e.GetEvent()
			if event != nil {
				tags := map[string]string{
					"host":       ipt.vcenter.Host,
					ResourceType: resourceType,
				}
				client.populateTags(obj, res, tags, performance.MetricSeries{})

				m := &Log{
					source:   EventMeasurementName,
					tags:     tags,
					election: ipt.Election,
				}

				pt := m.Point()

				if eventEx, ok := e.(*types.EventEx); ok {
					pt.AddTag(EventTypeID, eventEx.EventTypeId)
					pt.AddTag(ObjectName, eventEx.ObjectName)
					switch eventEx.Severity {
					case Warning:
						pt.SetTag(Status, Warning)
					case Error:
						pt.SetTag(Status, Error)
					}
				}

				pt.SetTag(Status, Info)
				pt.AddTag(ObjectName, obj.name)
				pt.AddTag(EventTypeID, fmt.Sprintf("%T", e))
				pt.AddTag(UserName, event.UserName)
				pt.SetTime(event.CreatedTime)
				pt.Add(ChangeTag, event.ChangeTag)
				pt.Add(ChainID, event.ChainId)
				pt.Add(Message, event.FullFormattedMessage)
				pt.Add(EventKey, event.Key)
				if event.CreatedTime.After(latestTime) {
					latestTime = event.CreatedTime
				}

				// custom tags
				for k, v := range ipt.Tags {
					pt.AddTag(k, v)
				}

				pts = append(pts, pt)
			}
		}
		obj.lastLogTime[resourceType] = latestTime
	}

	if len(pts) > 0 {
		ipt.mutex.Lock()
		defer ipt.mutex.Unlock()
		ipt.collectLogs = append(ipt.collectLogs, pts...)
	}
}

func (ipt *Input) collectResourceObject(resourceType string) {
	client := ipt.client
	res := client.resourceKinds[resourceType]
	pts := []*point.Point{}
	for _, obj := range res.objects {
		tags, fields := obj.objectTags, obj.objectFields

		if len(tags) == 0 && len(fields) == 0 {
			continue
		}

		tags["host"] = ipt.vcenter.Host

		client.populateTags(obj, res, tags, performance.MetricSeries{})

		m := &Object{
			class:    fmt.Sprintf("vsphere_%s", resourceType),
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}
		pt := m.Point()

		pt.SetTime(ntp.Now())

		// custom tags
		for k, v := range ipt.Tags {
			pt.AddTag(k, v)
		}

		pts = append(pts, pt)
	}
	if len(pts) > 0 {
		ipt.mutex.Lock()
		defer ipt.mutex.Unlock()
		ipt.collectObjects = append(ipt.collectObjects, pts...)
	}
}

func (ipt *Input) collectResource(resourceType string) {
	ctx := context.Background()
	client := ipt.client
	res := client.resourceKinds[resourceType]

	maxMetrics := ipt.MaxQueryMetrics
	if maxMetrics < 1 {
		maxMetrics = 1
	}

	if res.name == "cluster" && maxMetrics > 10 {
		maxMetrics = 10
	}

	now, err := client.GetServerTime(ctx)
	if err != nil {
		l.Errorf("Failed to get server time: %s", err.Error())
		return
	}

	// Estimate the interval at which we're invoked. Use local time (not server time)
	// since this is about how we got invoked locally.
	localNow := time.Now()
	estInterval := time.Minute
	if !res.lastColl.IsZero() {
		s := time.Duration(res.sampling) * time.Second
		rawInterval := localNow.Sub(res.lastColl)
		paddedInterval := rawInterval + time.Duration(res.sampling/2)*time.Second
		estInterval = paddedInterval.Truncate(s)
		if estInterval < s {
			estInterval = s
		}
	}

	res.lastColl = localNow

	pqs := make(queryChunk, 0, ipt.MaxQueryObjects)
	numQs := 0

	for _, obj := range res.objects {
		specs := make([]*types.PerfQuerySpec, 0)
		makeSpec := func() *types.PerfQuerySpec {
			spec := types.PerfQuerySpec{
				Entity:     obj.ref,
				MaxSample:  maxSampleConst,
				MetricId:   make([]types.PerfMetricId, 0),
				IntervalId: res.sampling,
				Format:     "normal",
			}
			if res.realTime {
				spec.MaxSample = 1
			} else {
				start := now.Add(-2 * time.Hour)
				spec.StartTime = &start
			}

			return &spec
		}

		var spec *types.PerfQuerySpec

		for _, metric := range res.metrics {
			metricName := client.getMetricNameForID(metric.CounterId)
			if metricName == "" {
				l.Debugf("Unable to find metric name for counter ID %d", metric.CounterId)
				continue
			}
			if spec == nil {
				spec = makeSpec()
			}
			spec.MetricId = append(spec.MetricId, metric)

			if (!res.realTime && len(spec.MetricId) >= maxMetrics) || len(spec.MetricId) > maxRealtimeMetrics {
				specs = append(specs, spec)
				spec = nil
			}
		}
		if spec != nil {
			specs = append(specs, spec)
		}

		for _, spec := range specs {
			pqs = append(pqs, *spec)
			numQs += len(spec.MetricId)
			if (!res.realTime && numQs > ipt.MaxQueryObjects) || numQs > maxRealtimeMetrics {
				ipt.collectResourcePoints(pqs, res, estInterval)
				pqs = make(queryChunk, 0, ipt.MaxQueryObjects)
				numQs = 0
			}
		}
	}

	if len(pqs) > 0 {
		ipt.collectResourcePoints(pqs, res, estInterval)
	}
}

type queryChunk []types.PerfQuerySpec

func (ipt *Input) collectResourcePoints(specs queryChunk, res *resourceKind, interval time.Duration) {
	latestSample := time.Time{}
	count := 0
	resourceType := res.name
	prefix := "vsphere_" + resourceType

	client := ipt.client

	metricInfo, err := client.CounterInfoByName(context.Background())
	if err != nil {
		l.Warnf("Failed to get counter info: %s", err.Error())
		return
	}
	ems, err := client.QueryMetrics(context.Background(), specs)
	if err != nil {
		l.Warnf("Failed to query metrics: %s", err.Error())
		return
	}

	l.Debugf("Query for %s returned metrics for %d objects", resourceType, len(ems))

	// Iterate through results
	for _, em := range ems {
		moid := em.Entity.Reference().Value
		instInfo, found := res.objects[moid]
		if !found {
			l.Errorf("MOID %s not found in cache. Skipping! (This should not happen!)", moid)
			continue
		}
		buckets := make(map[string]metricEntry)
		for _, v := range em.Value {
			name := v.Name
			t := map[string]string{
				"host":   ipt.vcenter.Host,
				"source": instInfo.name,
				"moid":   moid,
			}

			// Populate tags
			objectRef, ok := res.objects[moid]
			if !ok {
				l.Errorf("MOID %s not found in cache. Skipping", moid)
				continue
			}
			client.populateTags(objectRef, res, t, v)

			nValues := 0
			alignedInfo, alignedValues := client.alignSamples(em.SampleInfo, v.Value, interval)

			for idx, sample := range alignedInfo {
				// According to the docs, SampleInfo and Value should have the same length, but we've seen corrupted
				// data coming back with missing values. Take care of that gracefully!
				if idx >= len(alignedValues) {
					l.Debugf("Len(SampleInfo)>len(Value) %d > %d", len(alignedInfo), len(alignedValues))
					break
				}
				ts := sample.Timestamp
				if ts.After(latestSample) {
					latestSample = ts
				}
				nValues++

				// Organize the metrics into a bucket per measurement.
				mn, fn := client.makeMetricIdentifier(prefix, name)
				bKey := mn + " " + v.Instance + " " + strconv.FormatInt(ts.UnixNano(), 10)
				bucket, found := buckets[bKey]
				if !found {
					bucket = metricEntry{name: mn, ts: ts, fields: make(map[string]interface{}), tags: t}
					buckets[bKey] = bucket
				}

				// Percentage values must be scaled down by 100.
				info, ok := metricInfo[name]
				if !ok {
					l.Errorf("Could not determine unit for %s. Skipping", name)
				}
				v := alignedValues[idx]
				if info.UnitInfo.GetElementDescription().Key == "percent" {
					bucket.fields[fn] = v / 100.0
				} else {
					bucket.fields[fn] = v
				}
				count++
			}
			if nValues == 0 {
				l.Debugf("Missing value for: %s, %s", name, objectRef.name)
				continue
			}
		}

		ipt.makePoints(buckets)
	}
	if latestSample.After(res.latestSample) && !latestSample.IsZero() {
		res.latestSample = latestSample
	}
}

func (ipt *Input) makePoints(buckets map[string]metricEntry) {
	pts := make([]*point.Point, 0)

	for _, bucket := range buckets {
		m := &Measurement{
			name:     bucket.name,
			tags:     bucket.tags,
			fields:   bucket.fields,
			election: ipt.Election,
		}

		pt := m.Point()
		pt.SetTime(bucket.ts)

		for k, v := range ipt.Tags {
			pt.AddTag(k, v)
		}

		pts = append(pts, pt)
	}

	ipt.mutex.Lock()
	defer ipt.mutex.Unlock()
	ipt.collectCache = append(ipt.collectCache, pts...)
}

func (ipt *Input) Run() {
	ipt.setup()

	ipt.duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	ipt.timeout = config.ProtectedInterval(10*time.Second, time.Minute, ipt.Timeout.Duration)

	tick := time.NewTicker(ipt.duration)
	defer tick.Stop()

	l.Infof("%s input started", inputName)

	for {
		if err := ipt.init(); err != nil {
			l.Errorf("failed to init servers: %s", err.Error())
		} else {
			break
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			log.Info("vsphere input exit")

			return
		case <-ipt.semStop.Wait():
			ipt.exit()
			log.Info("vsphere input return")

			return
		case <-tick.C:
		case ipt.pause = <-ipt.pauseCh:
		}
	}

	ipt.ptsTime = ntp.Now()

	for {
		if !ipt.pause {
			collectStart := time.Now()
			l.Debugf("vsphere input gathering...")

			if err := ipt.Collect(); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
				)

				l.Errorf("collect failed: %s", err.Error())
			} else {
				l.Debugf("collect cache length: %d", len(ipt.collectCache))
				if len(ipt.collectCache) > 0 {
					if err := ipt.feeder.Feed(point.Metric, ipt.collectCache,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(inputName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
						)
					}
				}

				l.Debugf("collect log length: %d", len(ipt.collectLogs))
				if len(ipt.collectLogs) > 0 {
					if err := ipt.feeder.Feed(point.Logging, ipt.collectLogs,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(inputName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
							metrics.WithLastErrorSource(inputName),
						)
						l.Errorf("feed logging: %s", err)
					}
				}

				l.Debugf("collect object length: %d", len(ipt.collectObjects))
				if len(ipt.collectObjects) > 0 {
					if err := ipt.feeder.Feed(point.CustomObject, ipt.collectObjects,
						dkio.WithCollectCost(time.Since(collectStart)),
						dkio.WithElection(ipt.Election),
						dkio.WithSource(dkio.FeedSource(inputName, "CO")),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(inputName),
						)
					}
				}

				ipt.collectLogs = ipt.collectLogs[:0]
				ipt.collectCache = ipt.collectCache[:0]
				ipt.collectObjects = ipt.collectObjects[:0]
			}
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			log.Info("vsphere input exit")

			return
		case <-ipt.semStop.Wait():
			ipt.exit()
			log.Info("vsphere input return")

			return
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	ipt.pauseCh = make(chan bool, inputs.ElectionPauseChannelLength)
	ipt.semStop = cliutils.NewSem()
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil

	case <-datakit.Exit.Wait():
		log.Info("pause vsphere interrupted by global exit.")
		return nil

	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		log.Info("vsphere log exits")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		DatacenterInclude: []string{"/*"},
		ClusterInclude:    []string{"/*/host/**"},
		HostInstances:     true,
		HostInclude:       []string{"/*/host/**"},
		// ResourcePoolInclude:     []string{"/*/host/**"},
		VMInstances:             true,
		VMInclude:               []string{"/*/vm/**"},
		DatastoreInclude:        []string{"/*/datastore/**"},
		ObjectDiscoveryInterval: datakit.Duration{Duration: time.Second * 300},
		Timeout:                 datakit.Duration{Duration: time.Second * 60},
		HistoricalInterval:      datakit.Duration{Duration: time.Second * 300},
		MaxQueryObjects:         256,
		MaxQueryMetrics:         256,
		feeder:                  dkio.DefaultFeeder(),
		semStop:                 cliutils.NewSem(),
		Election:                true,
		g:                       goroutine.NewGroup(goroutine.Option{Name: goroutine.GetInputName("vsphere")}),
		vmLastLogTimes:          make(map[string]time.Time),
		hostLastLogTimes:        make(map[string]time.Time),
		datastoreLastLogTimes:   make(map[string]time.Time),
		networkLastLogTimes:     make(map[string]time.Time),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
