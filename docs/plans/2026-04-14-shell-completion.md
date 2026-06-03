# DataKit Shell Completion 实施计划

> **面向执行代理：** 实施本计划时，必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 逐项执行任务，并使用复选框 `- [ ]` 跟踪进度。

**目标：** 统一采用 `datakit completion` / `datakit completion <bash|powershell|fish|zsh>` 作为补全入口，默认自动识别当前 shell 并安装到本机，同时允许用户显式指定 shell 覆盖自动识别结果；安装完成后打印实际安装路径，并移除对 `tool` completer flags 与手写静态补全脚本的依赖。除宿主机场景外，还要覆盖 DataKit 运行在 Docker 容器中的 completion 安装。

**架构：** 以 Cobra root command 作为 completion 生成的唯一事实来源，在 `completion` 子命令中统一处理 shell 自动识别、显式覆盖、四种 shell 的脚本生成、安装路径选择、文件写入和成功提示。安装路径选择需要同时覆盖宿主机与 Docker 容器场景，并根据运行环境选择合适的默认目标。遗留 completer 入口与静态脚本直接删除，不保留兼容层，通过 breaking change changelog 通知用户迁移。

**技术栈：** Go、Cobra、testify、现有 `cmd/datakit/cmd/*` 命令树、`internal/cmds` 遗留 completer 逻辑

---

## 文件结构

### 新增文件

- `cmd/datakit/cmd/completion_support.go`
  - 抽取共享 completion 生成 helper，并处理 shell 自动识别
- `cmd/datakit/cmd/completion_install.go`
  - 处理安装路径选择、文件写入、覆盖检查和成功提示
- `cmd/datakit/cmd/completion_install_test.go`
  - 覆盖安装路径、覆盖逻辑、print 模式和成功消息

### 修改文件

- `cmd/datakit/cmd/completion.go`
  - 扩展 shell 支持到 `bash`、`powershell`、`fish`、`zsh`，并增加自动识别、`--print`、`--path`、`--force`
- `cmd/datakit/cmd/completion_test.go`
  - 增加四种 shell 的参数校验与输出测试
- `cmd/datakit/cmd/root.go`
  - 保持命令注册稳定，按需暴露 completion helper 所需上下文
- `cmd/datakit/cmd/tool/tool.go`
  - 直接删除旧 completer flags
- `cmd/datakit/cmd/tool/tool_test.go`
  - 删除旧 completer flags 相关断言
- `internal/cmds/tools.go`
  - 删除旧 completer flag 相关调用
- `internal/cmds/completer.go`
  - 直接删除
- `datakit-completer.sh`
  - 直接删除

### 文档文件

- `docs/specs/2026-04-14-shell-completion-spec.md`
- `docs/plans/2026-04-14-shell-completion.md`
- 实施阶段如需更新用户手册，修改 `internal/export/doc/*`

## 任务 1：审计遗留 completer 依赖

**涉及文件：**
- 修改：`internal/cmds/completer.go`
- 修改：`internal/cmds/tools.go`
- 修改：`datakit-completer.sh`

- [x] **Step 1: 定位所有旧 completer 入口引用**

执行：`rg -n "setup-completer-script|completer-script|datakit-completer\\.sh|completion " .`
预期：找出代码、文档、打包流程中仍依赖旧 completer 路径的地方

- [x] **Step 2: 记录删除影响范围**

预期结果：

- 哪些 CLI 入口会被直接删除
- 哪些文档和测试需要同步更新
- 是否需要在 changelog 中单列 breaking change 说明

- [x] **Step 3: 决定 `internal/cmds/completer.go` 的最终处理方式**

预期结果：

- 直接删除，并同步清理引用

## 任务 2：扩展 Cobra completion 到四种 shell，并支持自动识别

**涉及文件：**
- 新增：`cmd/datakit/cmd/completion_support.go`
- 修改：`cmd/datakit/cmd/completion.go`
- 测试：`cmd/datakit/cmd/completion_test.go`

- [x] **Step 1: 先写失败测试，覆盖四种 shell 与自动识别**

示例测试形态：

```go
func TestGenerateCompletionScript(t *testing.T) {
	for _, shell := range []string{"bash", "powershell", "fish", "zsh"} {
		script, err := generateCompletionScript(rootCmd, shell)
		require.NoError(t, err)
		assert.NotEmpty(t, script)
	}
}
```

- [x] **Step 2: 运行新测试，确认当前失败**

执行：`go test ./cmd/datakit/cmd -run 'TestGenerateCompletionScript' -v`
预期：FAIL，因为 `fish` 和 `zsh` 尚未接入

- [x] **Step 3: 实现共享生成 helper**

实现目标：

- 提供 shell 自动识别 helper，例如 `detectCurrentShell()`
- 提供 `generateCompletionScript(cmd *cobra.Command, shell string) (string, error)`
- 支持：
  - `bash`
  - `powershell`
  - `fish`
  - `zsh`
- 对未知 shell 返回清晰错误

