---
title     : '浏览器拨测'
summary   : '通过浏览器引擎模拟页面访问、交互和断言'
tags:
  - '拨测'
  - '网络'
__int_icon      : 'icon/dialtesting'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

[:octicons-tag-24: Version-2.1.0](../datakit/changelog-2026.md#cl-2.1.0)

---

浏览器拨测属于 `inputs.dialtesting` 采集器中的 `BROWSER` 任务类型，用于通过 Lightpanda 浏览器引擎模拟页面访问、交互和断言，并上报页面性能、步骤结果和失败原因。内置的 Lightpanda 不支持截图。

基础拨测节点配置请参考[网络拨测](dialtesting.md)。本文只说明浏览器拨测相关的额外配置、部署和排查方式。

## 使用要求 {#requirements}

浏览器拨测在 Linux 拨测节点上默认开启；非 Linux 环境下，DataKit 服务模式不会执行 `BROWSER` 任务，本地 debug 验证模式除外。

DataKit 使用 Lightpanda 执行 `BROWSER` 任务，任务中的 `advance_options.engine` 会被节点配置覆盖。`[inputs.dialtesting.browser].engine` 未配置时默认为 `lightpanda`，当前也只支持该值。

DataKit 按以下顺序查找 Lightpanda：

1. `[inputs.dialtesting.browser].engine_path`
1. `LIGHTPANDA_EXECUTABLE_PATH`
1. `PATH` 中的 `lightpanda`

如需显式关闭浏览器拨测，可在 `dialtesting.conf` 中设置：

```toml
[[inputs.dialtesting]]
  [inputs.dialtesting.browser]
    enabled = false
```

可通过 `max_concurrency` 限制同一时间执行的浏览器任务数。`0` 表示不限制；资源受限节点建议设置为 `1`。

## Kubernetes 部署 {#kubernetes}

Kubernetes 中推荐直接使用 DataKit 镜像：

```text
<<<% if custom_key.brand_key == 'guance' %>>>
pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:<version>
<<<% else %>>>
pubrepo.<<<custom_key.brand_main_domain>>>/truewatch/datakit:<version>
<<<% endif %>>>
```

DataKit 镜像内置 Lightpanda，可直接执行 `BROWSER` 任务。如需使用自定义 Lightpanda 二进制，可通过挂载方式提供可执行文件，并配置 `engine_path`：

```toml
[[inputs.dialtesting]]
  [inputs.dialtesting.browser]
    enabled = true
    engine = "lightpanda"
    engine_path = "/opt/datakit-browser/bin/lightpanda"
    max_concurrency = 10
```

## 主机部署 {#host}

主机部署时，需要先安装 Lightpanda。以下示例以 Linux 主机为例。

### 安装 Lightpanda {#host-install-lightpanda}

DataKit 镜像使用观测云发布的 Lightpanda `0.3.6-g2` 版本。在 x86_64 Linux 主机上安装相同版本：

```shell
curl -fL -o lightpanda \
  https://github.com/GuanceCloud/browser/releases/download/0.3.6-g2/lightpanda-x86_64-linux
echo "c68f7f340252156fa954fa1e2603769e3fcdb1dd6d07bce9d8b9f034545f09ba  lightpanda" | sha256sum -c -
sudo install -m 0755 lightpanda /usr/local/bin/lightpanda
rm lightpanda
```

arm64/aarch64 Linux 可使用：

```shell
curl -fL -o lightpanda \
  https://github.com/GuanceCloud/browser/releases/download/0.3.6-g2/lightpanda-aarch64-linux
echo "76f13c2debc88b5b7de91dbb1a540c0de97189fa134a8c84369097f7551e2566  lightpanda" | sha256sum -c -
sudo install -m 0755 lightpanda /usr/local/bin/lightpanda
rm lightpanda
```

安装完成后确认版本：

```shell
lightpanda version
lightpanda serve --help
```

### 配置 DataKit {#host-config-datakit}

复制拨测采集器配置：

```shell
cd /usr/local/datakit/conf.d/samples
sudo cp dialtesting.conf.sample ../dialtesting.conf
```

编辑 `/usr/local/datakit/conf.d/dialtesting.conf`，建议显式指定浏览器引擎和路径：

```toml
[[inputs.dialtesting]]
  server = "https://dflux-dial.<<<custom_key.brand_main_domain>>>"
  region_id = "<your-private-node-id>"
  ak = "<your-ak>"
  sk = "<your-sk>"
  pull_interval = "1m"
  time_out = "30s"

  [inputs.dialtesting.browser]
    engine = "lightpanda"
    engine_path = "/usr/local/bin/lightpanda"
    max_concurrency = 10

  [inputs.dialtesting.tags]
    region = "<your-region>"
```

也可以通过环境变量指定：

```shell
export LIGHTPANDA_EXECUTABLE_PATH=/usr/local/bin/lightpanda
```

如果 DataKit 以 systemd 服务方式运行，当前 shell 中的 `export` 通常不会传递给 DataKit 服务进程。主机部署时更推荐在 `dialtesting.conf` 中配置 `engine_path`。

修改配置后重启 DataKit：

```shell
sudo datakit service restart
```

## 自定义 CA 证书 {#custom-ca-certificates}

对于由企业或私有 CA 签发的站点证书，可在拨测节点导入 Lightpanda 信任的 CA：

```toml
[inputs.dialtesting.browser]
  ca_cert_file = "/etc/datakit/certs/internal-ca.pem"
  # ca_cert_dir = "/etc/datakit/certs"
```

请使用 PEM 证书。`ca_cert_dir` 会读取目录内的证书文件，两个字段可以同时配置。路径必须是拨测节点上的绝对路径，文件中不得包含私钥。

也可以使用 `ENV_INPUT_DIALTESTING_BROWSER_CA_CERT_FILE` 和 `ENV_INPUT_DIALTESTING_BROWSER_CA_CERT_DIR`。在 Kubernetes 中，可通过 ConfigMap 或 Secret 将 CA 证书挂载到 DataKit 容器，再配置容器内路径。

导入 CA 后仍会校验证书链、访问域名和证书有效期。该配置属于拨测节点级信任设置，不能由单个 `BROWSER` 任务覆盖。Lightpanda 在收到自定义 CA 参数时会替换当前信任池，因此 DataKit 会把检测到的系统 CA 目录与配置的自定义 CA 文件或目录作为不同参数同时传入。系统证书和自定义证书仍保留在各自目录中，只在 Lightpanda 内存中加载到同一个信任池。

## 内网拨测与代理 {#private-network-and-proxy}

DataKit 默认禁止拨测内网地址。私有拨测节点需要访问 loopback、RFC1918 或 link-local 地址时，在拨测采集器中关闭该限制：

```toml
[[inputs.dialtesting]]
  disable_internal_network_task = false
```

DataKit 还会把该设置传给 Lightpanda。保持默认的 `disable_internal_network_task = true` 且未配置自定义 CIDR 列表时，Lightpanda 使用 `--block-private-networks` 启动。配置 `disabled_internal_network_cidr_list` 后，DataKit 会通过 `--block-cidrs` 精确阻断这些范围，不再阻断全部私网范围。将 `disable_internal_network_task` 设置为 `false` 时，Lightpanda 允许私网请求，无需另外配置引擎专用环境变量。

Lightpanda `0.3.6-g2` 支持以下默认 HTTP 代理配置：

```toml
[inputs.dialtesting.browser]
  proxy_url = "http://proxy.example.com:8080"
  # proxy_url = "http://user:password@proxy.example.com:8080"
```

也可以使用环境变量 `ENV_INPUT_DIALTESTING_BROWSER_PROXY_URL`。代理的生效优先级为：任务 `advance_options.proxy_url` > `browser_config` 中的 `proxy_url` > 节点 `browser.proxy_url`。如果代理会解密 HTTPS 流量，还需要通过 `ca_cert_file` 或 `ca_cert_dir` 导入代理 CA。

## 本地验证 {#local-test}

如果暂时没有通过页面下发 `BROWSER` 任务，可以用本地 JSON 任务验证浏览器执行链路。`browser_config` 是 YAML 字符串。

浏览器脚本示例：

```yaml
name: browser-homepage
target: https://example.com
timeout_ms: 60000
viewport:
  width: 1280
  height: 720
steps:
  - name: open page
    action: goto
    url: https://example.com
  - name: assert title
    action: assert_title
    contains: Example
```

创建 `/tmp/dialtesting-browser-task.json`。写入 JSON 时，需要将上面的 YAML 作为字符串放到 `browser_config` 中，换行用 `\n` 表示：

```json
{
  "BROWSER": [
    {
      "name": "browser-homepage",
      "url": "https://example.com",
      "status": "OK",
      "frequency": "1m",
      "post_url": "https://openway.<<<custom_key.brand_main_domain>>>?token=<your-token>",
      "browser_config": "name: browser-homepage\ntarget: https://example.com\ntimeout_ms: 60000\nviewport:\n  width: 1280\n  height: 720\nsteps:\n  - name: open page\n    action: goto\n    url: https://example.com\n  - name: assert title\n    action: assert_title\n    contains: Example\n"
    }
  ]
}
```

将 `dialtesting.conf` 中的 `server` 临时改成本地文件地址：

```toml
[[inputs.dialtesting]]
  server = "file:///tmp/dialtesting-browser-task.json"
  pull_interval = "10s"

  [inputs.dialtesting.browser]
    engine = "lightpanda"
    engine_path = "/usr/local/bin/lightpanda"
    max_concurrency = 10
```

使用 debug 运行：

```shell
datakit debug --input-conf /usr/local/datakit/conf.d/dialtesting.conf
```

正常情况下，指标中应能看到 `BROWSER` 任务：

```shell
curl -s http://127.0.0.1:9529/metrics | grep datakit_dialtesting
```

验证完成后，请将 `server`、`region_id`、`ak`、`sk` 等配置恢复为真实拨测节点配置。

## BROWSER 任务示例 {#task}

`BROWSER` 任务通过 `browser_config` 定义浏览器脚本。`browser_config` 是 YAML 字符串，常用字段如下：

浏览器拨测配置 YAML 可通过页面录制生成，具体操作请参考[浏览器拨测录制说明](../synthetic-tests/request-task/browser-recorder.md)。

| 字段 | 类型 | 是否必须 | 说明 |
| --- | --- | --- | --- |
| `name` | string | N | 脚本名称 |
| `target` | string | N | 默认目标地址，`goto` 步骤未配置 URL 时使用 |
| `timeout_ms` | int | N | 脚本总超时时间，单位为毫秒 |
| `viewport.width` | int | N | 浏览器视口宽度 |
| `viewport.height` | int | N | 浏览器视口高度 |
| `tags` | object | N | 自定义标签 |
| `steps` | array | Y | 浏览器执行步骤 |

`steps` 中可使用 `goto`、`click`、`fill`、`wait_for_selector`、`wait_for_url`、`assert_title`、`assert_url`、`assert_text` 等动作和断言。

`wait_for_url` 支持 `contains`、`equals` 或 `text`，会持续轮询，直到 URL 匹配或步骤/脚本超时。`assert_title`、`assert_url` 和 `assert_text` 也始终轮询；配置了步骤 `timeout_ms` 时优先使用步骤超时，否则使用脚本总超时。DataKit 默认的步骤超时时间为 60 秒。例如：

```yaml
- name: 等待跳转到仪表板
  action: wait_for_url
  contains: https://console.example.com/dashboard
  timeout_ms: 15000
- name: 校验仪表板标题
  action: assert_title
  contains: Dashboard
  timeout_ms: 5000
```

完整任务 JSON 中，`browser_config` 位于 `BROWSER` 任务对象内：

```json
{
  "BROWSER": [
    {
      "name": "browser-homepage",
      "url": "https://example.com",
      "status": "OK",
      "frequency": "1m",
      "post_url": "https://openway.<<<custom_key.brand_main_domain>>>?token=<your-token>",
      "browser_config": "<browser_config YAML string>"
    }
  ]
}
```

## 截图支持 {#screenshot}

内置的 Lightpanda 不支持截图。Lightpanda 任务会忽略 `advance_options.screenshot_on_failure`，不会生成 `steps[].screenshot`。

## 排查方式 {#troubleshooting}

通过 DataKit 指标确认任务和上报状态：

```shell
curl -s http://127.0.0.1:9529/metrics | grep datakit_dialtesting
```

浏览器引擎环境可通过以下命令确认：

```shell
echo $LIGHTPANDA_EXECUTABLE_PATH
$LIGHTPANDA_EXECUTABLE_PATH version
command -v lightpanda
```

常见问题：

- 拉不到任务：确认 `server`、`region_id`、`ak`、`sk` 配置正确，且 `datakit_dialtesting_task_number{protocol="BROWSER"}` 大于 0。
- 页面已下发 BROWSER 任务但节点没有执行：确认没有显式配置 `[inputs.dialtesting.browser].enabled = false`，并检查 DataKit 日志中是否出现 `browser.enabled is false or unsupported`。
- 任务不上报：确认任务 `post_url` 可访问，且发送失败、缓存、丢弃相关指标未持续增长。
- 浏览器无法启动：确认 `engine_path`、`LIGHTPANDA_EXECUTABLE_PATH` 或 `PATH` 中的 `lightpanda` 可被 DataKit 进程访问。
- 浏览器依赖缺失：Kubernetes 中建议直接使用 `datakit:<version>` 镜像；主机部署时确认 Lightpanda 已正确安装。
- 截图未上传：内置的 Lightpanda 不会生成截图。
