# 华为云虚拟互联网接入

## 简介 {#introduction}

目前**华为云**已上架的服务为 **cn-south-guangzhou** 地域，其他地域的也即将上架，架构如下：

| 接入站点         | 您的服务器所在 Region | 接入终端节点服务的名称 ID                                 |
| ---------------- | --------------------- | --------------------------------------------------------- |
| 中国区 4（广州） | `cn-south`            | `cn-south-1.openway.af005322-387a-47cb-a21f-758b1c6c7ee7` |



## VPC 终端节点连接服务 {#hw-connect}

### 步骤一：帐号 ID 授权 {#Account-authorization}

登录管理控制台。

单击帐号下的「我的凭证」。

![img](https://static.guance.com/images/datakit/vpcep_01.png)

进入「我的凭证」页面，即可查看到所属租户的「帐号 ID」，如下所示。

![img](https://static.guance.com/images/datakit/vpcep_02.png)

复制该「账号 ID」并告知我方观测云的客户经理，加入其到我方白名单中。

### 步骤二：购买终端节点 {#Purchase-nodes}

登录管理控制台。

在管理控制台左上角选择**广州**区域和项目。

单击「服务列表」中的「网络 > VPC 终端节点」，进入「终端节点」页面。

在「终端节点」页面，单击「购买终端节点」。

![img](https://static.guance.com/images/datakit/vpcep_03.png)

根据界面提示配置参数。（**服务名称**需要用到上方表格中的**「接入终端节点服务的名称 ID」**）

参数配置完成，单击「立即购买」，进行规格确认。

- 规格确认无误，单击「提交」，任务提交成功。

- 参数信息配置有误，需要修改，单击「上一步」，修改参数，然后单击「提交」。

请求连接

- 到终端节点列表查看终端节点状态是「待接受」，**这时请告知我方客户经理**，验证连接请求。

- 告知接受连接后，再返回终端节点列表查看终端节点状态变为「已接受」，表示终端节点已成功连接至终端节点服务。

单击终端节点 ID ，即可查看终端节点的详细信息。

- 终端节点创建成功后，会生成一个「节点 IP」（就是私有 IP）和「内网域名」（如果在创建终端节点时您勾选了「创建内网域名」）。

![img](https://static.guance.com/images/datakit/vpcep_04.png)

- 您可以使用节点 IP 或内网域名访问终端节点服务，进行跨 VPC 资源通信。

## DNS 解析内网域名 {#dns-resolution}

### 添加内网域名解析 {#add-dns-resolution}

导航栏中点击「云解析服务 DNS」

选择内网域名

点击右上角「创建域名」，根据提示填入相应数据，域名填写主域名，建议使用 `guance.com` 作为主域名。

填入终端节点所在的 VPC。如下图所示：

![img](https://static.guance.com/images/datakit/vpcep_05.png)

创建成功后，点击管理解析

接着点击右上角，添加记录集。

- 主机记录填写服务名( `cn4-openway` )

- 值填写成终端节点的 IP 地址。

然后点击确定。如下图所示：

![img](https://static.guance.com/images/datakit/vpcep_06.png)

登陆 VPC 内网的随便一台机器，用以下命令验证一下

```shell
curl --insecure https://cn4-openway.guance.com
```
