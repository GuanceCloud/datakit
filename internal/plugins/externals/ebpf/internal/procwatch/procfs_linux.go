//go:build linux
// +build linux

package procwatch

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type mountNamespace struct {
	dev uint64
	ino uint64
}

type nonRegularExecutablePathError struct {
	pid  int
	path string
}

func (e *nonRegularExecutablePathError) Error() string {
	return fmt.Sprintf("non-regular executable path for pid %d: %s", e.pid, e.path)
}

type mountEntry struct {
	dev        string
	root       string
	mountPoint string
	fsType     string
	source     string
	options    map[string]string
}

type mountSnapshot struct {
	mounts []*mountEntry
}

type pathResolverCache struct {
	rootMounts *mountSnapshot
	rootNS     mountNamespace
	rootByDev  map[string][]*mountEntry

	mu          sync.RWMutex
	nsMounts    map[mountNamespace]*mountSnapshot
	resolved    map[mountNamespace]map[string]resolvedPathCacheEntry
	lastAccess  map[mountNamespace]time.Time
	lastCleanup time.Time
}

type resolvedPathCacheEntry struct {
	path string
	ts   time.Time
}

var sharedResolverCache = newPathResolverCache(HostProc())

const (
	deletedPathSuffix           = " (deleted)"
	resolvedPathCacheTTL        = 5 * time.Minute
	resolvedPathNamespaceTTL    = 10 * time.Minute
	resolvedPathCacheMaxPerNS   = 256
	resolvedPathCleanupInterval = time.Minute
)

func newPathResolverCache(procRoot string) *pathResolverCache {
	rootMounts := readMounts(filepath.Join(procRoot, "1", "mountinfo"))
	return &pathResolverCache{
		rootMounts: rootMounts,
		rootNS:     readMountNamespace(filepath.Join(procRoot, "1")),
		rootByDev:  indexMountsByDev(rootMounts),
		nsMounts:   make(map[mountNamespace]*mountSnapshot),
		resolved:   make(map[mountNamespace]map[string]resolvedPathCacheEntry),
		lastAccess: make(map[mountNamespace]time.Time),
	}
}

func (c *pathResolverCache) Resolve(pidPath string, target string) string {
	if c == nil || c.rootMounts == nil || target == "" {
		return ""
	}

	now := time.Now()
	ns := readMountNamespace(pidPath)
	if ns == (mountNamespace{}) {
		return ""
	}
	if ns == c.rootNS {
		return target
	}

	c.mu.RLock()
	if cache := c.resolved[ns]; cache != nil {
		if resolved, ok := cache[target]; ok && now.Sub(resolved.ts) < resolvedPathCacheTTL {
			c.mu.RUnlock()
			c.touchNamespace(ns, now)
			return resolved.path
		}
	}
	nsMounts := c.nsMounts[ns]
	c.mu.RUnlock()

	if nsMounts == nil {
		nsMounts = readMounts(filepath.Join(pidPath, "mountinfo"))
		if nsMounts == nil {
			return ""
		}
		c.mu.Lock()
		if cached := c.nsMounts[ns]; cached == nil {
			c.nsMounts[ns] = nsMounts
		} else {
			nsMounts = cached
		}
		c.lastAccess[ns] = now
		c.mu.Unlock()
	}

	resolved := resolveMountPath(c.rootMounts, c.rootByDev, nsMounts, target)
	if resolved == "" {
		c.touchNamespace(ns, now)
		return ""
	}

	c.mu.Lock()
	cache := c.resolved[ns]
	if cache == nil {
		cache = make(map[string]resolvedPathCacheEntry)
		c.resolved[ns] = cache
	} else if len(cache) >= resolvedPathCacheMaxPerNS {
		cache = make(map[string]resolvedPathCacheEntry)
		c.resolved[ns] = cache
	}
	cache[target] = resolvedPathCacheEntry{
		path: resolved,
		ts:   now,
	}
	c.lastAccess[ns] = now
	c.maybeCleanupLocked(now)
	c.mu.Unlock()
	return resolved
}

func (c *pathResolverCache) touchNamespace(ns mountNamespace, now time.Time) {
	c.mu.Lock()
	c.lastAccess[ns] = now
	c.maybeCleanupLocked(now)
	c.mu.Unlock()
}

