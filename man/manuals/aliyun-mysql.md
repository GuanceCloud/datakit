
# PolarDB Mysql
---

## 视图预览

阿里云 PolarDB Mysql 指标展示，包括 CPU 使用率，内存命中率，网络流量，连接数，QPS，TPS，只读节点延迟等

![image](imgs/input-aliyun-mysql-1.png)

![image](imgs/input-aliyun-mysql-2.png)

## 版本支持

操作系统支持：Linux

## 前置条件

- 服务器 <[安装 Datakit](../datakit/datakit-install.md)>
- 服务器 <[安装 Func 携带版](../dataflux-func/quick-start.md)>
- 阿里云 RAM 访问控制账号授权

### RAM 访问控制

1、登录 RAM 控制台  [https://ram.console.aliyun.com/users](https://ram.console.aliyun.com/users)
2、新建用户：人员管理 - 用户 - 创建用户

![image](imgs/input-aliyun-mysql-3.png)

3、保存或下载 **AccessKey** **ID** 和 **AccessKey Secret** 的 CSV 文件 (配置文件会用到)
4、用户授权 (云监控只读/时序指标数据权限)

![image](imgs/input-aliyun-mysql-4.png)

## 安装配置

说明：

- 示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)
- 通过一台服务器采集所有阿里云 PolarDB Mysql 数据

### 部署实施

#### 脚本市场

1、登录 Func，地址 http://ip:8088

![image](imgs/input-aliyun-mysql-5.png)

2、开启脚本市场，管理 - 实验性功能 - 开启脚本市场模块

![image](imgs/input-aliyun-mysql-6.png)

3、载入阿里云数据同步脚本，管理 - 脚本市场 - 阿里云数据同步 (云监控)

![image](imgs/input-aliyun-mysql-7.png)

#### 添加脚本

1、阿里云数据同步 (云监控) - 添加脚本

![image](imgs/input-aliyun-mysql-8.png)

2、输入标题/描述信息

![image](imgs/input-aliyun-mysql-9.png)

3、编辑脚本并复制代码，从 (同步阿里云监控数据) 到当前脚本
4、修改阿里云账号配置 (Ram 访问控制)

```bash
    'aliyun_ak_id'    : 'AccessKey ID',
    'aliyun_ak_secret': 'AccessKey Secret',
```

5、修改阿里云 PolarDB Mysql 指标

```bash
    'metric_targets': [
        {
            'namespace': 'acs_polardb',
            'metrics': 'ALL',
         }           
                      ]
```

6、**保存** 配置并 **发布**

![image](imgs/input-aliyun-mysql-10.png)

#### 定时任务

1、添加自动触发任务，管理 - 自动触发配置 - 新建任务

![image](imgs/input-aliyun-mysql-11.png)

2、自动触发配置，执行函数中添加此脚本，其他默认即可

![image](imgs/input-aliyun-mysql-12.png)

3、指标预览

![image](imgs/input-aliyun-mysql-13.png)

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - 阿里云 PolarDB Mysql>

## 监控规则

<监控 - 模板新建 - 阿里云 PolarDB Mysql 检测库>

## 指标详解

<[阿里云 PoarlDB Mysql 指标列表](https://help.aliyun.com/document_detail/165836.htm?spm=a2c4g.11186623.0.0.d5e33fe4KVrSrg#concept-2497787)>

## 故障排查

- 查看日志：Func 日志路径 /usr/local/dataflux-func/data/logs/dataflux-func.log
- 代码调试：选择主函数，直接运行 (可以看到脚本输出)

![image](imgs/input-aliyun-mysql-14.png)

- 连接配置：Func 无法连接 Datakit，请检查数据源配置

![image](imgs/input-aliyun-mysql-15.png)

## 进一步阅读

<[DataFlux Func 开发手册](https://func.guance.com/#/read?q=%2Fdataflux%2Ffunc%2Fdevelopment-guide.md)>

