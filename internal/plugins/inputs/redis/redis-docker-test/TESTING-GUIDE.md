# Redis 多版本多模式测试指南

本目录提供了四种 Redis 部署模式的测试环境，每种模式包含 5 个 Redis 版本，**使用不同端口可同时运行所有版本**。

## 统一配置

- **IP 地址**: `192.168.139.162`
- **密码**: `abc123456`
- **网络模式**: host

## 端口分配方案

### 一、单机模式

| 版本 | 端口 |
|------|------|
| 7.0.11 | 6379 |
| 6.2.12 | 6380 |
| 6.0.8 | 6381 |
| 5.0.14 | 6382 |
| 4.0.14 | 6383 |

**启动所有版本：**
```bash
docker-compose -f docker-compose-standalone-multi-version.yml up -d
```

**单独启动某个版本：**
```bash
docker-compose -f docker-compose-standalone-multi-version.yml up -d redis-7
docker-compose -f docker-compose-standalone-multi-version.yml up -d redis-6.2
```

**验证连接：**
```bash
redis-cli -h 192.168.139.162 -p 6379 -a abc123456 PING  # Redis 7
redis-cli -h 192.168.139.162 -p 6380 -a abc123456 PING  # Redis 6.2
redis-cli -h 192.168.139.162 -p 6381 -a abc123456 PING  # Redis 6.0
```

### 二、主从复制模式

每个版本使用 3 个端口（1主2从）：

| 版本 | 主节点 | 从节点1 | 从节点2 |
|------|--------|---------|---------|
| 7.0.11 | 6379 | 6479 | 6579 |
| 6.2.12 | 6380 | 6480 | 6580 |
| 6.0.8 | 6381 | 6481 | 6581 |
| 5.0.14 | 6382 | 6482 | 6582 |
| 4.0.14 | 6383 | 6483 | 6583 |

**启动所有版本：**
```bash
docker-compose -f docker-compose-replication-multi-version.yml up -d
```

**单独启动某个版本：**
```bash
# Redis 7
docker-compose -f docker-compose-replication-multi-version.yml up -d \
  redis-7-master redis-7-slave-1 redis-7-slave-2

# Redis 6.2
docker-compose -f docker-compose-replication-multi-version.yml up -d \
  redis-6.2-master redis-6.2-slave-1 redis-6.2-slave-2
```

**验证主从复制：**
```bash
# Redis 7 主节点
redis-cli -h 192.168.139.162 -p 6379 -a abc123456 INFO replication

# Redis 6.2 主节点
redis-cli -h 192.168.139.162 -p 6380 -a abc123456 INFO replication
```

### 三、哨兵模式

每个版本使用 6 个端口（1主2从+3哨兵）：

| 版本 | Redis主 | Redis从1 | Redis从2 | Sentinel1 | Sentinel2 | Sentinel3 | Master Name |
|------|---------|----------|----------|-----------|-----------|-----------|-------------|
| 7.0.11 | 6379 | 6479 | 6579 | 26379 | 26479 | 26579 | mymaster-7 |
| 6.2.12 | 6380 | 6480 | 6580 | 26380 | 26480 | 26580 | mymaster-6.2 |
| 6.0.8 | 6381 | 6481 | 6581 | 26381 | 26481 | 26581 | mymaster-6.0 |
| 5.0.14 | 6382 | 6482 | 6582 | 26382 | 26482 | 26582 | mymaster-5 |
| 4.0.14 | 6383 | 6483 | 6583 | 26383 | 26483 | 26583 | mymaster-4 |

**启动所有版本：**
```bash
docker-compose -f docker-compose-sentinel-multi-version.yml up -d
```

**单独启动某个版本：**
```bash
# Redis 7 哨兵集群
docker-compose -f docker-compose-sentinel-multi-version.yml up -d \
  redis-7-master redis-7-slave-1 redis-7-slave-2 \
  sentinel-7-1 sentinel-7-2 sentinel-7-3

# Redis 6.2 哨兵集群
docker-compose -f docker-compose-sentinel-multi-version.yml up -d \
  redis-6.2-master redis-6.2-slave-1 redis-6.2-slave-2 \
  sentinel-6.2-1 sentinel-6.2-2 sentinel-6.2-3
```

**验证哨兵状态：**
```bash
# Redis 7 哨兵
redis-cli -h 192.168.139.162 -p 26379 SENTINEL get-master-addr-by-name mymaster-7

# Redis 6.2 哨兵
redis-cli -h 192.168.139.162 -p 26380 SENTINEL get-master-addr-by-name mymaster-6.2
```

### 四、集群模式

每个版本使用 6 个端口（3主3从）：

| 版本 | 节点端口 |
|------|----------|
| 7.0.11 | 7001-7006 |
| 6.2.12 | 7101-7106 |
| 6.0.8 | 7201-7206 |
| 5.0.14 | 7301-7306 |
| 4.0.14 | 7401-7406 |

