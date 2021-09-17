{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`linux/amd64`

# 简介

本文介绍如何在容器环境安装 DataKit。

## Docker 安装


容器启动命令如下：

```shell
sudo docker run -d \
	-v "/path/to/your/local/conf/dir":"/usr/local/cloudcare/guance/datakit/conf.d" \
	-e ENV_DATAWAY="${dataway}" \
	-e ENV_WITHIN_DOCKER=1 \
	-e ENV_ENABLE_INPUTS='cpu,mem,disk,diskio,swap,system,net' \
	-e ENV_UUID="<your-datakit-uuid>" \
	--privileged \
	--publish 9529:9529 \
	pubrepo.jiagouyun.com/datakit/datakit:v1.1.6-rc1
```

几点说明：

- DataKit 中有很多配置文件，这些文件最好在主机上编辑好挂载进去
- DataKit 需要一个 UUID 才能运行，主机安装的情况下，是通过安装程序自动生成的。容器安装时，只能提前准备好，可通过如下命令生成
- 如果无需启用 9529 端口绑定，可移除 `--publish` 设定

```shell
# Linux
$ echo -n 'dkid_'; tr -dc A-Za-z0-9 </dev/urandom | head -c 20 ; echo ''
dkid_4OZrX57npugtrri24rdP

# Mac
$ echo -n 'dkid_'; openssl rand -hex 10
dkid_facc219347e914506d25
```

- DataKit 启动过程中，支持如下几个环境变量读取：

	- `ENV_DATAWAY`：DataKit 地址
	- `ENV_WITHIN_DOCKER`：这个必须指定，否则容器中的 DataKit 运行方式会不同
	- `ENV_ENABLE_INPUTS`：指定默认开启的采集器（即无需额外配置即可生效），可按照实际需求增删。但某些采集器必须配置，就不适合做成默认开启的，比如 MySQL/Nginx 等采集器，因为它们需要一些额外的配置，如用户名、密码等。
	- `ENV_GLOBAL_TAGS`：注入 global-tag，即给所有采集的数据添加全局 tags，支持多对`key=value`，用英文逗号分隔
