// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package hostobject collect host object.
package hostobject

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	dto "github.com/prometheus/client_model/go"
	diskutil "github.com/shirou/gopsutil/disk"
	netutil "github.com/shirou/gopsutil/net"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName              = "hostobject"
	maxInterval            = 30 * time.Minute
	minInterval            = 10 * time.Second
	hostObjMeasurementName = "HOST"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

type (
	NetIOCounters  func(bool) ([]netutil.IOCountersStat, error)
	DiskIOCounters func(names ...string) (map[string]diskutil.IOCountersStat, error)
)

type Input struct {
	Name  string `toml:"name,omitempty"`        // deprecated
	Class string `toml:"class,omitempty"`       // deprecated
	Desc  string `toml:"description,omitempty"` // deprecated

	PipelineDeprecated string `toml:"pipeline,omitempty"`

	Tags                              map[string]string `toml:"tags,omitempty"`
	EnableCloudHostTagsGlobalElection bool              `toml:"enable_cloud_host_tags_global_election"`
	EnableCloudHostTagsGlobalHost     bool              `toml:"enable_cloud_host_tags_global_host"`

	Interval                 time.Duration `toml:"interval,omitempty"`
	IgnoreInputsErrorsBefore time.Duration `toml:"ignore_inputs_errors_before,omitempty"`
	DeprecatedIOTimeout      time.Duration `toml:"io_timeout,omitempty"`

	EnableNetVirtualInterfaces bool     `toml:"enable_net_virtual_interfaces"`
	IgnoreZeroBytesDisk        bool     `toml:"ignore_zero_bytes_disk"`
	OnlyPhysicalDevice         bool     `toml:"only_physical_device"`
	ExtraDevice                []string `toml:"extra_device"`
	ExcludeDevice              []string `toml:"exclude_device"`

	DisableCloudProviderSync bool              `toml:"disable_cloud_provider_sync"`
	CloudInfo                map[string]string `toml:"cloud_info,omitempty"`
	lastSync                 time.Time

	netIOCounters  NetIOCounters
	diskIOCounters DiskIOCounters
	lastDiskIOInfo diskIOInfo
	lastNetIOInfo  netIOInfo
	cloudHostTags  map[string]string

	collectCache []*point.Point
	semStop      *cliutils.Sem // start stop signal
	feeder       dkio.Feeder
	mergedTags   map[string]string
	tagger       datakit.GlobalTagger

	mfs []*dto.MetricFamily
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		l.Debugf("start collecting...")
		start := time.Now()
		if err := ipt.collect(); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorCategory(point.Object),
			)
		} else {
			if err := ipt.feeder.FeedV2(point.Object, ipt.collectCache,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithElection(false),
				dkio.WithInputName(inputName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s exit on sem", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s return on sem", inputName)
			return
		case <-tick.C:
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	ts := time.Now()
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(ts))

	var kvs point.KVs

	if mfs, err := metrics.Gather(); err == nil {
		ipt.mfs = mfs
	}

	message, err := ipt.getHostObjectMessage()
	if err != nil {
		return err
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		l.Errorf("json marshal err:%s", err.Error())
		return err
	}

	l.Debugf("messageData len: %d", len(messageData))

	kvs = kvs.Add("message", string(messageData), false, true)
	kvs = kvs.Add("start_time", message.Host.HostMeta.BootTime*1000, false, true)
	kvs = kvs.Add("datakit_ver", datakit.Version, false, true)
	kvs = kvs.Add("cpu_usage", message.Host.cpuPercent, false, true)
	kvs = kvs.Add("mem_used_percent", message.Host.Mem.usedPercent, false, true)
	kvs = kvs.Add("load", message.Host.load5, false, true)
	kvs = kvs.Add("disk_used_percent", message.Host.diskUsedPercent, false, true)
	kvs = kvs.Add("diskio_read_bytes_per_sec", message.Host.diskIOReadBytesPerSec, false, true)
	kvs = kvs.Add("diskio_write_bytes_per_sec", message.Host.diskIOWriteBytesPerSec, false, true)
	kvs = kvs.Add("net_recv_bytes_per_sec", message.Host.netRecvBytesPerSec, false, true)
	kvs = kvs.Add("net_send_bytes_per_sec", message.Host.netSendBytesPerSec, false, true)
	kvs = kvs.Add("logging_level", message.Host.loggingLevel, false, true)
	kvs = kvs.Add("name", message.Host.HostMeta.HostName, true, true)
	kvs = kvs.Add("os", message.Host.HostMeta.OS, true, true)

	if !datakit.IsTestMode {
		kvs = kvs.Add("Scheck", message.Collectors[0].Version, false, true)
	}

	isDocker := 0
	if datakit.Docker {
		isDocker = 1
	}
	kvs = kvs.Add("is_docker", isDocker, false, true)

	// check if dk upgrader is available
	// TODO: check response message whether is valid
	if res, err := http.Get(fmt.Sprintf("http://%s",
		net.JoinHostPort(config.Cfg.DKUpgrader.Host, fmt.Sprintf("%d", config.Cfg.DKUpgrader.Port)))); err != nil {
		l.Warnf("get dk upgrader failed: %s", err.Error())
	} else {
		_ = res.Body.Close()
		kvs = kvs.Add("dk_upgrader", fmt.Sprintf("%s:%d", config.Cfg.DKUpgrader.Host, config.Cfg.DKUpgrader.Port), false, true)
	}

	// append extra cloud fields: all of them as tags
	for k, v := range message.Host.cloudInfo {
		switch tv := v.(type) {
		case string:
			if tv != Unavailable {
				kvs = kvs.Add(k, tv, true, true)
			}
		default:
			l.Warnf("ignore non-string cloud extra field: %s: %v, ignored", k, v)
		}
	}

	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	ipt.collectCache = append(ipt.collectCache, point.NewPointV2(hostObjMeasurementName, kvs, opts...))

	needUpdateGlobalTags := false
	if field := kvs.Get("region"); field != nil {
		if s := field.GetS(); s != "" && s != ipt.cloudHostTags["region"] {
			ipt.cloudHostTags["region"] = s
			needUpdateGlobalTags = true
		}
	}
	if field := kvs.Get("zone_id"); field != nil {
		if s := field.GetS(); s != "" && s != ipt.cloudHostTags["zone_id"] {
			ipt.cloudHostTags["zone_id"] = s
			needUpdateGlobalTags = true
		}
	}
	if needUpdateGlobalTags {
		if ipt.EnableCloudHostTagsGlobalHost {
			httpapi.UpdateHostTags(ipt.cloudHostTags, "cloud_host_meta")
		}
		if ipt.EnableCloudHostTagsGlobalElection {
			httpapi.UpdateElectionTags(ipt.cloudHostTags, "cloud_host_meta")
		}
	}

	return nil
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return "host" }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOS }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "EnableNetVirtualInterfaces", ENVName: "INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES", ConfField: "enable_net_virtual_interfaces", Type: doc.Boolean, Default: `false`, Desc: "Enable collect network virtual interfaces", DescZh: "允许采集虚拟网卡"},
		{FieldName: "IgnoreZeroBytesDisk", ENVName: "INPUT_HOSTOBJECT_IGNORE_ZERO_BYTES_DISK", ConfField: "ignore_zero_bytes_disk", Type: doc.Boolean, Default: `false`, Desc: "Ignore the disk which space is zero", DescZh: "忽略大小为 0 的磁盘"},
		{FieldName: "OnlyPhysicalDevice", ENVName: "INPUT_HOSTOBJECT_ONLY_PHYSICAL_DEVICE", ConfField: "only_physical_device", Type: doc.Boolean, Default: `false`, Desc: "Physical devices only, any string", DescZh: "忽略非物理磁盘（如网盘、NFS），任意非空字符串"},
		{FieldName: "ExcludeDevice", ENVName: "INPUT_HOSTOBJECT_EXCLUDE_DEVICE", ConfField: "exclude_device", Type: doc.List, Example: `/dev/loop0,/dev/loop1`, Desc: "Exclude some with dev prefix", DescZh: "忽略的 device"},
		{FieldName: "ExtraDevice", ENVName: "INPUT_HOSTOBJECT_EXTRA_DEVICE", ConfField: "extra_device", Type: doc.List, Example: "`/nfsdata,other`", Desc: "Additional device", DescZh: "额外增加的 device"},
		{FieldName: "EnableCloudHostTagsGlobalElection", ENVName: "ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_ELECTION_TAGS", ConfField: "enable_cloud_host_tags_global_election", Type: doc.Boolean, Default: "true", Desc: "Enable put cloud provider region/zone_id information into global election tags", DescZh: "将云服务商 region/zone_id 信息放入全局选举标签"},
		{FieldName: "EnableCloudHostTagsGlobalHost", ENVName: "ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_HOST_TAGS", ConfField: "enable_cloud_host_tags_global_host", Type: doc.Boolean, Default: "true", Desc: "Enable put cloud provider region/zone_id information into global host tags", DescZh: "将云服务商 region/zone_id 信息放入全局主机标签"},
		{FieldName: "Tags", ENVName: "INPUT_HOSTOBJECT_TAGS", ConfField: "tags"},
		{FieldName: "ENVCloud", ENVName: "CLOUD_PROVIDER", ConfField: "none", Type: doc.String, Example: "`aliyun/aws/tencent/hwcloud/azure`", Desc: "Designate cloud service provider", DescZh: "指定云服务商"},
	}

	return doc.SetENVDoc("ENV_", infos)
}

