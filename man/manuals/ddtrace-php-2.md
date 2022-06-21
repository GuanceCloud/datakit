# PHP 示例
---

# 视图预览

![image.png](imgs/input-php-01.png)<br />
![image.png](imgs/input-php-02.png)<br />
![image.png](imgs/input-php-03.png)<br />
![image.png](imgs/input-php-04.png)<br />
![image.png](imgs/input-php-05.png)

# 安装部署<ddtrace>

DF默认支持所有采用opentracing协议的APM监控手段，例如<**skywalking**><**jaeger**><**zipkin**>等，此处官方推荐ddtrace接入方式，ddtrace为开源的APM监控方式，相较于其他方式，支持更多的自定义字段，也就意味着可以有足够多的标签与其他的组件进行关联，ddtrace具体接入方式详细如下：

### 前置条件

- 需要进行链路追踪的应用服务器<[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- [下载ddtrace-php-agent](https://github.com/DataDog/dd-trace-php/releases)，可根据需求下载x86、arm64或者其他版本的agent。
- <[ddtrace -php -agent 框架兼容列表](https://docs.datadoghq.com/tracing/setup_overview/compatibility_requirements/php)>


### 配置实施
php所有的部署方式均是在应用启动的环境变量中添加ddtrace-agent相关启动参数。


#### 开启datakit.conf中链路追踪inputs
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


#### 安装php拓展
```
# using RPM package (RHEL/Centos 6+, Fedora 20+)
 rpm -ivh datadog-php-tracer.rpm

# using DEB package (Debian Jessie+ , Ubuntu 14.04+ on supported PHP versions)
 dpkg -i datadog-php-tracer.deb

# using APK package (Alpine)
 apk add datadog-php-tracer.apk --allow-untrusted
```
上述命令将为默认 PHP 版本安装扩展。要安装特定 PHP 版本的扩展，请在安装前使用DD_TRACE_PHP_BIN环境变量设置目标 PHP 二进制文件的位置。
```
 export DD_TRACE_PHP_BIN=$(which php-fpm7)
```

#### Apache环境下添加php参数
如果您使用的是php-fpm的Apache，请在www.conf文件中添加ddtrace相关环境变量
```
 env[DD_AGENT_HOST] = localhost （必填）
 env[DD_TRACE_AGENT_PORT] = 9529 （必填）
 env[DD_SERVICE] = xxx    (xxx为您应用在df平台上展示的名称)
 env[DD_ENV] = ENV  (可选)
 env[DD_VERSION] = 1.0.0 (可选) 
```
**参数释义：**

- DD_ENV：自定义环境类型，可选项。
- DD_SERVICE：自定义应用名称 ，必填项。
- DD_TRACE_AGENT_PORT：数据上传端口（默认9529 ），必填项。
- DD_VERSION:应用版本，可选项。
- DD_TRACE_SAMPLE_RATE：设置采样率（默认是全采），可选项，如需采样，可设置0~1之间的数，例如0.6，即采样60%。
- DD_SERVICE_MAPPING：当前应用调用到的redis、mysql等，可通过此参数添加别名，用以和其他应用调用到的redis、mysql进行区分，可选项，应用场景：例如项目A项目B都调用了mysql，且分别调用的mysql-a，mysql-b，如没有添加mapping配置项，在df平台上会展现项目A项目B调用了同一个名为mysql的数据库，如果添加了mapping配置项，配置为mysql-a，mysql-b，则在df平台上会展现项目A调用mysql-a，项目B调用mysql-b。
- DD_AGENT_HOST：数据传输目标IP，默认为本机localhost，可选项。


#### NGINX环境下添加php参数
如果您使用的是php-fpm的Nginx，请在www.conf文件中添加ddtrace相关环境变量
```
 env[DD_AGENT_HOST] = localhost （必填）
 env[DD_TRACE_AGENT_PORT] = 9529 （必填）
 env[DD_SERVICE] = xxx    (xxx为您应用在df平台上展示的名称)
 env[DD_ENV] = ENV  (可选)
 env[DD_VERSION] = 1.0.0 (可选) 
```


#### 重启PHP服务。

浏览器访问phpinfo输出的相关页面，查看ddtrace模块是否已安装成功。<br />![image.png](imgs/input-php-06.png)

### 链路分析
<[服务](https://www.yuque.com/dataflux/doc/te4k3x)><br /><[链路分析](https://www.yuque.com/dataflux/doc/qp1efz)>

# 场景视图
DF平台已内置 应用性能监测模块，无需手动创建

# 异常检测
暂无

# 相关术语说明
<[链路追踪-字段说明](https://www.yuque.com/dataflux/doc/vc48iq#1d644644)>

# 最佳实践
<[链路追踪（APM）最佳实践](https://www.yuque.com/dataflux/bp/apm)>

# 故障排查
暂无
