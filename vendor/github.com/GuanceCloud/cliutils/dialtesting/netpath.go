// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

const (
	netPathMetricName               = "netpath_dial_testing"
	defaultNetPathTimeout           = time.Second
	maxNetPathTimeout               = 30 * time.Second
	defaultNetPathTTL               = 30
	maxNetPathTTL                   = 60
	maxUDPNetPathTTL                = 255
	defaultNetPathE2EQueries        = 10
	maxNetPathE2EQueries            = 100
	defaultNetPathTracerouteQueries = 3
	maxNetPathTracerouteQueries     = 10
)

// NetPathAdvanceOptions contains optional NetPath probe settings.
type NetPathAdvanceOptions struct {
	Timeout           string `json:"timeout"`
	SourceName        string `json:"source_name,omitempty"`
	TargetName        string `json:"target_name,omitempty"`
	MaxTTL            int    `json:"max_ttl,omitempty"`
	E2EQueries        int    `json:"e2e_queries,omitempty"`
	TracerouteQueries int    `json:"traceroute_queries,omitempty"`
}

// NetPathCondition describes one result assertion in a task definition.
type NetPathCondition struct {
	Op     string          `json:"op"`
	Target json.RawMessage `json:"target"`
}

// NetPathSuccess contains the assertions supported by a NetPath task.
type NetPathSuccess struct {
	E2ERTTAvg           []*NetPathCondition `json:"e2e_rtt_avg,omitempty"`
	E2ERTTMin           []*NetPathCondition `json:"e2e_rtt_min,omitempty"`
	E2ERTTMax           []*NetPathCondition `json:"e2e_rtt_max,omitempty"`
	E2ERTTVariationAvg  []*NetPathCondition `json:"e2e_rtt_variation_avg,omitempty"`
	E2ERTTVariationMax  []*NetPathCondition `json:"e2e_rtt_variation_max,omitempty"`
	E2EProbeLossPercent []*NetPathCondition `json:"e2e_probe_loss_percent,omitempty"`
	HopCount            []*NetPathCondition `json:"hop_count,omitempty"`
	E2EStatus           []*NetPathCondition `json:"e2e_status,omitempty"`
	TracerouteStatus    []*NetPathCondition `json:"traceroute_status,omitempty"`
}

// NetPathProbeConfig is the normalized task configuration passed to the executor.
type NetPathProbeConfig struct {
	ID                string
	Name              string
	Host              string
	Port              uint16
	Protocol          string
	Timeout           time.Duration
	MaxTTL            int
	TracerouteQueries int
	E2EQueries        int
	ScheduledAt       time.Time
}

// NetPathExecutor lets the caller own probe execution and result processing.
type NetPathExecutor interface {
	Run(context.Context, NetPathProbeConfig) error
	Clear()
	CheckResult() ([]string, bool)
	GetResults() (map[string]string, map[string]any)
	SetError(string)
}

// NetPathTask implements the common dialtesting task lifecycle for NetPath.
type NetPathTask struct {
	*Task

	Protocol         string                `json:"protocol"`
	Host             string                `json:"host"`
	Port             string                `json:"port,omitempty"`
	AdvanceOptions   NetPathAdvanceOptions `json:"advance_options"`
	SuccessWhen      []*NetPathSuccess     `json:"success_when"`
	SuccessWhenLogic string                `json:"success_when_logic,omitempty"`

	rawTask  *NetPathTask
	executor NetPathExecutor

	mu     sync.Mutex
	cancel context.CancelFunc
}

var (
	_ TaskChild = (*NetPathTask)(nil)
	_ ITask     = (*NetPathTask)(nil)
)

// SetExecutor sets the caller-owned executor used when the task runs.
func (t *NetPathTask) SetExecutor(executor NetPathExecutor) {
	t.mu.Lock()
	t.executor = executor
	t.mu.Unlock()
}

