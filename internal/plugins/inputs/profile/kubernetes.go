// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"context"
	"fmt"
	"hash/fnv"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

const (
	defaultKubernetesProfileInterval        = 10 * time.Minute
	defaultKubernetesProfileDuration        = 10 * time.Second
	defaultKubernetesProfileMonitorInterval = 10 * time.Second
	defaultKubernetesProfileTriggerWindow   = time.Minute
	defaultKubernetesProfileCooldown        = 10 * time.Minute
	defaultKubernetesProfilePath            = "/debug/pprof"
	defaultKubernetesProfileConcurrency     = 2
	kubernetesProfileInformerListLimit      = int64(50)
)

// KubernetesProfiler discovers Go pprof endpoints from Kubernetes Pods and
// controls scheduled and resource-triggered collection for the matched targets.
type KubernetesProfiler struct {
	NodeLocal  *bool    `toml:"node_local"`
	Namespaces []string `toml:"namespaces"`
	Selector   string   `toml:"selector"`
	Container  string   `toml:"container"`
	Scheme     string   `toml:"scheme"`
	Port       string   `toml:"port"`
	Path       string   `toml:"path"`

	Service string            `toml:"service"`
	Env     string            `toml:"env"`
	Version string            `toml:"version"`
	Tags    map[string]string `toml:"tags"`

	PodLabelAsTags      map[string]string `toml:"pod_label_as_tags"`
	PodAnnotationAsTags map[string]string `toml:"pod_annotation_as_tags"`

	Interval          time.Duration `toml:"interval"`
	ScheduledTypes    []string      `toml:"scheduled_types"`
	TriggerTypes      []string      `toml:"trigger_types"`
	ProfileDuration   time.Duration `toml:"profile_duration"`
	EmergencyDuration time.Duration `toml:"emergency_duration"`
	RequestTimeout    time.Duration `toml:"request_timeout"`

	MonitorInterval time.Duration `toml:"monitor_interval"`
	TriggerWindow   time.Duration `toml:"trigger_window"`
	Cooldown        time.Duration `toml:"cooldown"`
	MaxConcurrency  int           `toml:"max_concurrency"`

	CPUUsageBaseLimit  float64 `toml:"cpu_usage_base_limit"`
	CPUUsageMillicores int64   `toml:"cpu_usage_millicores"`
	MemUsageBaseLimit  float64 `toml:"mem_usage_base_limit"`
	MemUsageBytes      int64   `toml:"mem_usage_bytes"`

	CPUEmergencyBaseLimit  float64 `toml:"cpu_emergency_base_limit"`
	CPUEmergencyMillicores int64   `toml:"cpu_emergency_millicores"`
	MemEmergencyBaseLimit  float64 `toml:"mem_emergency_base_limit"`
	MemEmergencyBytes      int64   `toml:"mem_emergency_bytes"`

	TLSOpen            bool   `toml:"tls_open"`
	CacertFile         string `toml:"tls_ca"`
	CertFile           string `toml:"tls_cert"`
	KeyFile            string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
}

type kubernetesProfileRule struct {
	config     *KubernetesProfiler
	selector   labels.Selector
	namespaces map[string]struct{}
	client     *http.Client
	semaphore  chan struct{}
	index      int
}

