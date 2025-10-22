# Redis cluster test

通过本地部署 redis 集群（cluster/master-slave）来采集 Redis 观测数据，同时利用 python 脚本给 redis 施加业务压力，以反应真实的采集效果。

## cluster 测试用法

此文档用于构建一个测试的 redis 集群，它有 6 个节点，同时准备了一个 python 脚本来构建一些业务数据，触发各种指标的采集。

- **启动 redis cluster**

```shell
docker compose -f docker-compose-cluster.yml up -d
```

- **启动 python 脚本**

```shell
# 先安装一点 python 库: pip install redis-py-cluster
python3 redis-load-gen.py
```

- **停止 redis cluster**

```shell
docker compose -f docker-compose-cluster.yml down -v
```

## master/slave 测试用法

- **启动 redis master-slave 节点**

```shell
docker compose -f docker-compose-master-slave.yml up -d
```

- **启动 python 脚本**

```shell
# 先安装一点 python 库: pip install redis-py-cluster
python3 redis-load-gen-ms.py
```

- **停止 redis cluster**

```shell
docker compose -f docker-compose-master-slave.yml down -v
```
