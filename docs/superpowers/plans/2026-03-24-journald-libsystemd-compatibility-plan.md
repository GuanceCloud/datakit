# Journald libsystemd 兼容性检测实施计划

> **For agentic workers:** 必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 按任务逐步执行本计划。步骤使用复选框（`- [ ]`）语法跟踪。

**目标：** 为 journald external collector 增加启动期兼容性检测，使主机 journal 文件与当前运行时不兼容时，collector 会明确告警并保持 inactive，而不是静默零采集。

**架构：** 变更主要位于 `internal/plugins/externals/journald/collect/` 和 `internal/plugins/inputs/journald/`。在现有诊断层之外，再引入一段启动前 host library prepare 逻辑：当显式配置项开启时，DataKit 在启动 journald external collector 前，无条件从 `/rootfs` 的候选系统库目录复制配置指定的动态库文件列表到 DataKit 自己的 `external-libs` 目录，并自动把 `LD_LIBRARY_PATH` 优先指向该目录。之后仍然通过 probe 决定 collector 是否继续启动。

**技术栈：** Go、`github.com/coreos/go-systemd/v22/sdjournal`、标准库文件系统工具、现有 journald 单元测试、Markdown 文档。

**测试环境：** journald collector package 仅支持 Linux。计划中的 `go test ./internal/plugins/externals/journald/...` 和 `go test ./internal/plugins/inputs/journald` 命令都必须在 Linux 环境、Linux 容器或 Linux CI 中执行，不能直接在 macOS 上验证。

---

## 文件结构

- Create: `docs/superpowers/plans/2026-03-24-journald-libsystemd-compatibility-plan.md`
- Create: `internal/plugins/externals/journald/collect/diagnose.go`
- Create: `internal/plugins/externals/journald/collect/diagnose_test.go`
- Modify: `internal/plugins/inputs/journald/input.go`
- Modify: `internal/plugins/inputs/journald/input_test.go`
- Modify: `internal/plugins/inputs/journald/sample.go`
- Modify: `internal/plugins/externals/journald/collect/journald.go`
- Modify: `internal/plugins/externals/journald/collect/journald_funcs_test.go`
- Modify: `internal/export/doc/en/inputs/journald.md`
- Modify: `internal/export/doc/zh/inputs/journald.md`

`diagnose.go` 负责兼容性诊断相关类型和辅助逻辑，使 `journald.go` 继续只聚焦 collector 生命周期。`internal/plugins/inputs/journald/` 负责新增配置项、启动前的 host library copy prepare，以及把 `LD_LIBRARY_PATH` 定向到 `external-libs`。已有测试文件可继续承载路径、cursor 等低层逻辑，新增的 `diagnose_test.go` 负责结果分类和启动降级测试。

### Task 0: 增加 host library copy 配置与启动前准备

**Files:**
- Modify: `internal/plugins/inputs/journald/input.go`
- Modify: `internal/plugins/inputs/journald/input_test.go`
- Modify: `internal/plugins/inputs/journald/sample.go`

- [ ] **Step 1: 先写失败测试，覆盖新配置项和启动前 copy 语义**

测试至少覆盖：
- `copy_node_libs = false` 时，不追加 host library prepare 逻辑
- `copy_node_libs = true` 时，无条件触发 copy，不以兼容性判断作为前置条件
- `copy_node_libs_files` 可配置，并会传递给 copy helper
- `copy_node_libs_files` 未填写或为空时，会自动回退到程序内置默认列表
- 当 copy 启用时，external collector 的 `envs` 会自动追加 `LD_LIBRARY_PATH=<external-libs>:$LD_LIBRARY_PATH`

- [ ] **Step 2: 实现配置项与 prepare helper**

新增配置项：
- `copy_node_libs`：`bool`，默认 `false`
- `copy_node_libs_files`：`[]string`，支持显式覆盖；如果未填写或为空，则程序必须使用一组内置默认动态库文件列表

实现要求：
- 一旦 `copy_node_libs = true`，启动 external collector 前就从 `/rootfs` 的候选系统库目录无条件复制 `copy_node_libs_files` 指定的文件到 `external-libs`
- 如果 `copy_node_libs_files` 未填写或为空，则回退到内置默认列表执行 copy
- copy 阶段本身不做兼容性判断，不根据 probe 结果决定是否 copy
- 复制失败要有明确日志，但是否继续启动仍由设计约定决定
- `LD_LIBRARY_PATH` 自动优先指向 `external-libs`
- `sample.go` 中应直接给出 `copy_node_libs_files` 的默认文件列表示例，方便用户按需覆盖