func normalizeKubernetesProfileRules(configs []*KubernetesProfiler) ([]*kubernetesProfileRule, error) {
	rules := make([]*kubernetesProfileRule, 0, len(configs))
	for index, cfg := range configs {
		if cfg == nil {
			return nil, fmt.Errorf("kubernetes profile rule %d cannot be nil", index)
		}

		if cfg.NodeLocal != nil && !*cfg.NodeLocal {
			return nil, fmt.Errorf("kubernetes profile rule %d: only node_local=true is supported", index)
		}
		if cfg.NodeLocal == nil {
			nodeLocal := true
			cfg.NodeLocal = &nodeLocal
		}
		if strings.TrimSpace(cfg.Port) == "" {
			return nil, fmt.Errorf("kubernetes profile rule %d: port cannot be empty", index)
		}
		cfg.Port = strings.TrimSpace(cfg.Port)
		if port, err := strconv.ParseInt(cfg.Port, 10, 32); err == nil && (port < 1 || port > 65535) {
			return nil, fmt.Errorf("kubernetes profile rule %d: numeric port must be between 1 and 65535", index)
		}

		selector := labels.Everything()
		if cfg.Selector != "" {
			parsed, err := labels.Parse(cfg.Selector)
			if err != nil {
				return nil, fmt.Errorf("kubernetes profile rule %d: invalid selector: %w", index, err)
			}
			selector = parsed
		}

		cfg.Scheme = strings.ToLower(strings.TrimSpace(cfg.Scheme))
		if cfg.Scheme == "" {
			cfg.Scheme = "http"
		}
		if cfg.Scheme != "http" && cfg.Scheme != "https" {
			return nil, fmt.Errorf("kubernetes profile rule %d: unsupported scheme %q", index, cfg.Scheme)
		}
		if cfg.Path == "" {
			cfg.Path = defaultKubernetesProfilePath
		}
		if !strings.HasPrefix(cfg.Path, "/") {
			return nil, fmt.Errorf("kubernetes profile rule %d: path must start with '/'", index)
		}

		if cfg.Interval == 0 {
			cfg.Interval = defaultKubernetesProfileInterval
		}
		if cfg.Interval < 10*time.Second {
			return nil, fmt.Errorf("kubernetes profile rule %d: interval must be at least 10s", index)
		}
		if cfg.ProfileDuration == 0 {
			cfg.ProfileDuration = defaultKubernetesProfileDuration
		}
		if cfg.ProfileDuration < time.Second {
			return nil, fmt.Errorf("kubernetes profile rule %d: profile_duration must be at least 1s", index)
		}
		if cfg.EmergencyDuration == 0 {
			cfg.EmergencyDuration = cfg.ProfileDuration
		}
		if cfg.EmergencyDuration < time.Second {
			return nil, fmt.Errorf("kubernetes profile rule %d: emergency_duration must be at least 1s", index)
		}

		if len(cfg.ScheduledTypes) == 0 {
			cfg.ScheduledTypes = []string{"heap", "goroutine"}
		}
		if err := validateProfileTypes(cfg.ScheduledTypes); err != nil {
			return nil, fmt.Errorf("kubernetes profile rule %d scheduled_types: %w", index, err)
		}

		if cfg.MonitorInterval == 0 {
			cfg.MonitorInterval = defaultKubernetesProfileMonitorInterval
		}
		if cfg.MonitorInterval < time.Second {
			return nil, fmt.Errorf("kubernetes profile rule %d: monitor_interval must be at least 1s", index)
		}
		if cfg.TriggerWindow == 0 {
			cfg.TriggerWindow = defaultKubernetesProfileTriggerWindow
		}
		if cfg.TriggerWindow < cfg.MonitorInterval {
			return nil, fmt.Errorf("kubernetes profile rule %d: trigger_window must not be shorter than monitor_interval", index)
		}
		if cfg.Cooldown == 0 {
			cfg.Cooldown = defaultKubernetesProfileCooldown
		}
		if cfg.Cooldown < 0 {
			return nil, fmt.Errorf("kubernetes profile rule %d: cooldown cannot be negative", index)
		}
		if cfg.MaxConcurrency == 0 {
			cfg.MaxConcurrency = defaultKubernetesProfileConcurrency
		}
		if cfg.MaxConcurrency < 1 {
			return nil, fmt.Errorf("kubernetes profile rule %d: max_concurrency must be positive", index)
		}

		if hasResourceTrigger(cfg) && len(cfg.TriggerTypes) == 0 {
			cfg.TriggerTypes = []string{"cpu", "heap", "goroutine"}
		}
		if len(cfg.TriggerTypes) > 0 {
			if err := validateProfileTypes(cfg.TriggerTypes); err != nil {
				return nil, fmt.Errorf("kubernetes profile rule %d trigger_types: %w", index, err)
			}
		}
		if err := validateThresholds(cfg, index); err != nil {
			return nil, err
		}

		maxDuration := cfg.ProfileDuration
		if cfg.EmergencyDuration > maxDuration {
			maxDuration = cfg.EmergencyDuration
		}
		if cfg.RequestTimeout == 0 {
			cfg.RequestTimeout = maxDuration + 5*time.Second
			if cfg.RequestTimeout < 15*time.Second {
				cfg.RequestTimeout = 15 * time.Second
			}
		}
		if cfg.RequestTimeout <= maxDuration {
			return nil, fmt.Errorf("kubernetes profile rule %d: request_timeout must exceed profile durations", index)
		}

		profiler := &GoProfiler{
			timeout:            cfg.RequestTimeout,
			TLSOpen:            cfg.TLSOpen,
			CacertFile:         cfg.CacertFile,
			CertFile:           cfg.CertFile,
			KeyFile:            cfg.KeyFile,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}
		client, err := profiler.createHTTPClient()
		if err != nil {
			return nil, fmt.Errorf("kubernetes profile rule %d: create HTTP client: %w", index, err)
		}
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		if client.Transport == nil {
			client.Transport = http.DefaultTransport.(*http.Transport).Clone()
		}
		transport := client.Transport.(*http.Transport)
		transport.Proxy = nil
		transport.MaxIdleConns = cfg.MaxConcurrency * 2
		transport.IdleConnTimeout = 90 * time.Second

		namespaces := make(map[string]struct{}, len(cfg.Namespaces))
		for _, namespace := range cfg.Namespaces {
			namespace = strings.TrimSpace(namespace)
			if namespace != "" {
				namespaces[namespace] = struct{}{}
			}
		}
		if err := validateTagMappings(cfg.PodLabelAsTags); err != nil {
			return nil, fmt.Errorf("kubernetes profile rule %d pod_label_as_tags: %w", index, err)
		}
		if err := validateTagMappings(cfg.PodAnnotationAsTags); err != nil {
			return nil, fmt.Errorf("kubernetes profile rule %d pod_annotation_as_tags: %w", index, err)
		}

		rules = append(rules, &kubernetesProfileRule{
			config:     cfg,
			selector:   selector,
			namespaces: namespaces,
			client:     client,
			semaphore:  make(chan struct{}, cfg.MaxConcurrency),
			index:      index,
		})
	}
	return rules, nil
}

