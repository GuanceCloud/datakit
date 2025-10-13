# AWS 虚拟互联网接入

---

## 概述 {#overview}

Amazon PrivateLink 是一项高度可用的可扩展技术，使您能够将 VPC 私密地连接到服务，如同这些服务就在您自己的 VPC 中一样。您无需使用互联网网关、NAT 设备、公有 IP 地址、Amazon Direct Connect 连接或 Amazon Site-to-Site VPN 连接来允许与私有子网中的服务进行通信。因此，您可以控制可从 VPC 访问的特定 API 端点、站点和服务。Amazon PrivateLink 可以帮你节省部分流量费用。

建立私网连接优点：

- **更高带宽**：不占用业务系统的公网带宽，通过终端节点服务实现更高带宽
- **更安全**：数据不经过公网，完全在私网内流转，数据更安全
- **更低资费**：相比公网带宽的高费用，虚拟互联网的资费成本更低

架构如下：

```mermaid
flowchart LR
  subgraph Customer_VPC
    dk_a[Availability Zone A - dk]
    dk_b[Availability Zone B - dk]
    dk_c[Availability Zone C - dk]
    plc[Endpoints]

    dk_a --> plc
    dk_b --> plc
    dk_c --> plc
  end

  subgraph <<<custom_key.brand_main_domain>>>_VPC
    pls[Endpoints Service]
    nlb[NLB]
    dw[DW - Availability Zone C]
    pls --> nlb --> dw
  end
  plc --> pls
```

## 前提条件 {#prerequisite}

1. 首先选择订阅地域，必须与您待接入<<<custom_key.brand_name>>>的系统所部署云资源的同一地域。
1. 选择待接入系统所部署云资源的同一个 VPC 网络，**如果涉及到多个 VPC 需要接入终端节点服务，可多次订阅，每个 VPC 订阅一次。**

## 订阅服务 {#sub-service}

### 服务部署链接 {#service-dep}

<<<% if custom_key.brand_key == "truewatch" %>>>
| **接入站点**      | **您的服务器所在 Region** | **接入终端节点服务的名称**                         |
| --------          | ----------------------    | -----------                          |
| 亚太区 1（新加坡）  | `ap-southeast-1` (新加坡)     |  `com.amazonaws.vpce.ap-southeast-1.vpce-svc-08465b643241dce58` |
<<<% else %>>>
| **接入站点**      | **您的服务器所在 Region** | **接入终端节点服务的名称**                         |
| --------          | ----------------------    | -----------                          |
| 中国区 2（宁夏）  | `cn-northwest-1` (宁夏)   | `cn.com.amazonaws.vpce.cn-northwest-1.vpce-svc-070f9283a2c0d1f0c` |
| 海外区 1（俄勒冈）  | `us-west-2` (俄勒冈)     |  `com.amazonaws.vpce.us-west-2.vpce-svc-084745e0ec33f0b44` |
| 亚太区 1（新加坡）  | `ap-southeast-1` (新加坡)     |  `com.amazonaws.vpce.ap-southeast-1.vpce-svc-070194ed9d834d571` |
<<<% endif %>>>


### 不同 Region 的私网数据网关默认 Endpoint {#region-endpoint}

<<<% if custom_key.brand_key == "truewatch" %>>>
| **接入站点**      | **您的服务器所在 Region** | **Endpoint**                         |
| --------          | ----------------------    | -----------                          |
| 亚太区 1（新加坡）  |  `ap-southeast-1` (新加坡)         | `https://ap1-openway.<<<custom_key.brand_main_domain>>>` |
<<<% else %>>>
| **接入站点**      | **您的服务器所在 Region** | **Endpoint**                         |
| --------          | ----------------------    | -----------                          |
| 中国区 2（宁夏）  | `cn-northwest-1` (宁夏)   | `https://aws-openway.<<<custom_key.brand_main_domain>>>`         |
| 海外区 1（俄勒冈）  |  `us-west-2` (俄勒冈)          | `https://us1-openway.<<<custom_key.brand_main_domain>>>` |
| 亚太区 1（新加坡）  |  `ap-southeast-1` (新加坡)         | `https://ap1-openway.<<<custom_key.brand_main_domain>>>` |
<<<% endif %>>>

### 配置服务订阅 {#config-sub}

