// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package container collect container metrics/loggings/objects.
package container

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	objectInterval     = time.Minute * 5
	metricInterval     = time.Second * 60
	goroutineGroupName = "inputs_container"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	g                  = datakit.G(goroutineGroupName)
)

type Input struct {
	DepercatedEndpoint string `toml:"endpoint"`
	DockerEndpoint     string `toml:"docker_endpoint"`
	ContainerdAddress  string `toml:"containerd_address"`

	EnableContainerMetric        bool `toml:"enable_container_metric"`
	EnableK8sMetric              bool `toml:"enable_k8s_metric"`
	EnablePodMetric              bool `toml:"enable_pod_metric"`
	LoggingRemoveAnsiEscapeCodes bool `toml:"logging_remove_ansi_escape_codes"`
	LoggingBlockingMode          bool `toml:"logging_blocking_mode"`
	ExcludePauseContainer        bool `toml:"exclude_pause_container"`
	Election                     bool `toml:"election"`

	K8sURL                              string `toml:"kubernetes_url"`
	K8sBearerToken                      string `toml:"bearer_token"`
	K8sBearerTokenString                string `toml:"bearer_token_string"`
	DisableK8sEvents                    bool   `toml:"disable_k8s_events"`
	AutoDiscoveryOfK8sServicePrometheus bool   `toml:"auto_discovery_of_k8s_service_prometheus"`
	ExtractK8sLabelAsTags               bool   `toml:"extract_k8s_label_as_tags"`

	ContainerIncludeLog               []string          `toml:"container_include_log"`
	ContainerExcludeLog               []string          `toml:"container_exclude_log"`
	LoggingExtraSourceMap             map[string]string `toml:"logging_extra_source_map"`
	LoggingSourceMultilineMap         map[string]string `toml:"logging_source_multiline_map"`
	LoggingAutoMultilineDetection     bool              `toml:"logging_auto_multiline_detection"`
	LoggingAutoMultilineExtraPatterns []string          `toml:"logging_auto_multiline_extra_patterns"`

	Tags map[string]string `toml:"tags"`

	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	DepercatedConf

	semStop *cliutils.Sem // start stop signal

	dockerInput     *dockerInput
	containerdInput *containerdInput
	k8sInput        *kubernetesInput

	chPause chan bool
	pause   bool

	discovery *discovery
}

func (i *Input) Singleton() {
}

var (
	l = logger.DefaultSLogger(inputName)

	maxPauseCh = inputs.ElectionPauseChannelLength
)

func newInput() *Input {
	return &Input{
		DockerEndpoint:            dockerEndpoint,
		ContainerdAddress:         containerdAddress,
		Tags:                      make(map[string]string),
		LoggingExtraSourceMap:     make(map[string]string),
		LoggingSourceMultilineMap: make(map[string]string),
		Election:                  true,
		chPause:                   make(chan bool, maxPauseCh),
		semStop:                   cliutils.NewSem(),
	}
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catelog }

func (*Input) PipelineConfig() map[string]string { return nil }

func (*Input) GetPipeline() []*tailer.Option { return nil }

func (*Input) RunPipeline() { /*nil*/ }

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s, datakit.LabelDocker}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return measurements
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("container input startd")

	if i.setup() {
		return
	}

	l.Debugf("container input, dockerInput: %#v, containerdInput: %#v", i.dockerInput, i.containerdInput)

	objectTick := time.NewTicker(objectInterval)
	defer objectTick.Stop()
	metricTick := time.NewTicker(metricInterval)
	defer metricTick.Stop()

	if datakit.Docker && !i.DisableK8sEvents {
		g.Go(func(ctx context.Context) error {
			i.watchingK8sEventLog()
			return nil
		})
	}

	if datakit.Docker {
		g := datakit.G("kubernetes-autodiscovery")
		g.Go(func(ctx context.Context) error {
			if i.discovery == nil {
				l.Warnf("unrechable, discovery not initialized")
				return nil
			}
			i.discovery.start()
			return nil
		})
	}

	i.collectObject()
	i.collectMetric()

	for {
		select {
		case <-datakit.Exit.Wait():
			i.stop()
			l.Info("container exit")
			return

		case <-i.semStop.Wait():
			i.stop()
			l.Info("container terminate")
			return

		case <-metricTick.C:
			l.Debug("collect mertric")
			i.collectMetric()

		case <-objectTick.C:
			l.Debug("collect object")
			i.collectObject()

		case i.pause = <-i.chPause:
			if i.discovery != nil {
				i.discovery.chPause <- i.pause
			}
			globalPause.set(i.pause)
		}
	}
}

