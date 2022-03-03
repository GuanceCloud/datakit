### 简介
`Elasticsearch`指标采集，参考`telegraf`的采集方式，是基于http请求来获取相关指标，无侵入性。

### 配置
```
[[inputs.elasticsearch]]
  ## Elasticsearch服务器配置
  # 支持Basic认证:
  # servers = ["http://user:pass@localhost:9200"]
  servers = ["http://localhost:9200"]

  ## 采集间隔
  # 单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## HTTP超时设置
  http_timeout = "5s"

  ## 默认local是开启的，只采集当前Node自身指标，如果需要采集集群所有Node，需要将local设置为false
  local = true

  ## 设置为true可以采集cluster health
  cluster_health = false

  ## cluster health level 设置，indices (默认) 和 cluster
  # cluster_health_level = "indices"

  ## 设置为true时可以采集cluster stats.
  cluster_stats = false

  ## 只从master Node获取cluster_stats，这个前提是需要设置 local = true
  cluster_stats_only_from_master = true

  ## 需要采集的Indices, 默认为 _all
  indices_include = ["_all"]

  ## indices级别，可取值："shards", "cluster", "indices"
  indices_level = "shards"

  ## node_stats可支持配置选项有"indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http", "breaker"
  # 默认是所有
  # node_stats = ["jvm", "http"]

  ## HTTP Basic Authentication 用户名和密码
  # username = ""
  # password = ""

  ## TLS Config
  tls_open = false
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

[inputs.elasticsearch.log]
  #	files = []
  ## grok pipeline script path
  #	pipeline = "elasticsearch.p"
	
  [inputs.elasticsearch.tags]
   # a = "b"
`
```

### 指标集
当前采集器主要采集的是node、indices和cluster相关的指标，具体会产生以下指标集:
- `elasticsearch_node_stats`
- `elasticsearch_indices_stats`
- `elasticsearch_cluster_stats`
- `elasticsearch_cluster_health`


## 测试
### 安装ES(Docker)

**docker run**

```
docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:6.8.20
```

**docker-compose**

`docker-compose.yml`

```
version: '2.2'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:6.8.20
    container_name: elasticsearch
    environment:
      - cluster.name=docker-cluster
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - esdata1:/usr/share/elasticsearch/data
    ports:
      - 9200:9200
    networks:
      - esnet
  elasticsearch2:
    image: docker.elastic.co/elasticsearch/elasticsearch:6.8.20
    container_name: elasticsearch2
    environment:
      - cluster.name=docker-cluster
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
      - "discovery.zen.ping.unicast.hosts=elasticsearch"
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - esdata2:/usr/share/elasticsearch/data
    networks:
      - esnet

volumes:
  esdata1:
    driver: local
  esdata2:
    driver: local

networks:
  esnet:
```

```
# 启动
docker-compose up -d
# 删除
docker-compose down
```