func (c *pathResolverCache) maybeCleanupLocked(now time.Time) {
	if !c.lastCleanup.IsZero() && now.Sub(c.lastCleanup) < resolvedPathCleanupInterval {
		return
	}

	for ns, last := range c.lastAccess {
		if now.Sub(last) > resolvedPathNamespaceTTL {
			delete(c.lastAccess, ns)
			delete(c.nsMounts, ns)
			delete(c.resolved, ns)
			continue
		}

		cache := c.resolved[ns]
		for path, entry := range cache {
			if now.Sub(entry.ts) > resolvedPathCacheTTL {
				delete(cache, path)
			}
		}
		if len(cache) == 0 {
			delete(c.resolved, ns)
		}
	}

	c.lastCleanup = now
}

func resolveMountPath(rootMounts *mountSnapshot, rootByDev map[string][]*mountEntry,
	nsMounts *mountSnapshot, target string,
) string {
	if rootMounts == nil || nsMounts == nil || target == "" {
		return ""
	}

	nsMount := nsMounts.find(target)
	if nsMount == nil {
		return ""
	}
	if resolved := resolveOverlayMountPath(nsMount, target); resolved != "" {
		return resolved
	}

	relTarget, err := filepath.Rel(nsMount.mountPoint, target)
	if err != nil {
		return ""
	}

	var rootMount *mountEntry
	for _, candidate := range rootMountCandidates(rootMounts, rootByDev, nsMount.dev) {
		if candidate.dev == nsMount.dev && strings.HasPrefix(nsMount.root, candidate.root) {
			rootMount = candidate
			break
		}
	}
	if rootMount == nil {
		return ""
	}

	relRoot, err := filepath.Rel(nsMount.root, rootMount.root)
	if err != nil {
		return ""
	}

	return filepath.Join(rootMount.mountPoint, relRoot, relTarget)
}

func resolveOverlayMountPath(mount *mountEntry, target string) string {
	if mount == nil || mount.fsType != "overlay" || mount.options == nil {
		return ""
	}

	relTarget, err := filepath.Rel(mount.mountPoint, target)
	if err != nil {
		return ""
	}
	if relTarget == "." {
		relTarget = ""
	}

	if upper := unescapeProcPathField([]byte(mount.options["upperdir"])); upper != "" {
		candidate := filepath.Join(upper, relTarget)
		if isRegularFile(candidate) {
			return candidate
		}
	}

	for _, key := range []string{"lowerdir", "lowerdir+"} {
		for _, layer := range splitEscapedMountPathList(mount.options[key], ':') {
			candidate := filepath.Join(layer, relTarget)
			if isRegularFile(candidate) {
				return candidate
			}
		}
	}

	return ""
}

func splitEscapedMountPathList(raw string, sep byte) []string {
	if raw == "" {
		return nil
	}

	paths := make([]string, 0, 4)
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i < len(raw) && raw[i] == '\\' && i+3 < len(raw) &&
			isOctalDigit(raw[i+1]) && isOctalDigit(raw[i+2]) && isOctalDigit(raw[i+3]) {
			i += 3
			continue
		}
		if i < len(raw) && raw[i] != sep {
			continue
		}

		path := unescapeProcPathField([]byte(raw[start:i]))
		if path == "" {
			start = i + 1
			continue
		}
		paths = append(paths, path)
		start = i + 1
	}
	return paths
}

func rootMountCandidates(rootMounts *mountSnapshot, rootByDev map[string][]*mountEntry, dev string) []*mountEntry {
	if len(rootByDev) > 0 {
		if mounts, ok := rootByDev[dev]; ok {
			return mounts
		}
	}
	if rootMounts == nil {
		return nil
	}
	return rootMounts.mounts
}

func indexMountsByDev(snapshot *mountSnapshot) map[string][]*mountEntry {
	if snapshot == nil || len(snapshot.mounts) == 0 {
		return nil
	}

	index := make(map[string][]*mountEntry, len(snapshot.mounts))
	for _, mount := range snapshot.mounts {
		if mount == nil || mount.dev == "" {
			continue
		}
		index[mount.dev] = append(index[mount.dev], mount)
	}
	for dev := range index {
		sort.SliceStable(index[dev], func(i, j int) bool {
			return len(index[dev][i].root) > len(index[dev][j].root)
		})
	}
	return index
}

func readMounts(path string) *mountSnapshot {
	// Path is derived from procfs roots managed by this package.
	//nolint:gosec
	file, err := os.Open(path)
	if err != nil {
		return nil
	}

	mounts := make([]*mountEntry, 0, 32)
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		mount, ok := parseMountInfoLine(line)
		if !ok {
			continue
		}
		mounts = append(mounts, mount)
	}

	_ = file.Close()

	sortMountEntries(mounts)
	return &mountSnapshot{mounts: mounts}
}