func (i *Input) stop() {
	if i.dockerInput != nil {
		i.dockerInput.stop()
	}
	if i.containerdInput != nil {
		i.containerdInput.stop()
	}
}

func (i *Input) collectObject() {
	timeNow := time.Now()
	defer func() {
		l.Debugf("collect object, cost %s", time.Since(timeNow))
	}()

	if err := i.gatherDockerContainerObject(); err != nil {
		l.Errorf("failed to collect docker container object: %s", err)
	}

	if err := i.gatherContainerdObject(); err != nil {
		l.Errorf("failed to collect containerd object: %s", err)
	}

	if !datakit.Docker {
		return
	}
	if i.pause {
		l.Debug("not leader, skipped")
		return
	}

	if i.k8sInput == nil {
		l.Error("unrechable, k8s input is empty pointer")
		return
	}

	l.Debug("collect k8s resource object")

	if err := i.gatherK8sResourceObject(); err != nil {
		l.Errorf("failed to collect resource object: %s", err)
	}
}

func (i *Input) collectMetric() {
	timeNow := time.Now()
	defer func() {
		l.Debugf("collect metric and logging, cost %s", time.Since(timeNow))
	}()

	if i.EnableContainerMetric {
		if err := i.gatherDockerContainerMetric(); err != nil {
			l.Errorf("failed to collect docker container metric: %s", err)
		}

		if err := i.gatherContainerdMetric(); err != nil {
			l.Errorf("failed to collect containerd metric: %s", err)
		}
	}

	if i.discovery != nil {
		i.discovery.updateGlobalCRDLogsConfList()
	}

	if err := i.watchNewDockerContainerLogs(); err != nil {
		l.Errorf("failed to watch container log: %s", err)
	}

	if err := i.watchNewContainerdLogs(); err != nil {
		l.Errorf("failed to watch containerd log: %s", err)
	}

	if !datakit.Docker {
		return
	}

	if i.pause {
		l.Debug("not leader, skipped")
		return
	}

	if i.k8sInput == nil {
		l.Errorf("unrechable, k8s input is empty pointer")
		return
	}

	l.Debug("collect k8s-pod metric")

	if err := i.gatherK8sResourceMetric(); err != nil {
		l.Errorf("failed to collect resource metric: %s", err)
	}
}

func (i *Input) gatherDockerContainerMetric() error {
	if i.dockerInput == nil {
		return nil
	}
	start := time.Now()

	res, err := i.dockerInput.gatherMetric()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		l.Debug("container metric: no point")
		return nil
	}

	return inputs.FeedMeasurement("container-metric", datakit.Metric, res,
		&io.Option{CollectCost: time.Since(start)})
}

func (i *Input) gatherDockerContainerObject() error {
	if i.dockerInput == nil {
		return nil
	}
	start := time.Now()

	res, err := i.dockerInput.gatherObject()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		l.Debugf("container object: no point")
		return nil
	}

	return inputs.FeedMeasurement("container-object", datakit.Object, res,
		&io.Option{CollectCost: time.Since(start)})
}

func (i *Input) gatherContainerdMetric() error {
	if i.containerdInput == nil {
		return nil
	}
	start := time.Now()

	res, err := i.containerdInput.gatherMetric()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		l.Debugf("containerd metric: no point")
		return nil
	}

	return inputs.FeedMeasurement("containerd-metric", datakit.Metric, res,
		&io.Option{CollectCost: time.Since(start)})
}