首版内置默认列表至少包含：
- `libsystemd.so*`
- `liblz4.so*`
- `libzstd.so*`
- `liblzma.so*`
- `libcap.so*`
- `libgcrypt.so*`
- `libgpg-error.so*`
- `libselinux.so*`
- `libmount.so*`
- `libblkid.so*`
- `libacl.so*`
- `libpcre2-8.so*`

- [ ] **Step 3: 运行 journald input 层测试**

Run in Linux: `go test ./internal/plugins/inputs/journald -v`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add internal/plugins/inputs/journald/input.go internal/plugins/inputs/journald/input_test.go internal/plugins/inputs/journald/sample.go
git commit -m "feat: add journald host library copy prepare config"
```

### Task 1: 增加显式诊断类型与测试 seam

**Files:**
- Create: `internal/plugins/externals/journald/collect/diagnose.go`
- Create: `internal/plugins/externals/journald/collect/diagnose_test.go`
- Modify: `internal/plugins/externals/journald/collect/journald.go`

- [ ] **Step 1: 先写失败测试，覆盖诊断结果与版本信息**

```go
func TestProbeSeverityOrder(t *testing.T) {
	results := []probeResult{
		{reason: reasonNoJournalFiles, target: "/a"},
		{reason: reasonUnsupportedFormat, target: "/b"},
	}

	got := selectProbeFailure(results)
	if got.reason != reasonUnsupportedFormat {
		t.Fatalf("reason = %s, want %s", got.reason, reasonUnsupportedFormat)
	}
}

func TestDetectReaderVersion_ParsesJournalctlVersion(t *testing.T) {
	runJournalctlVersion = func() ([]byte, error) {
		return []byte("systemd 249 (249.11-0ubuntu3.19)\n"), nil
	}
	t.Cleanup(func() { runJournalctlVersion = defaultRunJournalctlVersion })

	got := detectReaderVersion()
	if got != "249" {
		t.Fatalf("version = %q, want %q", got, "249")
	}
}
```

- [ ] **Step 2: 运行定向测试，确认它们先失败**

Run: `go test ./internal/plugins/externals/journald/collect -run 'TestProbeSeverityOrder|TestDetectReaderVersion_ParsesJournalctlVersion' -v`
Expected: FAIL，因为当前还没有诊断结果类型和版本检测 helper。

- [ ] **Step 3: 实现最小诊断骨架**

```go
type probeReason string

const (
	reasonOK                probeReason = "ok"
	reasonUnsupportedFormat probeReason = "unsupported-format"
	reasonPermissionDenied  probeReason = "permission-denied"
	reasonUnexpectedOpen    probeReason = "unexpected-open-error"
	reasonNoJournalFiles    probeReason = "no-journal-files"
)

type probeResult struct {
	reason  probeReason
	target  string
	message string
}

var runJournalctlVersion = defaultRunJournalctlVersion
```

实现内容：
- 显式严重级别排序
- 确定性的失败选择 helper
- `detectReaderVersion()`，在 `journalctl --version` 不可用或无法解析时返回 `""`
- 不应假设 Kubernetes 中的 DataKit 容器一定安装了 `journalctl`；缺失 `journalctl` 只能让 reader version 为空，不能影响 probe 和启动 gating
- 明确的 package-level 测试 seam，例如：

```go
var (
	systemdCheckFn = checkSystemd
	probeJournalTargetFn = probeJournalTarget
	initJournalFn = func(ipt *Input) error { return ipt.initJournal() }
	doCollectFn = func(ipt *Input) []*point.Point { return ipt.doCollect() }
)
```

- [ ] **Step 4: 再次运行定向测试，确认通过**

Run: `go test ./internal/plugins/externals/journald/collect -run 'TestProbeSeverityOrder|TestDetectReaderVersion_ParsesJournalctlVersion' -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/plugins/externals/journald/collect/diagnose.go internal/plugins/externals/journald/collect/diagnose_test.go internal/plugins/externals/journald/collect/journald.go
git commit -m "feat: add journald compatibility diagnosis types"
```

### Task 2: 实现确定性的 journal 目标解析

**Files:**
- Modify: `internal/plugins/externals/journald/collect/diagnose.go`
- Modify: `internal/plugins/externals/journald/collect/diagnose_test.go`
- Modify: `internal/plugins/externals/journald/collect/journald_funcs_test.go`

- [ ] **Step 1: 先写失败测试，覆盖解析器**

```go
func TestResolveProbeTargets_ExpandsJournalRoot(t *testing.T) {
	root := t.TempDir()
	machineDir := filepath.Join(root, "machine-id-1")
	require.NoError(t, os.MkdirAll(machineDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(machineDir, "system.journal"), []byte("x"), 0o644))

	got := resolveProbeTargets([]string{root})
	require.Equal(t, []string{machineDir}, got)
}

