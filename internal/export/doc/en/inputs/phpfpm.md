---
title     : 'PHP-FPM'
summary   : 'Collect metrics of PHP-FPM'
tags:
  - 'MIDDLEWARE'
__int_icon      : 'icon/php_fpm'
dashboard :
  - desc  : 'PHP-FPM'
    path  : 'dashboard/phpfpm'
---

{{.AvailableArchs}}

---

PHP-FPM collector is used for PHP-FPM metrics collection, such as the number of active processes, total number of idle processes, the number of requests in the queue of pending connections, the maximum number of active processes, etc.

## Configuration {#config}

### Preconditions {#reqirement}

1. Enable PHP-FPM status page

    - Edit the PHP-FPM configuration file(typically located at `/etc/php/8.x/fpm/pool.d/www.conf` or `/etc/php-fpm.d/www.conf` ). Enable `pm.status_path` :

    ```shell
    ; Enable status page
    pm.status_path = /status
   ```

    - Restart PHP-FPM

   ```shell
    sudo systemctl restart php-fpm
    sudo systemctl restart php8.x-fpm 
    ```

2. Configure Web Server

    - Nginx example：

    ```nginx
    location /status {
        include fastcgi_params;
        fastcgi_pass unix:/var/run/php/php8.x-fpm.sock; // unix socket mode
        # fastcgi_pass 127.0.0.1:9000; // TCP mode
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }
    ```

    - Restart Nginx:

    ```shell
    sudo systemctl restart nginx
    ```

    - Access the status page( `/status` ) to ensure it returns PHP-FPM status data correctly

### Collector Configuration {#input-config}
<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Adjust the following configurations according to your environment:
    - When using the HTTP by default (e.g., `http://localhost/status`),keep `use_fastcgi = false`
    - When using `Unix socket` or `TCP` :
      * Address formats: `unix:///socket/path;/status` or `tcp://ip:port/status`
      * Must set `use_fastcgi = true`

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
