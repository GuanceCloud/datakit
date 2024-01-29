---
title     : 'New Relic'
summary   : 'Receive data from New Relic Agent'
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

New Relic's .Net Agent is an open source project based on the .Net technology framework, which can be used to conduct comprehensive performance observations of apps based on the .NET technology framework. It can also be used for all languages compatible with the .NET technology framework such as: C#, VB.NET, CLI.

---

## Configuration {#config}

### Collector Config {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Enter the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [Restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Currently, the collector can be enabled through [ConfigMap method to inject collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

After completing the configuration, restart `Datakit` and `IIS`

```powershell
PS> datakit service -R
PS> iisreset
```

### Preconditions {#requrements}

- Domain name preparation and certificate generation and installation
- [Sign up for a New Relic account](https://newrelic.com/signup?via=login){:target="_blank"}
- Install New Relic Agent. The current supported version is 6.27.0
- Install .Net Framework. The current supported version is 3.0

#### Install and configure New Relic .NET Agent {#install-and-configure-new-relic-dotnet-agent}

First confirm the `DotNet Framework` version currently installed on `Windows OS`, run `cmd` and enter `reg query "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\NET Framework Setup\NDP"` to view all versions installed on the current OS.

Then install New Relic Agent:

- You can [log in to your personal `New Relic` account](https://one.newrelic.com){:target="_blank"} to install:

After entering the account, click the Create Data `+ Add Data` subdirectory under the directory column on the left, and then select `.Net` in the `Application monitoring` in the `Data source` subdirectory on the right and follow the installation guide to install it. .

- It can also be installed via the installer:

Open [download directory](https://download.newrelic.com/dot_net_agent/6.x_release/){:target="_blank"} to download `dotnet agent` version 6.27.0 and select the corresponding installation program.

Configure `New Relic Agent`

- Configure necessary environment variables

Right-click the `Windows` logo in the lower left corner of the desktop, select System, select Advanced System Settings, select Environment Variables, and check whether the System Variables list contains the following environment variable configuration:

<!-- markdownlint-disable MD046 -->
    - `COR_ENABLE_PROFILING`: Numeric value 1 enables by default
    - `COR_PROFILER`: character value, the default is `ID` automatically filled in by the system
    - `CORECLR_ENABLE_PROFILING`: Numeric value 1 enables by default
    - `NEW_RELIC_APP_NAME`: character value, fill in the name of the observed `APP` (optional)
    - `NEWRELIC_INSTALL_PATH`: `New Relic Agent` installation path
<!-- markdownlint-enable -->

- Configure `New Relic` through configuration file

Open `newrelic.config` in the `New Relic Agent` installation directory. Replace `{example value}` in the following example with the real value, and fill in other values according to the examples.

```xml
<?xml version="1.0"?>
<!-- Copyright (c) 2008-2017 New Relic, Inc.  All rights reserved. -->
<!-- For more information see: https://newrelic.com/docs/dotnet/dotnet-agent-configuration -->
<configuration xmlns="urn:newrelic-config" agentEnabled="true" agentRunID="{agent id (You can make your own or leave it blank)}">
  <service licenseKey="{license key}" ssl="true" host="{www.your-domain-name.com}" port="{DataKit Port}" />
  <application>
    <name>{Detected APP name}</name>
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

#### Configure host {#configure-host-for-newrelic}

Since `New Relic Agent` needs to configure `HTTPS` to complete data transmission, first complete the [certificate application] (certificate.md#self-signed-certificate-with-openssl) before configuring the host. Due to the `New Relic Agent` startup process The certificate validity verification needs to be completed. Here, the self-signing of `CA` and the issuance of the self-signed `CA` certificate need to be completed. After completing the issuance of the certificate authentication chain, refer to [Observation Cloud Access NewRelic .NET Probe](https://blog.csdn.net/liurui_wuhan/article/details/132889536){:target="_blank"} and [Windows Server How to import root and intermediate certificates?](https://baijiahao.baidu.com/s?id=1738111820379111942&wfr=spider&for=pc){:target="_blank"} to deploy the certificate.

After completing the certificate deployment, you need to configure the `hosts` file accordingly to meet the local ability to resolve domain names. The `hosts` configuration is as follows:

```config
127.0.0.1    www.your-domain-name.com
```

Where `www.your-domain-name.com` is the domain name specified in the `service.host` item in the `newrelic.config` configuration file

## Metric {#metric}

All the following data collection will add a global tag named `host` by default (the tag value is the host name of DataKit). You can also specify other tags through `[inputs.{{.InputName}}.tags]` in the configuration:

``` toml
[inputs.{{.InputName}}.tags]
 # some_tag = "some_value"
 # more_tag = "some_other_value"
 # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tags

{{$m.TagsMarkdownTable}}

- fields

{{$m.FieldsMarkdownTable}}

{{ end }}

## FAQ {#faq}

### Where is the New Relic license key? {#where-license-key}

If you install from the `New Relic` official website, the `license key` will be filled in automatically. If you install it manually, you will be asked to fill in the `license key` during the installation process. The `license key` is in [Create Account](https ://newrelic.com/signup?via=login){:target="_blank"} or [create data](newrelic.md#install-and-configure-new-relic-dotnet-agent), a suggestion to save will appear.

### TLS version incompatible {#tls-version}

During the deployment of `New Relic Agent`, if no data is reported, and an `ERROR` message similar to the following is seen in the `New Relic` log:

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

Please refer to the documentation[No data appears after disabling TLS 1.0](https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/){:target="_blank"} to troubleshoot the issue

## References {#newrelic-references}

- [Official Document](https://docs.newrelic.com/){:target="_blank"}
- [Code Warehouse](https://github.com/newrelic/newrelic-dotnet-agent){:target="_blank"}
- [Download](https://download.newrelic.com/){:target="_blank"}
- [Observation Cloud Access NewRelic .NET Probe](https://blog.csdn.net/liurui_wuhan/article/details/132889536){:target="_blank"}
- [How to import root certificates and intermediate certificates on Windows servers?](https://baijiahao.baidu.com/s?id=1738111820379111942&wfr=spider&for=pc){:target="_blank"}
