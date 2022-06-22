{{.CSS}}
# Nginx
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

NGINX 采集器可以从 NGINX 实例中采取很多指标，比如请求总数连接数、缓存等多种指标，并将指标采集到观测云 ，帮助监控分析 NGINX 各种异常情况。

## 视图预览
Nginx 性能指标展示：请求数、处理请求数、活跃请求数、等待连接数等。

![image.png](../imgs/nginx-1.png)

## 版本支持
操作系统：Linux / Windows

Nginx 版本：ALL

## 前置条件

- NGINX 版本 >= 1.19.6

### 指标采集（必选）

#### 开启 VTS 相关数据采集
- NGINX 默认采集 `http_stub_status_module` 模块的数据，开启 `http_stub_status_module` 模块参见[这里](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html)，开启了以后会上报 NGINX 指标集的数据

- 如果您正在使用 [VTS](https://github.com/vozlt/nginx-module-vts) 或者想监控更多数据，建议开启 VTS 相关数据采集，可在 `{{.InputName}}.conf` 中将选项 `use_vts` 设置为 `true`。如何开启 VTS 参见[这里](https://github.com/vozlt/nginx-module-vts#synopsis)。

- 开启 VTS 功能后，能产生如下指标集：

    - `nginx`
    - `nginx_server_zone`
    - `nginx_upstream_zone` (NGINX 需配置 `upstream` 相关配置)
    - `nginx_cache_zone`    (NGINX 需配置 `cache` 相关配置)

- 以产生 `nginx_upstream_zone` 指标集为例，NGINX 相关配置示例如下：

```
    ...
    http {
       ...
       upstream your-upstreamname {
         server upstream-ip:upstream-port;
      }
       server {
       ...
       location / {
       root  html;
       index  index.html index.htm;
       proxy_pass http://yourupstreamname;
     }}}

```

- 已经开启了 VTS 功能以后，不必再去采集 `http_stub_status_module` 模块的数据，因为 VTS 模块的数据会包括 `http_stub_status_module` 模块的数据

#### 开启 `http_stub_status_module` 模块的数据采集


1. 开启 nginx_status 页面，修改主配置文件 /etc/nginx/nginx.conf (以实际路径为准)

参数说明：

- 在主站点的 server 配置里添加 location /nginx_status
- stub_status：开启 nginx_status 页面
- access_log：关闭访问日志
- allow：只允许本机访问 (127.0.0.1)
- deny all：拒绝其他访问连接
```
    server {
        listen       80 default_server;
        listen       [::]:80 default_server;
        server_name  _;
        root         /usr/share/nginx/html;

        # Load configuration files for the default server block.
        include /etc/nginx/default.d/*.conf;
        
        location /nginx_status {
            stub_status  on;
            access_log   off;
            allow 127.0.0.1;
            deny all;
            }
```

2. 使用 nginx -t 测试配置文件语法

![image.png](../imgs/nginx-2.png)

3. 重载 nginx
```
systemctl reload nginx
```

4. 查看监控数据 curl http://127.0.0.1/nginx_status (Windows 浏览器访问)

(如果配置了 server_name，使用 curl http://域名:端口/nginx_status )

![image.png](../imgs/nginx-3.png)

5. 开启 Datakit nginx 插件，复制 sample 文件
```
cd /usr/local/datakit/conf.d/nginx/
cp nginx.conf.sample nginx.conf
```

6. 修改 nginx.conf 配置文件

主要参数说明

- url：nginx status 页面地址
- interval：采集频率
- insecure_skip_verify：是否忽略安全验证 (如果是 https，请设置为 true)
- response_timeout：响应超时时间 (默认5秒)
```
[[inputs.nginx]]
        url = "http://127.0.0.1/nginx_status"
        interval = "60s"
        insecure_skip_verify = false
        response_timeout = "5s"
```

7. 重启 Datakit (如果需要开启日志，请配置日志采集再重启)
```
systemctl restart datakit
```


### 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

### 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

#### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 


### 日志采集(非必选)

如需采集 NGINX 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 NGINX 日志文件的绝对路径。比如：

```
    [[inputs.nginx]]
      ...
      [inputs.nginx.log]
		files = ["/var/log/nginx/access.log","/var/log/nginx/error.log"]
```


开启日志采集以后，默认会产生日志来源（`source`）为 `nginx` 的日志。

>注意：必须将 DataKit 安装在 NGINX 所在主机才能采集 NGINX 日志


#### 日志 pipeline 功能切割字段说明

- NGINX 错误日志切割

错误日志文本示例：
```
2021/04/21 09:24:04 [alert] 7#7: *168 write() to "/var/log/nginx/access.log" failed (28: No space left on device) while logging request, client: 120.204.196.129, server: localhost, request: "GET / HTTP/1.1", host: "47.98.103.73"
```

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
|  status   | error     | 日志等级(alert转成了error) |
|  client_ip   | 120.204.196.129     | client ip地址 |
|  server   | localhost     | server 地址 |
|  http_method   | GET     | http 请求方式 |
|  http_url   | /     | http 请求url |
|  http_version   | 1.1     | http version |
|  ip_or_host   | 47.98.103.73     | 请求方ip或者host |
|  msg   | 7#7: *168 write()...host: \"47.98.103.73     | 日志内容 |
|  time   | 1618968244000000000     | 纳秒时间戳（作为行协议时间）|

错误日志文本示例：

```
2021/04/29 16:24:38 [emerg] 50102#0: unexpected ";" in /usr/local/etc/nginx/nginx.conf:23
```

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
|  status   | error     | 日志等级(emerg转成了error) |
|  msg   | 50102#0: unexpected \";\" in /usr/local/etc/nginx/nginx.conf:23    | 日志内容 |
|  time   | 1619684678000000000     | 纳秒时间戳（作为行协议时间）|

- NGINX 访问日志切割

访问日志文本示例:
```
127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"
```

切割后的字段列表如下：

| 字段名  |  字段值  | 说明 |
| ---    | ---     | --- |
|  client_ip   | 127.0.0.1     | 日志等级(emerg转成了error) |
|  status   | ok    | 日志等级 |
|  status_code   | 200    | http code |
|  http_method   | GET     | http 请求方式 |
|  http_url   | /basic_status     | http 请求url |
|  http_version   | 1.1     | http version |
|  agent   | Mozilla/5.0... Safari/537.36     | User-Agent |
|  browser   |   Chrome   | 浏览器 |
|  browserVer   |   89.0.4389.72   | 浏览器版本 |
|  isMobile   |   false   | 是否手机 |
|  engine   |   AppleWebKit   | 引擎 |
|  os   |   Intel Mac OS X 11_1_0   | 系统 |
|  time   | 1619243659000000000     | 纳秒时间戳（作为行协议时间）|

### 链路采集(非必选)
某些场景下，我们需要将前端负载均衡也纳入到全链路观测中，用于分析用户请求从系统入口位置到后端服务结束这一完整过程的链路调用及耗时情况。这时就需要安装Nginx链路追踪模块来实现该功能。

安装Nginx链路追踪有两个前置条件，首先是安装Nginx的OpenTracing插件 [linux-amd64-nginx-${NGINX_VERSION}-ot16-ngx_http_module.so.tgz](https://github.com/opentracing-contrib/nginx-opentracing/releases/latest), 点击链接打开git目录后，在Asset中根据自己的Nginx版本选择对应的模块包进行下载，将这个包解压到nginx模块目录下，通常为/usr/lib/nginx/modules。也可解压到其他目录，区别是在下面操作load_module时，需要引用绝对路径。

其次是需要安装ddagent运行所依赖的C++插件 [linux-amd64-libdd_opentracing_plugin.so.gz](https://github.com/DataDog/dd-opentracing-cpp/releases/latest)。 这个包需要解压到nginx可访问的某个目录下，例如/usr/local/lib。

完成插件包下载后，可使用下面的命令解压：<br />Nginx OpenTracing包：<br />tar zxf linux-amd64-nginx-<填写您下载的.so的版本号>-ot16-ngx_http_module.so.tgz -C /usr/lib/nginx/modules

ddagent Cpp支持包：<br />gunzip linux-amd64-libdd_opentracing_plugin.so.gz -c > /usr/local/lib/libdd_opentracing_plugin.so 

解压完成后，首先配置nginx.conf，加载Nginx OpenTracing 模块：

> #ps -ef | grep nginx 、nginx -V或whereis nginx命令查找您环境中nginx的安装位置<br />#cd到该目录下,进入conf文件夹，vi 编辑nginx.conf<br />#增加如下命令，加载Nginx OpenTracing 模块。注意需要加在event配置之前：<br />`load_module modules/ngx_http_opentracing_module.so;`

在nginx.conf的http配置项中，增加如下内容：

```
opentracing on; # Enable OpenTracing
opentracing_tag http_user_agent $http_user_agent; 
opentracing_trace_locations off; 

opentracing_load_tracer /usr/local/lib/libdd_opentracing_plugin.so /etc/nginx/dd-config.json;
```

其中opentracing_load_tracer的配置需要注意，第一个参数是C++插件的位置，这个已经在拷贝命令中添加好了。第二个参数dd-config.json需要手动添加。以示例中的位置为例，我们在/etc/nginx/目录下，vi dd-config.json并填写以下内容：
```
{
  "environment": "prod",
  "service": "nginx",
  "operation_name_override": "nginx.handle",
  "agent_host": "localhost",
  "agent_port": 9529
}

```
其中,agent_host需填写本地可访问的datakit地址，agent_port须填写datakit端口号9529。

下一步，编辑nginx日志格式，将Trace信息注入到Nginx日志中。可按如下示例编辑：

```
log_format with_trace_id '$remote_addr - $http_x_forwarded_user [$time_local] "$request" '
                         '$status $body_bytes_sent "$http_referer" '
                         '"$http_user_agent" "$http_x_forwarded_for" '
                         '"$opentracing_context_x_datadog_trace_id" "$opentracing_context_x_datadog_parent_id"';

access_log /var/log/nginx/access-with-trace.log with_trace_id;
```

说明:log_format关键字告诉nginx这里定义了一套日志规则，with_trace_id是规则名，可以自己修改，注意在下方指定日志路径时要用一样的名字来关联该日志的规则。access_log中的路径和文件名可以更换。通常情况下原nginx是配有日志规则的。我们可以配置多条规则，并将不同的日志格式输出到不同的文件，即保留原access_log规则及路径不变，新增一个包含trace信息的日志规则，命名为不同的日志文件，供不同的日志工具读取。

完成上述配置后，在http.server需要进行追踪的location配置中，增加如下内容：
```
opentracing_operation_name "$request_method $uri";             
opentracing_propagate_context;

opentracing_tag "custom-tag" "special value";#用户自定义标签，可选
```

配置完成后保存并退出nginx.conf，首先使用nginx -t进行基本的语法检查，在注入Nginx Trace模块之前，检查结果仅显示nginx本身的内容：
![](../imgs/nginx-5.png)

如成功配置Nginx Trace模块，则再次使用nginx -t进行语法检查时，会提示ddtrace的相关配置信息：

![](../imgs/nginx-6.png)


使用nginx -s reload重新加载nginx，使tracing功能生效。登录观测云的应用性能监控界面，查看Nginx Tracing信息：

![image.png](../imgs/nginx-7.png)



可能遇到的问题：<br />1、在进行nginx语法检查时报错，提示没有找到OpenTracing的module

![](../imgs/nginx-8.png)

这个报错说明您环境中的nginx保存Modules的路径并不是/usr/lib/nginx/modules，这时可以根据报错提示的路径，将Nginx OpenTracing包拷贝到您环境中nginx的模块引用位置。或在配置nginx.conf时，使用OpenTrace so文件所在位置的绝对路径。

2、在进行nginx语法检查时报错，提示“Nginx is not binary compatible...”类错误。<br />产生这个错误的可能原因为您本地使用的Nginx为编译安装版本，与本例中提供的OpenTracing模块的包签名不一致，导致出现兼容性问题。建议的解决方法为：通过本例提供的Module下载链接，找到Nginx_OpenTracing的代码仓库，将代码下载到本地。注意需要根据您现在所使用的Nginx版本来进行选择，例如Nginx-Opentracing Release 0.24.x版本支持的Nginx最低要求为1.13.x(可以通过github项目中记录的已经编译好的包的版本号来确认),如果低于这个版本的Nginx，需要在历史Release版本中查找对应的源代码版本。

找到对应版本后，停用本地nginx。将Nginx-Opentracing的代码拷贝到本地并解压。进入到nginx代码路径使用configure重建objt时，增加--add-dynamic-module=/path/to/your/module(该路径指向您本地保存ddagent代码)，可以在nginx目录下使用./configure命令直接添加。另外需要注意，tracing模块的重新编译依赖OpenTracingCPP公共包，需要一并下载这个包用于编译：<br />相关帮助信息：<br />[https://github.com/opentracing-contrib/nginx-opentracing](https://github.com/opentracing-contrib/nginx-opentracing)

![image.png](../imgs/nginx-9.png)

OpenTracingCPP下载地址：

[https://github.com/opentracing/opentracing-cpp/releases/tag/v1.6.0](https://github.com/opentracing/opentracing-cpp/releases/tag/v1.6.0)

编译安装步骤简述：<br />1、编译 OpenTracing CPP 库，生成 libopentracing.so，这个库后续用于 Nginx 调用 OT 接口生成 trace 信息。使用上面的 opentracing-cpp 链接将代码下载到编译环境本地。进入代码目录，按顺序执行以下命令：<br />`mkdir .build `<br />`cd .build `<br />`cmake .. `<br />`make sudo make install`<br />第一步操作生成编译临时目录。cd进入该目录后调用cmake执行编译，编译结果将保存在.bulid目录中。这里需要注意最新版本的opentracing-cpp编译需要cmake版本高于3.1，如使用操作系统默认版本的cmake，可能报版本过低的错误，可以通过yum install -7 cmake3的方式安装3.x版本。如安装提示找不到包，可尝试将yum源配置为国内安装源后重试。

2、下载 nginx-opentracing 代码，解压到本地并编译Nginx以加载该模块<br />下载 nginx-opentracing 代码：[https://github.com/opentracing-contrib/nginx-opentracing](https://github.com/opentracing-contrib/nginx-opentracing)<br />保存在本地任意路径<br />下载您所需要版本的Nginx，下载后解压到本地目录，增加opentracing模块并启动编译：<br />`tar zxvf nginx-1.xx.x.tar.gz `<br />`cd nginx-1.x.x `<br />增加模块时，--add参数后填写指向之前下载的nginx-opentracing代码目录中的opentracing<br />`./configure --add-dynamic-module=/absolute/path/to/nginx-opentracing/opentracing`<br />操作完成后启动编译<br />`make && sudo make install`<br />这一步操作可能会遇到较多头文件问题，因为ng opentracing的头文件路径以<>包含，编译器默认在/usr/include下查找文件，找不到即会编译报错。处理方式是在/usr/include下，创建指向头文件所在目录的软连接。使make可以找到这些文件。具体的头文件在哪个地方，可使用find等检索命令查找。<br />3、下载ddagent的nginx trace插件，在nginx.conf中开启trace_plugin使链路追踪生效：<br />wget -O - [https://github.com/DataDog/dd-opentracing-cpp/releases/download/v0.3.0/linux-amd64-libdd_opentracing_plugin.so.gz](https://github.com/DataDog/dd-opentracing-cpp/releases/download/v0.3.0/linux-amd64-libdd_opentracing_plugin.so.gz) | gunzip -c > /usr/local/lib/libdd_opentracing_plugin.so<br />注意：这里下载的ddplugin版本需要与opentracing lib版本对应，如opentracing版本与libdd_opentracing_plugin有差异，需要下载ddtrace代码到本地，基于本地环境opentracing库重新编译生成插件。本方案提供的版本：[https://github.com/opentracing/opentracing-cpp/releases/tag/v1.6.0](https://github.com/opentracing/opentracing-cpp/releases/tag/v1.6.0) 与ddagent插件环境匹配，下载的插件可以直接使用。后续不排除otracing库及ddagent同步更新的情况，请在使用前注意核对版本号。

本例中的 nginx 版本使用yum安装，考虑到编译安装的实施及调试成本，建议采用 yum 安装的版本进行尝试。安装方法 (CentOS7 为例)：<br />增加 Nginx 安装源：

rpm -Uvh http://nginx.org/packages/centos/7/noarch/RPMS/nginx-release-centos-7-0.el7.ngx.noarch.rpm<br />yum 安装指定版本 nginx,版本号1.x.xx：<br /> yum -y install nginx-1.x.xx

3、在接入 nginx tracing 后，发现观测云界面上 nginx 的 tracing 数据没有和 location 指向的后端应用关联起来。

这个问题产生的原因是 nginx tracing 模块生成的链路追踪 ID 没有被同时转发到后端服务，导致后端生成了新的 traceid，在界面上被识别为两个不同的链路，解决方法是在需要追踪的 Location 中，配置 http_header 的转发：

proxy_set_header X-datadog-trace-id $opentracing_context_x_datadog_trace_id;<br />proxy_set_header X-datadog-parent-id $opentracing_context_x_datadog_parent_id;

这里的参数为固定值，X-datadog-***为ddagent识别header转发字段的参数，opentracing_context_*为OpenTracing模块的traceid参数。

配置完成后保存并退出nginx.conf,使用nginx -s reload重启服务。

## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Nginx 监控视图>

## 常见问题排查
- [无数据上报排查](why-no-data.md)
## 进一步阅读
- [观测云 Nginx 可观测最佳实践](/best-practices/integrations/nginx.md)

- [八分钟带你深入浅出搞懂 Nginx](https://zhuanlan.zhihu.com/p/34943332)