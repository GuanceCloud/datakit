// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Telegraf (https://github.com/influxdata/telegraf.git).

package vsphere

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
)

var isolateLUN = regexp.MustCompile(`.*/([^/]+)/?$`)

var isIPv4 = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)

var isIPv6 = regexp.MustCompile(`^(?:[A-Fa-f0-9]{0,4}:){1,7}[A-Fa-f0-9]{1,4}$`)

const maxSampleConst = 10 // Absolute maximum number of samples regardless of period

const maxMetadataSamples = 100 // Number of resources to sample for metric metadata

const maxRealtimeMetrics = 50000 // Absolute maximum metrics per realtime query

type Client struct {
	Client           *govmomi.Client
	Perf             *performance.Manager
	Timeout          time.Duration
	resourceKinds    map[string]*resourceKind
	apiVersion       string
	metricNameLookup map[int32]string
	collectMux       sync.RWMutex
	metricNameMux    sync.RWMutex
	busy             sync.Mutex
	lun2ds           map[string]string
}

type objectMap map[string]*objectRef

type objectRef struct {
	name         string
	altID        string
	ref          types.ManagedObjectReference
	parentRef    *types.ManagedObjectReference // Pointer because it must be nillable
	guest        string
	dcname       string
	rpname       string // ResourcePool name, default Resources
	lookup       map[string]string
	objectTags   map[string]string
	objectFields map[string]interface{}
	lastLogTime  map[string]time.Time
}

type metricEntry struct {
	tags   map[string]string
	name   string
	ts     time.Time
	fields map[string]interface{}
}

type resourceKind struct {
	name             string
	vcName           string
	pKey             string
	parentTag        string
	enabled          bool
	realTime         bool
	sampling         int32
	objects          objectMap
	filters          filter.Filter
	paths            []string
	excludePaths     []string
	collectInstances bool
	getObjects       func(context.Context, *Client, *ResourceFilter) (objectMap, error)
	include          []string
	simple           bool
	metrics          performance.MetricList
	parent           string
	latestSample     time.Time
	lastColl         time.Time
}

// CounterInfoByKey wraps performance.CounterInfoByKey to give it proper timeouts.
func (c *Client) CounterInfoByKey(ctx context.Context) (map[int32]*types.PerfCounterInfo, error) {
	return c.Perf.CounterInfoByKey(ctx)
}

func (c *Client) getMetricNameForID(id int32) string {
	c.metricNameMux.RLock()
	defer c.metricNameMux.RUnlock()
	return c.metricNameLookup[id]
}

func (c *Client) reloadMetricNameMap(ctx context.Context) error {
	c.metricNameMux.Lock()
	defer c.metricNameMux.Unlock()
	mn, err := c.CounterInfoByKey(ctx)
	if err != nil {
		return err
	}
	c.metricNameLookup = make(map[int32]string)
	for key, m := range mn {
		c.metricNameLookup[key] = m.Name()
	}
	return nil
}

func (c *Client) discover(ctx context.Context) error {
	c.busy.Lock()
	defer c.busy.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}

	err := c.reloadMetricNameMap(ctx)
	if err != nil {
		return err
	}

	// get the vSphere API version
	c.apiVersion = c.Client.ServiceContent.About.ApiVersion

	dcNameCache := make(map[string]string)

	numRes := int64(0)

	// Populate resource objects, and endpoint instance info.
	newObjects := make(map[string]objectMap)

	for k, res := range c.resourceKinds {
		// Need to do this for all resource types even if they are not enabled
		if res.enabled || k != "vm" {
			rf := ResourceFilter{
				finder:       &Finder{c},
				resType:      res.vcName,
				paths:        res.paths,
				excludePaths: res.excludePaths,
			}

			ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
			objects, err := res.getObjects(ctx1, c, &rf)
			cancel1()
			if err != nil {
				return err
			}

			// Fill in datacenter names where available (no need to do it for Datacenters)
			if res.name != "datacenter" {
				for k, obj := range objects {
					if obj.parentRef != nil {
						obj.dcname, _ = c.getDatacenterName(ctx, c, dcNameCache, *obj.parentRef)
						objects[k] = obj
					}
				}
			}

			// No need to collect metric metadata if resource type is not enabled
			if res.enabled {
				if res.simple {
					c.simpleMetadataSelect(ctx, res)
				} else {
					c.complexMetadataSelect(ctx, res, objects)
				}
			}
			newObjects[k] = objects

			numRes += int64(len(objects))
		}
	}

	// Build lun2ds map
	dss := newObjects["datastore"]
	l2d := make(map[string]string)
	for _, ds := range dss {
		lunID := ds.altID
		m := isolateLUN.FindStringSubmatch(lunID)
		if m != nil {
			l2d[m[1]] = ds.name
		}
	}

	// Atomically swap maps
	c.collectMux.Lock()
	defer c.collectMux.Unlock()

	for k, v := range newObjects {
		c.resourceKinds[k].objects = v
	}
	c.lun2ds = l2d

	return nil
}

