# PHP profiling

---

从 Datakit [:octicons-tag-24: Version-1.13.0](../datakit/changelog.md#cl-1.13.0){:target="_blank"} 开始支持使用 [`dd-trace-php`](https://github.com/DataDog/dd-trace-php){:target="_blank"} 作为 `PHP` 项目的应用性能监测工具。

## 前置条件 {#prerequisites}

- `Linux X64 with glibc 2.17+，Linux X64 with musl v1.2+`
- `PHP 7.1+ NTS(Non-Thread Safe)`

## 安装 `dd-trace-php` {#install-dd-trace-php}

下载安装脚本 [*datadog-setup.php*](https://github.com/DataDog/dd-trace-php/releases/download/0.90.0/datadog-setup.php){:target="_blank"} 并执行：

```shell
wget https://github.com/DataDog/dd-trace-php/releases/download/0.90.0/datadog-setup.php
php datadog-setup.php --enable-profiling
```

安装过程中脚本会自动检测当前系统安装的 `php` 和 `php-fpm` 路径，并让你选择哪些程序需要开启 profiling，根据你的需要输入相应的序号即可。

```shell
Searching for available php binaries, this operation might take a while.
Multiple PHP binaries detected. Please select the binaries the datadog library will be installed to:

   1. php --> /usr/bin/php8.1
   2. php8.1 --> /usr/bin/php8.1
   3. php-fpm8.1 --> /usr/sbin/php-fpm8.1

Select binaries using their number. Multiple binaries separated by space (example: 1 3): 1 2 3
```


如果你知道当前系统的 `php` 或 `php-fpm` 的安装路径，也可以在执行安装脚本的时候通过 `--php-bin` 参数直接指定 `php` 的路径，这样可以跳过上述检测和选择的步骤，例如：

```shell
php datadog-setup.php --enable-profiling --php-bin=/usr/bin/php8.1 --php-bin=/usr/sbin/php-fpm8.1
```

程序安装过程中需要从 `github.com` 下载安装包，根据你的网络情况可能需要一些时间，请等待安装程序成功退出，之后可以执行命令 `php --ri "datadog-profiling"` 验证安装是否成功。

```shell
php --ri "datadog-profiling"

datadog-profiling

Version => 0.90.0
Profiling Enabled => true
Experimental CPU Time Profiling Enabled => true
Allocation Profiling Enabled => true
...
```


<!-- markdownlint-disable MD046 -->
???+ Note

    在编写此文档时 `Datakit` 最高支持到 [`dd-trace-php v0.90.0`](https://github.com/DataDog/dd-trace-php/releases/tag/0.90.0){:target="_blank"} 版本，更高的版本没有经过系统性测试，兼容性未知，如您在使用中遇到任何问题，可随时与我们联系。
<!-- markdownlint-enable -->


## 开启 Profiling {#start-profiling}

<!-- markdownlint-disable MD046 -->
=== "PHP CLI"

    启动 PHP 脚本时设置如下环境变量：
    ```shell
    DD_PROFILING_ENABLED=1 \
    DD_PROFILING_ENDPOINT_COLLECTION_ENABLED=1 \
    DD_PROFILING_ALLOCATION_ENABLED=1 \
    DD_PROFILING_EXPERIMENTAL_CPU_TIME_ENABLED=1 \
    DD_PROFILING_EXPERIMENTAL_TIMELINE_ENABLED=1 \
    DD_PROFILING_LOG_LEVEL=info \
    DD_SERVICE=my-php-cli-app \
    DD_ENV=prod \
    DD_VERSION=1.22.333 \
    DD_AGENT_HOST=127.0.0.1 \
    DD_TRACE_AGENT_PORT=9529 \
    php <your-path-to>.php
    ```

=== "PHP-FPM"

    在 `PHP-FPM` 的进程池配置目录 *pool.d* 下的配置文件中（默认只有一个 *www.conf* 配置文件）使用 `env` 指令在 *www.conf* 添加如下环境变量：

    ???+ Note
        可以尝试使用 `php-fpm -i | grep Configuration` 来确认 `php-fpm` 当前加载配置文件的位置：
        ```shell
        php-fpm8.1 -i | grep Configuration
        Configuration File (php.ini) Path => /etc/php/8.1/fpm
        ...
        ls -l /etc/php/8.1/fpm/pool.d
        total 48
        -rw-r--r-- 1 root root 20901 Aug  7 21:22 www.conf
        ```
        如果你的 `php-fpm` 程序被注册为 `systemd` 服务，也可以尝试使用命令 `systemctl status php-fpm.service | grep .conf` 来查看当前的配置文件路径：
        ```shell
        systemctl status php8.1-fpm.service | grep .conf
        Process: 629814 ExecStart=/usr/sbin/php-fpm8.1 --nodaemonize --fpm-config /etc/php/8.1/fpm/php-fpm.conf (code=exited, status=0/SUCCESS)
        Process: 629818 ExecStartPost=/usr/lib/php/php-fpm-socket-helper install /run/php/php-fpm.sock /etc/php/8.1/fpm/pool.d/www.conf 81 (code=exited, status=0/SUCCESS)
        Process: 630872 ExecStopPost=/usr/lib/php/php-fpm-socket-helper remove /run/php/php-fpm.sock /etc/php/8.1/fpm/pool.d/www.conf 81 (code=exited, status=0/SUCCESS)
        ...
        ```


    ```shell
    ...
    ; Pass environment variables like LD_LIBRARY_PATH. All $VARIABLEs are taken from
    ; the current environment.
    ; Default Value: clean env
    ;env[HOSTNAME] = $HOSTNAME
    ;env[PATH] = /usr/local/bin:/usr/bin:/bin
    ;env[TMP] = /tmp
    ;env[TMPDIR] = /tmp
    ;env[TEMP] = /tmp

    env[DD_PROFILING_ENABLED]=true
    env[DD_PROFILING_ENDPOINT_COLLECTION_ENABLED]=true
    env[DD_PROFILING_ALLOCATION_ENABLED]=true
    env[DD_PROFILING_EXPERIMENTAL_CPU_TIME_ENABLED]=true
    env[DD_PROFILING_EXPERIMENTAL_TIMELINE_ENABLED]=true
    env[DD_PROFILING_LOG_LEVEL]=info
    env[DD_SERVICE]=my-fpm-app
    env[DD_ENV]=dev
    env[DD_VERSION]=1.2.33
    env[DD_AGENT_HOST]=127.0.0.1
    env[DD_TRACE_AGENT_PORT]=9529

    ```

    重启 `php-fpm` 并访问你的项目。
<!-- markdownlint-enable -->

稍等几分钟后便可以在 [观测云控制台](https://console.guance.com/tracing/profile){:target="_blank"} 查看相关数据。