func TestResolveProbeTargets_PreservesDirectJournalFile(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "system.journal")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))

	got := resolveProbeTargets([]string{file})
	require.Equal(t, []string{file}, got)
}

func TestResolveProbeTargets_PreservesDirectMachineIDDirectory(t *testing.T) {
	root := t.TempDir()
	machineDir := filepath.Join(root, "ec2f02bf505ce9dd2cc7dce0561ccd18")
	require.NoError(t, os.MkdirAll(machineDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(machineDir, "system.journal"), []byte("x"), 0o644))

	got := resolveProbeTargets([]string{machineDir})
	require.Equal(t, []string{machineDir}, got)
}
```

- [ ] **Step 2: 运行解析器相关测试，确认先失败**

Run in Linux: `go test ./internal/plugins/externals/journald/collect -run 'TestResolveProbeTargets|TestResolvePaths' -v`
Expected: FAIL，因为 `resolveProbeTargets` 还不存在。

- [ ] **Step 3: 实现目标展开逻辑**

```go
func resolveProbeTargets(paths []string) []string {
	// 1. 保留直接 *.journal 文件
	// 2. 保留直接 machine-id 目录
	// 3. 将 journal 根目录展开成子 machine-id 目录
	// 4. 排序并去重，保证探测顺序稳定
}
```

实现说明：
- 保留现有 `resolvePaths()` 对可访问配置路径的行为
- 新的 probe target resolver 要直接基于原始 `config.Paths`，不能基于 `resolvePaths()` 的结果
- 对于不可访问的配置路径，要保留足够信息，以便 preflight 区分 `permission-denied` 和 `no-journal-files`
- 包含 `*.journal` 文件的目录应视为可直接探测目标
- 返回顺序必须稳定，便于日志与测试具备确定性

- [ ] **Step 4: 运行解析器测试及已有 helper 测试**

Run: `go test ./internal/plugins/externals/journald/collect -run 'TestResolveProbeTargets|TestResolvePaths|TestLoadCursor|TestSaveCursor' -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/plugins/externals/journald/collect/diagnose.go internal/plugins/externals/journald/collect/diagnose_test.go internal/plugins/externals/journald/collect/journald_funcs_test.go
git commit -m "feat: resolve journald compatibility probe targets"
```

### Task 3: 接入启动探测与告警式降级

**Files:**
- Modify: `internal/plugins/externals/journald/collect/diagnose.go`
- Modify: `internal/plugins/externals/journald/collect/diagnose_test.go`
- Modify: `internal/plugins/externals/journald/collect/journald.go`

- [ ] **Step 1: 先写失败测试，覆盖启动行为**

```go
func TestRun_SkipsCollectionWhenProbeFails(t *testing.T) {
	systemdCheckFn = func() error { return nil }
	probeJournalTargetFn = func(target string) probeResult {
		return probeResult{reason: reasonUnsupportedFormat, target: target, message: "unsupported feature"}
	}
	t.Cleanup(func() {
		systemdCheckFn = checkSystemd
		probeJournalTargetFn = probeJournalTarget
	})

	ipt := &Input{config: &config{Paths: []string{"/rootfs/var/log/journal"}}, done: make(chan bool)}
	got := ipt.preflightCompatibility()
	require.Equal(t, reasonUnsupportedFormat, got.reason)
}

func TestRun_ExitsBeforeInitJournalWhenProbeFails(t *testing.T) {
	systemdCheckFn = func() error { return nil }
	probeJournalTargetFn = func(target string) probeResult {
		return probeResult{reason: reasonUnsupportedFormat, target: target, message: "unsupported feature"}
	}
	initJournalFn = func(*Input) error {
		t.Fatal("initJournal should not run when compatibility probe fails")
		return nil
	}
	doCollectFn = func(*Input) []*point.Point {
		t.Fatal("doCollect should not run when compatibility probe fails")
		return nil
	}
	t.Cleanup(func() {
		systemdCheckFn = checkSystemd
		probeJournalTargetFn = probeJournalTarget
		initJournalFn = func(ipt *Input) error { return ipt.initJournal() }
		doCollectFn = func(ipt *Input) []*point.Point { return ipt.doCollect() }
	})

	ipt := &Input{config: &config{Paths: []string{"/rootfs/var/log/journal"}}, done: make(chan bool)}
	ipt.Run()
}

