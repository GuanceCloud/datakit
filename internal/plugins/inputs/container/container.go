// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"syscall"
	"time"

	gcpmonitoring "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cloudprovider/gcp/monitoring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type containerCollector struct {
	ipt       *Input
	runtime   runtime.ContainerRuntime
	k8sClient k8sclient.Client

	localNodeName        string
	maxConcurrent        int
	enableCollectLogging bool
	election             bool
	leaderOnly           bool

	enableExtractK8sLabelAsTagsV1 bool
	podLabelAsTagsForNonMetric    labelsOption
	podLabelAsTagsForMetric       labelsOption

	logFilter      filter.Filter
	logCoordinator *containerLogCoordinator
	loggingScanCh  <-chan struct{}

	extraTags map[string]string
	feeder    dkio.Feeder

	ptsTime time.Time
}

func newECSFargate(ipt *Input, agentURL string) (Collector, error) {
	r, err := runtime.NewECSFargateRuntime(agentURL)
	if err != nil {
		return nil, err
	}

	tags := inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")

	return &containerCollector{
		ipt:                  ipt,
		runtime:              r,
		maxConcurrent:        ipt.ContainerMaxConcurrent,
		enableCollectLogging: false,
		extraTags:            tags,
		feeder:               ipt.Feeder,
	}, nil
}

func newGCPCloudMonitoringCollector(ipt *Input, httpClient *http.Client, k8sClient k8sclient.Client) (Collector, error) {
	r, err := gcpmonitoring.NewRuntime(gcpmonitoring.Config{
		ProjectID:   ipt.GCPProjectID,
		ClusterName: ipt.GCPClusterName,
		Location:    ipt.GCPClusterLocation,
	}, httpClient)
	if err != nil {
		return nil, err
	}
	if err := validateRuntimeUniqueness(r); err != nil {
		closeContainerRuntime(r)
		return nil, err
	}

	labelOptions := buildLabelOptions(ipt)
	tags := inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, "")
	tags["cluster_name_k8s"] = ipt.GCPClusterName
	tags["gcp_project_id"] = ipt.GCPProjectID
	tags["gcp_location"] = ipt.GCPClusterLocation
	return &containerCollector{
		ipt:           ipt,
		runtime:       r,
		k8sClient:     k8sClient,
		maxConcurrent: ipt.ContainerMaxConcurrent,

		enableCollectLogging:          false,
		enableExtractK8sLabelAsTagsV1: ipt.EnableExtractK8sLabelAsTags,
		podLabelAsTagsForNonMetric:    labelOptions.nonMetric,
		podLabelAsTagsForMetric:       labelOptions.metric,
		election:                      true,
		leaderOnly:                    true,

		extraTags: tags,
		feeder:    ipt.Feeder,
	}, nil
}

var existingRuntimes sync.Map

// nolint:lll
func newContainerCollector(ipt *Input, endpoint string, mountPoint string, k8sClient k8sclient.Client, logCoordinator *containerLogCoordinator) (Collector, error) {
	logFilter, err := createLogFilter(ipt)
	if err != nil {
		return nil, err
	}

	runtime, err := createContainerRuntime(endpoint, mountPoint)
	if err != nil {
		return nil, err
	}

	if err := validateRuntimeUniqueness(runtime); err != nil {
		closeContainerRuntime(runtime)
		return nil, err
	}

	labelOptions := buildLabelOptions(ipt)

	return &containerCollector{
		ipt:           ipt,
		runtime:       runtime,
		k8sClient:     k8sClient,
		localNodeName: datakit.DKHost,
		maxConcurrent: ipt.ContainerMaxConcurrent,

		enableCollectLogging:          true,
		enableExtractK8sLabelAsTagsV1: ipt.EnableExtractK8sLabelAsTags,
		podLabelAsTagsForNonMetric:    labelOptions.nonMetric,
		podLabelAsTagsForMetric:       labelOptions.metric,

		logFilter:      logFilter,
		logCoordinator: logCoordinator,
		loggingScanCh:  logCoordinator.registerLoggingScanSignal(),

		extraTags: inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, ""),
		feeder:    ipt.Feeder,
	}, nil
}

func closeContainerRuntime(rt runtime.ContainerRuntime) {
	if closer, ok := rt.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			l.Warnf("close duplicate container runtime failed: %s", err)
		}
	}
}

func createLogFilter(ipt *Input) (filter.Filter, error) {
	return filter.NewFilter(ipt.ContainerIncludeLog, ipt.ContainerExcludeLog)
}

func createContainerRuntime(endpoint, mountPoint string) (runtime.ContainerRuntime, error) {
	if verifyErr := runtime.VerifyDockerRuntime(endpoint); verifyErr == nil {
		return runtime.NewDockerRuntime(endpoint, mountPoint)
	}
	return runtime.NewCRIRuntime(endpoint, mountPoint)
}

func isRuntimeConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrNotExist) ||
		errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) {
		return true
	}
	for current := err; current != nil; current = errors.Unwrap(current) {
		if code := status.Code(current); code == codes.Unavailable || code == codes.DeadlineExceeded {
			return true
		}
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func validateRuntimeUniqueness(rt runtime.ContainerRuntime) error {
	version, err := rt.Version()
	if err != nil {
		return fmt.Errorf("get runtime version: %w", err)
	}

	l.Infof("runtime platform %s, api-version %s", version.PlatformName, version.APIVersion)

	key := fmt.Sprintf("%s:%s", version.PlatformName, version.APIVersion)
	if _, exist := existingRuntimes.LoadOrStore(key, struct{}{}); exist {
		return fmt.Errorf("runtime %s already exists", key)
	}
	return nil
}

type labelOptions struct {
	nonMetric labelsOption
	metric    labelsOption
}

func buildLabelOptions(ipt *Input) labelOptions {
	return labelOptions{
		nonMetric: buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2, config.Cfg.Dataway.GlobalCustomerKeys),
		metric:    buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2ForMetric, config.Cfg.Dataway.GlobalCustomerKeys),
	}
}

