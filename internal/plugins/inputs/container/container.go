// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type containerCollector struct {
	ipt       *Input
	runtime   runtime.ContainerRuntime
	k8sClient k8sclient.Client

	localNodeName                 string
	maxConcurrent                 int
	enableCollectLogging          bool
	enableExtractK8sLabelAsTagsV1 bool
	podLabelAsTagsForNonMetric    labelsOption
	podLabelAsTagsForMetric       labelsOption

	logFilter filter.Filter
	logTable  *logTable
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
		enableCollectLogging: false,
		extraTags:            tags,
		feeder:               ipt.Feeder,
	}, nil
}

var containerExistList sync.Map

func newContainerCollector(ipt *Input, endpoint string, mountPoint string, k8sClient k8sclient.Client) (Collector, error) {
	logFilter, err := filter.NewFilter(ipt.ContainerIncludeLog, ipt.ContainerExcludeLog)
	if err != nil {
		return nil, err
	}

	var r runtime.ContainerRuntime

	// docker not supported CRI
	if verifyErr := runtime.VerifyDockerRuntime(endpoint); verifyErr == nil {
		r, err = runtime.NewDockerRuntime(endpoint, mountPoint)
	} else {
		r, err = runtime.NewCRIRuntime(endpoint, mountPoint)
	}

	if err != nil {
		return nil, err
	}

	version, err := r.Version()
	if err != nil {
		return nil, fmt.Errorf("get runtime version err: %w", err)
	}
	l.Infof("runtime platform %s, api-version %s", version.PlatformName, version.APIVersion)

	key := fmt.Sprintf("%s:%s", version.PlatformName, version.APIVersion)
	if _, exist := containerExistList.Load(key); exist {
		return nil, fmt.Errorf("runtime %s already exists", key)
	} else {
		containerExistList.Store(key, nil)
	}

	tags := inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")

	optForNonMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2, config.Cfg.Dataway.GlobalCustomerKeys)
	optForMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2ForMetric, config.Cfg.Dataway.GlobalCustomerKeys)

	return &containerCollector{
		ipt:           ipt,
		runtime:       r,
		k8sClient:     k8sClient,
		localNodeName: datakit.DKHost,
		maxConcurrent: ipt.ContainerMaxConcurrent,

		enableCollectLogging:          true,
		enableExtractK8sLabelAsTagsV1: ipt.EnableExtractK8sLabelAsTags,
		podLabelAsTagsForNonMetric:    optForNonMetric,
		podLabelAsTagsForMetric:       optForMetric,

		logFilter: logFilter,
		logTable:  newLogTable(),
		extraTags: tags,
		feeder:    ipt.Feeder,
	}, nil
}

func (c *containerCollector) StartCollect() {
	tickers := []*time.Ticker{
		time.NewTicker(metricInterval),
		time.NewTicker(c.ipt.LoggingSearchInterval),
		time.NewTicker(objectInterval),
	}

	for _, t := range tickers {
		defer t.Stop()
	}

	c.gatherLogging()
	time.Sleep(time.Second * 3) // window time

	c.gatherObject()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("container collect exit")
			return

		case tt := <-tickers[0].C:
			if c.ipt.EnableContainerMetric {
				c.ptsTime = inputs.AlignTime(tt, c.ptsTime, metricInterval)
				c.gatherMetric()
			}

		case <-tickers[1].C:
			if c.enableCollectLogging {
				c.gatherLogging()
			}

		case <-tickers[2].C:
			c.gatherObject()
		}
	}
}

func (c *containerCollector) ReloadConfigKV(_ map[string]string) error {
	l.Info("reload kv")
	c.logTable.closeAll()
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
		return fmt.Errorf("endpoint %s does not exist, maybe it is not running", endpoint)
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
