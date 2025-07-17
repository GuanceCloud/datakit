---
title     : 'PHP-FPM'
summary   : '采集 PHP-FPM 的指标数据'
tags:
  - '中间件'
__int_icon      : 'icon/php_fpm'
dashboard :
  - desc  : 'PHP-FPM'
    path  : 'dashboard/phpfpm'
---

{{.AvailableArchs}}

---

PHP-FPM 采集器用于 PHP-FPM 指标采集，如活动进程总数、空闲进程总数、挂起连接队列中的请求数量、活动进程数的最大值等。

## 配置 {#config}

### 前置条件 {#reqirement}

1. 启用 PHP-FPM 状态页面
    - 编辑 PHP-FPM 配置文件（路径通常为 `/etc/php/8.x/fpm/pool.d/www.conf` 或 `/etc/php-fpm.d/www.conf` ），启用 `pm.status_path` ：

    ```shell
    ; 启用状态页面
    pm.status_path = /status
   ```

    - 重启 PHP-FPM

   ```shell
    sudo systemctl restart php-fpm
    sudo systemctl restart php8.x-fpm 
    ```

2. 配置 Web 服务器

    - Nginx 示例：

    ```nginx
    location /status {
        include fastcgi_params;
        fastcgi_pass unix:/var/run/php/php8.x-fpm.sock; // unix socket 形式
        # fastcgi_pass 127.0.0.1:9000; // TCP 形式
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }
    ```

    - 重启 Nginx：

    ```shell
    sudo systemctl restart nginx
    ```

    - 访问状态页（ `/status` ），确保可以正常访问并返回 PHP-FPM 的状态信息

### 采集器配置 {#input-config}
<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 *conf.d/{{.Catalog}}* 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    请根据实际环境调整以下配置：
    - 默认使用 HTTP 协议时（如 `http://localhost/status`），需保持 `use_fastcgi = false`
    - 使用 `Unix socket` 或 `TCP` 协议时：
      * 地址格式分别为 `unix:///socket/path;/status` 和 `tcp://ip:port/status`
      * 必须设置 `use_fastcgi = true`

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