// ReadEnv used to read ENVs while running under DaemonSet.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES to bool: %s, ignore", err)
		} else {
			ipt.EnableNetVirtualInterfaces = b
		}
	}

	if _, ok := envs["ENV_INPUT_HOSTOBJECT_ONLY_PHYSICAL_DEVICE"]; ok {
		l.Info("setup OnlyPhysicalDevice...")
		ipt.OnlyPhysicalDevice = true
	}
	if fsList, ok := envs["ENV_INPUT_HOSTOBJECT_EXTRA_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add extra_device from ENV: %v", fsList)
		ipt.ExtraDevice = append(ipt.ExtraDevice, list...)
	}
	if fsList, ok := envs["ENV_INPUT_HOSTOBJECT_EXCLUDE_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add exlude_device from ENV: %v", fsList)
		ipt.ExcludeDevice = append(ipt.ExcludeDevice, list...)
	}
	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/505
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_ZERO_BYTES_DISK"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_ZERO_BYTES_DISK to bool: %s, ignore", err)
		} else {
			ipt.IgnoreZeroBytesDisk = b
		}
	}
	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/2076
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_ELECTION_TAGS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_ELECTION_TAGS to bool: %s, ignore", err)
		} else {
			ipt.EnableCloudHostTagsGlobalElection = b
		}
	}
	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/2136
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_HOST_TAGS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_CLOUD_META_AS_HOST_TAGS to bool: %s, ignore", err)
		} else {
			ipt.EnableCloudHostTagsGlobalHost = b
		}
	}
	if tagsStr, ok := envs["ENV_INPUT_HOSTOBJECT_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	// ENV_CLOUD_PROVIDER 会覆盖 ENV_INPUT_HOSTOBJECT_TAGS 中填入的 cloud_provider
	if tagsStr, ok := envs["ENV_CLOUD_PROVIDER"]; ok {
		cloudProvider := dkstring.TrimString(tagsStr)
		cloudProvider = strings.ToLower(cloudProvider)
		switch cloudProvider {
		case "aliyun", "tencent", "aws", "hwcloud", "azure":
			ipt.Tags["cloud_provider"] = cloudProvider
		}
	} // ENV_CLOUD_PROVIDER
}

func defaultInput() *Input {
	return &Input{
		Interval:                          5 * time.Minute,
		IgnoreInputsErrorsBefore:          30 * time.Second,
		IgnoreZeroBytesDisk:               true,
		EnableCloudHostTagsGlobalElection: true,
		EnableCloudHostTagsGlobalHost:     true,
		diskIOCounters:                    diskutil.IOCounters,
		netIOCounters:                     netutil.IOCounters,

		semStop:       cliutils.NewSem(),
		feeder:        dkio.DefaultFeeder(),
		Tags:          make(map[string]string),
		tagger:        datakit.DefaultGlobalTagger(),
		cloudHostTags: make(map[string]string),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

func SetLog() {
	l = logger.SLogger(inputName)
}
