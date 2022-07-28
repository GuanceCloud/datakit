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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

const (
	objectInterval = time.Minute * 5
	metricInterval = time.Second * 60
)

type Input struct {
	DepercatedEndpoint string `toml:"endpoint"`
	DockerEndpoint     string `toml:"docker_endpoint"`
	ContainerdAddress  string `toml:"containerd_address"`

	EnableContainerMetric bool `toml:"enable_container_metric"`
	EnableK8sMetric       bool `toml:"enable_k8s_metric"`
	EnablePodMetric       bool `toml:"enable_pod_metric"`

	LoggingRemoveAnsiEscapeCodes bool `toml:"logging_remove_ansi_escape_codes"`
	ExcludePauseContainer        bool `toml:"exclude_pause_container"`

	ContainerIncludeLog []string `toml:"container_include_log"`
	ContainerExcludeLog []string `toml:"container_exclude_log"`

	K8sURL               string `toml:"kubernetes_url"`
	K8sBearerToken       string `toml:"bearer_token"`
	K8sBearerTokenString string `toml:"bearer_token_string"`
	DisableK8sEvents     bool   `toml:"disable_k8s_events"`

	LoggingExtraSourceMap     map[string]string `toml:"logging_extra_source_map"`
	LoggingSourceMultilineMap map[string]string `toml:"logging_source_multiline_map"`
	Tags                      map[string]string `toml:"tags"`

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
	return []string{datakit.OSLabelLinux}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return measurements
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
		go i.watchingK8sEventLog()
	}

	g := datakit.G("kubernetes-autodiscovery")
	g.Go(func(ctx context.Context) error {
		if i.k8sInput == nil {
			l.Errorf("unrechable, not found k8s-client")
			return nil
		}
		d := newDiscovery(i.k8sInput.client, i.Tags)
		d.start()
		return nil
	})

	i.collectObject()

	for {
		select {
		case <-datakit.Exit.Wait():
			i.stop()
			l.Info("container exit success")
			return

		case <-i.semStop.Wait():
			l.Info("container exit return")
			return

		case <-metricTick.C:
			l.Debug("collect mertric")
			i.collectMetric()

		case <-objectTick.C:
			l.Debug("collect object")
			i.collectObject()

		case i.pause = <-i.chPause:
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
	l.Debug("collect object in func")
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
		l.Errorf("unrechable, k8s input is empty pointer")
		return
	}

	l.Debug("collect k8s resource object")

	if err := i.gatherK8sResourceObject(); err != nil {
		l.Errorf("failed to collect resource object: %s", err)
	}
}

func (i *Input) collectMetric() {
	l.Debug("collect mertric in func")

	if i.EnableContainerMetric {
		if err := i.gatherDockerContainerMetric(); err != nil {
			l.Errorf("failed to collect docker container metric: %s", err)
		}

		if err := i.gatherContainerdMetric(); err != nil {
			l.Errorf("failed to collect containerd metric: %s", err)
		}
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
		l.Debugf("container metric: no point")
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
	i.k8sInput.watchingEventLog(datakit.Exit.Wait())
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

		if d, err := newDockerInput(&dockerInputConfig{
			endpoint:               i.DockerEndpoint,
			excludePauseContainer:  i.ExcludePauseContainer,
			removeLoggingAnsiCodes: i.LoggingRemoveAnsiEscapeCodes,
			containerIncludeLog:    i.ContainerIncludeLog,
			containerExcludeLog:    i.ContainerExcludeLog,
			extraTags:              i.Tags,
			extraSourceMap:         i.LoggingExtraSourceMap,
			sourceMultilineMap:     i.LoggingSourceMultilineMap,
		}); err != nil {
			l.Warnf("create docker input err: %s", err)
		} else {
			i.dockerInput = d
		}

		if i.dockerInput == nil {
			if c, err := newContainerdInput(&containerdInputConfig{
				endpoint:            i.ContainerdAddress,
				extraTags:           i.Tags,
				extraSourceMap:      i.LoggingExtraSourceMap,
				sourceMultilineMap:  i.LoggingSourceMultilineMap,
				containerIncludeLog: i.ContainerIncludeLog,
				containerExcludeLog: i.ContainerExcludeLog,
			}); err != nil {
				l.Warnf("create containerd input err: %s", err)
			} else {
				i.containerdInput = c
			}
		}

		if datakit.Docker {
			if k, err := newKubernetesInput(&kubernetesInputConfig{
				url:               i.K8sURL,
				bearerToken:       i.K8sBearerToken,
				bearerTokenString: i.K8sBearerTokenString,
				extraTags:         i.Tags,
				enablePodMetric:   i.EnablePodMetric,
				enableK8sMetric:   i.EnableK8sMetric,
			}); err != nil {
				l.Errorf("create k8s input err: %s", err)
				continue
			} else {
				i.k8sInput = k
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

	if enable, ok := envs["ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_ENABLE_CONTAINER_METRIC to bool: %s, ignore", err)
		} else {
			i.EnableContainerMetric = b
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

	if exclude, ok := envs["ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER"]; ok {
		b, err := strconv.ParseBool(exclude)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CONTAINER_EXCLUDE_PAUSE_CONTAINER to bool: %s, ignore", err)
		} else {
			i.ExcludePauseContainer = b
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

//nolint:gochecknoinits
func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newInput()
	})
}
