---
icon: zy/datakit
---

# DataKit
---

## Overview {#intro}

DataKit is a data collection client running on your server. It sends the collected data to <<<custom_key.brand_name>>>. On the <<<custom_key.brand_name>>> Studio, you can view and analyze this data.

DataKit is an open-source software, and you can obtain its source code from [GitHub](https://github.com/GuanceCloud/datakit){:target="_blank"}.

## Quick Start {#quick-start}

On mainstream platforms, you can install DataKit with a single command. After the installation is complete, [some collectors](datakit-input-conf.md#default-enabled-inputs) are enabled by default. Through these collectors, you can collect some basic data of the host.

<div class="grid cards" markdown>
- :fontawesome-solid-computer: [Host Installation](datakit-install.md#get-install)
- :fontawesome-brands-docker: [Docker Installation](datakit-docker-deploy.md)
- :material-kubernetes: [Kubernetes Installation](datakit-daemonset-deploy.md)
- :simple-amazoneks: [AWS EKS Installation](datakit-eks-deploy.md)
- :simple-awsfargate: [AWS Fargate Installation](ecs-fargate.md)
- :octicons-cloud-offline-24: [Offline Installation](datakit-offline-install.md)
- [Lite Version for IoT](datakit-install.md#lite-install)
</div>

This section mainly includes the following content:

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Basic Usage of DataKit</u>: How to manage the DataKit service </font>](datakit-service-how-to.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit Tool Commands</u>: DataKit provides many convenient tools to assist your daily use</font>](datakit-tools-how-to.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Monitor</u>: View the running status of DataKit</font>](datakit-monitor.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Kubernetes Operator</u>: Automate the collection configuration in Kubernetes through the Operator</font>](datakit-operator.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Proxy</u>: If there is a bandwidth limit, you can use a network proxy to upload DataKit traffic</font>](datakit-proxy.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Log Collection</u>: Collect your application logs through DataKit</font>](../integrations/logging.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Security</u>: Instructions on some security issues involved in the configuration of DataKit</font>](datakit-conf.md#public-apis)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Resource Limits</u>: Limit the resource consumption of DataKit </font>](datakit-conf.md#resource-limit)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Troubleshooting</u>: Debug issues during the DataKit collection process</font>](why-no-data.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>API</u>: DataKit HTTP API](apis.md)
</div>
</font>

## Instructions {#spec}

### Experimental Features {#experimental}

When DataKit is released, some experimental features are included. These features are often newly released functions, and some implementations may have some considerations or inaccuracies:

- The functions may be unstable.
- Some function configurations may not guarantee compatibility during subsequent iterations.
- The functions may be removed, but there will be corresponding alternative measures to meet the corresponding requirements.

### Icon Instructions {#legends}

| Icon        | Description                                                                                                  |
| ---                                                                                                                        | ---                                                                                                          |
| :fontawesome-solid-flag-checkered:                                                                                         | Indicates that the collector supports election                                                               |
| :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker: | For example, they are respectively used to represent Linux, Windows, macOS, Kubernetes, and Docker           |
| :octicons-beaker-24:                                                                                                       | Indicates experimental features (Refer to the description of [experimental features](index.md#experimental)) |