func validateProfileTypes(types []string) error {
	seen := make(map[string]struct{}, len(types))
	for _, profileType := range types {
		if _, ok := profileConfigMap[profileType]; !ok {
			return fmt.Errorf("unsupported profile type %q", profileType)
		}
		if _, ok := seen[profileType]; ok {
			return fmt.Errorf("duplicate profile type %q", profileType)
		}
		seen[profileType] = struct{}{}
	}
	return nil
}

func validateThresholds(cfg *KubernetesProfiler, index int) error {
	values := []struct {
		name  string
		value float64
	}{
		{"cpu_usage_base_limit", cfg.CPUUsageBaseLimit},
		{"mem_usage_base_limit", cfg.MemUsageBaseLimit},
		{"cpu_emergency_base_limit", cfg.CPUEmergencyBaseLimit},
		{"mem_emergency_base_limit", cfg.MemEmergencyBaseLimit},
	}
	for _, item := range values {
		if item.value < 0 {
			return fmt.Errorf("kubernetes profile rule %d: %s cannot be negative", index, item.name)
		}
	}
	absolute := []struct {
		name  string
		value int64
	}{
		{"cpu_usage_millicores", cfg.CPUUsageMillicores},
		{"mem_usage_bytes", cfg.MemUsageBytes},
		{"cpu_emergency_millicores", cfg.CPUEmergencyMillicores},
		{"mem_emergency_bytes", cfg.MemEmergencyBytes},
	}
	for _, item := range absolute {
		if item.value < 0 {
			return fmt.Errorf("kubernetes profile rule %d: %s cannot be negative", index, item.name)
		}
	}
	return nil
}

func validateTagMappings(mappings map[string]string) error {
	for source, target := range mappings {
		if strings.TrimSpace(source) == "" || strings.TrimSpace(target) == "" {
			return fmt.Errorf("tag mapping keys and values cannot be empty")
		}
	}
	return nil
}

func hasResourceTrigger(cfg *KubernetesProfiler) bool {
	return cfg.CPUUsageBaseLimit > 0 || cfg.CPUUsageMillicores > 0 ||
		cfg.MemUsageBaseLimit > 0 || cfg.MemUsageBytes > 0 ||
		cfg.CPUEmergencyBaseLimit > 0 || cfg.CPUEmergencyMillicores > 0 ||
		cfg.MemEmergencyBaseLimit > 0 || cfg.MemEmergencyBytes > 0
}