// CounterInfoByName wraps performance.CounterInfoByName to give it proper timeouts.
func (c *Client) CounterInfoByName(ctx context.Context) (map[string]*types.PerfCounterInfo, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	return c.Perf.CounterInfoByName(ctx1)
}

// QueryMetrics wraps performance.Query to give it proper timeouts.
func (c *Client) QueryMetrics(ctx context.Context, pqs []types.PerfQuerySpec) ([]performance.EntityMetric, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	metrics, err := c.Perf.Query(ctx1, pqs)
	if err != nil {
		return nil, err
	}

	ctx2, cancel2 := context.WithTimeout(ctx, c.Timeout)
	defer cancel2()
	return c.Perf.ToMetricSeries(ctx2, metrics)
}

func (c *Client) populateTags(objectRef *objectRef, resource *resourceKind, t map[string]string, v performance.MetricSeries) {
	if t == nil {
		return
	}

	// Map name of object.
	if resource.pKey != "" {
		t[resource.pKey] = objectRef.name
		delete(t, "source")
	}

	// Map parent reference
	currentRef := objectRef
	currentResource := resource
	for {
		if currentRef == nil || currentResource == nil {
			break
		}

		parent, found := c.getParent(currentRef, currentResource)
		if found && currentResource.parentTag != "" {
			t[currentResource.parentTag] = parent.name
			currentResource = c.resourceKinds[currentResource.parent]
			currentRef = parent
		} else {
			break
		}
	}

	// Fill in Datacenter name
	if objectRef.dcname != "" {
		t["dcname"] = objectRef.dcname
	}

	// Determine which point tag to map to the instance
	if v.Instance != "" {
		t["instance"] = v.Instance
	}
}

func (c *Client) getParent(obj *objectRef, res *resourceKind) (*objectRef, bool) {
	if pKind, ok := c.resourceKinds[res.parent]; ok {
		if p, ok := pKind.objects[obj.parentRef.Value]; ok {
			return p, true
		}
	}
	return nil, false
}

func (c *Client) alignSamples(info []types.PerfSampleInfo, values []int64, interval time.Duration) ([]types.PerfSampleInfo, []float64) {
	rInfo := make([]types.PerfSampleInfo, 0, len(info))
	rValues := make([]float64, 0, len(values))
	bi := 1.0
	var lastBucket time.Time
	for idx := range info {
		// According to the docs, SampleInfo and Value should have the same length, but we've seen corrupted
		// data coming back with missing values. Take care of that gracefully!
		if idx >= len(values) {
			l.Debugf("len(SampleInfo)>len(Value) %d > %d during alignment", len(info), len(values))
			break
		}
		v := float64(values[idx])
		if v < 0 {
			continue
		}
		ts := info[idx].Timestamp
		roundedTS := ts.Truncate(interval)

		// Are we still working on the same bucket?
		if roundedTS == lastBucket {
			bi++
			p := len(rValues) - 1
			rValues[p] = ((bi-1)/bi)*rValues[p] + v/bi
		} else {
			rValues = append(rValues, v)
			roundedInfo := types.PerfSampleInfo{
				Timestamp: roundedTS,
				Interval:  info[idx].Interval,
			}
			rInfo = append(rInfo, roundedInfo)
			bi = 1.0
			lastBucket = roundedTS
		}
	}
	return rInfo, rValues
}

func (c *Client) simpleMetadataSelect(ctx context.Context, res *resourceKind) {
	m, err := c.CounterInfoByName(ctx)
	if err != nil {
		l.Errorf("Getting metric metadata. Discovery will be incomplete. Error: %s", err.Error())
		return
	}
	res.metrics = make(performance.MetricList, 0, len(res.include))
	for _, s := range res.include {
		if pci, ok := m[s]; ok {
			cnt := types.PerfMetricId{
				CounterId: pci.Key,
			}
			if res.collectInstances {
				cnt.Instance = "*"
			} else {
				cnt.Instance = ""
			}
			res.metrics = append(res.metrics, cnt)
		} else {
			l.Warnf("Metric name %s is unknown. Will not be collected", s)
		}
	}
}

