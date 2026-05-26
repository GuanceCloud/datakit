// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultProcRoot                = "/proc"
	defaultCgroupRoot              = "/sys/fs/cgroup"
	defaultCgroupPollInterval      = 250 * time.Millisecond
	defaultCgroupEmergencyPercent  = 95.0
	defaultCgroupEmergencyCooldown = time.Minute
	cgroupVersionV1                = "v1"
	cgroupVersionV2                = "v2"
)

type cgroupMemoryStats struct {
	Current uint64
	Max     uint64
	OOMKill uint64
}

func parseProcCgroupV2Path(content string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}
		if parts[1] == "" && strings.HasPrefix(parts[2], "/") {
			return parts[2], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan proc cgroup: %w", err)
	}

	return "", errors.New("cgroup v2 path not found")
}

func parseProcCgroupV1Path(content string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}

		controllers := strings.Split(parts[1], ",")
		for _, controller := range controllers {
			if controller == "memory" && strings.HasPrefix(parts[2], "/") {
				return parts[2], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan proc cgroup: %w", err)
	}

	return "", errors.New("cgroup v1 memory path not found")
}

func resolveCgroupV2Dir(procRoot, cgroupRoot string, pid int32) (string, error) {
	if procRoot == "" {
		procRoot = defaultProcRoot
	}
	if cgroupRoot == "" {
		cgroupRoot = defaultCgroupRoot
	}

	bts, err := os.ReadFile(filepath.Join(procRoot, strconv.Itoa(int(pid)), "cgroup")) //nolint:gosec
	if err != nil {
		return "", err
	}

	cgroupPath, err := parseProcCgroupV2Path(string(bts))
	if err != nil {
		return "", err
	}

	return resolveCgroupDir(procRoot, cgroupRoot, pid, cgroupPath,
		[]string{"memory.current", "memory.max", "memory.events"}), nil
}

func resolveCgroupV1Dir(procRoot, cgroupRoot string, pid int32) (string, error) {
	if procRoot == "" {
		procRoot = defaultProcRoot
	}
	if cgroupRoot == "" {
		cgroupRoot = defaultCgroupRoot
	}

	bts, err := os.ReadFile(filepath.Join(procRoot, strconv.Itoa(int(pid)), "cgroup")) //nolint:gosec
	if err != nil {
		return "", err
	}

	cgroupPath, err := parseProcCgroupV1Path(string(bts))
	if err != nil {
		return "", err
	}

	return resolveCgroupDir(procRoot, filepath.Join(cgroupRoot, "memory"), pid, cgroupPath,
		[]string{"memory.usage_in_bytes", "memory.limit_in_bytes", "memory.oom_control"}), nil
}

func resolveCgroupDir(procRoot, cgroupRoot string, pid int32, cgroupPath string, requiredFiles []string) string {
	candidates := make([]string, 0, 4)
	addCandidate := func(dir string) {
		if dir == "" {
			return
		}
		for _, existing := range candidates {
			if existing == dir {
				return
			}
		}
		candidates = append(candidates, dir)
	}

	procCgroupRoot := filepath.Join(procRoot, strconv.Itoa(int(pid)), "root", strings.TrimPrefix(cgroupRoot, "/"))
	addCandidate(cgroupRoot)
	addCandidate(procCgroupRoot)

	if cgroupPath != "" && cgroupPath != "/" && !strings.Contains(cgroupPath, "..") {
		rel := strings.TrimPrefix(cgroupPath, "/")
		addCandidate(filepath.Join(cgroupRoot, rel))
		addCandidate(filepath.Join(procCgroupRoot, rel))
	}

	for _, candidate := range candidates {
		if hasAllFiles(candidate, requiredFiles...) {
			return candidate
		}
	}

	if cgroupPath == "" || cgroupPath == "/" {
		return cgroupRoot
	}
	return filepath.Join(cgroupRoot, strings.TrimPrefix(cgroupPath, "/"))
}

func hasAllFiles(dir string, names ...string) bool {
	if dir == "" {
		return false
	}
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return false
		}
	}
	return true
}

func resolveCgroupWatcherTarget(procRoot, cgroupRoot string, pid int32) (string, string, error) {
	if dir, err := resolveCgroupV2Dir(procRoot, cgroupRoot, pid); err == nil {
		return dir, cgroupVersionV2, nil
	}
	if dir, err := resolveCgroupV1Dir(procRoot, cgroupRoot, pid); err == nil {
		return dir, cgroupVersionV1, nil
	}
	return "", "", errors.New("unable to resolve cgroup watcher target")
}