func (r *kubernetesProfileRule) matches(pod *corev1.Pod) bool {
	if len(r.namespaces) > 0 {
		if _, ok := r.namespaces[pod.Namespace]; !ok {
			return false
		}
	}
	return r.selector.Matches(labels.Set(pod.Labels))
}

type kubeletProfileStatsProvider interface {
	GetStatsSummaryWithContext(context.Context) (*statsv1alpha1.Summary, error)
}

type profileTargetCollector interface {
	collect(context.Context, []string, time.Duration, map[string]string) error
}

type goProfileTargetCollector struct {
	profiler *GoProfiler
}

func (c *goProfileTargetCollector) collect(
	ctx context.Context,
	types []string,
	duration time.Duration,
	tags map[string]string,
) error {
	return c.profiler.collect(ctx, types, duration, tags)
}

type profileCollectorFactory func(*kubernetesProfileRule, string, *Input) (profileTargetCollector, error)

type kubernetesProfileTarget struct {
	mu sync.Mutex

	id        string
	podKey    string
	podUID    string
	container string
	endpoint  string
	rule      *kubernetesProfileRule
	collector profileTargetCollector
	pod       *corev1.Pod
	tags      map[string]string

	nextScheduled time.Time
	nextMonitor   time.Time
	lastTriggered time.Time
	backoffUntil  time.Time
	failures      int
	busy          bool
	removed       bool
	ready         bool
	cancel        context.CancelFunc
	triggerState  resourceTriggerState
}

type desiredProfileTarget struct {
	id        string
	podKey    string
	podUID    string
	container string
	endpoint  string
	rule      *kubernetesProfileRule
	pod       *corev1.Pod
	tags      map[string]string
}

type kubernetesProfileManager struct {
	input           *Input
	client          kubernetes.Interface
	stats           kubeletProfileStatsProvider
	rules           []*kubernetesProfileRule
	nodeName        string
	newCollector    profileCollectorFactory
	ctx             context.Context
	deltaCache      *profileDeltaCache
	semaphore       chan struct{}
	activeMu        sync.Mutex
	activeEndpoints map[string]struct{}

	mu       sync.RWMutex
	targets  map[string]*kubernetesProfileTarget
	workers  sync.WaitGroup
	stopping bool
}

func newKubernetesProfileManager(
	input *Input,
	client kubernetes.Interface,
	stats kubeletProfileStatsProvider,
	nodeName string,
	configs []*KubernetesProfiler,
) (*kubernetesProfileManager, error) {
	if input == nil || stats == nil {
		return nil, fmt.Errorf("profile Kubernetes dependencies cannot be nil")
	}
	if input.KubernetesMaxConcurrency == 0 {
		input.KubernetesMaxConcurrency = defaultKubernetesProfileConcurrency
	}
	if input.KubernetesDeltaCacheMB == 0 {
		input.KubernetesDeltaCacheMB = defaultKubernetesDeltaCacheMB
	}
	if input.KubernetesMaxConcurrency < 1 || input.KubernetesDeltaCacheMB < 1 {
		return nil, fmt.Errorf("kubernetes_max_concurrency and kubernetes_delta_cache_mb must be positive")
	}
	rules, err := normalizeKubernetesProfileRules(configs)
	if err != nil {
		return nil, err
	}
	manager := &kubernetesProfileManager{
		input:           input,
		client:          client,
		stats:           stats,
		rules:           rules,
		nodeName:        nodeName,
		targets:         make(map[string]*kubernetesProfileTarget),
		deltaCache:      newProfileDeltaCache(int64(input.KubernetesDeltaCacheMB) * MiB),
		semaphore:       make(chan struct{}, input.KubernetesMaxConcurrency),
		activeEndpoints: make(map[string]struct{}),
	}
	manager.newCollector = manager.newGoCollector
	return manager, nil
}

func (m *kubernetesProfileManager) newGoCollector(
	rule *kubernetesProfileRule,
	endpoint string,
	input *Input,
) (profileTargetCollector, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	profiler := &GoProfiler{
		url:              parsed,
		client:           rule.client,
		deltaCache:       m.deltaCache,
		validateResponse: true,
		input:            input,
		tags:             map[string]string{},
	}
	return &goProfileTargetCollector{profiler: profiler}, nil
}

