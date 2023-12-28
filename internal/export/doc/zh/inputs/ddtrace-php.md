
# PHP

---

## 安装依赖 {#dependence}

首先[下载](https://github.com/DataDog/dd-trace-php/releases){:target="_blank"} 需要的 PHP DDTrace 扩展，下载完成后安装扩展。

### RPM package (CentOS/Fedora) {#rpm}

```shell
rpm -ivh datadog-php-tracer.rpm
```

### Deb package (Debian/Ubuntu) {#dev}

```shell
dpkg -i datadog-php-tracer.deb
```

### APK package (Alpine) {#apk}

```shell
apk add datadog-php-tracer.apk --allow-untrusted
```

通过以上方式将安装默认版本的 PHP 扩展。可以通过配置 `DD_TRACE_PHP_BIN` 到扩展包路径来安装制定的扩展包。

```shell
export DD_TRACE_PHP_BIN=$(which version of php-fpm7)
```

安装完成后重启 PHP ([PHP-FPM](https://www.php.net/manual/en/install.fpm.php){:target="_blank"} 或 [Apache SAPI](https://en.wikipedia.org/wiki/Server_application_programming_interface){:target="_blank"}) 然后访问启动了 tracing 的 endpoint。

> 注意：如果你的 PHP 应用没有使用 Composer 或使用 `spl_autoload_register()` 注册了 autoloader，你需要设置环境变量 `DD_TRACE_NO_AUTOLOADER=true`, 用来开启自动检测。

## 配置 {#config}

PHP tracer 可以通过环境变量和 ini 配置文件进行配置。ini 可以进行全局配置，例如：使用 php.ini 配置特定的 web server 或 virtual host。

<!-- markdownlint-disable MD046 -->
???+ attention

    如果你使用了自动检测（建议方案），需要注意的是用于自动检测的代码会在任何业务代码前运行。那么以下环境变量和 ini 配置需要在相应服务器上进行配置，并且能被 PHP runtime 访问到。例如： `putenv()` 函数和 *.env* 文件会失效。
<!-- markdownlint-enable -->

### Apache {#apache}

Apache 搭配 php-fpm, 在 `www.conf` 配置文件中配置环境变量。

```ini
; Example of passing the host environment variable SOME_ENV
; to the PHP process as DD_AGENT_HOST
env[DD_AGENT_HOST] = $SOME_ENV
; Example of passing the value 'my-app' to the PHP
; process as DD_SERVICE
env[DD_SERVICE] = my-app
; Or using the equivalent INI setting
php_value[datadog.service] = my-app
```

还可以在 server config, virtual host, directory, or .htaccess 文件中使用 SetEnv。

``` not-set
# In a virtual host configuration as an environment variable
SetEnv DD_TRACE_DEBUG true
# In a virtual host configuration as an INI setting
php_value datadog.service my-app
```

### Nginx {#ngx}

Nginx 搭配 php-fpm, 在 `www.conf` 配置文件中配置环境变量。

```ini
; Example of passing the host environment variable SOME_ENV
; to the PHP process as DD_AGENT_HOST
env[DD_AGENT_HOST] = $SOME_ENV
; Example of passing the value 'my-app' to the PHP
; process as DD_SERVICE
env[DD_SERVICE] = my-app
; Or using the equivalent INI setting
php_value datadog.service my-app
```

## 运行 {#run}

打开 shell 运行下面命令

```shell
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
DD_TRACE_DEBUG=true \
php -d datadog.service=my-php-app -S localhost:8888
```

## 环境变量支持 {#envs}

- `DD_AGENT_HOST`
  INI: `datadog.agent_host`
  Datakit 监听的主机地址，默认 localhost。

- `DD_TRACE_AGENT_PORT`
  INI: `datadog.trace.agent_port`
  Datakit 监听端口号，默认 9529。

- `DD_ENV`
  INI: `datadog.env`
  设置程序环境变量。

- `DD_SERVICE`
  INI: `datadog.service`
  设置 APP 名字， 版本小于 0.47.0 使用 `DD_SERVICE_NAME`。

- `DD_SERVICE_MAPPING`
  INI: `datadog.service_mapping`
  重命名 APM 服务名。

- `DD_TRACE_AGENT_ATTEMPT_RETRY_TIME_MSEC`
  INI: `datadog.trace.agent_attempt_retry_time_msec`
  基于 IPC 可配置的电路断点重试时间间隔 (milliseconds)， 默认 5000。

- `DD_TRACE_AGENT_CONNECT_TIMEOUT`
  INI: `datadog.trace.agent_connect_timeout`
  Agent 连接超时 (milliseconds)，默认 100。

- `DD_TRACE_AGENT_MAX_CONSECUTIVE_FAILURES`
  INI: `datadog.trace.agent_max_consecutive_failures`
  基于 IPC 可配置的电路最大连续断点次数，默认 3。

- `DD_TRACE_AGENT_TIMEOUT`
  INI: `datadog.trace.agent_timeout`
  Agent 数据传输超时(milliseconds)，默认 500。

- `DD_TAGS`
  INI: `datadog.tags`
  设置默认 tags。

- `DD_VERSION`
  INI: `datadog.version`
  设置服务版本。

- `DD_TRACE_SAMPLE_RATE`
  INI: `datadog.trace.smaple_rate`
  设置采样率从 0.0(0%) ~ 1.0(100%)。
