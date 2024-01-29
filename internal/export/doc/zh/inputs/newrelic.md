---
title     : 'New Relic'
summary   : '接收来自 New Relic Agemt 的数据'
__int_icon      : ''
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# New Relic For .Net
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

New Relic 的 .Net Agent 是基于 .Net 技术框架的开源项目，可用于对 .NET 技术框架的 App 进行全面的性能观测。也可以用于所有兼容 .NET 技术框架的语言例如：C#, VB.NET, CLI。

---

## 配置 {#config}

### 采集器配置 {input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

<!-- markdownlint-enable -->

完成配置后重启 `Datakit` 和 `IIS`

```powershell
PS> datakit service -R
PS> iisreset
```

### 前置条件 {#requirements}

- 域名准备及证书生成与安装
- [注册 New Relic 账号](https://newrelic.com/signup?via=login){:target="_blank"}
- 安装 New Relic Agent 当前支持版本为 6.27.0
- 安装 .Net Framework 当前支持版本为 3.0

#### 安装并配置 New Relic .NET Agent {#install-and-configure-new-relic-dotnet-agent}

首先确认当前 `Windows OS` 安装的 `DotNet Framework` 版本，运行 `cmd` 输入 `reg query "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\NET Framework Setup\NDP"` 查看当前 OS 上安装的所有版本。

然后进行 New Relic Agent 的安装：

- 可以[登陆个人 `New Relic` 账号](https://one.newrelic.com){:target="_blank"}进行安装：

进入账号后，单击左侧目录栏下的创建数据 `+ Add Data` 子目录，后在右侧的 `Data source` 子目录里的 `Application monitoring` 中选择 `.Net` 后安装指引进行安装。

- 也可以通过安装程序进行安装：

打开[下载目录](https://download.newrelic.com/dot_net_agent/6.x_release/){:target="_blank"} 下载 `dotnet agent` 版本 6.27.0 选择对应的安装程序。

配置 `New Relic Agent`

- 配置必要的环境变量

右键单击桌面左下角 `Windows` 徽标选择 系统，选择 高级系统设置，选择 环境变量 ，查看 系统变量 列表是否包含一下环境变量配置：

<!-- markdownlint-disable MD046 -->
    - `COR_ENABLE_PROFILING`: 数字值 1 默认开启
    - `COR_PROFILER`: 字符值，默认为系统自动填写的 `ID`
    - `CORECLR_ENABLE_PROFILING`: 数字值 1 默认开启
    - `NEW_RELIC_APP_NAME`: 字符值，填写被观测的 `APP` 名字 （可选）
    - `NEWRELIC_INSTALL_PATH`: `New Relic Agent` 安装路径
<!-- markdownlint-enable -->

- 通过配置文件配置 `New Relic`

打开 `New Relic Agent` 安装目录下的 `newrelic.config` 将以下示例中 `{示例值}` 替换为真实值，其他值按照示例中对照填写

```xml
<?xml version="1.0"?>
<!-- Copyright (c) 2008-2017 New Relic, Inc.  All rights reserved. -->
<!-- For more information see: https://newrelic.com/docs/dotnet/dotnet-agent-configuration -->
<configuration xmlns="urn:newrelic-config" agentEnabled="true" agentRunID="{agent id (可自己制定也可不填)}">
  <service licenseKey="{真实的 license key}" ssl="true" host="{www.your-domain-name.com}" port="{Datakit 端口号}" />
  <application>
    <name>{被检测的 APP 名字}</name>
  </application>
  <log level="debug" />
  <transactionTracer enabled="true" transactionThreshold="apdex_f" stackTraceThreshold="500" recordSql="obfuscated" explainEnabled="false" explainThreshold="500" />
  <crossApplicationTracer enabled="true" />
  <errorCollector enabled="true">
    <ignoreErrors>
      <exception>System.IO.FileNotFoundException</exception>
      <exception>System.Threading.ThreadAbortException</exception>
    </ignoreErrors>
    <ignoreStatusCodes>
      <code>401</code>
      <code>404</code>
    </ignoreStatusCodes>
  </errorCollector>
  <browserMonitoring autoInstrument="true" />
  <threadProfiling>
    <ignoreMethod>System.Threading.WaitHandle:InternalWaitOne</ignoreMethod>
    <ignoreMethod>System.Threading.WaitHandle:WaitAny</ignoreMethod>
  </threadProfiling>
</configuration>
```

#### 配置主机 {#configure-host-for-newrelic}

由于 `New Relic Agent` 需要配置 `HTTPS` 完成数据传输所以进行主机配置前首先完成[证书的申请](certificate.md#self-signed-certificate-with-openssl)，由于 `New Relic Agent` 启动过程中需要完成证书合法性验证，这里需要完成 `CA` 的自签和自签 `CA` 的证书签发。完成证书认证链的签发后参考[观测云接入 NewRelic .NET 探针](https://blog.csdn.net/liurui_wuhan/article/details/132889536){:target="_blank"}和[Windows 服务器如何导入根证书和中间证书？](https://baijiahao.baidu.com/s?id=1738111820379111942&wfr=spider&for=pc){:target="_blank"}进行证书的部署。

完成证书部署后需要对 `hosts` 文件进行相应的配置已满足本地解析域名的能力 `hosts` 配置如下：

```config
127.0.0.1    www.your-domain-name.com
```

其中 `www.your-domain-name.com` 为 `newrelic.config` 配置文件中 `service.host` 项中制定的域名

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
[inputs.{{.InputName}}.tags]
 # some_tag = "some_value"
 # more_tag = "some_other_value"
 # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## FAQ {#faq}

### New Relic license key 在哪里？ {#where-license-key}

如果是从 `New Relic` 官网指引安装，`license key` 将自动完成填写，如果是自行手动安装 `license key` 将在安装程序运行过程中要求填写，`license key` 在[创建账号](https://newrelic.com/signup?via=login){:target="_blank"}或者[创建数据](newrelic.md#install-and-configure-new-relic-dotnet-agent)时会出现建议保存。

### TLS 版本不兼容 {#tls-version}

部署 `New Relic Agent` 过程中如果遇到没有数据上报，并且在 `New Relic` 日志中看到类似以下 `ERROR` 信息：

```log
NewRelic ERROR: Unable to connect to the New Relic service at collector.newrelic.com:443 : System.Net.WebException:
The request was aborted: Could not create SSL/TLS secure channel.
...
NewRelic ERROR: Unable to connect to the New Relic service at collector.newrelic.com:443 : System.Net.WebException:
The underlying connection was closed: An unexpected error occurred on a send. ---> System.IO.IOException:
Received an unexpected EOF or 0 bytes from the transport stream.
...
NewRelic ERROR: Unable to connect to the New Relic service at collector.newrelic.com:443 : System.Net.WebException:
The underlying connection was closed: An unexpected error occurred on a receive. ---> System.ComponentModel.Win32Exception:
The client and server cannot communicate, because they do not possess a common algorithm.
```

请参考文档[No data appears after disabling TLS 1.0](https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/){:target="_blank"}排查问题

## 参考资料 {#newrelic-references}

- [官方文档](https://docs.newrelic.com/){:target="_blank"}
- [代码仓库](https://github.com/newrelic/newrelic-dotnet-agent){:target="_blank"}
- [下载](https://download.newrelic.com/){:target="_blank"}
- [观测云接入 NewRelic .NET 探针](https://blog.csdn.net/liurui_wuhan/article/details/132889536){:target="_blank"}
- [Windows 服务器如何导入根证书和中间证书？](https://baijiahao.baidu.com/s?id=1738111820379111942&wfr=spider&for=pc){:target="_blank"}