#### 步骤一：帐号 ID 授权 {#accredit-id}
<!-- markdownlint-disable MD032 -->
通过以下链接打开 Amazon  控制台：
<<<% if custom_key.brand_key == "truewatch" %>>>
- [Console Web](https://console.aws.amazon.com/console/home){:target="_blank"}
<<<% else %>>>
- [中国区](https://console.amazonaws.cn/console/home){:target="_blank"}
- [海外区](https://console.aws.amazon.com/console/home){:target="_blank"}
<<<% endif %>>>
获取控制台右上角账号 ID, 复制该「账号 ID」并**告知**我方<<<custom_key.brand_name>>>的客户经理，加入到我方白名单中。


#### 步骤二：创建终端节点 {#create-endpoint}

1. 确认业务 VPC 设置：
    - DNS hostname: Enabled（启用 DNS 主机名）
    - DNS resolution: Enabled（启用 DNS 支持）
1. 通过以下链接打开 Amazon VPC 控制台：
<<<% if custom_key.brand_key == "truewatch" %>>>
    - [VPC](https://console.amazonaws.cn/vpc/){:target="_blank"}
<<<% else %>>>
    - [中国区](https://console.amazonaws.cn/vpc/){:target="_blank"}
    - [海外区](https://console.amazonaws.cn/vpc/){:target="_blank"}
<<<% endif %>>>
<!-- markdownlint-disable MD051 -->
1. 新建安全组：
    - 安全组名称：private-link
    - 入站规则：HTTPS
    - 目标：0.0.0.0/0
1. 在导航窗格中，选择 **Endpoint**（端点服务）。
1. 创建终点节点
    - 终端节点设置
        - 类型：**Endpoint services that use NLBs and GWLBs** （使用 NLB 和 GWLB 的端点服务）
    - 服务设置

        - 服务名称：当前地域的[接入终端节点服务的名称](aws-access.md#service-dep){:target="_blank"}
        - 验证服务
    - 网络设置
        - VPC：业务服务的 VPC
        - 子网：选择业务子网
        - 安全组：private-link
<!-- markdownlint-enable -->
1. 通知 <<<custom_key.brand_name>>> 的客户经理审核
1. 等待创建成功，点击终端节点的「操作」- 「修改私有 DNS 名称」，设置「为此终端节点启用」

<!-- markdownlint-enable -->

#### 验证 {#verify}

EC2 执行命令：

```shell
dig xxx-openway.<<<custom_key.brand_main_domain>>>
```

结果：

```shell
; <<>> DiG 9.16.38-RH <<>> xxx-openway.<<<custom_key.brand_main_domain>>>
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 22545
;; flags: qr rd ra; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;xxx-openway.<<<custom_key.brand_main_domain>>>.      IN    A

;; ANSWER SECTION:
xxx-openway.<<<custom_key.brand_main_domain>>>. 296 IN  CNAME    172.31.16.128 

;; Query time: 0 msec
;; SERVER: 172.31.0.2#53(172.31.0.2)
;; WHEN: Thu May 18 11:23:04 UTC 2023
;; MSG SIZE  rcvd: 176
```

### 资费情况 {#cost}

以俄勒冈为例：

| 名称                                                         | 费用     | 文档                                                         | 备注                   |
| ------------------------------------------------------------ | -------- | ------------------------------------------------------------ | ---------------------- |
| 数据自 Amazon EC2  传出至互联网                              | $0.09/GB | [文档](https://aws.amazon.com/ec2/pricing/on-demand/#Data_Transfer){:target="_blank"} | 按流量收费             |
| 接口终端节点                                                 | $0.01/H  | [文档](https://aws.amazon.com/privatelink/pricing/?nc1=h_ls){:target="_blank"} | 按可用区数量和小时收费 |
| 终端接口节点流量传出                                         | $0.01/GB | [文档](https://aws.amazon.com/privatelink/pricing/?nc1=h_ls){:target="_blank"} | 按流量收费    |

资费情况主要是以下部分：

1. 第一部分是接口终端节点的[服务费用](https://aws.amazon.com/privatelink/pricing/?nc1=h_ls){:target="_blank"}
1. 第二部分是终端节点流量费用

对比如下：

假设客户方**每日**传出流量 **200GB**，传入流量 **10GB** 为例。

|          |           互联网           | PrivateLink                                                  |
| :------: | :------------------------: | ------------------------------------------------------------ |
|   公式   | 互联网传出流量 x 互联网传出流量费用 x 30 | 接口终端节点服务 x 3 可用区 x 24 小时 x 30 天 +（ 终端接口节点传出费用 x 终端接口节点传出流量 + 终端接口节点传入费用 x 终端接口节点传入流量 ）x 30 |
|   计算   |       0.09 x 200 x 30       | 0.01 x 3 x 24 x 30 +(0.01 x 200  + 10 x 0.01) x 30 |
| 每月总额 |           $540.0           | $84.6 |
