## DCA web 

### 安装

DCA web 是 DCA 客户端的 web 版本，目前服务仅支持 docker 镜像安装。

#### 下载镜像

运行容器之前，首先通过 `docker pull` 下载 DCA 镜像。

```shell
$ docker pull pubrepo.guance.com/tools/dca
```

#### 运行容器

通过 `docker run` 命令来创建和启动 DCA 容器，容器默认暴露访问端口是 80。

```shell
$ docker run -d --name dca -p 8000:80 pubrepo.guance.com/tools/dca
```

>-d # 表示后台运行
>
>--name # 创建的容器名称
>
>-p 8000:80 # 端口映射，即将本地 8000 端口映射到容器内部的 80 端口

容器运行成功后，可以通过浏览器进行访问，http://localhost:8000。

#### 环境变量

默认情况下，DCA 会采用系统默认的配置，如果需要自定义配置，可以通过注入环境变量方式来进行修改。目前支持以下环境变量：

- **`DCA_INNER_HOST`**

  观测云的 auth API 地址，默认值为 `https://auth-api.guance.com`

- **`DCA_FRONT_HOST`**

  观测云 console API 地址，默认值为 `https://console-api.guance.com`

- **`DCA_LOG_LEVEL`**

  日志等级，取值为 `NONE | DEBUG | INFO | WARN | ERROR`，如果不需要记录日志，可设置为 `NONE`

- **`DCA_LOG_ENABLE_STDOUT`**

  默认为 `false`，日志会输出至文件中，位于 `/usr/src/dca/logs` 下。如果需要将日志写到 `stdout`，可以设置为 `true`。

示例：

```shell
$ docker run -d --name dca -p 8000:80 -e DCA_LOG_ENABLE_STDOUT=true -e DCA_LOG_LEVEL=WARN pubrepo.guance.com/tools/dca
```