func (c *Client) complexMetadataSelect(ctx context.Context, res *resourceKind, objects objectMap) {
	// We're only going to get metadata from maxMetadataSamples resources. If we have
	// more resources than that, we pick maxMetadataSamples samples at random.
	sampledObjects := make([]*objectRef, 0, len(objects))
	for _, obj := range objects {
		sampledObjects = append(sampledObjects, obj)
	}
	if len(sampledObjects) > maxMetadataSamples {
		// Shuffle samples into the maxMetadataSamples positions
		for i := 0; i < maxMetadataSamples; i++ {
			j := int(rand.Int31n(int32(i + 1))) //nolint:gosec // G404: not security critical
			sampledObjects[i], sampledObjects[j] = sampledObjects[j], sampledObjects[i]
		}
		sampledObjects = sampledObjects[0:maxMetadataSamples]
	}

	for _, obj := range sampledObjects {
		metrics, err := c.getMetadata(ctx, obj, res.sampling)
		if err != nil {
			l.Errorf("Getting metric metadata. Discovery will be incomplete. Error: %s", err.Error())
		}
		mMap := make(map[string]types.PerfMetricId)
		for _, m := range metrics {
			if m.Instance != "" && res.collectInstances {
				m.Instance = "*"
			} else {
				m.Instance = ""
			}
			if res.filters.Match(c.getMetricNameForID(m.CounterId)) {
				mMap[strconv.Itoa(int(m.CounterId))+"|"+m.Instance] = m
			}
		}
		if len(mMap) > len(res.metrics) {
			res.metrics = make(performance.MetricList, len(mMap))
			i := 0
			for _, m := range mMap {
				res.metrics[i] = m
				i++
			}
		}
	}
}

// GetServerTime returns the time at the vCenter server.JJ.
func (c *Client) GetServerTime(ctx context.Context) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	t, err := methods.GetCurrentTime(ctx, c.Client)
	if err != nil {
		return time.Time{}, err
	}
	return *t, nil
}

func (c *Client) getDatacenterName(ctx context.Context, client *Client, cache map[string]string, r types.ManagedObjectReference) (string, bool) {
	return c.getAncestorName(ctx, client, "Datacenter", cache, r)
}

func (c *Client) getAncestorName(
	ctx context.Context,
	client *Client,
	resourceType string,
	cache map[string]string,
	r types.ManagedObjectReference,
) (string, bool) {
	path := make([]string, 0)
	returnVal := ""
	here := r
	done := false
	for !done {
		done = func() bool {
			if name, ok := cache[here.Reference().String()]; ok {
				// Populate cache for the entire chain of objects leading here.
				returnVal = name
				return true
			}
			path = append(path, here.Reference().String())
			o := object.NewCommon(client.Client.Client, r)
			var result mo.ManagedEntity
			ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
			defer cancel1()
			err := o.Properties(ctx1, here, []string{"parent", "name"}, &result)
			if err != nil {
				l.Warnf("Error while resolving parent. Assuming no parent exists. Error: %s", err.Error())
				return true
			}
			if result.Reference().Type == resourceType {
				// Populate cache for the entire chain of objects leading here.
				returnVal = result.Name
				return true
			}
			if result.Parent == nil {
				return true
			}
			here = result.Parent.Reference()
			return false
		}()
	}
	for _, s := range path {
		cache[s] = returnVal
	}
	return returnVal, returnVal != ""
}

func (c *Client) makeMetricIdentifier(prefix, metric string) (metricName string, fieldName string) {
	parts := strings.Split(metric, ".")
	if len(parts) == 1 {
		return prefix, parts[0]
	}
	return prefix, strings.Join(parts, "_")
}

func cleanGuestID(id string) string {
	return strings.TrimSuffix(id, "Guest")
}