func readCgroupV2MemoryStats(dir string) (*cgroupMemoryStats, error) {
	current, err := readUintFromFile(filepath.Join(dir, "memory.current"))
	if err != nil {
		return nil, err
	}
	max, err := readMaybeMaxUintFromFile(filepath.Join(dir, "memory.max"))
	if err != nil {
		return nil, err
	}
	oomKill, err := readOOMKillFromEvents(filepath.Join(dir, "memory.events"))
	if err != nil {
		return nil, err
	}

	return &cgroupMemoryStats{Current: current, Max: max, OOMKill: oomKill}, nil
}

func readCgroupV1MemoryStats(dir string) (*cgroupMemoryStats, error) {
	current, err := readUintFromFile(filepath.Join(dir, "memory.usage_in_bytes"))
	if err != nil {
		return nil, err
	}
	max, err := readMaybeMaxUintFromFile(filepath.Join(dir, "memory.limit_in_bytes"))
	if err != nil {
		return nil, err
	}
	oomKill, err := readOOMKillFromOOMControl(filepath.Join(dir, "memory.oom_control"))
	if err != nil {
		return nil, err
	}

	return &cgroupMemoryStats{Current: current, Max: max, OOMKill: oomKill}, nil
}

func readUintFromFile(path string) (uint64, error) {
	bts, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(bts)), 10, 64)
}

func readMaybeMaxUintFromFile(path string) (uint64, error) {
	bts, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return 0, err
	}
	val := strings.TrimSpace(string(bts))
	if val == "" || val == "max" {
		return 0, nil
	}
	return strconv.ParseUint(val, 10, 64)
}

func readOOMKillFromEvents(path string) (uint64, error) {
	bts, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(bts)))
	for scanner.Scan() {
		parts := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(parts) == 2 && parts[0] == "oom_kill" {
			return strconv.ParseUint(parts[1], 10, 64)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, nil
}

func readOOMKillFromOOMControl(path string) (uint64, error) {
	bts, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(bts)))
	for scanner.Scan() {
		parts := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(parts) == 2 && parts[0] == "oom_kill" {
			return strconv.ParseUint(parts[1], 10, 64)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, nil
}

func getCgroupEmergencyPercent(p *Process) float64 {
	if p == nil || p.MEMUsagePercentEmergency <= 0 {
		return defaultCgroupEmergencyPercent
	}
	return float64(p.MEMUsagePercentEmergency)
}

type cgroupTriggerCandidate struct {
	pm       *processM
	rssBytes uint64
}

func (m *monitor) watchCgroupMemory(ctx context.Context, watcher *cgroupWatcher) {
	if watcher == nil {
		return
	}

	var (
		lastOOMKill uint64
		initialized bool
	)
	ticker := time.NewTicker(m.cgroupPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := m.readCgroupMemoryStats(watcher.version, watcher.dir)
			if err != nil {
				log.Debugf("read cgroup %s memory stats failed for watcher=%s: %v", watcher.version, watcher.key, err)
				continue
			}
			if !initialized {
				lastOOMKill = stats.OOMKill
				initialized = true
				continue
			}
			lastOOMKill = m.handleCgroupMemoryStats(watcher, stats, lastOOMKill, nowFunc())
		}
	}
}

func (m *monitor) readCgroupMemoryStats(version string, dir string) (*cgroupMemoryStats, error) {
	switch version {
	case cgroupVersionV2:
		return readCgroupV2MemoryStats(dir)
	case cgroupVersionV1:
		return readCgroupV1MemoryStats(dir)
	default:
		return nil, fmt.Errorf("unsupported cgroup version: %s", version)
	}
}

var nowFunc = time.Now

