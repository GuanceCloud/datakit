# PHP 示例

---

# 视图预览

![image.png](imgs/input-php-01.png)<br />
![image.png](imgs/input-php-02.png)<br />
![image.png](imgs/input-php-03.png)<br />
![image.png](imgs/input-php-04.png)<br />
![image.png](imgs/input-php-05.png)

# 安装部署<ddtrace>

「观测云」默认支持所有采用 opentracing 协议的 APM 监控手段，例如<**skywalking**><**jaeger**><**zipkin**>等，此处官方推荐 ddtrace 接入方式，ddtrace 为开源的 APM 监控方式，相较于其他方式，支持更多的自定义字段，也就意味着可以有足够多的标签与其他的组件进行关联，ddtrace 具体接入方式详细如下：

### 前置条件

- 需要进行链路追踪的应用服务器<[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- [下载 ddtrace-php-agent](https://github.com/DataDog/dd-trace-php/releases)，可根据需求下载 x86、arm64 或者其他版本的 agent。
- <[ddtrace -php -agent 框架兼容列表](https://docs.datadoghq.com/tracing/setup_overview/compatibility_requirements/php)>

### 配置实施

php 所有的部署方式均是在应用启动的环境变量中添加 ddtrace-agent 相关启动参数。

#### 开启 datakit.conf 中链路追踪 inputs

**（必须开启）**

```
###########--------linux环境---------##########

 cd /usr/local/datakit/conf.d/
 cd /ddtrace
 cp ddtrace.conf.sample ddtrace.conf


## 复制完文件后，vim进入编辑模式，放开imputs的注释
## 举例:ddtrace    tags相关注释可根据需要进行开启操作，添加业务或其他相关的标签

#默认无需修改
 vim ddtrace.conf

 wq!

## 重启datakit
 systemctl restart datakit
```

#### 安装 php 拓展

```
# using RPM package (RHEL/Centos 6+, Fedora 20+)
 rpm -ivh datadog-php-tracer.rpm

# using DEB package (Debian Jessie+ , Ubuntu 14.04+ on supported PHP versions)
 dpkg -i datadog-php-tracer.deb

# using APK package (Alpine)
 apk add datadog-php-tracer.apk --allow-untrusted
```

上述命令将为默认 PHP 版本安装扩展。要安装特定 PHP 版本的扩展，请在安装前使用 DD_TRACE_PHP_BIN 环境变量设置目标 PHP 二进制文件的位置。

```
 export DD_TRACE_PHP_BIN=$(which php-fpm7)
```

#### Apache 环境下添加 php 参数

如果您使用的是 php-fpm 的 Apache，请在 www.conf 文件中添加 ddtrace 相关环境变量

```
 env[DD_AGENT_HOST] = localhost （必填）
 env[DD_TRACE_AGENT_PORT] = 9529 （必填）
 env[DD_SERVICE] = xxx    (xxx为您应用在「观测云」平台上展示的名称)
 env[DD_ENV] = ENV  (可选)
 env[DD_VERSION] = 1.0.0 (可选)
```

**参数释义：**

- DD_ENV：自定义环境类型，可选项。
- DD_SERVICE：自定义应用名称 ，必填项。
- DD_TRACE_AGENT_PORT：数据上传端口（默认 9529 ），必填项。
- DD_VERSION:应用版本，可选项。
- DD_TRACE_SAMPLE_RATE：设置采样率（默认是全采），可选项，如需采样，可设置 0~1 之间的数，例如 0.6，即采样 60%。
- DD_SERVICE_MAPPING：当前应用调用到的 redis、mysql 等，可通过此参数添加别名，用以和其他应用调用到的 redis、mysql 进行区分，可选项，应用场景：例如项目 A 项目 B 都调用了 mysql，且分别调用的 mysql-a，mysql-b，如没有添加 mapping 配置项，在「观测云」平台上会展现项目 A 项目 B 调用了同一个名为 mysql 的数据库，如果添加了 mapping 配置项，配置为 mysql-a，mysql-b，则在「观测云」平台上会展现项目 A 调用 mysql-a，项目 B 调用 mysql-b。
- DD_AGENT_HOST：数据传输目标 IP，默认为本机 localhost，可选项。

#### NGINX 环境下添加 php 参数

如果您使用的是 php-fpm 的 Nginx，请在 www.conf 文件中添加 ddtrace 相关环境变量

```
 env[DD_AGENT_HOST] = localhost （必填）
 env[DD_TRACE_AGENT_PORT] = 9529 （必填）
 env[DD_SERVICE] = xxx    (xxx为您应用在「观测云」平台上展示的名称)
 env[DD_ENV] = ENV  (可选)
 env[DD_VERSION] = 1.0.0 (可选)
```

#### 重启 PHP 服务。

浏览器访问 phpinfo 输出的相关页面，查看 ddtrace 模块是否已安装成功。<br />![image.png](imgs/input-php-06.png)

### 链路分析

<[服务](https://www.yuque.com/dataflux/doc/te4k3x)><br /><[链路分析](https://www.yuque.com/dataflux/doc/qp1efz)>

# 场景视图

「观测云」平台已内置 应用性能监测模块，无需手动创建

# 异常检测

暂无

# 相关术语说明

<[链路追踪-字段说明](https://www.yuque.com/dataflux/doc/vc48iq#1d644644)>

# 最佳实践

<[链路追踪（APM）最佳实践](https://www.yuque.com/dataflux/bp/apm)>

# 故障排查

暂无