func (c *Client) getMetadata(ctx context.Context, obj *objectRef, sampling int32) (performance.MetricList, error) {
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	metrics, err := c.Perf.AvailableMetric(ctx1, obj.ref.Reference(), sampling)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (ipt *Input) createVSphereClient(vSphereURL string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	u, err := url.Parse(vSphereURL)
	if err != nil {
		return nil, fmt.Errorf("fail to parse URL: %w", err)
	}
	u.User = url.UserPassword(ipt.Username, ipt.Password)

	var tlsConfig *tls.Config
	if ipt.TLSClientConfig != nil {
		if conf, err := ipt.TLSClientConfig.TLSConfig(); err != nil {
			return nil, fmt.Errorf("failed to get TLS config: %w", err)
		} else {
			tlsConfig = conf
		}
	}

	if tlsConfig == nil {
		tlsConfig = &tls.Config{} //nolint
	}

	soapClient := soap.NewClient(u, tlsConfig.InsecureSkipVerify)
	if len(tlsConfig.Certificates) > 0 {
		soapClient.SetCertificate(tlsConfig.Certificates[0])
	}

	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create vim25 client: %w", err)
	}

	sm := session.NewManager(vimClient)

	govmomiClient := &govmomi.Client{
		Client:         vimClient,
		SessionManager: sm,
	}

	if u.User != nil {
		if err := govmomiClient.Login(ctx, u.User); err != nil {
			return nil, fmt.Errorf("failed to login: %w", err)
		}
	}

	perf := performance.NewManager(govmomiClient.Client)

	client := &Client{
		Client: govmomiClient,
		Perf:   perf,
	}

	return client, nil
}

func getDatacenters(ctx context.Context, c *Client, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.Datacenter
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap, len(resources))
	for i := range resources {
		r := &resources[i]

		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:      r.Name,
			ref:       r.ExtensibleManagedObject.Reference(),
			parentRef: r.Parent,
			dcname:    r.Name,
		}
	}
	return m, nil
}

func getDatastores(ctx context.Context, c *Client, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.Datastore
	ctx1, cancel1 := context.WithTimeout(ctx, c.Timeout)
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for i := range resources {
		r := &resources[i]

		lunID := ""
		if r.Info != nil {
			info := r.Info.GetDatastoreInfo()
			if info != nil {
				lunID = info.Url
			}
		}
		tags, fields := getDatastoreTagsAndFields(r)
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:         r.Name,
			ref:          r.ExtensibleManagedObject.Reference(),
			parentRef:    r.Parent,
			altID:        lunID,
			objectTags:   tags,
			objectFields: fields,
		}
	}
	return m, nil
}

