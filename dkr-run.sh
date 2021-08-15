#!/bin/bash
version=`git describe --always --tags`
container_name=${USER}-datakit            # 以本地用户名命名的 datakit
host_confd=$HOME/datakit-confd            # 将 HOME 目录下的 datakit-confd 作为 datakit/confd 目录，便于在主机上编辑，不用再登入容器修改

# 大家自行配置 dataway
dataway="https://openway.dataflux.cn?token=tkn_f2b9920f05d84d6bb5b14d9d39db1dd3,https://openway.dataflux.cn?token=tkn_df4e067a4324460abd1378dcb58030d0"

# 绑定宿主机上的端口映射为 DataKit 的 HTTP 端口，自行改之
host_port=19529

# 将 datakit/agent 的配置文件和日志映射到 host 的 HOME 目录下
mkdir -p ${host_confd}
sudo truncate -s 0 $HOME/dk.log
sudo truncate -s 0 $HOME/dk-gin.log

# 停掉老的容器
sudo docker stop $container_name
sudo docker rm $container_name

# 从本地的编译包构建本地 docker 镜像
img="registry.jiagouyun.com/datakit/datakit:${version}"
sudo docker rmi $img
sudo docker build -t $img .

# 启动容器
sudo docker run -d --name=$container_name \
	-v "${host_confd}":"/usr/local/datakit/conf.d" \
	--mount type=bind,source="$HOME/dk.log",target="/var/log/datakit/log" \
	--mount type=bind,source="$HOME/dk-gin.log",target="/var/log/datakit/gin.log" \
	-e ENV_DATAWAY="${dataway}" \
	-e ENV_DEFAULT_ENABLED_INPUTS="cpu,mem,disk,hostobject" \
	-e ENV_DISABLE_PROTECT_MODE="true" \
	-e ENV_ENABLE_ELECTION="" \
	-e ENV_ENABLE_PPROF="true" \
	-e ENV_GLOBAL_TAGS="env:run-in-docker" \
	-e ENV_HOSTNAME="1024.coding" \
	-e ENV_HTTP_LISTEN="0.0.0.0:9529" \
	-e ENV_LOG_LEVEL="debug" \
	-e ENV_NAME="testing-datakit" \
	-e ENV_RUM_ORIGIN_IP_HEADER="X-Forwawrded-For" \
	--privileged \
	--publish $host_port:9529 \
	$img

# - ENV_DATAWAY":                "http://host1.org,http://host2.com",
# - ENV_DEFAULT_ENABLED_INPUTS": "cpu,mem,disk",
# - ENV_DISABLE_PROTECT_MODE":   "true",
# - ENV_ENABLE_ELECTION":        "1",
# - ENV_ENABLE_PPROF":           "true",
# - ENV_GLOBAL_TAGS":            "a=b,c=d",
# - ENV_HOSTNAME":               "1024.coding",
# - ENV_HTTP_LISTEN":            "localhost:9559",
# - ENV_LOG_LEVEL":              "debug",
# - ENV_NAME":                   "testing-datakit",
# - ENV_RUM_ORIGIN_IP_HEADER":   "not-set",
