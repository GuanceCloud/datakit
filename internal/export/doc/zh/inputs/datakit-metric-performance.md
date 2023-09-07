# Datakit 指标性能测试报告
---

## 测试环境参数

- CPU (4 核)：Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
- 内存：12 GB
    - 4GiB DIMM DDR4 Synchronous 2133 MHz (0.5 ns)
    - 8GiB DIMM DDR4 Synchronous 2133 MHz (0.5 ns)
- 操作系统：Ubuntu 22.04 LTS
- DataKit：1.14.1-3-gf792e9d

## 指标 (metrics) 测试结果

|  /   | 开启默认采集器  | 开启默认采集器 + 开启 1 个 MySQL 采集器  | 开启默认采集器 + 开启 100 个 MySQL 采集器  |
|  ----  | ----  | ----  | ----  |
| 平均占用 CPU  | 0.43%    | 0.38%    | 11.99% |
| 平均占用内存   | 22.91 MB | 21.22 MB | 42.19 MB |
| 上传字节数   | 150 K | 300 K | 3 M |

## CPU 的变化情况

采集时间：10 min

<!-- markdownlint-disable MD024 -->

### 开启默认采集器

![mp-1-cpu](imgs/mp-1-cpu.png)

### 开启默认采集器 + 开启 1 个 MySQL 采集器

![mp-2-cpu](imgs/mp-2-cpu.png)

### 开启默认采集器 + 开启 100 个 MySQL 采集器

![mp-3-cpu](imgs/mp-3-cpu.png)

## 内存的变化情况

采集时间：10 min

### 开启默认采集器

![mp-1-mem](imgs/mp-1-mem.png)

### 开启默认采集器 + 开启 1 个 MySQL 采集器

![mp-2-mem](imgs/mp-2-mem.png)

### 开启默认采集器 + 开启 100 个 MySQL 采集器

![mp-3-mem](imgs/mp-3-mem.png)

## 上传字节数的变化情况

采集时间：10 min

### 开启默认采集器

![mp-1-upload](imgs/mp-1-upload.png)

### 开启默认采集器 + 开启 1 个 MySQL 采集器

![mp-2-upload](imgs/mp-2-upload.png)

### 开启默认采集器 + 开启 100 个 MySQL 采集器

![mp-3-upload](imgs/mp-3-upload.png)

<!-- markdownlint-enable -->

## 其它测试结果

- [Datakit Trace Agent 性能报告](./datakit-trace-performance.md){:target="_blank"}
- [DataKit 日志采集器性能测试](./logging-pipeline-bench.md){:target="_blank"}