func (m *kubernetesProfileManager) run(ctx context.Context) error {
	defer func() {
		for _, rule := range m.rules {
			rule.client.CloseIdleConnections()
		}
	}()
	if m.client == nil {
		return fmt.Errorf("profile Kubernetes client cannot be nil")
	}
	m.ctx = ctx
	factory := informers.NewSharedInformerFactoryWithOptions(
		m.client,
		0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.Limit = kubernetesProfileInformerListLimit
			options.FieldSelector = fields.OneTermEqualSelector("spec.nodeName", m.nodeName).String()
		}),
	)
	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			m.reconcileObject(obj)
		},
		UpdateFunc: func(_, newObj interface{}) {
			m.reconcileObject(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			m.deleteObject(obj)
		},
	})

	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced) {
		return fmt.Errorf("profile Pod informer cache failed to sync")
	}

	log.Infof("Kubernetes profile discovery started on node %s", m.nodeName)
	// Track the schedule and monitor loops in the same WaitGroup as collection
	// workers: dispatch may still add a worker while a loop finishes its current
	// iteration, and the loop entries keep the counter positive so Add cannot
	// race with Wait below.
	m.workers.Add(2)
	go func() {
		defer m.workers.Done()
		m.runScheduleLoop(ctx)
	}()
	go func() {
		defer m.workers.Done()
		m.runMonitorLoop(ctx)
	}()

	<-ctx.Done()
	m.markAllTargetsRemoved()
	m.workers.Wait()
	return nil
}

func (m *kubernetesProfileManager) reconcileObject(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}
	m.reconcilePod(pod)
}

func (m *kubernetesProfileManager) deleteObject(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		tombstone, tombstoneOK := obj.(cache.DeletedFinalStateUnknown)
		if !tombstoneOK {
			return
		}
		pod, ok = tombstone.Obj.(*corev1.Pod)
		if !ok {
			return
		}
	}
	m.removePodTargets(pod.Namespace + "/" + pod.Name)
}

func (m *kubernetesProfileManager) reconcilePod(pod *corev1.Pod) {
	podKey := pod.Namespace + "/" + pod.Name
	desired := m.profileTargets(pod, true)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopping {
		return
	}

	for id, target := range m.targets {
		if target.podKey != podKey {
			continue
		}
		item, desiredTarget := desired[id]
		if !desiredTarget || item.rule != target.rule {
			m.retireTarget(target)
			delete(m.targets, id)
			observeKubernetesProfileDiscovery("removed")
		}
	}

	for id, item := range desired {
		if target, ok := m.targets[id]; ok {
			target.mu.Lock()
			target.ready = isContainerReady(item.pod, item.container)
			if !target.ready && target.cancel != nil {
				target.cancel()
			}
			target.pod = item.pod.DeepCopy()
			target.tags = copyTags(item.tags)
			target.mu.Unlock()
			continue
		}
		if !isContainerReady(item.pod, item.container) {
			continue
		}

		collector, err := m.newCollector(item.rule, item.endpoint, m.input)
		if err != nil {
			log.Warnf("create Kubernetes profile target %s failed: %s", id, err)
			continue
		}
		now := time.Now()
		target := &kubernetesProfileTarget{
			id:            item.id,
			podKey:        item.podKey,
			podUID:        item.podUID,
			container:     item.container,
			endpoint:      item.endpoint,
			rule:          item.rule,
			collector:     collector,
			pod:           item.pod.DeepCopy(),
			tags:          copyTags(item.tags),
			nextScheduled: now.Add(profileScheduleJitter(item.id, item.rule.config.Interval)),
			nextMonitor:   now,
			ready:         true,
		}
		m.targets[id] = target
		observeKubernetesProfileDiscovery("added")
		log.Infof("discovered Kubernetes profile target pod=%s container=%s endpoint=%s", podKey, item.container, item.endpoint)
	}
	setKubernetesProfileTargets(len(m.targets))
}

func (m *kubernetesProfileManager) desiredTargets(pod *corev1.Pod) map[string]*desiredProfileTarget {
	return m.profileTargets(pod, false)
}