func TestSelectProbeFailure_TieBreaksByProbeOrder(t *testing.T) {
	results := []probeResult{
		{reason: reasonPermissionDenied, target: "/a"},
		{reason: reasonPermissionDenied, target: "/b"},
	}

	got := selectProbeFailure(results)
	require.Equal(t, "/a", got.target)
}

func TestPreflightCompatibility_ClassifiesFailuresAndSuccess(t *testing.T) {
	tests := []struct {
		name    string
		results map[string]probeResult
		want    probeReason
	}{
		{name: "permission denied", results: map[string]probeResult{"/a": {reason: reasonPermissionDenied, target: "/a"}}, want: reasonPermissionDenied},
		{name: "no journal files", results: map[string]probeResult{"/a": {reason: reasonNoJournalFiles, target: "/a"}}, want: reasonNoJournalFiles},
		{name: "unexpected open error", results: map[string]probeResult{"/a": {reason: reasonUnexpectedOpen, target: "/a"}}, want: reasonUnexpectedOpen},
		{name: "one target succeeds", results: map[string]probeResult{"/a": {reason: reasonUnsupportedFormat, target: "/a"}, "/b": {reason: reasonOK, target: "/b"}}, want: reasonOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := aggregateProbeResults(tt.results)
			require.Equal(t, tt.want, got.reason)
		})
	}
}

func TestRun_LogsCompatibilityWarningDetails(t *testing.T) {
	var buf bytes.Buffer
	logger.InitRoot(&logger.Option{Level: "debug", Writer: &buf, Flags: 0})
	systemdCheckFn = func() error { return nil }
	probeJournalTargetFn = func(target string) probeResult {
		return probeResult{
			reason:  reasonUnsupportedFormat,
			target:  "/rootfs/var/log/journal/machine/system.journal",
			message: "unsupported feature",
		}
	}
	runJournalctlVersion = func() ([]byte, error) {
		return []byte("systemd 249 (249.11-0ubuntu3.19)\n"), nil
	}
	t.Cleanup(func() {
		systemdCheckFn = checkSystemd
		probeJournalTargetFn = probeJournalTarget
		runJournalctlVersion = defaultRunJournalctlVersion
	})

	ipt := &Input{config: &config{Paths: []string{"/rootfs/var/log/journal"}}, done: make(chan bool)}
	ipt.Run()

	out := buf.String()
	require.Contains(t, out, "reason=unsupported-format")
	require.Contains(t, out, "target=/rootfs/var/log/journal/machine/system.journal")
	require.Contains(t, out, "reader_version=249")
	require.Contains(t, out, "collector will stay inactive")
	require.Contains(t, out, "newer libsystemd")
}
```

- [ ] **Step 2: 运行启动相关测试，确认先失败**

Run in Linux: `go test ./internal/plugins/externals/journald/collect -run 'TestRun_SkipsCollectionWhenProbeFails|TestRun_ExitsBeforeInitJournalWhenProbeFails|TestSelectProbeFailure_TieBreaksByProbeOrder|TestPreflightCompatibility_ClassifiesFailuresAndSuccess|TestRun_LogsCompatibilityWarningDetails' -v`
Expected: FAIL，因为兼容性 preflight 还没有真正接到启动路径。

- [ ] **Step 3: 实现 probe 执行与启动 gating**

```go
func (ipt *Input) preflightCompatibility() probeResult {
	targets := resolveProbeTargets(ipt.config.Paths)
	version := detectReaderVersion()
	return collectProbeResults(targets, version)
}

