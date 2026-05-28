//go:build linux
// +build linux

package procwatch

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNormalizeProcfsBinaryPath(t *testing.T) {
	cases := map[string]string{
		"":                         "",
		"relative/path":            "",
		"[vdso]":                   "",
		"/tmp/demo (deleted)":      "/tmp/demo",
		" /usr/bin/bash ":          "/usr/bin/bash",
		"/usr/bin/../bin/python3":  "/usr/bin/python3",
		string(filepath.Separator): "",
	}

	for input, want := range cases {
		if got := normalizeProcfsBinaryPath(input); got != want {
			t.Fatalf("normalizeProcfsBinaryPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsRegularFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app")
	if err := os.WriteFile(file, []byte("bin"), 0o755); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	if !isRegularFile(file) {
		t.Fatalf("expected regular file for %s", file)
	}
	if isRegularFile(dir) {
		t.Fatalf("expected directory %s to be rejected", dir)
	}
	if isRegularFile(filepath.Join(dir, "missing")) {
		t.Fatal("expected missing file to be rejected")
	}
}

func TestReadLimitedProcFileRejectsOversizedInput(t *testing.T) {
	data, err := readLimitedProcFile(strings.NewReader("abcd"), 4)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(data) != "abcd" {
		t.Fatalf("read data = %q, want abcd", data)
	}

	if _, err := readLimitedProcFile(strings.NewReader("abcde"), 4); err == nil {
		t.Fatal("expected oversized proc file to be rejected")
	}
}

func TestReadTruncatedProcFileKeepsPrefix(t *testing.T) {
	data, err := readTruncatedProcFile(strings.NewReader("A=1\x00B=2345"), 8)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(data) != "A=1\x00" {
		t.Fatalf("read data = %q, want first complete env entry", data)
	}

	data, err = readTruncatedProcFile(strings.NewReader("A=1"), 8)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(data) != "A=1" {
		t.Fatalf("read data = %q, want full untruncated data", data)
	}
}

func TestReadProcessEnvironMapForKeysKeepsEarlyKeysWhenOversized(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	pidPath := filepath.Join(procRoot, "123")
	if err := os.MkdirAll(pidPath, 0o755); err != nil {
		t.Fatal(err)
	}
	data := append([]byte("SERVICE_NAME=api\x00"), []byte(strings.Repeat("x", maxProcessEnvironReadBytes+1))...)
	if err := os.WriteFile(filepath.Join(pidPath, "environ"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	env := readProcessEnvironMapForKeys(123, map[string]struct{}{"SERVICE_NAME": {}})
	if got := env["SERVICE_NAME"]; got != "api" {
		t.Fatalf("SERVICE_NAME = %q, want api; env=%#v", got, env)
	}
}

func TestScanSharedLibrariesHonorsLineLimit(t *testing.T) {
	t.Setenv(procMapsScanLineLimitEnv, "1")

	pidPath := t.TempDir()
	maps := strings.Join([]string{
		"00400000-00452000 r--p 00000000 00:00 0 /tmp/liba.so",
		"00600000-00652000 r--p 00000000 00:00 0 /tmp/libb.so",
	}, "\n")
	if err := os.WriteFile(filepath.Join(pidPath, "maps"), []byte(maps), 0o644); err != nil {
		t.Fatal(err)
	}

	libs := scanSharedLibraries(pidPath, regexp.MustCompile(`\.so$`))
	if len(libs) != 1 || libs[0] != "/tmp/liba.so" {
		t.Fatalf("expected one scanned library, got %#v", libs)
	}
}

func TestLimitLibraryScanPIDsSpreadsAcrossProcTable(t *testing.T) {
	pids := make([]int, 0, 100)
	for i := 1; i <= 100; i++ {
		pids = append(pids, i)
	}

	got := limitLibraryScanPIDs(pids, 5)
	want := []int{1, 21, 41, 61, 81}
	if len(got) != len(want) {
		t.Fatalf("sample len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sample[%d] = %d, want %d; got=%#v", i, got[i], want[i], got)
		}
	}
}

func TestFindLoadedLibraryHostPathsHonorsPIDLimit(t *testing.T) {
	t.Setenv(libraryScanPIDLimitEnv, "1")
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)
	nsFile := filepath.Join(procRoot, "mntns")
	if err := os.WriteFile(nsFile, []byte("ns"), 0o644); err != nil {
		t.Fatal(err)
	}

	libA := filepath.Join(procRoot, "liba.so")
	libB := filepath.Join(procRoot, "libb.so")
	if err := os.WriteFile(libA, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(libB, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeMaps := func(pid, lib string) {
		pidPath := filepath.Join(procRoot, pid)
		if err := os.MkdirAll(filepath.Join(pidPath, "ns"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(nsFile, filepath.Join(pidPath, "ns", "mnt")); err != nil {
			t.Fatal(err)
		}
		line := "00400000-00452000 r--p 00000000 00:00 0 " + lib
		if err := os.WriteFile(filepath.Join(pidPath, "maps"), []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeMaps("100", libA)
	writeMaps("101", libB)

	oldResolver := sharedResolverCache
	defer func() { sharedResolverCache = oldResolver }()
	sharedResolverCache = &pathResolverCache{
		rootMounts: &mountSnapshot{mounts: []*mountEntry{{mountPoint: string(filepath.Separator)}}},
		rootNS:     readMountNamespace(filepath.Join(procRoot, "100")),
		rootByDev:  map[string][]*mountEntry{},
		nsMounts:   map[mountNamespace]*mountSnapshot{},
		resolved:   map[mountNamespace]map[string]resolvedPathCacheEntry{},
		lastAccess: map[mountNamespace]time.Time{},
	}

	found := findLoadedLibraryHostPaths(regexp.MustCompile(`\.so$`))
	if _, ok := found[libA]; !ok {
		t.Fatalf("expected first pid library %s, got %#v", libA, found)
	}
	if _, ok := found[libB]; ok {
		t.Fatalf("expected second pid to be skipped by limit, got %#v", found)
	}
}

func TestParseMountInfoLine(t *testing.T) {
	line := []byte("36 25 0:32 /docker/containers /var/lib/docker rw,relatime - overlay overlay rw")

	mount, ok := parseMountInfoLine(line)
	if !ok {
		t.Fatal("expected mountinfo line to parse")
	}
	if mount.dev != "0:32" || mount.root != "/docker/containers" || mount.mountPoint != "/var/lib/docker" {
		t.Fatalf("unexpected mountinfo parse result: dev=%q root=%q mountPoint=%q", mount.dev, mount.root, mount.mountPoint)
	}
}

func TestParseMountInfoLineUnescapesPaths(t *testing.T) {
	line := []byte(`36 25 0:32 /docker\040root /var/lib/container\040mount rw,relatime - overlay overlay rw`)

	mount, ok := parseMountInfoLine(line)
	if !ok {
		t.Fatal("expected escaped mountinfo line to parse")
	}
	if mount.dev != "0:32" || mount.root != "/docker root" || mount.mountPoint != "/var/lib/container mount" {
		t.Fatalf("unexpected escaped mountinfo parse result: dev=%q root=%q mountPoint=%q", mount.dev, mount.root, mount.mountPoint)
	}
}

func TestResolveMountPath(t *testing.T) {
	rootMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "0:32", root: "/docker/containers/abcd/rootfs", mountPoint: "/var/lib/docker/containers/abcd/rootfs"},
			{dev: "8:1", root: "/", mountPoint: "/"},
		},
	}
	nsMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "0:32", root: "/docker/containers/abcd/rootfs", mountPoint: "/"},
		},
	}

	got := resolveMountPath(rootMounts, indexMountsByDev(rootMounts), nsMounts, "/usr/bin/app")
	want := "/var/lib/docker/containers/abcd/rootfs/usr/bin/app"
	if got != want {
		t.Fatalf("resolveMountPath() = %q, want %q", got, want)
	}
}

func TestMountSnapshotFindUsesPathBoundary(t *testing.T) {
	snapshot := &mountSnapshot{
		mounts: []*mountEntry{
			{mountPoint: "/var/lib/docker"},
			{mountPoint: "/var/lib/docker2"},
			{mountPoint: "/"},
		},
	}

	if got := snapshot.find("/var/lib/docker2/overlay2/merged/usr/bin/app"); got == nil || got.mountPoint != "/var/lib/docker2" {
		t.Fatalf("expected longest boundary-safe match, got %+v", got)
	}
	if got := snapshot.find("/var/lib/dockerx/usr/bin/app"); got == nil || got.mountPoint != "/" {
		t.Fatalf("expected root mount fallback for sibling prefix, got %+v", got)
	}
}

func TestRootMountCandidatesUsesDeviceIndex(t *testing.T) {
	rootMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "0:32", root: "/docker", mountPoint: "/var/lib/docker"},
			{dev: "8:1", root: "/", mountPoint: "/"},
		},
	}

	candidates := rootMountCandidates(rootMounts, indexMountsByDev(rootMounts), "0:32")
	if len(candidates) != 1 || candidates[0].mountPoint != "/var/lib/docker" {
		t.Fatalf("unexpected candidates: %+v", candidates)
	}
}

