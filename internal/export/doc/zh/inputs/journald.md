---
title     : 'Journald'
summary   : '收集 systemd 日志'
tags:
  - 'HOST'
  - 'LOG'
__int_icon      : 'icon/system'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

Journald 采集器用于在 Linux 系统上从 systemd journal (journald) 收集日志。它使用外部二进制包装器与 `libsystemd` 交互，高效地从 journal 收集结构化日志条目。

## 前置条件 {#prerequisites}

- **仅限 Linux**: 需要 `systemd` 和 `journald`
- **libsystemd**: 外部二进制需要 `libsystemd` 开发库
- **权限**: DataKit 需要 journal 文件的读取权限（通常需要加入 `systemd-journal` 组）

### 系统要求检查 {#system-requirements-check}

部署 journald 采集器之前，验证您的系统是否满足要求：

可通过如下命令快速检查：

```bash
systemctl --version >/dev/null 2>&1 && journalctl -n 1 >/dev/null 2>&1 && echo "Systemd OK" || echo "Systemd not available"
```

以下是全面的预检查脚本：

<!-- markdownlint-disable MD046 -->
??? info "*journald-prereq-check.sh*"

    ``` bash linenums="1"
    #!/bin/bash
    # journald-prereq-check.sh - 验证 systemd 要求
    
    echo "=== Journald 采集器前置条件检查 ==="
    echo
    
    # 1. 检查 systemctl 是否存在
    echo -n "1. systemctl 命令："
    if command -v systemctl >/dev/null 2>&1; then
        VERSION=$(systemctl --version | head -1)
        echo "✅ 已找到 - $VERSION"
    else
        echo "❌ 未找到 - 未安装 systemctl"
        exit 1
    fi
    
    # 2. 检查 libsystemd 库
    echo -n "2. libsystemd.so.0: "
    if ldconfig -p 2>/dev/null | grep -q "libsystemd.so.0"; then
        LIBPATH=$(ldconfig -p 2>/dev/null | grep "libsystemd.so.0" | head -1 | awk '{print $NF}')
        echo "✅ 已找到 - $LIBPATH"
    else
        echo "❌ 未找到 - 缺少 libsystemd.so.0"
        exit 1
    fi
    
    # 3. 检查 journalctl 访问权限
    echo -n "3. journalctl 访问："
    if journalctl -n 1 >/dev/null 2>&1; then
        echo "✅ 正常 - 可以读取 journal"
    else
        echo "⚠️  受限 - journalctl 存在但无读取权限"
    fi
    
    # 4. 检查 journal 目录
    echo "4. Journal 目录："
    for dir in "/var/log/journal" "/run/log/journal"; do
        echo -n "   $dir: "
        if [ -d "$dir" ]; then
            if [ -r "$dir" ]; then
                echo "✅ 存在且可读"
            else
                echo "⚠️  存在但不可读"
            fi
        else
            echo "❌ 未找到"
        fi
    done
    
    # 5. 检查 systemd 版本
    echo -n "5. systemd 版本："
    SYSTEMD_VERSION=$(systemctl --version | head -1 | grep -oP 'systemd \K\d+' || echo "0")
    if [ "$SYSTEMD_VERSION" -ge 205 ]; then
        echo "✅ v$SYSTEMD_VERSION (满足最低 v205 要求)"
    else
        echo "⚠️  v$SYSTEMD_VERSION (低于推荐版本 v205)"
    fi
    
    echo
    echo "=== 检查完成 ==="
    ```
<!-- markdownlint-enable -->

保存为 `journald-prereq-check.sh` 并运行：

```bash
chmod +x journald-prereq-check.sh
./journald-prereq-check.sh
```

预期输出：

``` txt
=== Journald 采集器前置条件检查 ===

1. systemctl 命令：✅ 已找到 - systemd 257 (257.3-1-arch)
2. libsystemd.so.0: ✅ 已找到 - /usr/lib/libsystemd.so.0
3. journalctl 访问：✅ 正常 - 可以读取 journal
4. Journal 目录：
   /var/log/journal: ✅ 存在且可读
   /run/log/journal: ✅ 存在且可读
5. systemd 版本：✅ v257 (满足最低 v205 要求)

=== 检查完成 ===
```

可能的故障排除方案：

| 问题 | 解决方案 |
|-------|----------|
| `systemctl: command not found` | 安装 systemd 或使用替代日志收集方式 |
| `libsystemd.so.0: cannot open` | 安装 systemd-libs：`apt install libsystemd0` 或 `yum install systemd-libs` |
| `journalctl: no read access` | 将用户添加到 `systemd-journal` 组：`usermod -aG systemd-journal $USER` |
| `/var/log/journal not found` | 启用持久化 journal：`mkdir -p /var/log/journal && systemd-tmpfiles --create` |

## 配置 {#config}

