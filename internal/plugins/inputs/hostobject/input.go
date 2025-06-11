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
	"regexp"
	"runtime"
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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkmetrics "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/pcommon"
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

	Tags                                        map[string]string `toml:"tags,omitempty"`
	EnableCloudHostTagsGlobalElection           bool              `toml:"enable_cloud_host_tags_as_global_election_tags"`
	EnableCloudHostTagsGlobalElectionDeprecated bool              `toml:"enable_cloud_host_tags_global_election"` // deprecated
	EnableCloudHostTagsGlobalHost               bool              `toml:"enable_cloud_host_tags_as_global_host_tags"`
	EnableCloudHostTagsGlobalHostDeprecated     bool              `toml:"enable_cloud_host_tags_global_host"` // deprecated

	Interval                 time.Duration `toml:"interval,omitempty"`
	IgnoreInputsErrorsBefore time.Duration `toml:"ignore_inputs_errors_before,omitempty"`
	DeprecatedIOTimeout      time.Duration `toml:"io_timeout,omitempty"`

	EnableNetVirtualInterfaces bool     `toml:"enable_net_virtual_interfaces"`
	IgnoreZeroBytesDisk        bool     `toml:"ignore_zero_bytes_disk"`
	ExtraDevice                []string `toml:"extra_device"`
	ExcludeDevice              []string `toml:"exclude_device"`

	IgnoreFSTypes    string `toml:"ignore_fstypes"`
	regIgnoreFSTypes *regexp.Regexp

	IgnoreMountpoints    string `toml:"ignore_mountpoints"`
	regIgnoreMountpoints *regexp.Regexp

	diskStats pcommon.DiskStats

	ConfigPath []string `toml:"config_path"`

	DisableCloudProviderSync bool              `toml:"disable_cloud_provider_sync"`
	EnableCloudAWSIMDSv2     bool              `toml:"enable_cloud_aws_imds_v2"`
	EnableCloudAWSIPv6       bool              `toml:"enable_cloud_aws_ipv6"`
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

	mfs               []*dto.MetricFamily
	hostRoot          string
	CloudMetaURL      map[string]string `toml:"cloud_meta_url,omitempty"`
	CloudMetaTokenURL map[string]string `toml:"cloud_meta_token_url,omitempty"`
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	start := ntp.Now()
	pcommon.SetLog()

	for {
		collectStart := time.Now()
		l.Debugf("start collecting...")
		if err := ipt.collect(start.UnixNano()); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				dkmetrics.WithLastErrorInput(inputName),
				dkmetrics.WithLastErrorCategory(point.Object),
			)
		} else {
			if err := ipt.feeder.FeedV2(point.Object, ipt.collectCache,
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithElection(false),
				dkio.WithInputName(inputName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkmetrics.WithLastErrorInput(inputName),
					dkmetrics.WithLastErrorCategory(point.Metric),
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
		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval)
		}
	}
}

func (ipt *Input) setup() {
	SetLog()
	ipt.EnableCloudHostTagsGlobalElection = ipt.EnableCloudHostTagsGlobalElection && ipt.EnableCloudHostTagsGlobalElectionDeprecated
	ipt.EnableCloudHostTagsGlobalHost = ipt.EnableCloudHostTagsGlobalHost && ipt.EnableCloudHostTagsGlobalHostDeprecated

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)

	if ipt.IgnoreFSTypes != "" {
		if re, err := regexp.Compile(ipt.IgnoreFSTypes); err != nil {
			l.Warnf("regexp.Compile(%q): %s, ignored", ipt.IgnoreFSTypes, err.Error())
		} else {
			ipt.regIgnoreFSTypes = re
		}
	}

	if ipt.IgnoreMountpoints != "" {
		if re, err := regexp.Compile(ipt.IgnoreMountpoints); err != nil {
			l.Warnf("regexp.Compile(%q): %s, ignored", ipt.IgnoreMountpoints, err.Error())
		} else {
			ipt.regIgnoreMountpoints = re
		}
	}
}

func (ipt *Input) collect(ptTS int64) error {
	ipt.collectCache = make([]*point.Point, 0)
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTimestamp(ptTS))

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

	kvs = kvs.Add("message", string(messageData), false, true).
		Add("start_time", message.Host.HostMeta.BootTime*1000, false, true).
		Add("datakit_ver", datakit.Version, false, true).
		Add("cpu_usage", message.Host.cpuPercent, false, true).
		Add("mem_used_percent", message.Host.Mem.usedPercent, false, true).
		Add("load", message.Host.load5, false, true).
		Add("disk_used_percent", message.Host.diskUsedPercent, false, true).
		Add("diskio_read_bytes_per_sec", message.Host.diskIOReadBytesPerSec, false, true).
		Add("diskio_write_bytes_per_sec", message.Host.diskIOWriteBytesPerSec, false, true).
		Add("net_recv_bytes_per_sec", message.Host.netRecvBytesPerSec, false, true).
		Add("net_send_bytes_per_sec", message.Host.netSendBytesPerSec, false, true).
		Add("logging_level", message.Host.loggingLevel, false, true).
		AddTag("name", message.Host.HostMeta.HostName).
		AddTag("os", message.Host.HostMeta.OS).
		Add("num_cpu", runtime.NumCPU(), false, false).
		AddTag("unicast_ip", message.Config.IP).
		Add("disk_total", message.Host.getDiskTotal(), false, true).
		AddTag("arch", message.Host.HostMeta.Arch)

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

func defaultInput() *Input {
	return &Input{
		Interval:                                    5 * time.Minute,
		IgnoreInputsErrorsBefore:                    30 * time.Second,
		EnableCloudHostTagsGlobalElection:           true,
		EnableCloudHostTagsGlobalElectionDeprecated: true,
		EnableCloudHostTagsGlobalHost:               true,
		EnableCloudHostTagsGlobalHostDeprecated:     true,
		IgnoreZeroBytesDisk:                         true,
		diskIOCounters:                              diskutil.IOCounters,
		netIOCounters:                               netutil.IOCounters,

		diskStats:     &pcommon.DiskStatsImpl{},
		IgnoreFSTypes: `^(tmpfs|autofs|binfmt_misc|devpts|fuse.lxcfs|overlay|proc|squashfs|sysfs)$`,
		// default ignore config map and loggging collector related mount points
		IgnoreMountpoints: `^(/usr/local/datakit/.*|/run/containerd/.*)$`,

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