func getClusters(ctx context.Context, client *Client, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	ctx1, cancel1 := context.WithTimeout(ctx, client.Timeout)
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]*types.ManagedObjectReference)
	m := make(objectMap, len(resources))
	for i := range resources {
		r := &resources[i]

		// Wrap in a function to make defer work correctly.
		err := func() error {
			// We're not interested in the immediate parent (a folder), but the data center.
			p, ok := cache[r.Parent.Value]
			if !ok {
				o := object.NewFolder(client.Client.Client, *r.Parent)
				var folder mo.Folder
				ctx3, cancel3 := context.WithTimeout(ctx, client.Timeout)
				defer cancel3()
				err = o.Properties(ctx3, *r.Parent, []string{"parent"}, &folder)
				if err != nil {
					l.Warnf("Error while getting folder parent: %s", err.Error())
					p = nil
				} else {
					pp := folder.Parent.Reference()
					p = &pp
					cache[r.Parent.Value] = p
				}
			}
			tags, fields := getClusterTagsAndFields(r)
			m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
				name:         r.Name,
				ref:          r.ExtensibleManagedObject.Reference(),
				parentRef:    p,
				objectTags:   tags,
				objectFields: fields,
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// noinspection GoUnusedParameterJJ.
func getResourcePools(ctx context.Context, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.ResourcePool
	err := resourceFilter.FindAll(ctx, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for i := range resources {
		r := &resources[i]

		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:      r.Name,
			ref:       r.ExtensibleManagedObject.Reference(),
			parentRef: r.Parent,
		}
	}
	return m, nil
}

func getResourcePoolName(rp types.ManagedObjectReference, rps objectMap) string {
	// Loop through the Resource Pools objectmap to find the corresponding one
	for _, r := range rps {
		if r.ref == rp {
			return r.name
		}
	}
	return "Resources" // Default value
}

// noinspection GoUnusedParameter.
func getHosts(ctx context.Context, c *Client, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.HostSystem
	err := resourceFilter.FindAll(ctx, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for i := range resources {
		r := &resources[i]
		tags, fields := getHostTagsAndFields(r)

		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:         r.Name,
			ref:          r.ExtensibleManagedObject.Reference(),
			parentRef:    r.Parent,
			objectTags:   tags,
			objectFields: fields,
		}
	}
	return m, nil
}

func getVMs(ctx context.Context, client *Client, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.VirtualMachine
	ctx1, cancel1 := context.WithTimeout(ctx, client.Timeout)
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	// Create a ResourcePool Filter and get the list of Resource Pools
	rprf := ResourceFilter{
		finder:       &Finder{client},
		resType:      "ResourcePool",
		paths:        []string{"/*/host/**"},
		excludePaths: nil,
	}
	resourcePools, err := getResourcePools(ctx, &rprf)
	if err != nil {
		return nil, err
	}
	for i := range resources {
		r := &resources[i]

		if r.Runtime.PowerState != "poweredOn" {
			continue
		}
		guest := "unknown"
		uuid := ""
		lookup := make(map[string]string)
		// Get the name of the VM resource pool
		rpname := getResourcePoolName(*r.ResourcePool, resourcePools)

		// Extract host name
		if r.Guest != nil && r.Guest.HostName != "" {
			lookup["guesthostname"] = r.Guest.HostName
		}

		// Collect network information
		for _, net := range r.Guest.Net {
			if net.DeviceConfigId == -1 {
				continue
			}
			if net.IpConfig == nil || net.IpConfig.IpAddress == nil {
				continue
			}
			ips := make(map[string][]string)
			for _, ip := range net.IpConfig.IpAddress {
				addr := ip.IpAddress
				for _, ipType := range []string{"ipv6", "ipv4"} {
					if !(ipType == "ipv4" && isIPv4.MatchString(addr) ||
						ipType == "ipv6" && isIPv6.MatchString(addr)) {
						continue
					}

					// By convention, we want the preferred addresses to appear first in the array.
					if _, ok := ips[ipType]; !ok {
						ips[ipType] = make([]string, 0)
					}
					if ip.State == "preferred" {
						ips[ipType] = append([]string{addr}, ips[ipType]...)
					} else {
						ips[ipType] = append(ips[ipType], addr)
					}
				}
			}
			for ipType, ipList := range ips {
				lookup["nic/"+strconv.Itoa(int(net.DeviceConfigId))+"/"+ipType] = strings.Join(ipList, ",")
			}
		}

		// Sometimes Config is unknown and returns a nil pointer
		if r.Config != nil {
			guest = cleanGuestID(r.Config.GuestId)
			if r.Guest.GuestId != "" {
				guest = cleanGuestID(r.Guest.GuestId)
			}
			uuid = r.Config.Uuid
		}
		tags, fields := getVMTagsAndFields(r)
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:         r.Name,
			ref:          r.ExtensibleManagedObject.Reference(),
			parentRef:    r.Runtime.Host,
			guest:        guest,
			altID:        uuid,
			rpname:       rpname,
			lookup:       lookup,
			objectTags:   tags,
			objectFields: fields,
		}
	}
	return m, nil
}

func (ipt *Input) getClient() (*Client, error) {
	var client *Client
	var err error
	if ipt.TLSClientConfig != nil && ipt.InsecureSkipVerify {
		client, err = ipt.createVSphereClient2(ipt.vcenter.String())
	} else {
		client, err = ipt.createVSphereClient(ipt.vcenter.String())
	}

	if err != nil {
		return nil, err
	}

	ipt.setupResource(client)

	err = client.discover(context.Background())
	return client, err
}

func (ipt *Input) testClient(ctx context.Context) error {
	// Execute a dummy call against the server to make sure the client is
	// still functional. If not, try to log back in. If that doesn't work,
	// we give up.
	ctx1, cancel1 := context.WithTimeout(ctx, ipt.timeout)
	defer cancel1()
	if _, err := methods.GetCurrentTime(ctx1, ipt.client.Client); err != nil {
		l.Info("Client session seems to have time out. Reauthenticating!")
		ctx2, cancel2 := context.WithTimeout(ctx, ipt.timeout)
		defer cancel2()

		auth := url.UserPassword(ipt.Username, ipt.Password)

		if err := ipt.client.Client.SessionManager.Login(ctx2, auth); err != nil {
			return fmt.Errorf("renewing authentication failed: %w", err)
		}
	}

	return nil
}

func (ipt *Input) setupResource(client *Client) {
	client.resourceKinds = map[string]*resourceKind{
		"datacenter": {
			name:             "datacenter",
			vcName:           "Datacenter",
			pKey:             "dcname",
			parentTag:        "",
			enabled:          anythingEnabled(ipt.DatacenterMetricExclude),
			realTime:         false,
			sampling:         int32(ipt.HistoricalInterval.Duration.Seconds()),
			filters:          newFilterOrPanic(ipt.DatacenterMetricInclude, ipt.DatacenterMetricExclude),
			objects:          make(objectMap),
			paths:            ipt.DatacenterInclude,
			excludePaths:     ipt.DatacenterExclude,
			simple:           isSimple(ipt.DatacenterMetricInclude, ipt.DatacenterMetricExclude),
			collectInstances: ipt.DatacenterInstances,
			getObjects:       getDatacenters,
			parent:           "",
		},
		"cluster": {
			name:             "cluster",
			vcName:           "ClusterComputeResource",
			pKey:             "cluster_name",
			parentTag:        "dcname",
			enabled:          anythingEnabled(ipt.ClusterMetricExclude),
			realTime:         false,
			sampling:         int32((ipt.HistoricalInterval.Duration).Seconds()),
			objects:          make(objectMap),
			filters:          newFilterOrPanic(ipt.ClusterMetricInclude, ipt.ClusterMetricExclude),
			paths:            ipt.ClusterInclude,
			excludePaths:     ipt.ClusterExclude,
			simple:           isSimple(ipt.ClusterMetricInclude, ipt.ClusterMetricExclude),
			include:          ipt.ClusterMetricInclude,
			collectInstances: ipt.ClusterInstances,
			getObjects:       getClusters,
			parent:           "datacenter",
		},
		"host": {
			name:             "host",
			vcName:           "HostSystem",
			pKey:             "esx_hostname",
			parentTag:        "cluster_name",
			enabled:          anythingEnabled(ipt.HostMetricExclude),
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(ipt.HostMetricInclude, ipt.HostMetricExclude),
			paths:            ipt.HostInclude,
			excludePaths:     ipt.HostExclude,
			simple:           isSimple(ipt.HostMetricInclude, ipt.HostMetricExclude),
			include:          ipt.HostMetricInclude,
			collectInstances: ipt.HostInstances,
			getObjects:       getHosts,
			parent:           "cluster",
		},
		"vm": {
			name:             "vm",
			vcName:           "VirtualMachine",
			pKey:             "vm_name",
			parentTag:        "esx_hostname",
			enabled:          anythingEnabled(ipt.VMMetricExclude),
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(ipt.VMMetricInclude, ipt.VMMetricExclude),
			paths:            ipt.VMInclude,
			excludePaths:     ipt.VMExclude,
			simple:           isSimple(ipt.VMMetricInclude, ipt.VMMetricExclude),
			include:          ipt.VMMetricInclude,
			collectInstances: ipt.VMInstances,
			getObjects:       getVMs,
			parent:           "host",
		},
		"datastore": {
			name:             "datastore",
			vcName:           "Datastore",
			pKey:             "dsname",
			enabled:          anythingEnabled(ipt.DatastoreMetricExclude),
			realTime:         false,
			sampling:         int32(ipt.HistoricalInterval.Duration.Seconds()),
			objects:          make(objectMap),
			filters:          newFilterOrPanic(ipt.DatastoreMetricInclude, ipt.DatastoreMetricExclude),
			paths:            ipt.DatastoreInclude,
			excludePaths:     ipt.DatastoreExclude,
			simple:           isSimple(ipt.DatastoreMetricInclude, ipt.DatastoreMetricExclude),
			include:          ipt.DatastoreMetricInclude,
			collectInstances: ipt.DatastoreInstances,
			getObjects:       getDatastores,
			parent:           "",
		},
	}
}

func (ipt *Input) createVSphereClient2(vSphereURL string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	u, err := url.Parse(vSphereURL)
	if err != nil {
		return nil, fmt.Errorf("fail to parse URL: %w", err)
	}
	u.User = url.UserPassword(ipt.Username, ipt.Password)
	soapClient := soap.NewClient(u, true)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create vim25 client: %w", err)
	}

	sm := session.NewManager(vimClient)

	if u.User != nil {
		if err := sm.Login(ctx, u.User); err != nil {
			return nil, fmt.Errorf("failed to login: %w", err)
		}
	}

	govmomiClient := &govmomi.Client{
		Client:         vimClient,
		SessionManager: sm,
	}

	perf := performance.NewManager(govmomiClient.Client)

	client := &Client{
		Client:  govmomiClient,
		Perf:    perf,
		Timeout: ipt.timeout,
	}

	return client, nil
}
