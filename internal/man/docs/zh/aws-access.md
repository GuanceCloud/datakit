# AWS 虚拟互联网接入

---

## 概述 {#overview}

Amazon PrivateLink 是一项高度可用的可扩展技术，使您能够将 VPC 私密地连接到服务，如同这些服务就在您自己的 VPC 中一样。您无需使用互联网网关、NAT 设备、公有 IP 地址、Amazon Direct Connect 连接或 Amazon Site-to-Site VPN 连接来允许与私有子网中的服务进行通信。因此，您可以控制可从 VPC 访问的特定 API 端点、站点和服务。Amazon PrivateLink 可以帮你节省部分流量费用。

建立私网连接优点：

- **更高带宽**：不占用业务系统的公网带宽，通过终端节点服务实现更高带宽
- **更安全**：数据不经过公网，完全在私网内流转，数据更安全
- **更低资费**：相比公网带宽的高费用，虚拟互联网的资费成本更低

目前已上架的服务为 **cn-northwest-1、us-west-2** 两个地域，其他地域的也即将上架，架构如下：

![not-set](https://static.guance.com/images/datakit/aws_privatelink.png)

## 前提条件 {#prerequisite}

1. 首先选择订阅地域，必须与您待接入观测云的系统所部署云资源的同一地域。
1. 选择待接入系统所部署云资源的同一个 VPC 网络，**如果涉及到多个 VPC 需要接入终端节点服务，可多次订阅，每个 VPC 订阅一次。**

## 订阅服务 {#sub-service}

### 服务部署链接 {#service-dep}

| **接入站点**      | **您的服务器所在 Region** | **接入终端节点服务的名称**                         |
| --------          | ----------------------    | -----------                          |
| 中国区 2（宁夏）  | `cn-northwest-1` (宁夏)   | `cn.com.amazonaws.vpce.cn-northwest-1.vpce-svc-070f9283a2c0d1f0c` |
| 海外区 1（俄勒冈）  | `us-west-2` (俄勒冈)     |  `com.amazonaws.vpce.us-west-2.vpce-svc-084745e0ec33f0b44` |

### 不同 Region 的私网数据网关默认 Endpoint {#region-endpoint}

| **接入站点**      | **您的服务器所在 Region** | **Endpoint**                         |
| --------          | ----------------------    | -----------                          |
| 中国区 2（宁夏）  | `cn-northwest-1` (宁夏)   | `https://aws-openway.guance.com`         |
| 海外区 1（俄勒冈）  |  `us-west-2` (俄勒冈)          | `https://us1-openway.guance.com` |

### 配置服务订阅 {#config-sub}

#### 步骤一：帐号 ID 授权 {#accredit-id}

通过以下链接打开 Amazon  控制台：
    - [中国区](https://console.amazonaws.cn/console/home){:target="_blank"}
    - [海外区](https://console.aws.amazon.com/console/home){:target="_blank"}

获取右上角账号 ID, 复制该「账号 ID」并**告知**我方观测云的客户经理，加入其到我方白名单中。

![not-set](https://static.guance.com/images/datakit/aws_privatelink_id.png)

#### 步骤二：创建终端节点 {#create-endpoint}

1. 通过以下链接打开 Amazon VPC 控制台：
    - [中国区](https://console.amazonaws.cn/vpc/){:target="_blank"}
    - [海外区](https://console.amazonaws.cn/vpc/){:target="_blank"}
1. 在导航窗格中，选择 **Endpoint**（端点服务）。
1. 创建终点节点
1. **服务设置**输入服务名称，验证。选择 vpc，可用区，安全组开通 443
1. 等待创建成功，获取终端节点服务地址

![not-set](https://static.guance.com/images/datakit/aws-privatelink-dns.png)

#### 步骤三：Route 53 解析终端节点 {#route-53}

1. 通过以下链接打开 Amazon Route 53 控制台：
    - [中国区](https://console.amazonaws.cn/route53/v2/hostedzones/){:target="_blank"}
    - [海外区](https://console.aws.amazon.com/route53/v2/hostedzones/){:target="_blank"}
1. 创建托管区
1. 域名：`guance.com`，类型：私有托管区，区域：DK 所在的区域，VPC ID：客户方 VPC ID
1. 创建记录。
1. 记录名称：参考 [Endpoint](aws-access.md#region-endpoint) 地址 ，记录类型：`cname`，值： [创建终端节点](aws-access.md#create-endpoint)的服务地址

![not-set](https://static.guance.com/images/datakit/aws_privatelink_route53.png)

#### 验证 {#verify}

EC2 执行命令：

```shell script
dig aws-openway.guance.com
```

结果：

```shell script
; <<>> DiG 9.16.38-RH <<>> aws-openway.guance.com
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 22545
;; flags: qr rd ra; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;aws-openway.guance.com.      IN    A

;; ANSWER SECTION:
aws-openway.guance.com. 296 IN  CNAME    vpce-0d431e354cf9ad4f1-h1y2auf6.vpce-svc-070f9283a2c0d1f0c.cn-northwest-1.vpce.amazonaws.com.cn.
vpce-0d431e354cf9ad4f1-h1y2auf6.vpce-svc-070f9283a2c0d1f0c.cn-northwest-1.vpce.amazonaws.com.cn. 56 IN A 172.31.38.12

;; Query time: 0 msec
;; SERVER: 172.31.0.2#53(172.31.0.2)
;; WHEN: Thu May 18 11:23:04 UTC 2023
;; MSG SIZE  rcvd: 176
```

### 资费情况 {#cost}

以俄勒冈为例：

| 名称                                                         | 费用     | 文档                                                         | 备注                   |
| ------------------------------------------------------------ | -------- | ------------------------------------------------------------ | ---------------------- |
| 数据自 Amazon EC2  传出至互联网                              | $0.09/GB | [文档](https://aws.amazon.com/cn/ec2/pricing/on-demand/#Data_Transfer){:target="_blank"} | 按流量收费             |
| 接口终端节点                                                 | $0.01/H  | [文档](https://aws.amazon.com/cn/privatelink/pricing/?nc1=h_ls){:target="_blank"} | 按可用区数量和小时收费 |
| 终端接口节点流量传出                                         | $0.01/GB | [文档](https://aws.amazon.com/cn/privatelink/pricing/?nc1=h_ls){:target="_blank"} | 按流量收费    |

资费情况主要是以下部分：

1. 第一部分是接口终端节点的[服务费用](https://aws.amazon.com/cn/privatelink/pricing/?nc1=h_ls){:target="_blank"}
1. 第二部分是终端节点流量费用

对比如下：

假设客户方**每日**传出流量 **200GB**，传入流量 **10GB** 为例。

|          |           互联网           | PrivateLink                                                  |
| :------: | :------------------------: | ------------------------------------------------------------ |
|   公式   | 互联网传出流量 x 互联网传出流量费用 x 30 | 接口终端节点服务 x 3 可用区 x 24 小时 x 30 天 +（ 终端接口节点传出费用 x 终端接口节点传出流量 + 终端接口节点传入费用 x 终端接口节点传入流量 ）x 30 |
|   计算   |       0.09 x 200 x 30       | 0.01 x 3 x 24 x 30 +(0.01 x 200  + 10 x 0.01) x 30 |
| 每月总额 |           $540.0           | $84.6 |