func (m *kubernetesProfileManager) profileTargets(pod *corev1.Pod, includeNotReady bool) map[string]*desiredProfileTarget {
	result := make(map[string]*desiredProfileTarget)
	if pod.Spec.NodeName != m.nodeName || pod.Status.Phase != corev1.PodRunning ||
		net.ParseIP(pod.Status.PodIP) == nil || pod.DeletionTimestamp != nil {
		return result
	}
	endpoints := make(map[string]struct{})

	for _, rule := range m.rules {
		if !rule.matches(pod) {
			continue
		}
		containerName := rule.config.Container
		if containerName == "" {
			if numeric, err := strconv.ParseInt(rule.config.Port, 10, 32); err == nil {
				if len(pod.Spec.Containers) == 1 {
					containerName = pod.Spec.Containers[0].Name
				} else {
					matches := 0
					for _, container := range pod.Spec.Containers {
						for _, port := range container.Ports {
							if port.ContainerPort == int32(numeric) && (port.Protocol == "" || port.Protocol == corev1.ProtocolTCP) {
								containerName = container.Name
								matches++
								break
							}
						}
					}
					if matches != 1 {
						observeKubernetesProfileSkipped("ambiguous_port")
						continue
					}
				}
			}
		}
		for index := range pod.Spec.Containers {
			container := &pod.Spec.Containers[index]
			if containerName != "" && containerName != container.Name {
				continue
			}
			if !includeNotReady && !isContainerReady(pod, container.Name) {
				continue
			}
			port, ok := resolveContainerPort(container, rule.config.Port)
			if !ok {
				continue
			}
			host := net.JoinHostPort(pod.Status.PodIP, strconv.Itoa(int(port)))
			endpoint := (&url.URL{Scheme: rule.config.Scheme, Host: host, Path: rule.config.Path}).String()
			id := string(pod.UID) + "/" + container.Name + "/" + endpoint
			if _, exists := endpoints[endpoint]; exists {
				log.Warnf("overlapping Kubernetes profile rules for pod=%s/%s container=%s endpoint=%s; first rule wins",
					pod.Namespace, pod.Name, container.Name, endpoint)
				continue
			}
			endpoints[endpoint] = struct{}{}
			result[id] = &desiredProfileTarget{
				id:        id,
				podKey:    pod.Namespace + "/" + pod.Name,
				podUID:    string(pod.UID),
				container: container.Name,
				endpoint:  endpoint,
				rule:      rule,
				pod:       pod,
				tags:      kubernetesProfileTags(rule.config, pod, container.Name),
			}
		}
	}
	return result
}

func isContainerReady(pod *corev1.Pod, containerName string) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			return status.Ready && status.State.Running != nil
		}
	}
	return false
}

func resolveContainerPort(container *corev1.Container, configured string) (int32, bool) {
	if port, err := strconv.ParseInt(configured, 10, 32); err == nil {
		if port > 0 && port <= 65535 {
			return int32(port), true
		}
		return 0, false
	}
	for _, port := range container.Ports {
		if port.Name == configured && port.ContainerPort > 0 &&
			(port.Protocol == "" || port.Protocol == corev1.ProtocolTCP) {
			return port.ContainerPort, true
		}
	}
	return 0, false
}