func (m *monitor) handleCgroupMemoryStats(watcher *cgroupWatcher, stats *cgroupMemoryStats, lastOOMKill uint64, now time.Time) uint64 {
	if watcher == nil || stats == nil {
		return lastOOMKill
	}

	if stats.OOMKill > lastOOMKill {
		delta := stats.OOMKill - lastOOMKill
		event := m.buildOOMEvent(watcher, delta, now)
		if event != nil {
			for _, candidate := range event.Candidates {
				service := candidate.Service
				if service == "" {
					service = "unknown_service"
				}
				missedOOM.WithLabelValues(service).Add(float64(delta))
			}
			if len(event.Candidates) == 0 {
				log.Infof("oom detected for watcher=%s but no target process members remain", watcher.key)
			}
			if m != nil && m.config != nil && m.config.OOMHProfEnabled {
				m.oomChan <- event
			}
		}
		lastOOMKill = stats.OOMKill
	}

	if stats.Max == 0 {
		return lastOOMKill
	}

	percent := (float64(stats.Current) / float64(stats.Max)) * 100
	candidate := m.selectCgroupTriggerCandidate(watcher, percent)
	if candidate == nil || candidate.pm == nil || candidate.pm.configProcess == nil {
		return lastOOMKill
	}
	pm := candidate.pm

	if !pm.inCooldown(defaultCgroupEmergencyCooldown) {
		pm.markProfileTriggered(now)
		tags := make([]string, 0, len(m.config.Tags)+len(pm.configProcess.Tags)+3)
		tags = append(tags, m.config.Tags...)
		tags = append(tags, pm.configProcess.Tags...)
		tags = append(tags,
			fmt.Sprintf("service:%s", pm.configProcess.Service),
			"trigger:cgroup_memory_pressure",
			fmt.Sprintf("cgroup_mem_percent:%0.2f", percent),
		)

		req := newTriggerStats(pm.configProcess.Events, getEmergencyProfileDuration(pm.configProcess), tags)
		req.CommandName = pm.Name
		req.PID = pm.Pid
		req.Triggered = true
		req.Service = pm.configProcess.Service
		m.statsChan <- req
	}

	if m != nil && m.config != nil && m.config.JCmdSnapshotEnabled && !pm.inJcmdCooldown(defaultCgroupEmergencyCooldown) {
		pm.markJcmdTriggered(now)
		jcmdTags := make([]string, 0, len(m.config.Tags)+len(pm.configProcess.Tags)+2)
		jcmdTags = append(jcmdTags, m.config.Tags...)
		jcmdTags = append(jcmdTags, pm.configProcess.Tags...)
		jcmdTags = append(jcmdTags,
			fmt.Sprintf("service:%s", pm.configProcess.Service),
			"trigger:cgroup_memory_pressure",
		)
		m.jcmdChan <- &jcmdSnapshotRequest{
			Service:     pm.configProcess.Service,
			PID:         pm.Pid,
			ProcessName: pm.Name,
			DetectedAt:  now,
			MemPercent:  percent,
			Tags:        jcmdTags,
		}
	}

	return lastOOMKill
}

func (m *monitor) selectCgroupTriggerCandidate(watcher *cgroupWatcher, percent float64) *cgroupTriggerCandidate {
	if watcher == nil {
		return nil
	}

	members := watcher.snapshotMembers()
	var best *cgroupTriggerCandidate
	for _, pm := range members {
		if pm == nil || pm.configProcess == nil {
			continue
		}
		if percent < getCgroupEmergencyPercent(pm.configProcess) {
			continue
		}

		rssBytes, ok := pm.currentRSSBytes()
		if !ok {
			rssBytes = 0
		}

		if best == nil || rssBytes > best.rssBytes {
			best = &cgroupTriggerCandidate{
				pm:       pm,
				rssBytes: rssBytes,
			}
		}
	}

	return best
}

func (m *monitor) buildOOMEvent(watcher *cgroupWatcher, delta uint64, now time.Time) *OOMEvent {
	if watcher == nil {
		return nil
	}

	members := watcher.snapshotMembers()
	globalTags := []string{}
	if m != nil && m.config != nil {
		globalTags = m.config.Tags
	}
	event := &OOMEvent{
		DetectedAt:   now,
		OOMKillDelta: delta,
		Candidates:   make([]*oomProcessCandidate, 0, len(members)),
	}

	for _, pm := range members {
		if pm == nil || pm.configProcess == nil {
			continue
		}

		service := pm.configProcess.Service
		if service == "" {
			service = "unknown_service"
		}
		hprofPath := detectOOMHProfPath(pm)
		if hprofPath == "" {
			log.Infof("oom detected for pid=%d service=%s but HeapDumpOnOutOfMemoryError/HeapDumpPath not found in java args", pm.Pid, service)
		}

		tags := make([]string, 0, len(globalTags)+len(pm.configProcess.Tags)+1)
		tags = append(tags, globalTags...)
		tags = append(tags, pm.configProcess.Tags...)
		tags = append(tags, fmt.Sprintf("service:%s", service))

		event.Candidates = append(event.Candidates, &oomProcessCandidate{
			Service:     service,
			PID:         pm.Pid,
			ProcessName: pm.Name,
			HProfPath:   hprofPath,
			Tags:        tags,
		})
		log.Warnf("detected oom_kill increase for pid=%d service=%s delta=%d watcher=%s", pm.Pid, service, delta, watcher.key)
	}

	return event
}