func (ipt *Input) Run() {
	// 先执行现有 libsystemd 可用性检查
	// 再执行 compatibility preflight
	// 如果 result.reason != reasonOK，则记录 warning 并 return
	// 否则通过 initJournalFn(ipt) 和 doCollectFn(ipt) 继续正常路径
}
```

实现说明：
- 按确定性顺序探测目标
- 一旦有一个目标成功，就停止后续探测
- 如果没有成功结果，选择严重级别最高的失败
- probe 结果聚合必须基于原始配置路径构建，确保缺失和不可访问路径仍可诊断
- 启动检查应通过 Task 1 中定义的 package-level seam 接入
- 使用 `initJournal` / `doCollect` seam，使测试能够确定性地证明 `Run()` 在 probe 失败时会提前退出
- 显式分类 unsupported-format、permission-denied、no-journal-files 和兜底 unexpected error
- 输出一条醒目的 warning，包含 configured path、sampled target、reason、reader version（如有）以及 “collector will stay inactive”
- 添加日志断言测试，使诊断日志本身成为契约的一部分
- probe 成功后的 `initJournal()` 和 `doCollect()` 行为保持不变

- [ ] **Step 4: 跑完整个 journald collector 测试包**

Run in Linux: `go test ./internal/plugins/externals/journald/collect -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/plugins/externals/journald/collect/diagnose.go internal/plugins/externals/journald/collect/diagnose_test.go internal/plugins/externals/journald/collect/journald.go
git commit -m "feat: warn and disable journald on incompatible libsystemd"
```

### Task 4: 更新运维文档并做最终验证

**Files:**
- Modify: `internal/export/doc/en/inputs/journald.md`
- Modify: `internal/export/doc/zh/inputs/journald.md`

- [ ] **Step 1: 把文档缺失项先写成检查清单，并确认当前确实缺少这些内容**

确认文档当前尚未覆盖：
- 启动兼容性告警行为
- unsupported-format 诊断
- 不应把主机 `/usr/lib64` 直接作为 `LD_LIBRARY_PATH`

Run: `rg -n 'unsupported-format|LD_LIBRARY_PATH|collector will stay inactive|libsystemd' internal/export/doc/en/inputs/journald.md internal/export/doc/zh/inputs/journald.md`
Expected: 结果中缺少新 troubleshooting 文本。

- [ ] **Step 2: 更新英文和中文 journald 文档**

增加简洁说明，覆盖：
- 兼容性 preflight 行为
- 在 Kubernetes 中采集 node journal 时，当 Pod 内 `libsystemd` 低于宿主机 journal 格式要求时的典型症状
- 在 Kubernetes 容器中不应假设一定存在 `journalctl`；如果没有 `journalctl`，仍应以 DataKit 自身日志和 probe 结果作为诊断依据
- 支持的修复方向：
  - 从宿主机单独复制兼容的 `systemd` 相关库到独立目录后挂入 DataKit 容器
  - 在 collector 中优先从该独立目录加载 `libsystemd`
- 约束说明：
  - 宿主机上的 `libsystemd` 只是兼容性候选，不保证一定兼容当前 DataKit 使用的 journald external binary
  - 如果宿主机 `libsystemd` 版本过低，external binary 也可能在动态链接阶段因符号或版本不匹配而无法启动
  - 如果 `copy_node_libs` 开启，DataKit 会在启动前无条件复制 `copy_node_libs_files` 指定的动态库文件，但 copy 行为本身不代表这些库一定兼容
- 不支持的规避方式：不要把主机 `/usr/lib64` 直接放进 `LD_LIBRARY_PATH`

建议文案形态：

```md
If startup logs report `reason=unsupported-format`, the collector runtime is older than the target journal file format. In this case DataKit keeps the journald collector inactive and logs a warning instead of collecting partial or misleading results.
```

- [ ] **Step 3: 运行代码与文档的定向验证**

Run in Linux: `go test ./internal/plugins/externals/journald/collect -v`
Expected: PASS

Run: `git diff --check`
Expected: 无输出

- [ ] **Step 4: 提交**

```bash
git add internal/export/doc/en/inputs/journald.md internal/export/doc/zh/inputs/journald.md
git commit -m "docs: document journald libsystemd compatibility warnings"
```

### Task 5: 最终集成检查

**Files:**
- Modify: `internal/plugins/externals/journald/collect/diagnose.go`
- Modify: `internal/plugins/externals/journald/collect/diagnose_test.go`
- Modify: `internal/plugins/externals/journald/collect/journald.go`
- Modify: `internal/export/doc/en/inputs/journald.md`
- Modify: `internal/export/doc/zh/inputs/journald.md`

- [ ] **Step 1: 对受影响范围做最终验证**

Run in Linux: `go test ./internal/plugins/externals/journald/... -v`
Expected: PASS

Run in Linux: `go test ./internal/plugins/inputs/journald -v`
Expected: PASS

- [ ] **Step 2: 检查最终 diff 是否符合范围控制**

Run: `git diff --stat HEAD~4..HEAD`
Expected: 改动仅限 journald collector 逻辑、测试和相关文档。

- [ ] **Step 3: 如果过程中还有收尾修复，则补一个集成提交**

```bash
git add internal/plugins/externals/journald/collect internal/export/doc/en/inputs/journald.md internal/export/doc/zh/inputs/journald.md
git commit -m "test: finalize journald compatibility detection"
```

如果 Task 4 结束后分支已经是干净且已验证状态，则可以跳过这个 commit。
