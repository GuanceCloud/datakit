
# PHP Sample
---

## Install Dependency {#dependence}

First [download](https://github.com/DataDog/dd-trace-php/releases){:target="_blank"} the required PHP ddtrace extension, and install the extension after downloading.

**Using RPM package (RHEL/Centos 6+, Fedora 20+)**

```shell
rpm -ivh datadog-php-tracer.rpm
```

**Using DEB package (Debian Jessie+ , Ubuntu 14.04+ on supported PHP versions)**

```shell
dpkg -i datadog-php-tracer.deb
```

**Using APK package (Alpine)**

```shell
apk add datadog-php-tracer.apk --allow-untrusted
```

The default version of PHP extension will be installed in the above way. You can install the specified extension packs by configuring the DD_TRACE_PHP_BIN to the extension pack path.

```shell
export DD_TRACE_PHP_BIN=$(which version of php-fpm7)
```

After installation, restart PHP (PHP-FPM or the Apache SAPI) and visit the endpoint where tracing is started.

**Note:** If your PHP application does not use Composer or registers autoloader with spl_autoload_register (), you need to set the environment variable DD_TRACE_NO_AUTOLOADER=true to turn on automatic detection.

## Configuration {#config}

PHP tracer can be configured through environment variables and an ini configuration file.

Ini can be configured globally, for example, using php.ini to configure a specific web server or virtual host.

**Note:** If you use auto-detection (recommended), it should be noted that the code used for auto-detection will run before any business code. Then, the following environment variables and ini configuration need to be configured on the corresponding server and accessible by PHP runtime. For example, the putenv () function and the. env file will fail.

**Apache**

Apache works with php-fpm to configure environment variables in the www.conf configuration file.

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

You can also use SetEnv in server config, virtual host, directory, or. htaccess files.

```htaccess
# In a virtual host configuration as an environment variable
SetEnv DD_TRACE_DEBUG true
# In a virtual host configuration as an INI setting
php_value datadog.service my-app
```

**NGINX**

Nginx works with php-fpm to configure environment variables in the www.conf configuration file.

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

## Run {#run}

Open the shell and run the following command

```shell
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
DD_TRACE_DEBUG=true \
php -d datadog.service=my-php-app -S localhost:8888
```

## Environment Variable Support {#envs}

- DD_AGENT_HOST
  INI: datadog.agent_host
  The address of the host that Datakit listens on, default localhost.
- DD_TRACE_AGENT_PORT
  INI: datadog.trace.agent_port
  Datakit listening port number, default 9529.
- DD_ENV
  INI: datadog.env
  Set program environment variables.
- DD_SERVICE
  INI: datadog.service
  Set the APP name, version less than 0.47. 0 using DD_SERVICE_NAME.
- DD_SERVICE_MAPPING
  INI: datadog.service_mapping
  Rename the APM service name.
- DD_TRACE_AGENT_ATTEMPT_RETRY_TIME_MSEC
  INI: datadog.trace.agent_attempt_retry_time_msec
  IPC-based configurable circuit breakpoint retry intervals (milliseconds), default 5000.
- DD_TRACE_AGENT_CONNECT_TIMEOUT
  INI: datadog.trace.agent_connect_timeout
  Agent connection timeout (milliseconds), default 100.
- DD_TRACE_AGENT_MAX_CONSECUTIVE_FAILURES
  INI: datadog.trace.agent_max_consecutive_failures
  Maximum number of consecutive breakpoints for IPC-based configurable circuits, default 3.
- DD_TRACE_AGENT_TIMEOUT
  INI: datadog.trace.agent_timeout
  Agent data transfer timeout (milliseconds), the default is 500.
- DD_TAGS
  INI: datadog.tags
  Set the default tags.
- DD_VERSION
  INI: datadog.version
  Set the service version.
- DD_TRACE_SAMPLE_RATE
  INI: datadog.trace.smaple_rate
  Set the sampling rate from 0.0 (0%) to 1.0 (100%).