**启动所有版本（含自动初始化）：**
```bash
docker-compose -f docker-compose-cluster-multi-version.yml up -d
```

**单独启动某个版本：**
```bash
# Redis 7 集群
docker-compose -f docker-compose-cluster-multi-version.yml up -d \
  redis-7-node-1 redis-7-node-2 redis-7-node-3 \
  redis-7-node-4 redis-7-node-5 redis-7-node-6 \
  redis-7-cluster-creator

# Redis 6.2 集群
docker-compose -f docker-compose-cluster-multi-version.yml up -d \
  redis-6.2-node-1 redis-6.2-node-2 redis-6.2-node-3 \
  redis-6.2-node-4 redis-6.2-node-5 redis-6.2-node-6 \
  redis-6.2-cluster-creator
```

**验证集群状态：**
```bash
# Redis 7 集群
redis-cli -h 192.168.139.162 -p 7001 -a abc123456 CLUSTER INFO
redis-cli -h 192.168.139.162 -p 7001 -a abc123456 CLUSTER NODES

# Redis 6.2 集群
redis-cli -h 192.168.139.162 -p 7101 -a abc123456 CLUSTER INFO
redis-cli -h 192.168.139.162 -p 7101 -a abc123456 CLUSTER NODES

# 查看初始化日志
docker logs redis-cluster-7-creator
docker logs redis-cluster-6.2-creator
```

## DataKit 配置示例

### 单机模式

```toml
# Redis 7.0.11
[[inputs.redis]]
  host = "192.168.139.162"
  port = 6379
  password = "abc123456"
  interval = "10s"

# Redis 6.2.12
[[inputs.redis]]
  host = "192.168.139.162"
  port = 6380
  password = "abc123456"
  interval = "10s"
```

### 主从复制模式

```toml
# Redis 7.0.11 主从
[[inputs.redis]]
  [inputs.redis.master_slave]
    master = "192.168.139.162:6379"
    slaves = [
      "192.168.139.162:6479",
      "192.168.139.162:6579",
    ]
  password = "abc123456"
  interval = "10s"

# Redis 6.2.12 主从
[[inputs.redis]]
  [inputs.redis.master_slave]
    master = "192.168.139.162:6380"
    slaves = [
      "192.168.139.162:6480",
      "192.168.139.162:6580",
    ]
  password = "abc123456"
  interval = "10s"
```

### 哨兵模式

```toml
# Redis 7.0.11 哨兵
[[inputs.redis]]
  [inputs.redis.master_slave]
    [inputs.redis.master_slave.sentinel]
      master_name = "mymaster-7"
      sentinel_addrs = [
        "192.168.139.162:26379",
        "192.168.139.162:26479",
        "192.168.139.162:26579",
      ]
  password = "abc123456"
  interval = "10s"

# Redis 6.2.12 哨兵
[[inputs.redis]]
  [inputs.redis.master_slave]
    [inputs.redis.master_slave.sentinel]
      master_name = "mymaster-6.2"
      sentinel_addrs = [
        "192.168.139.162:26380",
        "192.168.139.162:26480",
        "192.168.139.162:26580",
      ]
  password = "abc123456"
  interval = "10s"
```

### 集群模式

```toml
# Redis 7.0.11 集群
[[inputs.redis]]
  [inputs.redis.cluster]
    addrs = [
      "192.168.139.162:7001",
      "192.168.139.162:7002",
      "192.168.139.162:7003",
    ]
  password = "abc123456"
  interval = "10s"

# Redis 6.2.12 集群
[[inputs.redis]]
  [inputs.redis.cluster]
    addrs = [
      "192.168.139.162:7101",
      "192.168.139.162:7102",
      "192.168.139.162:7103",
    ]
  password = "abc123456"
  interval = "10s"
```

## 批量测试脚本

### 快速测试所有单机版本

```bash
#!/bin/bash
echo "启动所有单机版本..."
docker-compose -f docker-compose-standalone-multi-version.yml up -d
sleep 5

# 测试所有版本
for port in 6379 6380 6381 6382 6383; do
  echo "测试端口 $port..."
  redis-cli -h 192.168.139.162 -p $port -a abc123456 PING
done

echo "测试完成"
```

### 快速测试所有主从版本

```bash
#!/bin/bash
echo "启动所有主从版本..."
docker-compose -f docker-compose-replication-multi-version.yml up -d
sleep 10

# 测试所有主节点
for port in 6379 6380 6381 6382 6383; do
  echo "测试主节点 $port..."
  redis-cli -h 192.168.139.162 -p $port -a abc123456 INFO replication | grep role
done

echo "测试完成"
```

### 快速测试所有哨兵版本

