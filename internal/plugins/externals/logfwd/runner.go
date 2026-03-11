// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	g           = goroutine.NewGroup(goroutine.Option{Name: "logfwd"})
	log         = logger.DefaultSLogger("logfwd")
	loggerLevel = os.Getenv("LOGFWD_LOG_LEVEL")
)

func initLogger() {
	lopt := &logger.Option{
		Level: "info",
		Flags: (logger.OPT_DEFAULT | logger.OPT_STDOUT),
	}

	if loggerLevel == "debug" {
		lopt.Level = "debug"
	}

	if err := logger.InitRoot(lopt); err != nil {
		return
	}

	log = logger.SLogger("logfwd")
}

func StartLogForwarding(ctx context.Context) error {
	initLogger()

	datakitEndpoint, operatorURL, err := getEndpointConfig()
	if err != nil {
		return fmt.Errorf("failed to get endpoint config: %w", err)
	}

	log.Infof("logfwd endpoints: datakit=%s operator=%s",
		datakitEndpoint, operatorURL)

	runner := newRunner(datakitEndpoint, operatorURL)

	// Read and parse pod labels from /etc/podinfo/labels
	podLabels, podLabelsStr, err := readPodLabels("/etc/podinfo/labels")
	if err != nil {
		log.Warnf("failed to read pod labels: %v", err)
		podLabels = make(map[string]string)
		podLabelsStr = ""
	} else {
		log.Infof("loaded %d pod labels", len(podLabels))
	}
	runner.podLabels = podLabels
	runner.podLabelsStr = podLabelsStr

	if err := runner.Start(); err != nil {
		return fmt.Errorf("failed to start runner: %w", err)
	}

	<-ctx.Done()

	if err := runner.Stop(); err != nil {
		log.Errorf("error stopping runner: %v", err)
		return err
	}

	log.Info("logfwd stopped")
	return nil
}

type runner struct {
	datakitEndpoint   string
	envTailers        []*tailer.Tailer
	operatorTailers   []*tailer.Tailer
	operatorConfigMD5 string
	wsClient          *websocketClient
	operatorClient    *operatorClient
	podLabels         map[string]string
	podLabelsStr      string
	mutex             sync.RWMutex
	stopChan          chan struct{}
}

func newRunner(datakitEndpoint, operatorURL string) *runner {
	var opClient *operatorClient
	if operatorURL != "" {
		opClient = newOperatorClient(operatorURL)
	}

	return &runner{
		datakitEndpoint: datakitEndpoint,
		envTailers:      make([]*tailer.Tailer, 0),
		operatorTailers: make([]*tailer.Tailer, 0),
		operatorClient:  opClient,
		stopChan:        make(chan struct{}),
	}
}

func (r *runner) Start() error {
	wsClient, err := createWebsocketClient(r.datakitEndpoint)
	if err != nil {
		return fmt.Errorf("failed to create websocket client: %w", err)
	}
	r.wsClient = wsClient

	if err := r.handleEnvConfigs(); err != nil {
		log.Errorf("failed to handle env configs: %v", err)
	}

	r.startOperatorWatcher()

	log.Info("logfwd runner started")
	return nil
}

func (r *runner) Stop() error {
	close(r.stopChan)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, t := range r.envTailers {
		t.Close()
	}
	for _, t := range r.operatorTailers {
		t.Close()
	}

	r.envTailers = nil
	r.operatorTailers = nil

	if r.wsClient != nil {
		if err := r.wsClient.close(); err != nil {
			log.Errorf("error closing websocket client: %v", err)
			return err
		}
	}

	log.Info("logfwd runner stopped")
	return nil
}

func (r *runner) handleEnvConfigs() error {
	if envLogConfigsStr == "" {
		return nil
	}

	log.Info("processing environment log configs")
	configs, err := parseLogConfigs(envLogConfigsStr)
	if err != nil {
		return fmt.Errorf("failed to parse env log configs: %w", err)
	}

	return r.updateEnvTailers(configs)
}

func (r *runner) createTailers(configs []*logConfig, name string) ([]*tailer.Tailer, error) {
	if len(configs) == 0 {
		log.Debugf("no %s configs to create tailers for", name)
		return []*tailer.Tailer{}, nil
	}

	tailers := make([]*tailer.Tailer, 0, len(configs))
	var errors []string

	for _, cfg := range configs {
		if cfg.Disable || cfg.Path == "" {
			continue
		}

		log.Infof("creating tailer for path: %s", cfg.Path)

		t, err := r.createAndStartTailer(cfg)
		if err != nil {
			errMsg := fmt.Sprintf("failed to create tailer for path %s: %v", cfg.Path, err)
			log.Warnf(errMsg)
			errors = append(errors, errMsg)
			continue
		}
		tailers = append(tailers, t)
	}

	// If all tailers failed to create, return an error
	if len(tailers) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to create any %s tailers: %v", name, errors)
	}

	log.Infof("created %d %s tailers", len(tailers), name)
	return tailers, nil
}