### 采集器配置 {#collector-configuration}

成功安装并启动 DataKit 后，通过复制配置文件启用 Journald 采集器：

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置完成后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service)。

=== "Kubernetes"

    可以通过 [ConfigMap 注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启。
<!-- markdownlint-enable -->

### 配置选项 {#config-options}

| 选项 | 类型 | 默认值 | 描述 |
|--------|------|---------|-------------|
| `paths` | []string | `["/var/log/journal", "/run/log/journal"]` | Journal 目录路径 |
| `units` | []string | `[]` | 按 systemd 单元名称过滤（支持 glob 模式，例如 `*.service`） |
| `priorities` | []string | `[]` | 按优先级过滤：`emerg`、`alert`、`crit`、`err`、`warning`、`notice`、`info`、`debug` |
| `exclude_fields` | []string | `[]` | 从收集中排除的 journal 字段（例如 `_BOOT_ID`、`_MACHINE_ID`） |
| `tail_only` | bool | `true` | 仅收集新条目（启动时跳过历史日志） |
| `max_entries_per_batch` | int | `1000` | 每批收集的最大条目数 |
| `save_cursor` | bool | `true` | 持久化读取位置以便重启后恢复 |
| `cursor_file` | string | `/usr/local/datakit/cache/journald.pos` | 存储游标位置的路径 |
| `mount_dir` | string | `"/rootfs"` | 仅在容器/Kubernetes 模式使用的 rootfs 挂载目录。DataKit 会用它作为绝对 `paths` 的前缀，以及宿主机动态库准备的源目录根路径 |
| `copy_node_libs` | bool | `false`（容器或 Kubernetes 下自动强制为 `true`） | 启动 external collector 前，是否从 mount_dir 复制宿主机动态库到 DataKit 自己的 `external-libs` 目录。在容器或 Kubernetes 环境（`datakit.Docker || config.IsKubernetes()`）中会自动启用 |
| `copy_node_libs_files` | []string | `[]` | 要复制的动态库文件名或 glob 列表。若显式配置则只复制该列表。若容器/Kubernetes 自动模式下为空，DataKit 会先复制 `libsystemd.so*`，再执行 `LD_LIBRARY_PATH=/usr/local/datakit/externals/systemd-libs ldd libsystemd.so.0` 风格依赖探测并自动补齐缺失 `.so`。若在非容器/Kubernetes 且 `copy_node_libs=true` 时留空，会以配置错误方式启动失败 |

## 日志字段 {#logging}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}

{{ end }}

## 常见用例 {#use-cases}

- 收集特定服务的日志

```toml
[[inputs.journald]]
  units = ["nginx.service", "mysql.service", "docker.service"]
  priorities = ["err", "crit", "alert", "emerg"]
  tail_only = true
```

- 排除冗余字段

```toml
[[inputs.journald]]
  exclude_fields = [
    "_BOOT_ID",
    "_MACHINE_ID",
    "__MONOTONIC_TIMESTAMP",
    "_AUDIT_SESSION",
    "_AUDIT_LOGINUID",
  ]
```

- Kubernetes 节点 journal 收集（自动模式）

```toml
[[inputs.journald]]
  paths = ["/var/log/journal", "/run/log/journal"]
  tail_only = true
```

说明：

- collector 会按配置顺序解析候选目录，并优先尝试打开第一个可读的 journal 目录
- 在容器或 Kubernetes 环境（`datakit.Docker || config.IsKubernetes()`）中，DataKit 会自动启用 journald rootfs 模式
- 在容器/Kubernetes 模式下，绝对路径会自动加上 `mount_dir` 前缀（默认 `"/rootfs"`）
- 如果路径本身是 `<mount_dir>/var/log/journal` 这类 journal 根目录，collector 会自动下钻到 machine-id 子目录后再打开
- 在 kind、k3d 等容器化节点环境中，要在 node 容器内验证 `logger`/`journalctl`，不要在宿主机直接验证

- Kubernetes 节点 journal 收集，并在启动前准备宿主机 systemd 相关库

```toml
[[inputs.journald]]
  mount_dir = "/rootfs"
  paths = ["/var/log/journal", "/run/log/journal"]
  tail_only = true
  copy_node_libs = true
  copy_node_libs_files = [
    "libsystemd.so*",
    "liblz4.so*",
    "libzstd.so*",
    "liblzma.so*",
    "libcap.so*",
    "libgcrypt.so*",
    "libgpg-error.so*",
    "libselinux.so*",
    "libmount.so*",
    "libblkid.so*",
    "libacl.so*",
    "libpcre2-8.so*",
    "libpcre.so*",
  ]
```

- 收集所有日志（调试）

```toml
[[inputs.journald]]
  tail_only = false
  max_entries_per_batch = 500
  exclude_fields = []
```

## 故障排除 {#troubleshooting}

### 权限错误 {#troubleshoot-permissions}

确保 DataKit 有 journal 文件的读取权限：

```bash
# 将 datakit 用户添加到 systemd-journal 组
sudo usermod -aG systemd-journal datakit

# 重启 DataKit
sudo systemctl restart datakit
```

### 未收集到日志 {#troubleshoot-no-logs}

1. 验证 journald 是否正在运行：

```bash
systemctl status systemd-journald
```

1. 检查 journal 文件是否存在：

```bash
ls -la /var/log/journal/
ls -la /run/log/journal/
```

1. 如果当前环境安装了 `journalctl`，可继续使用它做额外验证；如果容器里没有 `journalctl`，直接查看 DataKit 的兼容性告警和 probe 结果即可：

```bash
journalctl -n 10
```

如果启动日志中出现 `reason=unsupported-format`，说明当前 collector 运行时使用的 `libsystemd` 版本低于目标 journal 文件格式。此时 DataKit 会记录告警并让 journald collector 保持 inactive，而不是继续输出部分或具有误导性的采集结果。

这种情况并不只出现在 EKS。在 Kubernetes 中，只要 DataKit 需要采集 node 上的 journal，而容器镜像自带的 `libsystemd` 版本低于宿主机 journal 文件格式所需版本，就可能出现这个问题。典型现象包括：

- 如果 Pod 内安装了 `journalctl`，执行后可能报 `unsupported feature`
- DataKit 已启动，但 journald collector 在兼容性告警后保持 inactive

在容器或 Kubernetes 环境（`datakit.Docker || config.IsKubernetes()`）中，DataKit 已经自动启用宿主机 `systemd` 相关动态库准备能力；如果你希望在非容器场景也启用，可配置：

```toml
[[inputs.journald]]
  copy_node_libs = true
```

启用后，DataKit 会在启动 collector 前，从 `mount_dir`（默认 `"/rootfs"`）下的候选系统库目录复制动态库到自己的 `external-libs` 目录，并自动把该目录前置到 `LD_LIBRARY_PATH`。

复制行为细节：

- 如果 `copy_node_libs_files` 已配置且非空，则只复制该列表。
- 如果容器/Kubernetes 自动模式下 `copy_node_libs_files` 为空，DataKit 会先复制 `libsystemd.so*`，然后在复制目录下对 `libsystemd.so.0` 做 `ldd` 依赖探测，并自动补齐缺失 `.so`。
- 如果非容器且非 Kubernetes 且 `copy_node_libs=true` 且 `copy_node_libs_files` 为空，DataKit 会报配置错误并保持 collector inactive。
- 启用 `copy_node_libs` 后如果库准备失败，journald 采集器会保持 inactive（不影响 DataKit 其他采集器）。

collector 成功打开 journal 后，会在 external `journald.log` 中打印类似如下日志，帮助确认运行时到底加载了哪一套 `libsystemd`：

```text
loaded libsystemd paths: [/usr/local/datakit/externals/systemd-libs/libsystemd.so.0.35.0]
```

约束说明：

- 宿主机上的 `libsystemd` 并不保证一定兼容 DataKit 当前使用的 journald external binary
- 如果宿主机上的 `libsystemd` 版本过低，external binary 可能在动态链接阶段就因为符号或版本不匹配而无法启动
- 如果宿主机上的 `libsystemd` 版本更高，则也可能在读取 journal 文件时出现 `unsupported feature`
- 因此，`copy_node_libs` 只是一个前置准备能力，不代表复制后的库一定兼容；最终仍需结合启动日志与 probe 结果判断

不要把整个宿主机 `/usr/lib64` 直接加入 `LD_LIBRARY_PATH`。这样可能把不兼容的 glibc 组件一并带入 collector 进程，导致更难诊断的问题。

如果启动日志显示：

```text
resolved journal directory: target=...
opening journal from directory: ...
```

说明 collector 当前是按目录方式打开 journal，这也是当前 live journal 的推荐路径；不要手动把单个 `.journal` 文件路径作为主要配置方式。

### 游标文件问题 {#troubleshoot-cursor}

如果游标文件损坏（例如主机重启后），采集器会自动回退到 tail 模式并创建新游标。要手动重置：

```bash
# 删除游标文件
rm /usr/local/datakit/cache/journald.pos

# 重启 DataKit
sudo systemctl restart datakit
```

### 高内存使用 {#troubleshoot-memory}

默认批次大小为 1000 个条目。如果内存使用是问题，可以减少批次大小：

```toml
[[inputs.journald]]
  max_entries_per_batch = 100
```

## 相关文档 {#related}

- [DataKit 日志概述](../datakit/datakit-logging.md)
- [日志管道配置](../pipeline/pipeline.md)
- [Systemd Journal 文档](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html){:target="_blank"}