func (t *NetPathTask) init() error {
	if strings.EqualFold(t.CurStatus, StatusStop) {
		return nil
	}

	// CheckTask runs before variables are rendered by the debug API. Defer
	// value validation until RenderTemplateAndInit replaces every template.
	if t.rawTask == nil && t.hasUnrenderedTemplate() {
		return nil
	}

	if err := t.check(); err != nil {
		return err
	}

	config, err := t.probeConfig(time.Time{})
	if err != nil {
		return err
	}
	if err := validateNetPathProbeConfig(config); err != nil {
		return err
	}
	return t.validateAssertions()
}

func (t *NetPathTask) check() error {
	if t.rawTask == nil && t.hasUnrenderedTemplate() {
		return nil
	}

	if strings.TrimSpace(t.Host) == "" {
		return errors.New("host should not be empty")
	}

	protocol := strings.ToLower(strings.TrimSpace(t.Protocol))
	switch protocol {
	case "tcp", "udp", "icmp":
	default:
		return fmt.Errorf("unsupported protocol %q", t.Protocol)
	}

	if ip := net.ParseIP(strings.TrimSpace(t.Host)); ip != nil && ip.To4() == nil {
		return errors.New("IPv6 host is not supported")
	}

	if protocol == "icmp" {
		if strings.TrimSpace(t.Port) != "" {
			return errors.New("port must be omitted for ICMP")
		}
		return nil
	}

	port, err := strconv.ParseUint(strings.TrimSpace(t.Port), 10, 16)
	if err != nil || port == 0 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

func (t *NetPathTask) hasUnrenderedTemplate() bool {
	data, err := json.Marshal(struct {
		Protocol         string                `json:"protocol"`
		Host             string                `json:"host"`
		Port             string                `json:"port"`
		AdvanceOptions   NetPathAdvanceOptions `json:"advance_options"`
		SuccessWhen      []*NetPathSuccess     `json:"success_when"`
		SuccessWhenLogic string                `json:"success_when_logic"`
	}{
		Protocol:         t.Protocol,
		Host:             t.Host,
		Port:             t.Port,
		AdvanceOptions:   t.AdvanceOptions,
		SuccessWhen:      t.SuccessWhen,
		SuccessWhenLogic: t.SuccessWhenLogic,
	})
	return err == nil && hasTemplateTags(string(data))
}

func (t *NetPathTask) run() error {
	executor := t.getExecutor()
	if executor == nil {
		return errors.New("netpath executor is not configured")
	}

	config, err := t.probeConfig(time.Now())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.mu.Unlock()
	defer func() {
		cancel()
		t.mu.Lock()
		t.cancel = nil
		t.mu.Unlock()
	}()

	if err := executor.Run(ctx, config); err != nil {
		executor.SetError(err.Error())
	}
	return nil
}

func (t *NetPathTask) stop() {
	t.mu.Lock()
	cancel := t.cancel
	t.cancel = nil
	t.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (t *NetPathTask) clear() {
	if executor := t.getExecutor(); executor != nil {
		executor.Clear()
	}
}

func (t *NetPathTask) checkResult() ([]string, bool) {
	executor := t.getExecutor()
	if executor == nil {
		return []string{"netpath executor is not configured"}, false
	}
	reasons, success := executor.CheckResult()
	return append([]string(nil), reasons...), success
}

func (t *NetPathTask) getResults() (map[string]string, map[string]any) {
	executor := t.getExecutor()
	if executor == nil {
		return map[string]string{}, map[string]any{}
	}
	tags, fields := executor.GetResults()
	return maps.Clone(tags), maps.Clone(fields)
}

func (t *NetPathTask) getExecutor() NetPathExecutor {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.executor
}

func (t *NetPathTask) getVariableValue(Variable) (string, error) {
	return "", errors.New("NETPATH task does not extract variables")
}

func (t *NetPathTask) class() string {
	return ClassNetPath
}

func (t *NetPathTask) metricName() string {
	return netPathMetricName
}

func (t *NetPathTask) getHostName() ([]string, error) {
	host := strings.TrimSpace(t.Host)
	if host == "" {
		return nil, errors.New("host should not be empty")
	}
	return []string{host}, nil
}

func (t *NetPathTask) getRawTask(taskString string) (string, error) {
	task := NetPathTask{}
	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal NETPATH task failed: %w", err)
	}
	task.Task = nil

	data, err := json.Marshal(&task)
	if err != nil {
		return "", fmt.Errorf("marshal NETPATH task failed: %w", err)
	}
	return string(data), nil
}

func (t *NetPathTask) renderTemplate(fm template.FuncMap) error {
	if t.rawTask == nil {
		rawTask := &NetPathTask{}
		if err := t.NewRawTask(rawTask); err != nil {
			return fmt.Errorf("new raw NETPATH task failed: %w", err)
		}
		rawTask.Task = nil
		t.rawTask = rawTask
	}

	rawTask := t.rawTask
	if rawTask == nil {
		return errors.New("raw NETPATH task is nil")
	}

	var err error
	if t.Protocol, err = t.renderNetPathString("protocol", rawTask.Protocol, fm); err != nil {
		return err
	}
	if t.Host, err = t.renderNetPathString("host", rawTask.Host, fm); err != nil {
		return err
	}
	if t.Port, err = t.renderNetPathString("port", rawTask.Port, fm); err != nil {
		return err
	}
	if t.SuccessWhenLogic, err = t.renderNetPathString(
		"success_when_logic",
		rawTask.SuccessWhenLogic,
		fm,
	); err != nil {
		return err
	}

	t.AdvanceOptions = rawTask.AdvanceOptions
	if t.AdvanceOptions.Timeout, err = t.renderNetPathString(
		"advance_options.timeout",
		rawTask.AdvanceOptions.Timeout,
		fm,
	); err != nil {
		return err
	}
	if t.AdvanceOptions.SourceName, err = t.renderNetPathString(
		"advance_options.source_name",
		rawTask.AdvanceOptions.SourceName,
		fm,
	); err != nil {
		return err
	}
	if t.AdvanceOptions.TargetName, err = t.renderNetPathString(
		"advance_options.target_name",
		rawTask.AdvanceOptions.TargetName,
		fm,
	); err != nil {
		return err
	}

	t.SuccessWhen, err = cloneNetPathSuccessWhen(rawTask.SuccessWhen)
	if err != nil {
		return err
	}
	if err := t.renderNetPathSuccessWhen(fm); err != nil {
		return err
	}
	return nil
}

func (t *NetPathTask) renderNetPathString(
	field string,
	value string,
	fm template.FuncMap,
) (string, error) {
	rendered, err := t.GetParsedString(value, fm)
	if err != nil {
		return "", fmt.Errorf("render %s failed: %w", field, err)
	}
	return rendered, nil
}

func cloneNetPathSuccessWhen(successWhen []*NetPathSuccess) ([]*NetPathSuccess, error) {
	data, err := json.Marshal(successWhen)
	if err != nil {
		return nil, fmt.Errorf("marshal raw NETPATH success_when failed: %w", err)
	}
	var cloned []*NetPathSuccess
	if err := json.Unmarshal(data, &cloned); err != nil {
		return nil, fmt.Errorf("unmarshal raw NETPATH success_when failed: %w", err)
	}
	return cloned, nil
}

func (t *NetPathTask) renderNetPathSuccessWhen(fm template.FuncMap) error {
	for _, success := range t.SuccessWhen {
		if success == nil {
			continue
		}
		for _, assertions := range success.assertions() {
			for _, condition := range assertions.conditions {
				if condition == nil {
					continue
				}
				op, err := t.renderNetPathString(assertions.field+" operator", condition.Op, fm)
				if err != nil {
					return err
				}
				condition.Op = op
				if err := t.renderNetPathConditionTarget(assertions.field, condition, fm); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *NetPathTask) renderNetPathConditionTarget(
	field string,
	condition *NetPathCondition,
	fm template.FuncMap,
) error {
	var target string
	if err := json.Unmarshal(condition.Target, &target); err != nil {
		return nil
	}

	rendered, err := t.renderNetPathString(field+" target", target, fm)
	if err != nil {
		return err
	}

	// Config variables are strings. For numeric assertions, convert a
	// templated string such as "{{loss}}" back to a JSON number after
	// rendering; ordinary string targets remain strings and validation
	// reports their type mismatch as before.
	if !isNetPathStatusField(field) &&
		!isNetPathDurationField(field) &&
		hasTemplateTags(target) {
		var number *float64
		if err := json.Unmarshal([]byte(rendered), &number); err == nil && number != nil {
			condition.Target = json.RawMessage(rendered)
			return nil
		}
	}

	data, err := json.Marshal(rendered)
	if err != nil {
		return fmt.Errorf("marshal rendered %s target failed: %w", field, err)
	}
	condition.Target = data
	return nil
}

func (t *NetPathTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
}

func (t *NetPathTask) setReqError(message string) {
	if executor := t.getExecutor(); executor != nil {
		executor.SetError(message)
	}
}

func (t *NetPathTask) probeConfig(scheduledAt time.Time) (NetPathProbeConfig, error) {
	port := uint64(0)
	var err error
	if strings.TrimSpace(t.Port) != "" {
		port, err = strconv.ParseUint(strings.TrimSpace(t.Port), 10, 16)
		if err != nil || port == 0 {
			return NetPathProbeConfig{}, errors.New("port must be between 1 and 65535")
		}
	}

	timeout := defaultNetPathTimeout
	if value := strings.TrimSpace(t.AdvanceOptions.Timeout); value != "" {
		timeout, err = time.ParseDuration(value)
		if err != nil {
			return NetPathProbeConfig{}, fmt.Errorf("invalid timeout: %w", err)
		}
	}
	maxTTL := t.AdvanceOptions.MaxTTL
	if maxTTL == 0 {
		maxTTL = defaultNetPathTTL
	}
	e2eQueries := t.AdvanceOptions.E2EQueries
	if e2eQueries == 0 {
		e2eQueries = defaultNetPathE2EQueries
	}
	tracerouteQueries := t.AdvanceOptions.TracerouteQueries
	if tracerouteQueries == 0 {
		tracerouteQueries = defaultNetPathTracerouteQueries
	}

	return NetPathProbeConfig{
		ID:                t.GetExternalID(),
		Name:              t.Name,
		Host:              strings.TrimSpace(t.Host),
		Port:              uint16(port),
		Protocol:          strings.ToLower(strings.TrimSpace(t.Protocol)),
		Timeout:           timeout,
		MaxTTL:            maxTTL,
		TracerouteQueries: tracerouteQueries,
		E2EQueries:        e2eQueries,
		ScheduledAt:       scheduledAt,
	}, nil
}

func validateNetPathProbeConfig(config NetPathProbeConfig) error {
	if config.Timeout < time.Second || config.Timeout > maxNetPathTimeout {
		return fmt.Errorf("timeout must be between 1s and %s", maxNetPathTimeout)
	}

	ttlLimit := maxNetPathTTL
	if config.Protocol == "udp" {
		ttlLimit = maxUDPNetPathTTL
	}
	if config.MaxTTL < 1 || config.MaxTTL > ttlLimit {
		return fmt.Errorf("max_ttl must be between 1 and %d", ttlLimit)
	}
	if config.TracerouteQueries < 1 || config.TracerouteQueries > maxNetPathTracerouteQueries {
		return fmt.Errorf("traceroute_queries must be between 1 and %d", maxNetPathTracerouteQueries)
	}
	if config.E2EQueries < 1 || config.E2EQueries > maxNetPathE2EQueries {
		return fmt.Errorf("e2e_queries must be between 1 and %d", maxNetPathE2EQueries)
	}
	return nil
}

func (t *NetPathTask) validateAssertions() error {
	if len(t.SuccessWhen) == 0 {
		return errors.New("success_when is required")
	}

	logic := strings.ToLower(strings.TrimSpace(t.SuccessWhenLogic))
	if logic != "" && logic != "and" && logic != "or" {
		return errors.New("success_when_logic must be and or or")
	}

	count := 0
	for _, success := range t.SuccessWhen {
		if success == nil {
			continue
		}
		for _, assertions := range success.assertions() {
			for _, condition := range assertions.conditions {
				count++
				if err := validateNetPathCondition(assertions.field, condition); err != nil {
					return err
				}
			}
		}
	}
	if count == 0 {
		return errors.New("success_when must contain at least one assertion")
	}
	return nil
}

type netPathAssertions struct {
	field      string
	conditions []*NetPathCondition
}

func (s *NetPathSuccess) assertions() []netPathAssertions {
	return []netPathAssertions{
		{field: "e2e_rtt_avg", conditions: s.E2ERTTAvg},
		{field: "e2e_rtt_min", conditions: s.E2ERTTMin},
		{field: "e2e_rtt_max", conditions: s.E2ERTTMax},
		{field: "e2e_rtt_variation_avg", conditions: s.E2ERTTVariationAvg},
		{field: "e2e_rtt_variation_max", conditions: s.E2ERTTVariationMax},
		{field: "e2e_probe_loss_percent", conditions: s.E2EProbeLossPercent},
		{field: "hop_count", conditions: s.HopCount},
		{field: "e2e_status", conditions: s.E2EStatus},
		{field: "traceroute_status", conditions: s.TracerouteStatus},
	}
}

func validateNetPathCondition(field string, condition *NetPathCondition) error {
	if condition == nil {
		return fmt.Errorf("%s assertion is null", field)
	}

	op := strings.ToLower(strings.TrimSpace(condition.Op))
	if isNetPathStatusField(field) {
		if op != "eq" {
			return fmt.Errorf("%s only supports eq", field)
		}
		if _, err := condition.stringTarget(); err != nil {
			return fmt.Errorf("%s: %w", field, err)
		}
		return nil
	}

	switch op {
	case "eq", "lt", "leq", "gt", "geq":
	default:
		return fmt.Errorf("%s has unsupported operator %q", field, condition.Op)
	}
	if isNetPathDurationField(field) {
		if _, err := condition.durationTarget(); err != nil {
			return fmt.Errorf("%s: %w", field, err)
		}
	} else if _, err := condition.numberTarget(); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}
	return nil
}

func isNetPathStatusField(field string) bool {
	return field == "e2e_status" || field == "traceroute_status"
}

func isNetPathDurationField(field string) bool {
	return strings.HasPrefix(field, "e2e_rtt")
}

func (c *NetPathCondition) stringTarget() (string, error) {
	var target string
	if err := json.Unmarshal(c.Target, &target); err != nil || strings.TrimSpace(target) == "" {
		return "", errors.New("target must be a non-empty string")
	}
	return target, nil
}

func (c *NetPathCondition) durationTarget() (time.Duration, error) {
	target, err := c.stringTarget()
	if err != nil {
		return 0, err
	}
	duration, err := time.ParseDuration(target)
	if err != nil || duration < 0 {
		return 0, errors.New("target must be a duration")
	}
	return duration, nil
}

func (c *NetPathCondition) numberTarget() (float64, error) {
	var target *float64
	if err := json.Unmarshal(c.Target, &target); err != nil || target == nil {
		return 0, errors.New("target must be numeric")
	}
	return *target, nil
}