func kubernetesProfileTags(cfg *KubernetesProfiler, pod *corev1.Pod, containerName string) map[string]string {
	tags := copyTags(cfg.Tags)
	workloadKind, workloadName := podutil.PodOwner(pod)
	for source, target := range cfg.PodLabelAsTags {
		if value := pod.Labels[source]; value != "" {
			tags[target] = value
		}
	}
	for source, target := range cfg.PodAnnotationAsTags {
		if value := pod.Annotations[source]; value != "" {
			tags[target] = value
		}
	}
	// Discovery identity cannot be overridden by user-provided tag mappings.
	tags["namespace"] = pod.Namespace
	tags["pod_name"] = pod.Name
	tags["pod_uid"] = string(pod.UID)
	tags["node_name"] = pod.Spec.NodeName
	tags["container_name"] = containerName
	if _, ok := tags["cluster_name_k8s"]; !ok {
		tags["cluster_name_k8s"] = kubernetesClusterName()
	}
	if workloadKind != "" {
		tags["workload_kind"] = workloadKind
	}
	if workloadName != "" {
		tags["workload_name"] = workloadName
	}

	tags["service"] = firstNonEmpty(
		cfg.Service,
		tags["service"],
		pod.Labels["tags.datadoghq.com/service"],
		pod.Labels["app.kubernetes.io/name"],
		workloadName,
		containerName,
	)
	tags["env"] = firstNonEmpty(
		cfg.Env,
		tags["env"],
		pod.Labels["tags.datadoghq.com/env"],
		pod.Labels["app.kubernetes.io/environment"],
	)
	tags["version"] = firstNonEmpty(
		cfg.Version,
		tags["version"],
		pod.Labels["tags.datadoghq.com/version"],
		pod.Labels["app.kubernetes.io/version"],
	)
	for key, value := range tags {
		if key == "" || strings.ContainsAny(key, ",:\r\n") || strings.ContainsAny(value, ",\r\n") {
			delete(tags, key)
			observeKubernetesProfileSkipped("invalid_tag")
		}
	}
	return tags
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// kubernetesClusterName returns the Kubernetes cluster name injected through
// ENV_CLUSTER_NAME_K8S, defaulting to "default" like the container collector.
func kubernetesClusterName() string {
	if name := datakit.GetEnv("ENV_CLUSTER_NAME_K8S"); name != "" {
		return name
	}
	return "default"
}

func profileScheduleJitter(key string, interval time.Duration) time.Duration {
	maximum := interval / 10
	if maximum <= 0 {
		return 0
	}
	hash := fnv.New64a()
	if _, err := hash.Write([]byte(key)); err != nil {
		return 0
	}
	return time.Duration(hash.Sum64() % uint64(maximum))
}

func (m *kubernetesProfileManager) removePodTargets(podKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, target := range m.targets {
		if target.podKey != podKey {
			continue
		}
		m.retireTarget(target)
		delete(m.targets, id)
		observeKubernetesProfileDiscovery("removed")
	}
	setKubernetesProfileTargets(len(m.targets))
}

func (m *kubernetesProfileManager) markAllTargetsRemoved() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopping = true
	for id, target := range m.targets {
		m.retireTarget(target)
		delete(m.targets, id)
	}
	setKubernetesProfileTargets(0)
}

func (m *kubernetesProfileManager) retireTarget(target *kubernetesProfileTarget) {
	target.mu.Lock()
	defer target.mu.Unlock()
	target.removed = true
	if target.cancel != nil {
		target.cancel()
	}
	if collector, ok := target.collector.(*goProfileTargetCollector); ok {
		m.deltaCache.release(collector.profiler)
	}
}

func (m *kubernetesProfileManager) targetSnapshot() []*kubernetesProfileTarget {
	m.mu.RLock()
	defer m.mu.RUnlock()
	targets := make([]*kubernetesProfileTarget, 0, len(m.targets))
	for _, target := range m.targets {
		targets = append(targets, target)
	}
	return targets
}

func (m *kubernetesProfileManager) runScheduleLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			m.schedule(now)
		}
	}
}

func (m *kubernetesProfileManager) schedule(now time.Time) {
	for _, target := range m.targetSnapshot() {
		target.mu.Lock()
		if target.removed || !target.ready || now.Before(target.nextScheduled) {
			target.mu.Unlock()
			continue
		}
		target.nextScheduled = now.Add(target.rule.config.Interval)
		target.mu.Unlock()
		if !m.dispatch(target, "scheduled", target.rule.config.ScheduledTypes, target.rule.config.ProfileDuration, nil, now) {
			retry := target.rule.config.Interval / 10
			if retry > 10*time.Second {
				retry = 10 * time.Second
			}
			if retry < time.Second {
				retry = time.Second
			}
			target.mu.Lock()
			if !target.removed {
				target.nextScheduled = now.Add(retry)
			}
			target.mu.Unlock()
		}
	}
}