func parseMountInfoLine(line []byte) (*mountEntry, bool) {
	fieldIdx := 0
	start := 0
	entry := &mountEntry{}

	for start < len(line) {
		for start < len(line) && line[start] == ' ' {
			start++
		}
		if start >= len(line) {
			break
		}

		end := start
		for end < len(line) && line[end] != ' ' {
			end++
		}

		switch fieldIdx {
		case 2:
			entry.dev = string(line[start:end])
		case 3:
			entry.root = unescapeProcPathField(line[start:end])
		case 4:
			entry.mountPoint = unescapeProcPathField(line[start:end])
		}
		if fieldIdx >= 5 && end < len(line) && line[end] == ' ' && end+1 < len(line) && line[end+1] == '-' {
			parseMountInfoPostSeparator(entry, line[end+2:])
			return entry, entry.dev != "" && entry.root != "" && entry.mountPoint != ""
		}

		fieldIdx++
		start = end + 1
	}

	return nil, false
}

func parseMountInfoPostSeparator(entry *mountEntry, tail []byte) {
	if entry == nil || len(tail) == 0 {
		return
	}
	fields := bytes.Fields(bytes.TrimSpace(tail))
	if len(fields) < 3 {
		return
	}
	entry.fsType = string(fields[0])
	entry.source = unescapeProcPathField(fields[1])
	entry.options = parseMountOptionMap(fields[2])
}

func parseMountOptionMap(raw []byte) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	parts := bytes.Split(raw, []byte{','})
	options := make(map[string]string, len(parts))
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		key, value, ok := bytes.Cut(part, []byte{'='})
		if !ok {
			options[string(part)] = ""
			continue
		}
		options[string(key)] = string(value)
	}
	if len(options) == 0 {
		return nil
	}
	return options
}

func (s *mountSnapshot) find(target string) *mountEntry {
	for _, mount := range s.mounts {
		if pathWithinMount(target, mount.mountPoint) {
			return mount
		}
	}
	return nil
}

func pathWithinMount(target, mountPoint string) bool {
	if target == "" || mountPoint == "" {
		return false
	}
	if mountPoint == string(filepath.Separator) {
		return strings.HasPrefix(target, mountPoint)
	}
	return target == mountPoint || strings.HasPrefix(target, mountPoint+string(filepath.Separator))
}

func sortMountEntries(mounts []*mountEntry) {
	sort.SliceStable(mounts, func(i, j int) bool {
		return len(mounts[i].mountPoint) > len(mounts[j].mountPoint)
	})
}

func readMountNamespace(pidPath string) mountNamespace {
	var stat unix.Stat_t
	if err := unix.Stat(filepath.Join(pidPath, "ns", "mnt"), &stat); err != nil {
		return mountNamespace{}
	}
	return mountNamespace{dev: stat.Dev, ino: stat.Ino}
}

func scanSharedLibraries(pidPath string, filter *regexp.Regexp) []string {
	// Path is derived from procfs roots managed by this package.
	//nolint:gosec
	file, err := os.Open(filepath.Join(pidPath, "maps"))
	if err != nil {
		return nil
	}

	libs := make([]string, 0, 8)
	seen := make(map[string]struct{})
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		rawPath, ok := parseProcMapsPathname(line)
		if !ok {
			continue
		}

		pathname := unescapeProcPathField(rawPath)
		pathname = normalizeProcfsBinaryPath(pathname)
		if pathname == "" {
			continue
		}
		if filter != nil && !filter.MatchString(pathname) {
			continue
		}
		if _, ok := seen[pathname]; ok {
			continue
		}
		seen[pathname] = struct{}{}
		libs = append(libs, pathname)
	}

	_ = file.Close()

	return libs
}

func parseProcMapsPathname(line []byte) ([]byte, bool) {
	field := 0
	start := 0

	for start < len(line) {
		for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
			start++
		}
		if start >= len(line) {
			return nil, false
		}
		if field == 5 {
			return bytes.TrimSpace(line[start:]), true
		}

		end := start
		for end < len(line) && line[end] != ' ' && line[end] != '\t' {
			end++
		}
		field++
		start = end
	}

	return nil, false
}