func (i *Input) gatherContainerdObject() error {
	if i.containerdInput == nil {
		return nil
	}
	start := time.Now()

	res, err := i.containerdInput.gatherObject()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		l.Debugf("containerd object: no point")
		return nil
	}

	return inputs.FeedMeasurement("containerd-object", datakit.Object, res,
		&io.Option{CollectCost: time.Since(start)})
}

func (i *Input) watchNewContainerdLogs() error {
	if i.containerdInput == nil {
		return nil
	}
	return i.containerdInput.watchNewLogs()
}

func (i *Input) gatherK8sResourceMetric() error {
	start := time.Now()

	metricMeas, err := i.k8sInput.gatherResourceMetric()
	if err != nil {
		return err
	}

	return inputs.FeedMeasurement("k8s-metric", datakit.Metric, metricMeas,
		&io.Option{CollectCost: time.Since(start)})
}

func (i *Input) gatherK8sResourceObject() error {
	start := time.Now()

	objectMeas, err := i.k8sInput.gatherResourceObject()
	if err != nil {
		return err
	}

	return inputs.FeedMeasurement("k8s-object", datakit.Object, objectMeas,
		&io.Option{CollectCost: time.Since(start)})
}

func (i *Input) watchNewDockerContainerLogs() error {
	if i.dockerInput == nil {
		return nil
	}
	return i.dockerInput.watchNewLogs()
}

func (i *Input) watchingK8sEventLog() {
	i.k8sInput.watchingEventLog(i.semStop.Wait())
}