func (m *kubernetesProfileManager) dispatch(
	target *kubernetesProfileTarget,
	cause string,
	types []string,
	duration time.Duration,
	trigger *resourceTrigger,
	now time.Time,
) bool {
	target.mu.Lock()
	if target.removed {
		target.mu.Unlock()
		observeKubernetesProfileSkipped("removed")
		return false
	}
	if !target.ready {
		target.mu.Unlock()
		observeKubernetesProfileSkipped("not_ready")
		return false
	}
	if target.busy {
		target.mu.Unlock()
		observeKubernetesProfileSkipped("busy")
		return false
	}
	if now.Before(target.backoffUntil) {
		target.mu.Unlock()
		observeKubernetesProfileSkipped("backoff")
		return false
	}
	if cause == "resource_threshold" && now.Before(target.lastTriggered.Add(target.rule.config.Cooldown)) {
		target.mu.Unlock()
		observeKubernetesProfileSkipped("cooldown")
		return false
	}

	if !m.acquireCollection(target) {
		target.mu.Unlock()
		observeKubernetesProfileSkipped("concurrency")
		return false
	}

	target.busy = true
	if cause == "resource_threshold" {
		target.lastTriggered = now
	}
	tags := copyTags(target.tags)
	if trigger != nil {
		for key, value := range trigger.tags() {
			tags[key] = value
		}
	}
	tags["trigger_type"] = cause
	collector := target.collector
	endpoint := target.endpoint
	podKey := target.podKey
	container := target.container
	collectCtx := m.ctx
	if collectCtx == nil {
		collectCtx = context.Background()
	}
	collectCtx, cancel := context.WithCancel(collectCtx)
	target.cancel = cancel
	m.workers.Add(1)
	target.mu.Unlock()

	go func() {
		defer m.workers.Done()
		defer m.releaseCollection(target)
		defer cancel()
		started := time.Now()
		err := collectProfileSafely(collector, collectCtx, append([]string(nil), types...), duration, tags)
		elapsed := time.Since(started)

		target.mu.Lock()
		target.busy = false
		target.cancel = nil
		if err != nil {
			target.failures++
			target.backoffUntil = time.Now().Add(profileFailureBackoff(target.failures))
			if cause == "scheduled" && target.nextScheduled.After(target.backoffUntil) {
				target.nextScheduled = target.backoffUntil
			}
		} else {
			target.failures = 0
			target.backoffUntil = time.Time{}
		}
		target.mu.Unlock()

		status := "ok"
		if err != nil {
			status = "error"
			log.Warnf("collect Kubernetes profile pod=%s container=%s endpoint=%s cause=%s failed: %s",
				podKey, container, endpoint, cause, err)
		}
		observeKubernetesProfileCollection(cause, status, elapsed)
	}()
	return true
}

func (m *kubernetesProfileManager) acquireCollection(target *kubernetesProfileTarget) bool {
	m.activeMu.Lock()
	defer m.activeMu.Unlock()
	if _, exists := m.activeEndpoints[target.endpoint]; exists {
		return false
	}
	select {
	case m.semaphore <- struct{}{}:
	default:
		return false
	}
	select {
	case target.rule.semaphore <- struct{}{}:
	default:
		<-m.semaphore
		return false
	}
	m.activeEndpoints[target.endpoint] = struct{}{}
	return true
}

func (m *kubernetesProfileManager) releaseCollection(target *kubernetesProfileTarget) {
	m.activeMu.Lock()
	defer m.activeMu.Unlock()
	delete(m.activeEndpoints, target.endpoint)
	<-target.rule.semaphore
	<-m.semaphore
}

func profileFailureBackoff(failures int) time.Duration {
	if failures < 1 {
		return 0
	}
	if failures > 9 {
		failures = 9
	}
	backoff := time.Second * time.Duration(1<<(failures-1))
	if backoff > 5*time.Minute {
		return 5 * time.Minute
	}
	return backoff
}

// collectProfileSafely runs a collection and converts panics into errors so a
// faulty target cannot crash DataKit or leave the target busy forever.
func collectProfileSafely(
	collector profileTargetCollector,
	ctx context.Context,
	types []string,
	duration time.Duration,
	tags map[string]string,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			log.Errorf("panic during Kubernetes profile collection: %v\n%s", r, buf)
			err = fmt.Errorf("panic during Kubernetes profile collection: %v", r)
		}
	}()
	return collector.collect(ctx, types, duration, tags)
}

func (ipt *Input) runKubernetesProfilers(parent context.Context) error {
	nodeName, err := config.GetLocalNodeName()
	if err != nil {
		return err
	}
	client, err := k8sclient.NewKubernetesClientInCluster(k8sclient.ComponentProfile)
	if err != nil {
		return fmt.Errorf("create Kubernetes profile client: %w", err)
	}
	manager, err := newKubernetesProfileManager(
		ipt,
		client.KubernetesClientset(),
		client,
		nodeName,
		ipt.Kubernetes,
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	go func() {
		select {
		case <-ctx.Done():
		case <-datakit.Exit.Wait():
			cancel()
		case <-ipt.semStop.Wait():
			cancel()
		}
	}()
	return manager.run(ctx)
}