- [x] **Step 4: 更新 `completion` 命令的 `ValidArgs` 和 help**

实现目标：

- `Use` 体现 shell 参数可选
- `ValidArgs` 包含四种 shell
- help 示例体现自动识别、显式指定和 `--print`

- [x] **Step 5: 让 `completion` 命令复用共享 helper**

实现目标：

- 去掉重复分支代码
- 未显式指定 shell 时走自动识别
- 显式指定 shell 时覆盖自动识别
- 统一通过共享 helper 生成脚本

- [x] **Step 6: 运行聚焦测试**

执行：`go test ./cmd/datakit/cmd -run 'TestCompletion|TestGenerateCompletionScript|TestDetectCurrentShell' -v`
预期：PASS

- [ ] **Step 7: 提交四 shell completion 支持**

```bash
git add cmd/datakit/cmd/completion.go cmd/datakit/cmd/completion_support.go cmd/datakit/cmd/completion_test.go
git commit -m "feat: extend cobra completion to fish and zsh"
```

## 任务 3：实现安装路径选择与写入

**涉及文件：**
- 新增：`cmd/datakit/cmd/completion_install.go`
- 新增：`cmd/datakit/cmd/completion_install_test.go`
- 修改：`cmd/datakit/cmd/completion.go`

- [x] **Step 1: 先写失败测试，覆盖自动识别和各 shell 安装目标**

覆盖：

- `datakit completion` 自动识别当前 shell
- 显式 shell 覆盖自动识别结果
- DataKit 运行在 Docker 容器内时的默认安装目标
- bash 系统路径和用户回退路径
- fish 用户配置目录
- zsh completion 目录与文件命名
- PowerShell profile 或 completion 文件路径
- `--path` 显式指定路径

- [x] **Step 2: 先写失败测试，覆盖 `--force` 与 `--print`**

覆盖：

- 目标已存在但未传 `--force`
- 目标已存在且传入 `--force`
- `--print` 只输出不写文件

- [x] **Step 3: 定义安装选项模型**

建议类型：

```go
type completionInstallOptions struct {
	Shell string
	Print bool
	Path  string
	Force bool
}
```

- [x] **Step 4: 实现各 shell 的默认目标路径解析**

实现目标：

- 自动识别后将结果归一到四种 shell 之一
- Docker 容器场景下优先选择容器内可写的标准路径
- bash：
  - `/usr/share/bash-completion/completions/datakit`
  - `/etc/bash_completion.d/datakit`
  - `~/.local/share/bash-completion/completions/datakit`
- fish：
  - `~/.config/fish/completions/datakit.fish`
- zsh：
  - 标准 zsh completion 目录中的 `_datakit`
- powershell：
  - 当前用户 profile 相关路径

- [x] **Step 5: 实现文件写入与目录创建**

实现目标：

- 对用户目录自动创建父目录
- 对系统路径不可写时返回清晰错误或回退
- 写入成功后返回实际路径

- [x] **Step 6: 实现成功提示**

实现目标：

- 打印 shell 类型
- 打印实际安装路径
- 如果是自动识别得到的 shell，提示识别结果
- 如果当前运行在 Docker 容器内，提示安装发生在容器文件系统中
- 打印当前会话如何生效

- [x] **Step 7: 实现 `--print`、`--path`、`--force`**

实现目标：

- `--print` 直接输出脚本
- `--path` 覆盖默认安装路径
- `--force` 控制覆盖行为
- 不带 shell 参数时自动识别；带 shell 参数时强制使用指定 shell

- [x] **Step 8: 运行安装相关测试**

执行：`go test ./cmd/datakit/cmd -run 'TestInstall|TestCompletion|TestDetectCurrentShell' -v`
预期：PASS

- [ ] **Step 9: 提交安装逻辑**

```bash
git add cmd/datakit/cmd/completion.go cmd/datakit/cmd/completion_install.go cmd/datakit/cmd/completion_install_test.go
git commit -m "feat: install shell completion from cobra command"
```

## 任务 4：直接删除旧 completer 入口

**涉及文件：**
- 修改：`cmd/datakit/cmd/tool/tool.go`
- 修改：`cmd/datakit/cmd/tool/tool_test.go`
- 修改：`internal/cmds/tools.go`
- 修改：`internal/cmds/completer.go`
- 修改：`datakit-completer.sh`

- [x] **Step 1: 先写失败测试，覆盖旧 completer 入口处理**

覆盖：

- `--setup-completer-script`
- `--completer-script`

- [x] **Step 2: 从代码中删除旧 completer flags**

实现目标：

- 删除 `cmd/datakit/cmd/tool/tool.go` 中对应 flags
- 删除相关测试断言
- 删除旧路径调用链

- [x] **Step 3: 删除静态脚本与旧 completer 逻辑**

实现目标：

- 删除 `datakit-completer.sh`
- 删除 `internal/cmds/completer.go`
- 清理所有引用与构建残留

