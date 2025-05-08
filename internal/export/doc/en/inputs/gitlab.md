---
title     : 'GitLab'
summary   : 'Collect GitLab metrics and logs'
tags:
  - 'GITLAB'
  - 'CI/CD'
__int_icon      : 'icon/gitlab'
dashboard :
  - desc  : 'GitLab'
    path  : 'dashboard/en/gitlab'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

Collect GitLab operation data and report it to <<<custom_key.brand_name>>> in the form of metrics.

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

Ensure that the DataFlux Func platform is available.

By configuring GitLab Webhook, GitLab CI visualization can be achieved. Data reporting needs to be done through DataFlux Func, and the steps to enable it are as follows:

1. Install the GitLab CI integration (script ID: `guance_gitlab_ci`) on DataFlux Func. Follow the installation process as referenced in [GitLab CI Integration Configuration](https://func.<<<custom_key.brand_main_domain>>>/doc/script-market-guance-gitlab-ci/){:target="_blank"};
2. In GitLab go to `Settings` > `Webhooks`, configure the URL to the API address obtained from step one. Trigger configure Job events and Pipeline events, and click Add webhook to confirm the addition;

Triggering the GitLab CI process will allow you to log in to <<<custom_key.brand_name>>> to view the execution status of CI after completion.

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

We can specify additional tags for **Gitlab CI data** in the configuration by `[inputs.{{.InputName}}.ci_extra_tags]`:

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

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}