func findLoadedLibraryHostPaths(filter *regexp.Regexp) map[string]struct{} {
	entries, err := os.ReadDir(HostProc())
	if err != nil {
		return nil
	}

	found := make(map[string]struct{})
	seenNSPaths := make(map[string]struct{})

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}

		pidPath := HostProc(entry.Name())
		ns := readMountNamespace(pidPath)
		for _, libPath := range scanSharedLibraries(pidPath, filter) {
			key := strconv.FormatUint(ns.dev, 10) + ":" +
				strconv.FormatUint(ns.ino, 10) + ":" + libPath
			if _, ok := seenNSPaths[key]; ok {
				continue
			}
			seenNSPaths[key] = struct{}{}

			hostPath := sharedResolverCache.Resolve(pidPath, libPath)
			hostPath = normalizeProcfsBinaryPath(hostPath)
			if hostPath != "" && isRegularFile(hostPath) {
				found[hostPath] = struct{}{}
			}
		}
	}

	return found
}

func readProcessCmdline(pid int) []string {
	if pid <= 0 {
		return nil
	}

	data, err := os.ReadFile(HostProc(strconv.Itoa(pid), "cmdline"))
	if err != nil || len(data) == 0 {
		return nil
	}

	return parseProcessCmdline(data)
}

func readProcessCmdlineFromProcFD(dirfd int) []string {
	data, err := readProcFileAt(dirfd, "cmdline")
	if err != nil || len(data) == 0 {
		return nil
	}

	return parseProcessCmdline(data)
}

func parseProcessCmdline(data []byte) []string {
	if len(data) == 0 {
		return nil
	}

	parts := bytes.Split(data, []byte{0})
	args := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		args = append(args, string(part))
	}

	return args
}