- [x] **Step 4: 运行删除后的回归测试**

执行：`go test ./cmd/datakit/cmd/tool ./internal/cmds/... -v`
预期：PASS

- [x] **Step 5: 更新 breaking change changelog**

实现目标：

- 在对应 changelog 中明确说明：
  - 删除了旧 completer flags
  - 删除了静态补全脚本
  - 新入口为 `datakit completion`

- [ ] **Step 6: 提交删除调整**

```bash
git add cmd/datakit/cmd/tool/tool.go cmd/datakit/cmd/tool/tool_test.go internal/cmds/tools.go internal/cmds/completer.go datakit-completer.sh
git commit -m "chore: remove legacy completion entry points"
```

## 任务 5：验证生成与安装结果覆盖当前命令树

**涉及文件：**
- 修改：`cmd/datakit/cmd/completion_test.go`
- 修改：`cmd/datakit/cmd/completion_install_test.go`

- [x] **Step 1: 增加代表性命令覆盖断言**

至少覆盖：

- `dql`
- `pipeline`
- `service`
- `monitor`
- `install`
- `debug`
- `tool`
- `check`
- `version`
- `import`
- `run`

- [x] **Step 2: 分别验证自动识别与四种 shell 的安装、print 行为**

执行：

```bash
go test ./cmd/datakit/cmd -run 'TestCompletion|TestInstall' -v
```

预期：

- 四种 shell 生成结果均非空
- 自动识别路径能正确落到对应 shell
- Docker 容器场景能落到容器内默认路径
- 安装成功时能拿到实际路径
- `--print` 时不发生文件写入

- [ ] **Step 3: 提交回归测试**

```bash
git add cmd/datakit/cmd/completion_test.go cmd/datakit/cmd/completion_install_test.go
git commit -m "test: cover completion generation and install flow"
```

## 任务 6：更新用户文档

**涉及文件：**
- 按需修改：`internal/export/doc/zh/*`
- 按需修改：`internal/export/doc/en/*`

- [x] **Step 1: 定位 completion 相关用户文档**

执行：`rg -n "completion|completer|bash_completion|powershell|zsh|fish" internal/export/doc`
预期：找出描述旧 completer 流程或缺少新 shell 示例的文档位置

- [x] **Step 2: 更新为标准 completion 用法**

必须包含：

- `datakit completion`
- `datakit completion bash`
- `datakit completion powershell`
- `datakit completion fish`
- `datakit completion zsh`
- 默认自动识别当前 shell
- 显式 shell 会覆盖自动识别
- 默认安装到本机
- Docker 场景下默认安装到容器内文件系统
- 成功后会打印安装路径
- `--print` 用于只输出脚本
- 旧 completer 入口已删除
- breaking change changelog 迁移说明

- [x] **Step 3: 通过 grep 检查陈旧文案**

执行：`rg -n "setup-completer-script|completer-script|datakit tool --complete" internal/export/doc`
预期：仅保留有意留下的 deprecated 说明，不再出现已废弃的 `tool --complete`

- [ ] **Step 4: 提交文档更新**

```bash
git add internal/export/doc
git commit -m "docs: update completion install usage"
```

## 任务 7：最终验证

**涉及文件：**
- 验证所有改动文件

- [x] **Step 1: 运行聚焦单元测试**

执行：`go test ./cmd/datakit/cmd/... ./cmd/datakit/cmd/tool ./internal/cmds/... -v`
预期：PASS

- [x] **Step 2: 如条件允许，执行更大范围验证**

执行：`go test ./...`
预期：PASS，或明确记录与本次改动无关的已知失败

- [x] **Step 3: 手动验证四种 shell 安装与输出**

执行：

```bash
SHELL=/bin/bash go run ./cmd/datakit completion --path /tmp/datakit-auto --force
go run ./cmd/datakit completion bash --path /tmp/datakit-bash --force
go run ./cmd/datakit completion powershell --path /tmp/datakit-pwsh.ps1 --force
go run ./cmd/datakit completion fish --path /tmp/datakit.fish --force
go run ./cmd/datakit completion zsh --path /tmp/_datakit --force
```

预期：

- 目标文件成功创建
- 输出中包含实际安装路径
- 自动识别调用会打印识别出的 shell

- [x] **Step 4: 手动验证 Docker 容器场景**

执行：

```bash
docker run --rm -e SHELL=/bin/bash -v "$PWD":/work -w /work <datakit-image> datakit completion bash
```

预期：

- 容器内成功安装 completion
- 输出中明确指出安装路径位于容器文件系统

- [x] **Step 5: 手动验证 print 模式**

执行：

```bash
go run ./cmd/datakit completion bash --print | sed -n '1,20p'
go run ./cmd/datakit completion zsh --print | sed -n '1,20p'
```

预期：

- 输出非空
- 不写入文件

- [x] **Step 6: 完成后回填计划状态并准备评审**

预期结果：

- 已完成步骤全部打勾
- 如有跳过项，在对应步骤下明确说明原因