func (c *containerCollector) StartCollect() {
	if c.leaderOnly {
		c.startLeaderOnlyCollect()
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "container-collector"})

	if c.ipt.EnableContainerMetric {
		g.Go(func(_ context.Context) error {
			c.runMetricCollector()
			return nil
		})
	}

	g.Go(func(_ context.Context) error {
		c.runObjectCollector()
		return nil
	})

	if c.enableCollectLogging {
		g.Go(func(_ context.Context) error {
			c.runLoggingDiscovery()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		l.Warnf("container collector stopped with error: %s", err)
	}
	l.Info("container collector stopped")
}

func (c *containerCollector) runMetricCollector() {
	metricTicker := time.NewTicker(c.ipt.MetricCollecInterval)
	defer metricTicker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case tt := <-metricTicker.C:
			c.ptsTime = inputs.AlignTime(tt, c.ptsTime, c.ipt.MetricCollecInterval)
			c.gatherMetric()
		}
	}
}

func (c *containerCollector) runObjectCollector() {
	objectTicker := time.NewTicker(c.ipt.ObjectCollecInterval)
	defer objectTicker.Stop()

	initialTimer := time.NewTimer(3 * time.Second)
	defer initialTimer.Stop()

	select {
	case <-datakit.Exit.Wait():
		return
	case <-initialTimer.C:
		c.gatherObject()
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-objectTicker.C:
			c.gatherObject()
		}
	}
}

func (c *containerCollector) startLeaderOnlyCollect() {
	activationTicker := time.NewTicker(5 * time.Second)
	metricTicker := time.NewTicker(c.ipt.MetricCollecInterval)
	objectTicker := time.NewTicker(c.ipt.ObjectCollecInterval)
	defer activationTicker.Stop()
	defer metricTicker.Stop()
	defer objectTicker.Stop()

	active := false
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("cloud monitoring collector stopped")
			return

		case now := <-activationTicker.C:
			if !c.ipt.leaderGate().Allowed() {
				active = false
				continue
			}
			if !active {
				active = true
				if c.ipt.EnableContainerMetric {
					c.ptsTime = inputs.AlignTime(now, c.ptsTime, c.ipt.MetricCollecInterval)
					c.gatherMetric()
				}
				c.gatherObject()
			}

		case now := <-metricTicker.C:
			if !c.ipt.leaderGate().Allowed() || !c.ipt.EnableContainerMetric {
				continue
			}
			c.ptsTime = inputs.AlignTime(now, c.ptsTime, c.ipt.MetricCollecInterval)
			c.gatherMetric()

		case <-objectTicker.C:
			if !c.ipt.leaderGate().Allowed() {
				continue
			}
			c.gatherObject()
		}
	}
}

func (c *containerCollector) runLoggingDiscovery() {
	loggingTicker := time.NewTicker(c.ipt.LoggingSearchInterval)
	defer loggingTicker.Stop()

	c.gatherLogging("initial")

	var lastPodEventScan time.Time
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case scheduledAt := <-loggingTicker.C:
			loggingDiscoveryScheduleDelayVec.Observe(time.Since(scheduledAt).Seconds())
			c.gatherLogging("ticker")
		case <-c.loggingScanCh:
			if !lastPodEventScan.IsZero() {
				if wait := time.Second - time.Since(lastPodEventScan); wait > 0 {
					timer := time.NewTimer(wait)
					select {
					case <-datakit.Exit.Wait():
						timer.Stop()
						return
					case <-timer.C:
					}
				}
			}
			c.gatherLogging("pod-event")
			lastPodEventScan = time.Now()
		}
	}
}

func (c *containerCollector) ReloadConfigKV(_ map[string]string) error {
	l.Info("reloading container config")
	return nil
}

func checkEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint %s, err: %w", endpoint, err)
	}

	switch u.Scheme {
	case "unix":
		// nil
	default:
		return fmt.Errorf("using %s as endpoint is not supported protocol", endpoint)
	}

	info, err := os.Stat(u.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("endpoint %s does not exist, maybe it is not running: %w", endpoint, err)
	}
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("endpoint %s cannot be a directory", u.Path)
	}

	return nil
}

type labelsOption struct {
	all  bool
	keys []string
}

func buildLabelsOption(asTagKeys, customerKeys []string) labelsOption {
	// e.g. [""] (all)
	if len(asTagKeys) == 1 && asTagKeys[0] == "" {
		return labelsOption{all: true}
	}
	keys := unique(append(asTagKeys, customerKeys...))
	sort.Strings(keys)
	return labelsOption{keys: keys}
}

func getMountPoint() string {
	if !datakit.Docker {
		return ""
	}
	if n := os.Getenv("HOST_ROOT"); n != "" {
		return n
	}
	return "/rootfs"
}

func getClusterNameK8s() string {
	return os.Getenv("ENV_CLUSTER_NAME_K8S")
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
