//go:build linux
// +build linux

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package netpath sends dynamic network-path candidates from datakit-ebpf to
// the local DataKit netpath input.
package netpath

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
)

const (
	DefaultAPI           = "http://127.0.0.1:9529/v1/netpath/candidates"
	DefaultFlushInterval = 10 * time.Second
	DefaultBatchSize     = 100
	DefaultTimeout       = 3 * time.Second
	DefaultQueueSize     = 4096

	defaultRetryBackoff    = time.Second
	defaultMaxPostAttempts = 5
	maxRetryBackoff        = 30 * time.Second
	minFlushInterval       = 100 * time.Millisecond
	maxFlushInterval       = 10 * time.Minute
	maxBatchSize           = 10000
	minTimeout             = 100 * time.Millisecond
	maxTimeout             = 30 * time.Second
	maxQueueSize           = 100000
	maxPendingSize         = maxQueueSize + maxBatchSize
	tokenHeader            = "X-Datakit-Netpath-Token"
)

var l = logger.DefaultSLogger("ebpf-netpath")

type Config struct {
	Enabled       bool
	API           string
	Token         string
	FlushInterval time.Duration
	BatchSize     int
	Timeout       time.Duration
	QueueSize     int
	Host          string
	Tags          map[string]string

	retryBackoff    time.Duration
	maxPostAttempts int
}