```bash
#!/bin/bash
echo "启动所有哨兵版本..."
docker-compose -f docker-compose-sentinel-multi-version.yml up -d
sleep 15

# 测试所有哨兵
redis-cli -h 192.168.139.162 -p 26379 SENTINEL get-master-addr-by-name mymaster-7
redis-cli -h 192.168.139.162 -p 26380 SENTINEL get-master-addr-by-name mymaster-6.2
redis-cli -h 192.168.139.162 -p 26381 SENTINEL get-master-addr-by-name mymaster-6.0
redis-cli -h 192.168.139.162 -p 26382 SENTINEL get-master-addr-by-name mymaster-5
redis-cli -h 192.168.139.162 -p 26383 SENTINEL get-master-addr-by-name mymaster-4

echo "测试完成"
```

### 快速测试所有集群版本

```bash
#!/bin/bash
echo "启动所有集群版本..."
docker-compose -f docker-compose-cluster-multi-version.yml up -d
sleep 20

# 测试所有集群
for port in 7001 7101 7201 7301 7401; do
  echo "测试集群端口 $port..."
  redis-cli -h 192.168.139.162 -p $port -a abc123456 CLUSTER INFO | grep cluster_state
done

echo "测试完成"
```

## 停止服务

```bash
# 停止所有单机版本
docker-compose -f docker-compose-standalone-multi-version.yml down

# 停止所有主从版本
docker-compose -f docker-compose-replication-multi-version.yml down

# 停止所有哨兵版本
docker-compose -f docker-compose-sentinel-multi-version.yml down

# 停止所有集群版本
docker-compose -f docker-compose-cluster-multi-version.yml down

# 停止并清理所有数据
docker-compose -f docker-compose-standalone-multi-version.yml down -v
docker-compose -f docker-compose-replication-multi-version.yml down -v
docker-compose -f docker-compose-sentinel-multi-version.yml down -v
docker-compose -f docker-compose-cluster-multi-version.yml down -v
```

## 常见问题

### Q: 为什么需要指定 IP 地址 192.168.139.162？
A: 在 Mac 上使用 host 网络模式时，需要 bind 到具体的网卡 IP 地址，不能使用 0.0.0.0。

### Q: 如何查看我的 Mac IP 地址？
A:
```bash
ifconfig | grep "inet " | grep -v 127.0.0.1
```

### Q: 可以同时运行所有版本吗？
A: 可以！每个版口使用不同的端口，可以同时启动所有版本进行批量测试。

### Q: 集群初始化需要多久？
A: 通常 10-20 秒。可以通过 `docker logs redis-cluster-X-creator` 查看初始化进度。

### Q: 如何查看容器日志？
A:
```bash
# 单机模式
docker logs redis-standalone-7

# 主从模式
docker logs redis-replication-7-master

# 哨兵模式
docker logs redis-sentinel-7-1

# 集群模式
docker logs redis-cluster-7-node-1
```

### Q: 端口被占用怎么办？
A: 检查是否有其他 Redis 实例在运行：
```bash
lsof -i :6379
ps aux | grep redis
```

## 完整测试流程示例

```bash
# 1. 启动所有环境
docker-compose -f docker-compose-standalone-multi-version.yml up -d
docker-compose -f docker-compose-replication-multi-version.yml up -d
docker-compose -f docker-compose-sentinel-multi-version.yml up -d
docker-compose -f docker-compose-cluster-multi-version.yml up -d

# 2. 等待初始化
sleep 30

# 3. 验证所有环境
echo "=== 验证单机 ==="
redis-cli -h 192.168.139.162 -p 6379 -a abc123456 PING

echo "=== 验证主从 ==="
redis-cli -h 192.168.139.162 -p 6379 -a abc123456 INFO replication

echo "=== 验证哨兵 ==="
redis-cli -h 192.168.139.162 -p 26379 SENTINEL get-master-addr-by-name mymaster-7

echo "=== 验证集群 ==="
redis-cli -h 192.168.139.162 -p 7001 -a abc123456 CLUSTER INFO

# 4. 运行 DataKit 测试
# ./datakit --test

# 5. 清理环境
docker-compose -f docker-compose-standalone-multi-version.yml down -v
docker-compose -f docker-compose-replication-multi-version.yml down -v
docker-compose -f docker-compose-sentinel-multi-version.yml down -v
docker-compose -f docker-compose-cluster-multi-version.yml down -v
```

## 测试矩阵

| 版本 | 单机 | 主从 | 哨兵 | 集群 |
|------|------|------|------|------|
| 7.0.11 | ✓ (6379) | ✓ (6379-6579) | ✓ (6379+26379) | ✓ (7001-7006) |
| 6.2.12 | ✓ (6380) | ✓ (6380-6580) | ✓ (6380+26380) | ✓ (7101-7106) |
| 6.0.8  | ✓ (6381) | ✓ (6381-6581) | ✓ (6381+26381) | ✓ (7201-7206) |
| 5.0.14 | ✓ (6382) | ✓ (6382-6582) | ✓ (6382+26382) | ✓ (7301-7306) |
| 4.0.14 | ✓ (6383) | ✓ (6383-6583) | ✓ (6383+26383) | ✓ (7401-7406) |

**优势**：所有版本可以同时运行，大大提高测试效率！