func TestReadProcessEnvironMapForKeys(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	pidDir := filepath.Join(procRoot, "123")
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		t.Fatal(err)
	}

	payload := []byte("DD_SERVICE=checkout\x00IGNORED=value\x00OTEL_SERVICE_NAME=frontend\x00")
	if err := os.WriteFile(filepath.Join(pidDir, "environ"), payload, 0o644); err != nil {
		t.Fatal(err)
	}

	env := readProcessEnvironMapForKeys(123, map[string]struct{}{
		"DD_SERVICE":        {},
		"OTEL_SERVICE_NAME": {},
	})
	if len(env) != 2 {
		t.Fatalf("unexpected env size: %d", len(env))
	}
	if env["DD_SERVICE"] != "checkout" {
		t.Fatalf("unexpected DD_SERVICE value: %q", env["DD_SERVICE"])
	}
	if env["OTEL_SERVICE_NAME"] != "frontend" {
		t.Fatalf("unexpected OTEL_SERVICE_NAME value: %q", env["OTEL_SERVICE_NAME"])
	}
	if _, ok := env["IGNORED"]; ok {
		t.Fatal("expected unselected env var to be skipped")
	}
}

func TestScanSharedLibrariesPreservesSpacesInPath(t *testing.T) {
	pidPath := t.TempDir()
	mapsPath := filepath.Join(pidPath, "maps")
	line := "7f0000000000-7f0000001000 r-xp 00000000 08:02 12345 /tmp/lib with space.so\n"
	if err := os.WriteFile(mapsPath, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	libs := scanSharedLibraries(pidPath, nil)
	if len(libs) != 1 || libs[0] != "/tmp/lib with space.so" {
		t.Fatalf("unexpected libs: %+v", libs)
	}
}

func TestUnescapeProcPathField(t *testing.T) {
	raw := []byte(`/var/lib/docker\040merged/foo\134bar\012baz`)
	if got := unescapeProcPathField(raw); got != "/var/lib/docker merged/foo\\bar\nbaz" {
		t.Fatalf("unexpected unescaped path: %q", got)
	}
}

func TestResolveHostBinaryPathUsesProcRootFirst(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4321
	pidDir := filepath.Join(procRoot, "4321")
	rootfs := filepath.Join(procRoot, "rootfs")
	binPath := filepath.Join(rootfs, "usr", "bin", "app")

	if err := os.MkdirAll(filepath.Join(pidDir, "ns"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidDir, "ns", "mnt"), []byte("mntns"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(rootfs, filepath.Join(pidDir, "root")); err != nil {
		t.Fatal(err)
	}

	got := resolveHostBinaryPath(pid, "/usr/bin/app")
	if got != binPath {
		t.Fatalf("resolveHostBinaryPath() = %q, want %q", got, binPath)
	}
}

func TestResolveHostBinaryPathUsesProcRootPortal(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 5321
	binPath := filepath.Join(procRoot, "5321", "root", "usr", "bin", "app")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := resolveHostBinaryPath(pid, "/usr/bin/app")
	if got != binPath {
		t.Fatalf("resolveHostBinaryPath() = %q, want %q", got, binPath)
	}
}

func TestResolveMountPathOverlayReturnsHostLayerFile(t *testing.T) {
	layer1 := t.TempDir()
	layer2 := t.TempDir()
	binPath := filepath.Join(layer2, "usr", "local", "bin", "smrepro")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	nsMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{
				dev:        "0:44",
				root:       "/",
				mountPoint: "/",
				fsType:     "overlay",
				options: map[string]string{
					"lowerdir": layer1 + ":" + layer2,
				},
			},
		},
	}

	got := resolveMountPath(&mountSnapshot{}, nil, nsMounts, "/usr/local/bin/smrepro")
	if got != binPath {
		t.Fatalf("resolveMountPath() = %q, want %q", got, binPath)
	}
}

func TestSplitEscapedMountPathListPreservesEscapedSeparator(t *testing.T) {
	raw := `/var/lib/containerd/layer\072one:/var/lib/containerd/layer-two`
	got := splitEscapedMountPathList(raw, ':')
	want := []string{
		"/var/lib/containerd/layer:one",
		"/var/lib/containerd/layer-two",
	}
	if len(got) != len(want) {
		t.Fatalf("splitEscapedMountPathList() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splitEscapedMountPathList()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolveMountPathOverlayHandlesEscapedLowerdirPath(t *testing.T) {
	base := t.TempDir()
	layer1 := filepath.Join(base, "layer:one")
	layer2 := filepath.Join(base, "layer-two")
	binPath := filepath.Join(layer1, "usr", "local", "bin", "smrepro")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layer2, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	escapedLayer1 := strings.ReplaceAll(layer1, ":", `\072`)
	nsMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{
				dev:        "0:44",
				root:       "/",
				mountPoint: "/",
				fsType:     "overlay",
				options: map[string]string{
					"lowerdir": escapedLayer1 + ":" + layer2,
				},
			},
		},
	}

	got := resolveMountPath(&mountSnapshot{}, nil, nsMounts, "/usr/local/bin/smrepro")
	if got != binPath {
		t.Fatalf("resolveMountPath() = %q, want %q", got, binPath)
	}
}

func TestNormalizeProcessLinkTargetAllowsContainerAbsolutePath(t *testing.T) {
	got, err := normalizeProcessLinkTarget(1234, "executable", "/usr/local/bin/smrepro")
	if err != nil {
		t.Fatalf("normalizeProcessLinkTarget() error = %v", err)
	}
	if got != "/usr/local/bin/smrepro" {
		t.Fatalf("normalizeProcessLinkTarget() = %q, want %q", got, "/usr/local/bin/smrepro")
	}
}

func TestNormalizeProcessLinkTargetRejectsDeletedBinary(t *testing.T) {
	got, err := normalizeProcessLinkTarget(1234, "executable", "/usr/local/bin/smrepro (deleted)")
	if got != "" {
		t.Fatalf("normalizeProcessLinkTarget() = %q, want empty", got)
	}
	if !isNonRegularExecutablePathError(err) {
		t.Fatalf("expected non-regular executable error, got %v", err)
	}
}

func TestResolveDoesNotCacheEmptyPath(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "200")
	if err := os.MkdirAll(filepath.Join(pidPath, "ns"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidPath, "ns", "mnt"), []byte("pid-ns"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidPath, "mountinfo"), []byte("36 25 0:32 /container / rw - overlay overlay rw\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rootMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "8:1", root: "/", mountPoint: "/"},
		},
	}
	cache := &pathResolverCache{
		rootMounts: rootMounts,
		rootNS:     mountNamespace{dev: 1, ino: 1},
		rootByDev:  indexMountsByDev(rootMounts),
		nsMounts:   make(map[mountNamespace]*mountSnapshot),
		resolved:   make(map[mountNamespace]map[string]resolvedPathCacheEntry),
		lastAccess: make(map[mountNamespace]time.Time),
	}

	if got := cache.Resolve(pidPath, "/usr/bin/app"); got != "" {
		t.Fatalf("Resolve() = %q, want empty path", got)
	}

	ns := readMountNamespace(pidPath)
	if ns == (mountNamespace{}) {
		t.Fatal("expected test namespace to be readable")
	}
	if cache.resolved[ns] != nil {
		t.Fatalf("expected empty resolve result to avoid cache entry, got %+v", cache.resolved[ns])
	}
}

func TestMaybeCleanupLockedDropsExpiredNamespaceEntries(t *testing.T) {
	ns := mountNamespace{dev: 2, ino: 3}
	now := time.Now()
	cache := &pathResolverCache{
		nsMounts: map[mountNamespace]*mountSnapshot{
			ns: {mounts: []*mountEntry{{dev: "0:32", root: "/", mountPoint: "/"}}},
		},
		resolved: map[mountNamespace]map[string]resolvedPathCacheEntry{
			ns: {
				"/usr/bin/app": {path: "/host/app", ts: now.Add(-resolvedPathCacheTTL - time.Second)},
			},
		},
		lastAccess: map[mountNamespace]time.Time{
			ns: now.Add(-resolvedPathNamespaceTTL - time.Second),
		},
	}

	cache.maybeCleanupLocked(now)

	if _, ok := cache.nsMounts[ns]; ok {
		t.Fatal("expected expired namespace mounts to be removed")
	}
	if _, ok := cache.resolved[ns]; ok {
		t.Fatal("expected expired resolve cache to be removed")
	}
	if _, ok := cache.lastAccess[ns]; ok {
		t.Fatal("expected expired namespace access record to be removed")
	}
}
