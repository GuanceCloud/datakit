package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
)

const defaultServiceAccountPath = "/run/secrets/kubernetes.io/serviceaccount/token"

// Kubernetes represents the config object for the plugin
type Kubernetes struct {
	URL           string   `toml:"kubelet_url"`
	IgnorePodName []string `toml:"ignore_pod_name"`

	// Bearer Token authorization file path
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`

	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	roundTripper http.RoundTripper
}

func (k *Kubernetes) Init() error {
	l.Debugf("use kubelet_url %s", k.URL)
	u, err := url.Parse(k.URL)
	if err != nil {
		return err
	}

	// kubelet API 没有提供 ping 功能，此处手动检查该端口是否可以连接
	if err := net.RawConnect(u.Hostname(), u.Port(), time.Second); err != nil {
		l.Warnf("failed of kubelet connect(not collect kubelet): %s", err)
		return err
	}

	// If neither are provided, use the default service account.
	if k.BearerToken == "" && k.BearerTokenString == "" {
		k.BearerToken = defaultServiceAccountPath
	}

	if k.BearerToken != "" {
		if path.IsFileExists(k.BearerToken) {
			token, err := ioutil.ReadFile(k.BearerToken)
			if err != nil {
				return err
			}
			k.BearerTokenString = strings.TrimSpace(string(token))
		} else {
			l.Info("kubernetes bearerToken is not exist, use empty token")
		}
	}

	t := net.TlsClientConfig{
		CaCerts: func() []string {
			if k.TLSCA == "" {
				return nil
			}
			return []string{k.TLSCA}
		}(),
		Cert:               k.TLSCert,
		CertKey:            k.TLSKey,
		InsecureSkipVerify: k.InsecureSkipVerify,
	}

	tlsConfig, err := t.TlsConfig()
	if err != nil {
		return err
	}

	k.roundTripper = &http.Transport{
		TLSHandshakeTimeout:   apiTimeoutDuration,
		TLSClientConfig:       tlsConfig,
		ResponseHeaderTimeout: apiTimeoutDuration,
	}

	l.Debug("init k8s client success")
	return nil
}

func (k *Kubernetes) Stop() {
	return
}

func (k *Kubernetes) Metric(ctx context.Context, in chan<- []*job) {
	summary, err := k.getStatsSummary()
	if err != nil {
		l.Error(err)
		return
	}

	nodeName := summary.Node.NodeName

	var jobs []*job
	for _, podMetrics := range summary.Pods {
		if k.ignorePodName(podMetrics.PodRef.Name) {
			continue
		}

		result := k.gatherPodMetrics(&podMetrics)
		if result == nil {
			return
		}
		result.addTag("node_name", nodeName)
		result.setMetric()
		jobs = append(jobs, result)
	}

	l.Debugf("get len(%d) k8s metric", len(jobs))
	in <- jobs
}

func (k *Kubernetes) Object(ctx context.Context, in chan<- []*job) {
	var summary *SummaryMetrics
	var pods *Pods
	var err error

	summary, err = k.getStatsSummary()
	if err != nil {
		l.Error(err)
		return
	}

	pods, err = k.getPods()
	if err != nil {
		l.Error(err)
		return
	}

	nodeName := summary.Node.NodeName

	var jobs []*job
	for _, item := range pods.Items {
		if k.ignorePodName(item.Metadata.Name) {
			continue
		}

		result := k.gatherPodObject(&item)
		if result == nil {
			return
		}

		func() {
			podMetrics := k.findPodMetricsByUID(item.Metadata.UID, summary)
			if podMetrics == nil {
				return
			}

			resMetrics := k.gatherPodMetrics(podMetrics)
			if resMetrics == nil {
				return
			}

			if err := result.merge(resMetrics); err != nil {
				l.Warn(err)
			}
		}()

		result.addTag("node_name", nodeName)
		// 一定要在此处进行 add pod_name
		result.addTag("pod_name", item.Metadata.Name)

		if message, err := result.marshal(); err != nil {
			l.Warnf("failed of marshal json, %s", err)
		} else {
			result.addField("message", string(message))
		}

		result.setObject()
		jobs = append(jobs, result)
	}

	l.Debugf("get len(%d) k8s object/pod", len(jobs))
	in <- jobs
}

func (k *Kubernetes) Logging(ctx context.Context) {
	return
}

func (k *Kubernetes) ignorePodName(name string) bool {
	return regexpMatchString(k.IgnorePodName, name)
}

func (k *Kubernetes) gatherPodMetrics(pod *PodMetrics) *job {
	var tags = make(map[string]string)
	tags["namespace"] = pod.PodRef.Namespace
	tags["pod_name"] = pod.PodRef.Name

	var fields = make(map[string]interface{})
	fields["cpu_usage_nanocores"] = float64(pod.CPU.UsageNanoCores)
	fields["cpu_usage_core_nanoseconds"] = float64(pod.CPU.UsageCoreNanoSeconds)
	fields["memory_usage_bytes"] = float64(pod.Memory.UsageBytes)
	fields["memory_working_set_bytes"] = float64(pod.Memory.WorkingSetBytes)
	fields["memory_rss_bytes"] = float64(pod.Memory.RSSBytes)
	fields["memory_page_faults"] = float64(pod.Memory.PageFaults)
	fields["memory_major_page_faults"] = float64(pod.Memory.MajorPageFaults)
	fields["network_rx_bytes"] = float64(pod.Network.RXBytes())
	fields["network_rx_errors"] = float64(pod.Network.RXErrors())
	fields["network_tx_bytes"] = float64(pod.Network.TXBytes())
	fields["network_tx_errors"] = float64(pod.Network.TXErrors())

	if cpuPrecent, err := pod.CPU.Percent(); err == nil {
		fields["cpu_usage"] = cpuPrecent
	}

	return &job{measurement: kubeletPodName, tags: tags, fields: fields, ts: time.Now()}
}

func (k *Kubernetes) gatherPodObject(item *PodItem) *job {
	var tags = make(map[string]string)
	tags["name"] = item.Metadata.UID
	tags["state"] = item.Status.Phase

	fields := map[string]interface{}{
		"age":       item.Status.Age(),
		"restart":   item.Status.ContainerStatuses.RestartCount(),
		"ready":     item.Status.ContainerStatuses.Ready(),
		"available": item.Status.ContainerStatuses.Length(),
		// http://gitlab.jiagouyun.com/cloudcare-tools/kodo/-/issues/61#note_11580
		"df_label":            item.Metadata.LabelsJSON(),
		"df_label_premission": "read_only",
		"df_label_source":     "datakit",
	}

	return &job{measurement: kubeletPodName, tags: tags, fields: fields, ts: time.Now()}
}

func (k *Kubernetes) findPodMetricsByUID(uid string, summary *SummaryMetrics) *PodMetrics {
	for _, podMetrics := range summary.Pods {
		if podMetrics.PodRef.UID == uid {
			return &podMetrics
		}
	}
	return nil
}

func (k *Kubernetes) getPods() (*Pods, error) {
	var pods Pods
	err := k.LoadJson(fmt.Sprintf("%s/pods", k.URL), &pods)
	if err != nil {
		return nil, err
	}
	return &pods, nil
}

func (k *Kubernetes) getStatsSummary() (*SummaryMetrics, error) {
	var summary SummaryMetrics
	err := k.LoadJson(fmt.Sprintf("%s/stats/summary", k.URL), &summary)
	if err != nil {
		return nil, err
	}
	return &summary, err
}

func (k *Kubernetes) GetContainerPodNamespace(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid containerID, cannot be empty")
	}
	pods, err := k.getPods()
	if err != nil {
		return "", err
	}
	return pods.GetContainerPodNamespace(id), nil
}

func (k *Kubernetes) GetContainerPodName(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid containerID, cannot be empty")
	}
	pods, err := k.getPods()
	if err != nil {
		return "", err
	}
	return pods.GetContainerPodName(id), nil
}

func (k *Kubernetes) GetContainerDeploymentName(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid containerID, cannot be empty")
	}
	pods, err := k.getPods()
	if err != nil {
		return "", err
	}
	return pods.GetContainerDeploymentName(id), nil
}

func (k *Kubernetes) LoadJson(url string, v interface{}) error {
	var req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	var resp *http.Response
	req.Header.Set("Authorization", "Bearer "+k.BearerTokenString)
	req.Header.Add("Accept", "application/json")

	resp, err = k.roundTripper.RoundTrip(req)
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

type Pods struct {
	Kind       string    `json:"kind"`
	ApiVersion string    `json:"apiVersion"`
	Items      []PodItem `json:"items"`
}

type PodItem struct {
	Metadata PodItemMetadata `json:"metadata"`
	Status   PodItemStatus   `json:"status"`
}

type PodItemMetadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	UID       string            `json:"uid"`
	Labels    map[string]string `json:"labels"`
}

type PodItemStatus struct {
	Phase             string                  `json:"phase"`
	StartTime         string                  `json:"startTime"`
	ContainerStatuses PodItemStatusContainers `json:"containerStatuses"`
}

type PodItemStatusContainers []PodItemStatusContainer

type PodItemStatusContainer struct {
	ContainerID  string `json:"containerID"`
	RestartCount int64  `json:"restartCount"`
	Ready        bool   `json:"ready"`
}

type SummaryMetrics struct {
	Node NodeMetrics  `json:"node"`
	Pods []PodMetrics `json:"pods"`
}

type NodeMetrics struct {
	NodeName         string             `json:"nodeName"`
	SystemContainers []ContainerMetrics `json:"systemContainers"`
	CPU              CPUMetrics         `json:"cpu"`
	Memory           MemoryMetrics      `json:"memory"`
	Network          NetworkMetrics     `json:"network"`
	FileSystem       FileSystemMetrics  `json:"fs"`
	Runtime          RuntimeMetrics     `json:"runtime"`
}

type ContainerMetrics struct {
	Name   string            `json:"name"`
	CPU    CPUMetrics        `json:"cpu"`
	Memory MemoryMetrics     `json:"memory"`
	RootFS FileSystemMetrics `json:"rootfs"`
	LogsFS FileSystemMetrics `json:"logs"`
}

type RuntimeMetrics struct {
	ImageFileSystem FileSystemMetrics `json:"imageFs"`
}

type CPUMetrics struct {
	Time                 time.Time `json:"time"`
	UsageNanoCores       int64     `json:"usageNanoCores"`
	UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
}

type PodMetrics struct {
	PodRef     PodReference       `json:"podRef"`
	StartTime  *time.Time         `json:"startTime"`
	Containers []ContainerMetrics `json:"containers"`
	CPU        CPUMetrics         `json:"cpu"`
	Memory     MemoryMetrics      `json:"memory"`
	Network    NetworkMetrics     `json:"network"`
}

type PodReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}

type MemoryMetrics struct {
	Time            time.Time `json:"time"`
	AvailableBytes  int64     `json:"availableBytes"`
	UsageBytes      int64     `json:"usageBytes"`
	WorkingSetBytes int64     `json:"workingSetBytes"`
	RSSBytes        int64     `json:"rssBytes"`
	PageFaults      int64     `json:"pageFaults"`
	MajorPageFaults int64     `json:"majorPageFaults"`
}

type FileSystemMetrics struct {
	AvailableBytes int64 `json:"availableBytes"`
	CapacityBytes  int64 `json:"capacityBytes"`
	UsedBytes      int64 `json:"usedBytes"`
}

type NetworkMetrics struct {
	Time       time.Time `json:"time"`
	Interfaces []struct {
		Name     string `json:"name"`
		RXBytes  int64  `json:"rxBytes"`
		RXErrors int64  `json:"rxErrors"`
		TXBytes  int64  `json:"txBytes"`
		TXErrors int64  `json:"txErrors"`
	} `json:"interfaces"`
}

type VolumeMetrics struct {
	Name           string `json:"name"`
	AvailableBytes int64  `json:"availableBytes"`
	CapacityBytes  int64  `json:"capacityBytes"`
	UsedBytes      int64  `json:"usedBytes"`
}

func (p PodItemMetadata) LabelsJSON() string {
	// empty array
	labelsString := "[]"

	if len(p.Labels) != 0 {
		var lb []string
		for k, v := range p.Labels {
			lb = append(lb, k+":"+v)
		}

		b, err := json.Marshal(lb)
		if err == nil {
			labelsString = string(b)
		}
	}

	return labelsString
}

func (p PodItemStatus) Age() int64 {
	ts, err := time.Parse(time.RFC3339, p.StartTime)
	if err != nil {
		return -1
	}
	return time.Since(ts).Milliseconds() / 1e3 // 毫秒除以1000得秒数，不使用Second()因为它返回浮点
}

func (ps PodItemStatusContainers) RestartCount() int64 {
	var num int64
	for _, p := range ps {
		num += p.RestartCount
	}
	return num
}

func (ps PodItemStatusContainers) Length() int64 {
	return int64(len(ps))
}

func (ps PodItemStatusContainers) Ready() int64 {
	var num int64
	for _, p := range ps {
		if p.Ready {
			num++
		}
	}
	return num
}

func (n NetworkMetrics) RXBytes() int64 {
	var sum int64
	for _, i := range n.Interfaces {
		sum += i.RXBytes
	}
	return sum
}

func (n NetworkMetrics) RXErrors() int64 {
	var sum int64
	for _, i := range n.Interfaces {
		sum += i.RXErrors
	}
	return sum
}

func (n NetworkMetrics) TXBytes() int64 {
	var sum int64
	for _, i := range n.Interfaces {
		sum += i.TXBytes
	}
	return sum
}

func (n NetworkMetrics) TXErrors() int64 {
	var sum int64
	for _, i := range n.Interfaces {
		sum += i.TXErrors
	}
	return sum
}

func (c *CPUMetrics) Percent() (float64, error) {
	if c.UsageNanoCores == 0 {
		return -1, fmt.Errorf("cpu usageNanoCores cannot be zero")

	}
	// source link: https://github.com/kubernetes/heapster/issues/650#issuecomment-147795824
	// cpu_usage_core_nanoseconds / (cpu_usage_nanocores * 1000000000) * 100
	return float64(c.UsageCoreNanoSeconds) / float64(c.UsageNanoCores*1000000000) * 100, nil
}

func (m *MemoryMetrics) Percent() (float64, error) {
	if m.AvailableBytes+m.UsageBytes == 0 {
		return -1, fmt.Errorf("memory total cannot be zero")
	}
	// mem_usage_percent = memory_usage_bytes / (memory_usage_bytes + memory_available_bytes)
	return float64(m.UsageBytes) / float64(m.UsageBytes+m.AvailableBytes), nil
}

func (p *Pods) GetContainerDeploymentName(id string) string {
	podName := p.GetContainerPodName(id)
	if podName == "" {
		return ""
	}
	return getDeploymentFromPodName(podName)
}

func (p *Pods) GetContainerPodName(id string) string {
	for _, podMetadata := range p.Items {
		if len(podMetadata.Status.ContainerStatuses) == 0 {
			continue
		}
		for _, containerStauts := range podMetadata.Status.ContainerStatuses {
			if containerStauts.ContainerID == id {
				return podMetadata.Metadata.Name
			}
		}
	}
	return ""
}

func (p *Pods) GetContainerPodNamespace(id string) string {
	for _, podMetadata := range p.Items {
		if len(podMetadata.Status.ContainerStatuses) == 0 {
			continue
		}
		for _, containerStauts := range podMetadata.Status.ContainerStatuses {
			if containerStauts.ContainerID == id {
				return podMetadata.Metadata.Namespace
			}
		}
	}
	return ""
}

func (p *Pods) GetContainerPodUID(id string) string {
	for _, podMetadata := range p.Items {
		if len(podMetadata.Status.ContainerStatuses) == 0 {
			continue
		}
		for _, containerStauts := range podMetadata.Status.ContainerStatuses {
			if containerStauts.ContainerID == id {
				return podMetadata.Metadata.UID
			}
		}
	}
	return ""
}

func verifyIntegrityOfK8sConnect(k *Kubernetes) bool {
	return k != nil && k.roundTripper != nil
}

func getDeploymentFromPodName(name string) string {
	s := strings.Split(name, "-")
	if len(s) > 2 {
		return strings.Join(s[:len(s)-2], "-")
	}
	return name
}