func listProcessIDs() ([]int, error) {
	entries, err := os.ReadDir(HostProc())
	if err != nil {
		return nil, err
	}

	pids := make([]int, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func readProcessStat(pid int) (string, int, uint64, error) {
	if pid <= 0 {
		return "", 0, 0, fmt.Errorf("invalid pid %d", pid)
	}

	data, err := os.ReadFile(HostProc(strconv.Itoa(pid), "stat"))
	if err != nil {
		return "", 0, 0, err
	}

	return parseProcessStatData(pid, data)
}

func readProcessStatFromProcFD(dirfd int, pid int) (string, int, uint64, error) {
	if dirfd < 0 {
		return "", 0, 0, fmt.Errorf("invalid proc dir fd for pid %d", pid)
	}
	data, err := readProcFileAt(dirfd, "stat")
	if err != nil {
		return "", 0, 0, err
	}
	return parseProcessStatData(pid, data)
}

func parseProcessStatData(pid int, data []byte) (string, int, uint64, error) {
	if len(data) == 0 {
		return "", 0, 0, fmt.Errorf("empty stat for pid %d", pid)
	}

	line := string(bytes.TrimSpace(data))
	start := strings.IndexByte(line, '(')
	end := strings.LastIndexByte(line, ')')
	if start < 0 || end <= start {
		return "", 0, 0, fmt.Errorf("invalid stat format for pid %d", pid)
	}

	name := line[start+1 : end]
	rest := strings.Fields(line[end+1:])
	if len(rest) < 20 {
		return "", 0, 0, fmt.Errorf("invalid stat field count for pid %d", pid)
	}

	ppid, err := strconv.Atoi(rest[1])
	if err != nil {
		return "", 0, 0, err
	}

	startTime, err := strconv.ParseUint(rest[19], 10, 64)
	if err != nil {
		return "", 0, 0, err
	}

	return name, ppid, startTime, nil
}

func readProcessStartTime(pid int) (uint64, error) {
	_, _, startTime, err := readProcessStat(pid)
	if err != nil {
		return 0, err
	}
	return startTime, nil
}

func readProcessStartTimeFromProcFD(dirfd int, pid int) (uint64, error) {
	_, _, startTime, err := readProcessStatFromProcFD(dirfd, pid)
	if err != nil {
		return 0, err
	}
	return startTime, nil
}

func readProcessEnvironMapForKeys(pid int, keys map[string]struct{}) map[string]string {
	if pid <= 0 {
		return nil
	}

	data, err := os.ReadFile(HostProc(strconv.Itoa(pid), "environ"))
	if err != nil || len(data) == 0 {
		return nil
	}

	return parseProcessEnvironMapForKeys(data, keys)
}

func readProcessEnvironMapForKeysFromProcFD(dirfd int, keys map[string]struct{}) map[string]string {
	data, err := readProcFileAt(dirfd, "environ")
	if err != nil || len(data) == 0 {
		return nil
	}

	return parseProcessEnvironMapForKeys(data, keys)
}

func parseProcessEnvironMapForKeys(data []byte, keys map[string]struct{}) map[string]string {
	if len(data) == 0 {
		return nil
	}

	env := make(map[string]string, minInt(len(keys), 8))
	start := 0
	for start < len(data) {
		end := start
		for end < len(data) && data[end] != 0 {
			end++
		}
		part := data[start:end]
		start = end + 1

		if len(part) == 0 {
			continue
		}

		idx := bytes.IndexByte(part, '=')
		if idx <= 0 {
			continue
		}

		key := string(part[:idx])
		if len(keys) > 0 {
			if _, ok := keys[key]; !ok {
				continue
			}
		}

		env[key] = string(part[idx+1:])
		if len(keys) > 0 && len(env) == len(keys) {
			break
		}
	}

	if len(env) == 0 {
		return nil
	}
	return env
}

func readProcessExePath(pid int) (string, error) {
	if pid <= 0 {
		return "", fmt.Errorf("invalid pid %d", pid)
	}
	path, err := os.Readlink(HostProc(strconv.Itoa(pid), "exe"))
	if err != nil {
		return "", err
	}

	return normalizeProcessLinkTarget(pid, "executable", path)
}

func readProcessExePathFromProcFD(dirfd int, pid int) (string, error) {
	path, err := readProcSymlinkAt(dirfd, "exe")
	if err != nil {
		return "", err
	}

	return normalizeProcessLinkTarget(pid, "executable", path)
}

func normalizeProcessLinkTarget(pid int, kind string, path string) (string, error) {
	rawPath := strings.TrimSpace(path)
	path = normalizeProcfsBinaryPath(rawPath)
	if path == "" || strings.HasSuffix(rawPath, deletedPathSuffix) {
		return "", &nonRegularExecutablePathError{pid: pid, path: path}
	}
	return path, nil
}

func readProcessRootPath(pid int) (string, error) {
	if pid <= 0 {
		return "", fmt.Errorf("invalid pid %d", pid)
	}

	path, err := os.Readlink(HostProc(strconv.Itoa(pid), "root"))
	if err != nil {
		return "", err
	}

	path = normalizeProcfsBinaryPath(path)
	if path == "" {
		return "", fmt.Errorf("invalid root path for pid %d", pid)
	}
	return path, nil
}

func readProcessRootPathFromProcFD(dirfd int, pid int) (string, error) {
	path, err := readProcSymlinkAt(dirfd, "root")
	if err != nil {
		return "", err
	}

	path = normalizeProcfsBinaryPath(path)
	if path == "" {
		return "", fmt.Errorf("invalid root path for pid %d", pid)
	}
	return path, nil
}

func openProcessDir(pid int) (int, error) {
	if pid <= 0 {
		return -1, fmt.Errorf("invalid pid %d", pid)
	}
	return unix.Open(HostProc(strconv.Itoa(pid)), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
}

func readProcFileAt(dirfd int, name string) ([]byte, error) {
	if dirfd < 0 {
		return nil, fmt.Errorf("invalid dirfd")
	}
	fd, err := unix.Openat(dirfd, name, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	if file == nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("wrap proc file %s", name)
	}

	data, readErr := io.ReadAll(file)
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

func readProcSymlinkAt(dirfd int, name string) (string, error) {
	if dirfd < 0 {
		return "", fmt.Errorf("invalid dirfd")
	}

	bufSize := 256
	for {
		buf := make([]byte, bufSize)
		n, err := unix.Readlinkat(dirfd, name, buf)
		if err != nil {
			return "", err
		}
		if n < len(buf) {
			return string(buf[:n]), nil
		}
		bufSize *= 2
		if bufSize > 64*1024 {
			return "", fmt.Errorf("readlinkat %s too large", name)
		}
	}
}

func normalizeProcfsBinaryPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimSuffix(path, deletedPathSuffix)
	if path == "" || !filepath.IsAbs(path) {
		return ""
	}

	clean := filepath.Clean(path)
	if clean == "." || clean == string(filepath.Separator) {
		return ""
	}
	return clean
}

func unescapeProcPathField(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	buf := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		if raw[i] == '\\' && i+3 < len(raw) && isOctalDigit(raw[i+1]) && isOctalDigit(raw[i+2]) && isOctalDigit(raw[i+3]) {
			value := (raw[i+1]-'0')<<6 | (raw[i+2]-'0')<<3 | (raw[i+3] - '0')
			buf = append(buf, value)
			i += 3
			continue
		}
		buf = append(buf, raw[i])
	}
	return string(buf)
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func isProcessGoneError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, os.ErrNotExist)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isNonRegularExecutablePathError(err error) bool {
	if err == nil {
		return false
	}
	var target *nonRegularExecutablePathError
	return errors.As(err, &target)
}
