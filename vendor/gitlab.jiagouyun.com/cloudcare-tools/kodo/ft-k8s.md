# ft k8s 使用文档汇总

## k8s 镜像仓库(hub)使用手册

以线上 k8s demo 环境为例。登陆 [hub 地址](http://172.16.0.43:30002/harbor/sign-in?redirect_url=%2Fharbor%2Fprojects)

用户名：`admin`
密码： `Harbor12345`

本地 docker 设置：编辑 `/etc/docker/daemon.json`(一般这个目录) 如下：

	{
		"insecure-registries" : ["172.16.0.43:30002"]
	}

登陆这个私有仓库 `sudo docker login 172.16.0.43:30002`，输入上面的用户名/密码（可能需要）

假定你本地已经通过 Dockerfile 构建了一个名为 `ftdataway` 的镜像，如

	$ docker images
	REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
	ftdataway           latest              064b6528f4b5        4 days ago          84.9MB

执行 `docker tag SOURCE_IMAGE[:TAG] 172.16.0.43:30002/ft2.0/IMAGE[:TAG]`，将 ftdataway 镜像打个 tag，如:

	$ docker tag ftdataway:latest 172.16.0.43:30002/ft2.0/ftdataway:latest
	$ docker images
	REPOSITORY                          TAG                 IMAGE ID            CREATED             SIZE
	172.16.0.43:30002/ft2.0/ftdataway   latest              064b6528f4b5        4 days ago          84.9MB
	ftdataway                           latest              064b6528f4b5        4 days ago          84.9MB

这样镜像就已经准备好了，然后执行 `docker push 172.16.0.43:30002/ft2.0/IMAGE[:TAG]` 将镜像推到 hub 上:

	$ docker push 172.16.0.43:30002/ft2.0/ftdataway:latest

这样，在 hub 页面上，就能看到该镜像了。

下面再验证镜像的拉取（可在另一台机器上实验）。先删掉本地刚刚 tag 的镜像，然后再拉取 hub 上的镜像：

	$ docker rmi 172.16.0.43:30002/ft2.0/ftdataway
	$ docker pull 172.16.0.43:30002/ft2.0/ftdataway

如此本地就有了刚刚 push 上去的镜像。

## k8s 监控部署

目前暂定使用 telegraf 来采集 k8s 的运行指标。本文档是 telegraf 以及 k8s node 的配置说明。

### k8s 配置

在 k8s node 节点上，确保文件 `/var/lib/kubelet/kubeadm-flags.env` 中开启了 10255 这个只读端口，其配置如下。如果没有则追加上去。

	--read-only-port=10255

修改完成后，重启 `kubelet.service` 服务 (`systemctl restart kubelet.service`)

### telegraf 配置

编辑 telegraf 配置文件，找到 `[[inputs.kubernetes]]`，编辑其中的 `url` 以及 `bearer_token_string` 字段。如

	[[inputs.kubernetes]]
		url = "http://<k8s-node-host>:10255"
		bearer_token_string = "abc-123"

其中 `bearer_token_string` 通过在 k8s node 节点执行如下命令获取

	kubectl get secrets | grep ^default | cut -f1 -d ' '

需要注意的是，一个 telegraf [可以配置成抓取多个](https://docs.influxdata.com/telegraf/v1.12/administration/configuration/#multiple-inputs-of-the-same-type) k8s node，多加几个 `[[inputs.kubernetes]]` 即可。

然后再配置 telegraf 的 http output:

	[[outputs.http]]
		url = "http://<ft-dataway-host>:9528/v1/write/metrics"
		[outputs.http.headers]
			User-Agent = "telegraf"
			X-Datakit-UUID = "dkit-<UUID>"