type Candidate struct {
	Hostname          string            `json:"hostname"`
	TargetIP          string            `json:"target_ip,omitempty"`
	IP                string            `json:"ip,omitempty"`
	Port              uint16            `json:"port,omitempty"`
	DstIP             string            `json:"dst_ip,omitempty"`
	DstPort           uint16            `json:"dst_port,omitempty"`
	Protocol          string            `json:"protocol"`
	Origin            string            `json:"origin"`
	Namespace         string            `json:"namespace,omitempty"`
	SourceContainerID string            `json:"source_container_id,omitempty"`
	Source            Source            `json:"source,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

type Source struct {
	Hostname    string `json:"hostname,omitempty"`
	IP          string `json:"ip,omitempty"`
	Port        uint16 `json:"port,omitempty"`
	NetNS       string `json:"netns,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	PID         uint32 `json:"pid,omitempty"`
	ProcessName string `json:"process_name,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
}

type request struct {
	Source string            `json:"source"`
	Host   string            `json:"host,omitempty"`
	SentAt int64             `json:"sent_at,omitempty"`
	Tags   map[string]string `json:"tags,omitempty"`
	Tests  []Candidate       `json:"tests"`
}

type candidateResponse struct {
	Accepted    int            `json:"accepted"`
	Dropped     int            `json:"dropped"`
	DropReasons map[string]int `json:"drop_reasons,omitempty"`
}

type candidateResponseError struct {
	response  candidateResponse
	retryable bool
}

func (e *candidateResponseError) Error() string {
	dropType := "permanently"
	if e.retryable {
		dropType = "transiently"
	}
	return fmt.Sprintf(
		"datakit accepted %d and %s dropped %d netpath candidates: %v",
		e.response.Accepted, dropType, e.response.Dropped, e.response.DropReasons,
	)
}

type Scheduler struct {
	cfg     Config
	client  *http.Client
	ctx     context.Context
	cancel  context.CancelFunc
	ch      chan Candidate
	done    chan struct{}
	once    sync.Once
	wg      sync.WaitGroup
	mu      sync.Mutex
	closed  bool
	pending map[candidateKey]struct{}
}

type candidateKey struct {
	sourceHost string
	source     string
	namespace  string
	netns      string
	target     string
	targetIP   string
	protocol   string
	dstIP      string
	port       uint16
	dstPort    uint16
}

func NewScheduler(cfg Config) *Scheduler {
	cfg = normalizeConfig(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		ctx:     ctx,
		cancel:  cancel,
		ch:      make(chan Candidate, cfg.QueueSize),
		done:    make(chan struct{}),
		pending: make(map[candidateKey]struct{}, pendingCapacity(cfg.QueueSize, cfg.BatchSize)),
	}
	s.wg.Add(1)
	go s.run()
	return s
}

func normalizeConfig(cfg Config) Config {
	cfg.API = strings.TrimSpace(cfg.API)
	if cfg.API == "" {
		cfg.API = DefaultAPI
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = DefaultFlushInterval
	} else if cfg.FlushInterval < minFlushInterval {
		cfg.FlushInterval = minFlushInterval
	} else if cfg.FlushInterval > maxFlushInterval {
		cfg.FlushInterval = maxFlushInterval
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = DefaultBatchSize
	} else if cfg.BatchSize > maxBatchSize {
		cfg.BatchSize = maxBatchSize
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	} else if cfg.Timeout < minTimeout {
		cfg.Timeout = minTimeout
	} else if cfg.Timeout > maxTimeout {
		cfg.Timeout = maxTimeout
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = DefaultQueueSize
	} else if cfg.QueueSize > maxQueueSize {
		cfg.QueueSize = maxQueueSize
	}
	if cfg.retryBackoff <= 0 {
		cfg.retryBackoff = defaultRetryBackoff
	}
	if cfg.maxPostAttempts <= 0 {
		cfg.maxPostAttempts = defaultMaxPostAttempts
	}
	return cfg
}

func pendingCapacity(queueSize, batchSize int) int {
	if queueSize >= maxPendingSize || batchSize >= maxPendingSize-queueSize {
		return maxPendingSize
	}
	return queueSize + batchSize
}

func (s *Scheduler) Schedule(candidate Candidate) bool {
	if s == nil {
		return false
	}
	if strings.TrimSpace(candidate.Hostname) == "" && strings.TrimSpace(candidate.TargetIP) == "" &&
		strings.TrimSpace(candidate.IP) == "" {
		return false
	}
	key := makeCandidateKey(s.cfg.Host, candidate)
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return false
	}
	if s.pending == nil {
		s.pending = make(map[candidateKey]struct{})
	}
	if _, ok := s.pending[key]; ok {
		s.mu.Unlock()
		return true
	}
	select {
	case s.ch <- candidate:
		s.pending[key] = struct{}{}
		s.mu.Unlock()
		return true
	default:
		s.mu.Unlock()
		exporter.AddNetpathCandidatesDropped("queue_full", 1)
		l.Debug("drop netpath candidate: queue full")
		return false
	}
}

func makeCandidateKey(host string, candidate Candidate) candidateKey {
	targetIP := strings.TrimSpace(candidate.TargetIP)
	if targetIP == "" {
		targetIP = strings.TrimSpace(candidate.IP)
	}
	target := strings.TrimSpace(candidate.Hostname)
	if target == "" {
		target = targetIP
	}

	sourceHost := strings.TrimSpace(candidate.Source.Hostname)
	if sourceHost == "" {
		sourceHost = strings.TrimSpace(host)
	}
	source := strings.TrimSpace(candidate.Source.ServiceName)
	if source == "" {
		source = strings.TrimSpace(candidate.Source.ProcessName)
	}
	if source == "" {
		source = strings.TrimSpace(candidate.Source.ContainerID)
	}
	if source == "" {
		source = strings.TrimSpace(candidate.SourceContainerID)
	}
	if source == "" {
		source = sourceHost
	}
	key := candidateKey{
		sourceHost: sourceHost,
		source:     source,
		namespace:  strings.TrimSpace(candidate.Namespace),
		netns:      strings.TrimSpace(candidate.Source.NetNS),
		target:     target,
		targetIP:   targetIP,
		port:       candidate.Port,
		protocol:   strings.ToLower(strings.TrimSpace(candidate.Protocol)),
	}
	if candidateHasTranslation(candidate, targetIP) {
		key.dstIP = strings.TrimSpace(candidate.DstIP)
		key.dstPort = candidate.DstPort
	}
	return key
}

func candidateHasTranslation(candidate Candidate, targetIP string) bool {
	if targetIP == "" {
		return false
	}
	dstIP := strings.TrimSpace(candidate.DstIP)
	if dstIP != "" && !sameIP(dstIP, targetIP) {
		return true
	}
	return candidate.DstPort > 0 && candidate.DstPort != candidate.Port
}

func sameIP(left, right string) bool {
	leftIP := net.ParseIP(strings.TrimSpace(left))
	rightIP := net.ParseIP(strings.TrimSpace(right))
	if leftIP != nil && rightIP != nil {
		return leftIP.Equal(rightIP)
	}
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func (s *Scheduler) release(candidates []Candidate) {
	s.mu.Lock()
	for _, candidate := range candidates {
		delete(s.pending, makeCandidateKey(s.cfg.Host, candidate))
	}
	s.mu.Unlock()
}

func (s *Scheduler) clearPending() {
	s.mu.Lock()
	clear(s.pending)
	s.mu.Unlock()
}

func (s *Scheduler) Close() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
		if s.cancel != nil {
			s.cancel()
		}
		if s.done != nil {
			close(s.done)
		}
		s.wg.Wait()
	})
}

func (s *Scheduler) run() {
	defer s.wg.Done()
	defer s.clearPending()

	ticker := time.NewTicker(s.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]Candidate, 0, s.cfg.BatchSize)
	attempts := 0
	lastResponseDropped := 0
	var retryTimer *time.Timer
	var retryCh <-chan time.Time

	stopRetry := func() {
		if retryTimer != nil && !retryTimer.Stop() && retryCh != nil {
			select {
			case <-retryTimer.C:
			default:
			}
		}
		retryCh = nil
	}
	defer stopRetry()

	resetBatch := func() {
		s.release(batch)
		for i := range batch {
			batch[i] = Candidate{}
		}
		batch = batch[:0]
		attempts = 0
		lastResponseDropped = 0
		stopRetry()
	}
	observePending := func() {
		exporter.ObserveNetpathPendingCandidates(len(batch) + len(s.ch))
	}
	scheduleRetry := func(delay time.Duration) {
		stopRetry()
		if retryTimer == nil {
			retryTimer = time.NewTimer(delay)
		} else {
			retryTimer.Reset(delay)
		}
		retryCh = retryTimer.C
	}
	flush := func() {
		if len(batch) == 0 {
			return
		}
		retryable, err := s.post(batch)
		attempts++
		if err == nil {
			exporter.ObserveNetpathCandidateRequest("success")
			resetBatch()
			observePending()
			return
		}
		var responseErr *candidateResponseError
		hasResponse := errors.As(err, &responseErr)
		if hasResponse {
			lastResponseDropped = responseErr.response.Dropped
			if !retryable {
				exporter.ObserveNetpathCandidateRequest("partial")
				exporter.AddNetpathCandidatesDropped("server_rejected", responseErr.response.Dropped)
				l.Warnf("release netpath candidate batch after partial response: %s", err.Error())
				resetBatch()
				observePending()
				return
			}
		}

		result := "permanent_error"
		if retryable {
			result = "retryable_error"
		}
		exporter.ObserveNetpathCandidateRequest(result)
		if retryable && attempts < s.cfg.maxPostAttempts {
			delay := postRetryDelay(s.cfg.retryBackoff, attempts)
			l.Warnf("post netpath candidates failed (attempt %d/%d), retry in %s: %s",
				attempts, s.cfg.maxPostAttempts, delay, err.Error())
			scheduleRetry(delay)
			observePending()
			return
		}

		reason := "permanent_error"
		if retryable {
			reason = "retry_exhausted"
		}
		dropped := len(batch)
		if hasResponse {
			dropped = responseErr.response.Dropped
		} else if lastResponseDropped > 0 {
			dropped = lastResponseDropped
		}
		exporter.AddNetpathCandidatesDropped(reason, dropped)
		l.Warnf("drop %d netpath candidates after %d post attempt(s): %s", dropped, attempts, err.Error())
		resetBatch()
		observePending()
	}
	observePending()

	for {
		candidateCh := s.ch
		if len(batch) >= s.cfg.BatchSize || retryCh != nil {
			candidateCh = nil
		}
		select {
		case <-s.done:
			stopRetry()
			// Close prevents new Schedule calls before signaling done, so the
			// channel can be drained without closing it. Use one deadline for all
			// remaining batches to keep shutdown bounded.
			shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Timeout)
			for {
			drainQueue:
				// A batch with prior attempts may already have been partially
				// accepted. Keep it unchanged so lastResponseDropped continues to
				// describe exactly this batch if the shutdown attempt has no response.
				for attempts == 0 && len(batch) < s.cfg.BatchSize {
					select {
					case candidate := <-s.ch:
						batch = append(batch, candidate)
					default:
						break drainQueue
					}
				}
				if len(batch) == 0 {
					break
				}

				retryable, err := s.postContext(shutdownCtx, batch)
				if err == nil {
					exporter.ObserveNetpathCandidateRequest("success")
					resetBatch()
					observePending()
					continue
				}

				var responseErr *candidateResponseError
				if errors.As(err, &responseErr) && !retryable {
					// A non-retryable partial response fully accounts for this batch:
					// accepted candidates are stored and the rest are final rejects.
					exporter.ObserveNetpathCandidateRequest("partial")
					exporter.AddNetpathCandidatesDropped("server_rejected", responseErr.response.Dropped)
					l.Warnf("release netpath candidate batch after shutdown partial response: %s", err.Error())
					resetBatch()
					observePending()
					continue
				}

				dropped := len(batch)
				if responseErr != nil {
					dropped = responseErr.response.Dropped
				} else if lastResponseDropped > 0 {
					dropped = lastResponseDropped
				}
				queued := len(s.ch)
				totalDropped := dropped + queued
				result := "permanent_error"
				if retryable {
					result = "retryable_error"
				}
				exporter.ObserveNetpathCandidateRequest(result)
				exporter.AddNetpathCandidatesDropped("shutdown", totalDropped)
				l.Warnf("drop %d netpath candidates during shutdown (%d current, %d queued): %s",
					totalDropped, dropped, queued, err.Error())
				// The scheduler may remain referenced after Close. Empty the channel
				// so rejected candidates and their tag maps are not retained.
				for i := range batch {
					batch[i] = Candidate{}
				}
				batch = batch[:0]
				for i := 0; i < queued; i++ {
					<-s.ch
				}
				break
			}
			cancel()
			exporter.ObserveNetpathPendingCandidates(0)
			return
		case candidate := <-candidateCh:
			batch = append(batch, candidate)
			observePending()
			if len(batch) >= s.cfg.BatchSize && retryCh == nil {
				flush()
			}
		case <-retryCh:
			retryCh = nil
			flush()
		case <-ticker.C:
			if retryCh == nil {
				flush()
			}
		}
	}
}

func postRetryDelay(base time.Duration, attempts int) time.Duration {
	if base >= maxRetryBackoff {
		return maxRetryBackoff
	}
	delay := base
	for i := 1; i < attempts && delay < maxRetryBackoff; i++ {
		if delay > maxRetryBackoff/2 {
			return maxRetryBackoff
		}
		delay *= 2
	}
	return delay
}

func (s *Scheduler) post(candidates []Candidate) (bool, error) {
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return s.postContext(ctx, candidates)
}

func (s *Scheduler) postContext(ctx context.Context, candidates []Candidate) (bool, error) {
	payload := request{
		Source: "ebpf_netflow",
		Host:   s.cfg.Host,
		SentAt: time.Now().UnixMilli(),
		Tags:   s.cfg.Tags,
		Tests:  append([]Candidate(nil), candidates...),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.API, bytes.NewReader(data))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.Token != "" {
		req.Header.Set(tokenHeader, s.cfg.Token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close() //nolint:errcheck
	body, err := io.ReadAll(io.LimitReader(resp.Body, (64<<10)+1))
	if err != nil {
		return true, fmt.Errorf("drain response body: %w", err)
	}
	if len(body) > 64<<10 {
		return true, fmt.Errorf("response body exceeds 64 KiB")
	}
	if resp.StatusCode/100 != 2 {
		retryable := resp.StatusCode == http.StatusRequestTimeout ||
			resp.StatusCode == http.StatusTooEarly ||
			resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode >= http.StatusInternalServerError
		return retryable, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	// Older DataKit versions returned an empty 2xx response. Preserve that
	// behavior while requiring current versions to account for partial drops.
	if len(bytes.TrimSpace(body)) == 0 {
		return false, nil
	}
	var result candidateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return true, fmt.Errorf("decode successful response: %w", err)
	}
	if result.Accepted < 0 || result.Dropped < 0 || result.Accepted+result.Dropped != len(candidates) {
		return true, fmt.Errorf(
			"invalid candidate response counts: accepted=%d dropped=%d sent=%d",
			result.Accepted, result.Dropped, len(candidates),
		)
	}
	if result.Dropped == 0 {
		return false, nil
	}

	if result.DropReasons["input_chan_full"] > 0 ||
		result.DropReasons["scheduler_stopped"] > 0 {
		// The response does not identify individual candidates. Retrying the
		// entire batch is safe because DataKit stores tasks by schedule key and
		// updates an existing task when an accepted candidate is seen again.
		return true, &candidateResponseError{response: result, retryable: true}
	}
	return false, &candidateResponseError{response: result}
}