func (r *runner) updateEnvTailers(configs []*logConfig) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Close existing tailers
	for _, t := range r.envTailers {
		t.Close()
	}

	// Create new tailers
	newTailers, err := r.createTailers(configs, "env")
	if err != nil {
		return err
	}

	r.envTailers = newTailers
	log.Info("updated environment tailers")
	return nil
}

func (r *runner) updateOperatorTailers(configs []*logConfig) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Close existing tailers
	for _, t := range r.operatorTailers {
		t.Close()
	}

	// Create new tailers
	newTailers, err := r.createTailers(configs, "operator")
	if err != nil {
		return err
	}

	r.operatorTailers = newTailers
	log.Info("updated operator tailers")
	return nil
}

func (r *runner) handleOperatorConfigs() {
	if r.operatorClient == nil {
		return
	}

	log.Debugf("fetching operator configs for pod: namespace=%s, name=%s, labels=%s",
		podNamespace, podName, r.podLabelsStr)

	apiResp, err := r.operatorClient.fetchOperatorResponse(podNamespace, podName, r.podLabelsStr)
	if err != nil {
		log.Errorf("failed to fetch operator configs for pod %s/%s: %v", podNamespace, podName, err)
		return
	}

	// Parse configs from response
	var loggingConfigs []loggingConfig
	if apiResp.Configs != "" {
		if err := json.Unmarshal([]byte(apiResp.Configs), &loggingConfigs); err != nil {
			log.Errorf("failed to unmarshal configs: %v", err)
			return
		}
	}

	configName := "unknown"
	if apiResp.Name != "" {
		configName = apiResp.Name
	}

	// Calculate MD5 of new configs (marshal parsed configs to ensure consistent hash)
	configData, err := json.Marshal(loggingConfigs)
	if err != nil {
		log.Errorf("failed to marshal operator configs for MD5: %v", err)
		return
	}
	h := fnv.New32a()
	_, _ = h.Write(configData)
	newMD5 := fmt.Sprintf("%x", h.Sum(nil))

	// Check if configs have changed
	r.mutex.RLock()
	currentMD5 := r.operatorConfigMD5
	r.mutex.RUnlock()

	if newMD5 == currentMD5 {
		log.Debugf("operator configs unchanged for pod %s/%s, configName=%s, MD5: %s", podNamespace, podName, configName, currentMD5)
		return
	}

	log.Infof("operator configs changed for pod %s/%s, configName=%s, updating tailers", podNamespace, podName, configName)

	// Process configs
	configs := make([]*logConfig, len(loggingConfigs))
	for i, lc := range loggingConfigs {
		cfg := lc
		processConfig(&cfg)
		setTagsFromPodLabels(&cfg, r.podLabels, apiResp.PodTargetLabels)
		configs[i] = &cfg
	}

	// Update tailers with new configs
	if err := r.updateOperatorTailers(configs); err != nil {
		log.Errorf("failed to update operator tailers: %v", err)
		return
	}

	// Update stored MD5
	r.mutex.Lock()
	r.operatorConfigMD5 = newMD5
	r.mutex.Unlock()

	log.Infof("operator configs updated successfully for pod %s/%s, configName=%s, %d configs processed",
		podNamespace, podName, configName, len(configs))
}

func (r *runner) startOperatorWatcher() {
	if r.operatorClient == nil {
		return
	}

	watcherGo := goroutine.NewGroup(goroutine.Option{Name: "operator-watcher"})
	watcherGo.Go(func(ctx context.Context) error {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.handleOperatorConfigs()
			case <-r.stopChan:
				return nil
			}
		}
	})
}

func (r *runner) createAndStartTailer(cfg *logConfig) (*tailer.Tailer, error) {
	fn := forwardFunc(cfg, r.wsClient.writeMessage)
	opts := buildTailerOptions(cfg, fn)

	t, err := tailer.NewTailer([]string{cfg.Path}, opts...)
	if err != nil {
		log.Errorf("failed to create tailer for path %s: %v", cfg.Path, err)
		return nil, err
	}

	g.Go(func(ctx context.Context) error {
		t.Start()
		return nil
	})

	return t, nil
}

func createWebsocketClient(endpoint string) (*websocketClient, error) {
	wsURL := fmt.Sprintf("ws://%s/logfwd", endpoint)
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid websocket URL: %w", err)
	}

	wsClient := newWebsocketClient(u)
	wsClient.tryConnectWebsocketServer()

	wsGo := goroutine.NewGroup(goroutine.Option{Name: "websocket"})
	wsGo.Go(func(ctx context.Context) error {
		wsClient.start()
		return nil
	})

	return wsClient, nil
}
