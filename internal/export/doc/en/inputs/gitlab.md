---
title     : 'GitLab'
summary   : 'Collect Gitlab metrics and logs'
__int_icon      : 'icon/gitlab'
dashboard :
  - desc  : 'GitLab'
    path  : 'dashboard/en/gitlab'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# GitLab
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Collect GitLab operation data and report it to Guance Cloud in the form of metrics.

## Configuration {#config}

### Collector Configuration {#input-config}

First, you need to open the data collection function of GitLab service and set the white list. See the following sections for specific operations.

After the GitLab setup is complete, configure the DataKit. Note that the data collected may vary depending on the GitLab version and configuration.

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

### GitLab Turns on Data Collection {#enable-prom}

GitLab needs to turn on the Prometheus data collection function as follows (taking English page as an example):

- Log in to your GitLab page as an administrator account
- Go to `Admin Area` > `Settings` > `Metrics and profiling`
- Select `Metrics - Prometheus`, click `Enable Prometheus Metrics` and `save change`
- Restart the GitLab service

See [official configuration doc](https://docs.gitlab.com/ee/administration/monitoring/prometheus/gitlab_metrics.html#gitlab-prometheus-metrics){:target="_blank"}.

### Configure Data Access Whitelist {#white-list}

It is not enough to turn on the data collection function. GitLab is very strict with data management, so it is necessary to configure the white list on the access side. The opening mode is as follows:

- Modify the GitLab configuration file `/etc/gitlab/gitlab.rb`, find `gitlab_rails['monitoring_whitelist'] = ['::1/128']` and add the access IP of the DataKit to the array (typically the IP of the host where the DataKit resides, if the GitLab is running in a container, depending on the actual situation)
- Restart the GitLab service

See [official configuration doc](https://docs.gitlab.com/ee/administration/monitoring/ip_whitelist.html){:target="_blank"}.

### Turn on GitLab CI Visualization {#ci-visible}

Ensure that the current Datakit version (1.2. 13 and later) supports GitLab CI visualization.

GitLab CI visualization can be achieved by configuring GitLab Webhook. The opening steps are as follows:

- In GitLab go to `Settings` > `Webhooks`, configure the URL to http://Datakit_IP:PORT/v1/gitlab, Trigger configure Job events and Pipeline events, and click Add webhook to confirm the addition;

- You can Test whether the Webhook is configured correctly by clicking the Test button, and Datakit should return a status code of 200 when it receives the Webhook. After proper configuration, Datakit can successfully collect CI information of GitLab.

After Datakit receives the Webhook Event, it logs the data to the data center.

Note: Additional configuration of Gitlab is required if Gitlab data is sent to Datakit on the local network, see [allow requests to the local network](https://docs.gitlab.com/ee/security/webhooks.html){:target="_blank"}.

In addition, GitLab CI function does not participate in collector election, and users only need to configure the URL of GitLab Webhook as the URL of one of Datakit; If you only need GitLab CI visualization and do not need GitLab metrics collection, you can turn off metrics collection by configuring `enable_collect = false`.

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit).

You can specify additional labels for **Gitlab metrics data** in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

You can specify additional tags for **Gitlab CI data** in the configuration by `[inputs.{{.InputName}}.ci_extra_tags]`:

``` toml
 [inputs.{{.InputName}}.ci_extra_tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

Note: To ensure that GitLab CI functions properly, the extra tags specified for GitLab CI data do not overwrite tags already in its data (see below for a list of GitLab CI tags).



{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
