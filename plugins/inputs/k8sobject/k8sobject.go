package k8sobject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	"github.com/influxdata/telegraf/filter"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// Kubernetes represents the config object for the plugin
type K8sObject struct {
	URL string

	Interval string `toml:"interval"`

	// Bearer Token authorization file path
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`

	//LabelInclude []string `toml:"label_include"`
	//LabelExclude []string `toml:"label_exclude"`

	labelFilter filter.Filter

	// HTTP Timeout specified as a string - 3s, 1m, 1h
	ResponseTimeout time.Duration

	tls.ClientConfig

	RoundTripper http.RoundTripper
}

// SummaryMetrics represents all the summary data about a paritcular node retrieved from a kubelet
type SummaryMetrics struct {
	Node NodeMetrics  `json:"node"`
	Pods []PodMetrics `json:"pods"`
}

// NodeMetrics represents detailed information about a node
type NodeMetrics struct {
	NodeName         string             `json:"nodeName"`
	SystemContainers []ContainerMetrics `json:"systemContainers"`
	StartTime        time.Time          `json:"startTime"`
	CPU              CPUMetrics         `json:"cpu"`
	Memory           MemoryMetrics      `json:"memory"`
	Network          NetworkMetrics     `json:"network"`
	FileSystem       FileSystemMetrics  `json:"fs"`
	Runtime          RuntimeMetrics     `json:"runtime"`
}

// ContainerMetrics represents the metric data collect about a container from the kubelet
type ContainerMetrics struct {
	Name      string            `json:"name"`
	StartTime time.Time         `json:"startTime"`
	CPU       CPUMetrics        `json:"cpu"`
	Memory    MemoryMetrics     `json:"memory"`
	RootFS    FileSystemMetrics `json:"rootfs"`
	LogsFS    FileSystemMetrics `json:"logs"`
}

// RuntimeMetrics contains metric data on the runtime of the system
type RuntimeMetrics struct {
	ImageFileSystem FileSystemMetrics `json:"imageFs"`
}

// CPUMetrics represents the cpu usage data of a pod or node
type CPUMetrics struct {
	Time                 time.Time `json:"time"`
	UsageNanoCores       int64     `json:"usageNanoCores"`
	UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
}

// PodMetrics contains metric data on a given pod
type PodMetrics struct {
	PodRef     PodReference       `json:"podRef"`
	StartTime  *time.Time         `json:"startTime"`
	Containers []ContainerMetrics `json:"containers"`
	Network    NetworkMetrics     `json:"network"`
	Volumes    []VolumeMetrics    `json:"volume"`
}

// PodReference is how a pod is identified
type PodReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// MemoryMetrics represents the memory metrics for a pod or node
type MemoryMetrics struct {
	Time            time.Time `json:"time"`
	AvailableBytes  int64     `json:"availableBytes"`
	UsageBytes      int64     `json:"usageBytes"`
	WorkingSetBytes int64     `json:"workingSetBytes"`
	RSSBytes        int64     `json:"rssBytes"`
	PageFaults      int64     `json:"pageFaults"`
	MajorPageFaults int64     `json:"majorPageFaults"`
}

// FileSystemMetrics represents disk usage metrics for a pod or node
type FileSystemMetrics struct {
	AvailableBytes int64 `json:"availableBytes"`
	CapacityBytes  int64 `json:"capacityBytes"`
	UsedBytes      int64 `json:"usedBytes"`
}

// NetworkMetrics represents network usage data for a pod or node
type NetworkMetrics struct {
	Time     time.Time `json:"time"`
	RXBytes  int64     `json:"rxBytes"`
	RXErrors int64     `json:"rxErrors"`
	TXBytes  int64     `json:"txBytes"`
	TXErrors int64     `json:"txErrors"`
}

// VolumeMetrics represents the disk usage data for a given volume
type VolumeMetrics struct {
	Name           string `json:"name"`
	AvailableBytes int64  `json:"availableBytes"`
	CapacityBytes  int64  `json:"capacityBytes"`
	UsedBytes      int64  `json:"usedBytes"`
}

type Pods struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	Items      []Item `json:"items"`
}

type Item struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}

type K8sContent struct {
	PodRef    *PodReference     `json:"podRef"`
	Container *ContainerMetrics `json:"container"`
}

type K8sObj struct {
	Name    string `json:"name"`
	Class   string `json:"class"`
	Content string `json:"content"`
}

const (
	k8sSampleConfig = `
[[inputs.k8sobject]]
  ## URL for the kubelet
  url = "http://127.0.0.1:10255"

  ## Monitor interval
  interval = "60s"

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  ## If both of these are empty, we'll use the default serviceaccount:
  ## at: /run/secrets/kubernetes.io/serviceaccount/token
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`
	defaultInterval           = "60s"
	summaryEndpoint           = `%s/stats/summary`
	defaultServiceAccountPath = "/run/secrets/kubernetes.io/serviceaccount/token"
	pluginName                = "k8sobject"
)

var (
	log = logger.DefaultSLogger(pluginName)
)

func (k *K8sObject) Catalog() string {
	return pluginName
}

func (k *K8sObject) SampleConfig() string {
	return k8sSampleConfig
}

func (k *K8sObject) Run() {
	var err error

	log = logger.SLogger(pluginName)
	log.Infof("%s input started...", pluginName)

	if k.Interval == "" {
		k.Interval = defaultInterval
	}

	if err = k.getToken(); err != nil {
		log.Errorf("getToken err: %v", err)
	}

	k.Gather()
}

func init() {
	inputs.Add(pluginName, func() inputs.Input {
		s := &K8sObject{}
		return s
	})
}

func (k *K8sObject) getToken() error {
	// If neither are provided, use the default service account.
	if k.BearerTokenString != "" {
		return nil
	}

	if k.BearerToken != "" {
		token, err := ioutil.ReadFile(k.BearerToken)
		if err != nil {
			return err
		}
		k.BearerTokenString = strings.TrimSpace(string(token))
	}

	return nil
}

//Gather collects kubernetes object from a given URL
func (k *K8sObject) Gather() {
	d, err := time.ParseDuration(k.Interval)
	if err != nil {
		log.Errorf("ParseDuration err: %s", err)
		return
	}

	ticker := time.NewTicker(d)
	for {
		select {
		case <-ticker.C:
			rst, err := k.gatherSummary(k.URL)
			if err != nil {
				log.Errorf("gatherSummary err: %v", err)
			} else {
				log.Debugf("%s", string(rst))
				io.NamedFeed(rst, datakit.Object, pluginName)
			}

		case <-datakit.Exit.Wait():
			log.Infof("input %s exit", pluginName)
			return
		}
	}
}

func buildURL(endpoint string, base string) (*url.URL, error) {
	u := fmt.Sprintf(endpoint, base)
	addr, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse address '%s': %s", u, err)
	}
	return addr, nil
}

func (k *K8sObject) gatherSummary(baseURL string) ([]byte, error) {
	rst := make([][]byte, 0)
	summaryMetrics := &SummaryMetrics{}
	err := k.LoadJson(fmt.Sprintf("%s/stats/summary", baseURL), summaryMetrics)
	if err != nil {
		return nil, err
	}

	for _, pod := range summaryMetrics.Pods {
		for _, container := range pod.Containers {
			kc := K8sContent{
				&pod.PodRef,
				&container,
			}

			kcJson, _ := json.Marshal(kc)
			tag := map[string]string{"name": container.Name}
			field := map[string]interface{}{"message": string(kcJson)}

			pt, err := io.MakeMetric("k8sobject", tag, field, time.Now())
			if err != nil {
				return nil, err
			}

			rst = append(rst, pt)
		}
	}
	return bytes.Join(rst, []byte("\n")), nil
}

func (k *K8sObject) gatherPodInfo(baseURL string) ([]Metadata, error) {
	var podApi Pods
	err := k.LoadJson(fmt.Sprintf("%s/pods", baseURL), &podApi)
	if err != nil {
		return nil, err
	}
	var podInfos []Metadata
	for _, podMetadata := range podApi.Items {
		podInfos = append(podInfos, podMetadata.Metadata)
	}
	return podInfos, nil
}

func (k *K8sObject) LoadJson(url string, v interface{}) error {
	var req, err = http.NewRequest("GET", url, nil)
	var resp *http.Response
	tlsCfg, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	if k.RoundTripper == nil {
		if k.ResponseTimeout < time.Second {
			k.ResponseTimeout = time.Second * 5
		}
		k.RoundTripper = &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			TLSClientConfig:       tlsCfg,
			ResponseHeaderTimeout: k.ResponseTimeout,
		}
	}
	req.Header.Set("Authorization", "Bearer "+k.BearerTokenString)
	req.Header.Add("Accept", "application/json")
	resp, err = k.RoundTripper.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf(`Error parsing response: %s`, err)
	}

	return nil
}