func (i *Input) setup() bool {
	if i.DepercatedEndpoint != "" && i.DepercatedEndpoint != i.DockerEndpoint {
		i.DockerEndpoint = i.DepercatedEndpoint
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		time.Sleep(time.Second)

		if d, err := newDockerInput(i); err != nil {
			l.Warnf("create docker input err: %s", err)
		} else {
			i.dockerInput = d
		}

		if i.dockerInput == nil {
			if c, err := newContainerdInput(i); err != nil {
				l.Warnf("create containerd input err: %s", err)
			} else {
				i.containerdInput = c
			}
		}

		if datakit.Docker {
			if k, err := newKubernetesInput(i); err != nil {
				l.Errorf("create k8s input err: %s", err)
				continue
			} else {
				i.k8sInput = k

				i.discovery = newDiscovery(i.k8sInput.client, i.semStop.Wait())
				i.discovery.extraTags = i.Tags
				i.discovery.extractK8sLabelAsTags = i.ExtractK8sLabelAsTags
				i.discovery.autoDiscoveryOfK8sServicePrometheus = i.AutoDiscoveryOfK8sServicePrometheus

				if i.dockerInput != nil {
					i.dockerInput.k8sClient = i.k8sInput.client
				}
				if i.containerdInput != nil {
					i.containerdInput.k8sClient = i.k8sInput.client
				}
			}
		}

		break
	}

	return false
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case i.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

// ReadEnv , support envs：
//   ENV_INPUT_CONTAINER_DOCKER_ENDPOINT : string
//   ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS : string
//   ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES : booler
//   ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC : booler
//   ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC : booler
//   ENV_INPUT_CONTAINER_ENABLE_POD_METRIC : booler
//   ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_K8S_SERVICE_PROMETHEUS: booler
//   ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS: booler
//   ENV_INPUT_CONTAINER_TAGS : "a=b,c=d"
//   ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER : booler
//   ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG : []string
//   ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG : []string
//   ENV_INPUT_CONTAINER_MAX_LOGGING_LENGTH : int
//   ENV_INPUT_CONTAINER_KUBERNETES_URL : string
//   ENV_INPUT_CONTAINER_BEARER_TOKEN : string
//   ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING : string
//   ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP : string
//   ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON : string (JSON map)
//   ENV_INPUT_CONTAINER_LOGGING_BLOCKING_MODE : booler
//   ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION: booler
//   ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON : string (JSON string array)
func (i *Input) ReadEnv(envs map[string]string) {
	if endpoint, ok := envs["ENV_INPUT_CONTAINER_DOCKER_ENDPOINT"]; ok {
		i.DockerEndpoint = endpoint
	}

	if address, ok := envs["ENV_INPUT_CONTAINER_CONTAINERD_ADDRESS"]; ok {
		i.ContainerdAddress = address
	}

	if v, ok := envs["ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP"]; ok {
		i.LoggingExtraSourceMap = config.ParseGlobalTags(v)
	}

	if v, ok := envs["ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON"]; ok {
		if err := json.Unmarshal([]byte(v), &i.LoggingSourceMultilineMap); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON to map: %s, ignore", err)
		}
	}

	if remove, ok := envs["ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES"]; ok {
		b, err := strconv.ParseBool(remove)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES to bool: %s, ignore", err)
		} else {
			i.LoggingRemoveAnsiEscapeCodes = b
		}
	}

	if disable, ok := envs["ENV_INPUT_CONTAINER_LOGGING_BLOCKING_MODE"]; ok {
		b, err := strconv.ParseBool(disable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_BLOCKING_MODE to bool: %s, ignore", err)
		} else {
			i.LoggingBlockingMode = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC to bool: %s, ignore", err)
		} else {
			i.EnableContainerMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS to bool: %s, ignore", err)
		} else {
			i.ExtractK8sLabelAsTags = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_K8S_METRIC to bool: %s, ignore", err)
		} else {
			i.EnableK8sMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_POD_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_POD_METRIC to bool: %s, ignore", err)
		} else {
			i.EnablePodMetric = b
		}
	}

	if enable, ok := envs["ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_K8S_SERVICE_PROMETHEUS"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_AUTO_DISCOVERY_OF_K8S_SERVICE_PROMETHEUS to bool: %s, ignore", err)
		} else {
			i.AutoDiscoveryOfK8sServicePrometheus = b
		}
	}

	if exclude, ok := envs["ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER"]; ok {
		b, err := strconv.ParseBool(exclude)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER to bool: %s, ignore", err)
		} else {
			i.ExcludePauseContainer = b
		}
	}

	if open, ok := envs["ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_DETECTION to bool: %s, ignore", err)
		} else {
			i.LoggingAutoMultilineDetection = b
		}
	}

	if v, ok := envs["ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON"]; ok {
		if err := json.Unmarshal([]byte(v), &i.LoggingAutoMultilineExtraPatterns); err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_LOGGING_AUTO_MULTILINE_EXTRA_PATTERNS_JSON to map: %s, ignore", err)
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_CONTAINER_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			i.Tags[k] = v
		}
	}

	//   ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG : []string
	//   ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG : []string

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add CONTAINER_INCLUDE_LOG from ENV: %v", arrays)
		i.ContainerIncludeLog = append(i.ContainerIncludeLog, arrays...)
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add CONTAINER_EXCLUDE_LOG from ENV: %v", arrays)
		i.ContainerExcludeLog = append(i.ContainerExcludeLog, arrays...)
	}

	//   ENV_INPUT_CONTAINER_MAX_LOGGING_LENGTH : int
	//   ENV_INPUT_CONTAINER_KUBERNETES_URL : string
	//   ENV_INPUT_CONTAINER_BEARER_TOKEN : string
	//   ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING : string

	if str, ok := envs["ENV_INPUT_CONTAINER_MAX_LOGGING_LENGTH"]; ok {
		n, err := strconv.Atoi(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_MAX_LOGGING_LENGTH to int: %s, ignore", err)
		} else {
			i.MaxLoggingLength = n
		}
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_KUBERNETES_URL"]; ok {
		i.K8sURL = str
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN"]; ok {
		i.K8sBearerToken = str
	}

	if str, ok := envs["ENV_INPUT_CONTAINER_BEARER_TOKEN_STRING"]; ok {
		i.K8sBearerTokenString = str
	}
}

func (i *Input) getAutoMultilinePatterns() []string {
	if !i.LoggingAutoMultilineDetection {
		return nil
	}
	if len(i.LoggingAutoMultilineExtraPatterns) != 0 {
		return i.LoggingAutoMultilineExtraPatterns
	}
	return multiline.GlobalPatterns
}

//nolint:gochecknoinits
func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
