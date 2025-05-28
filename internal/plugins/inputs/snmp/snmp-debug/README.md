# SNMP 本地测试搭建

本文档描述 ubunt 上如何模拟一个 snmp 被采集数据源，并且在 DK 上配置 snmp.conf 来采集数据。


```shell
# 安装工具
sudo apt install snmp snmpd libsnmp-dev  # Debian/Ubuntu 
```

在当前目录下，有多个 conf 文件，都是针对 snmpd 的，其启动方式如下：


```
# ubuntu 启动 snmpd agent
sudo /usr/sbin/snmpd -f -Lo -C  -p x.pid -Ddump,usm,acl,header,context,pdu,snmpv3 -c <XXX.conf>
```

其中：

- *agent-context.conf*: 针对 v3 协议采集的配置
- *agent1.conf*: 默认的最简单的 agent 配置
- 其它几个都是测试用的

这里几个 conf 中的端口都不是默认的 161，注意区分。

启动后，我们就可以用 snmpwalk（详见脚本 *walk.sh*） 命令来测试本机 snmpd 的连接情况。其中：

- `-n` 相当于 `v3_context_name`
- `-E` 相当于 `v3_context_engine_id`

# DK 采集器配置

采集器 *snmp.conf* 的一个 v3 配置如下：

```toml 
[[inputs.snmp]]
  specific_devices = ["127.0.0.1"] # SNMP Device IP.
  snmp_version = 3
  port = 1161

  v3_user = "snmpv3user1"
  v3_auth_protocol = "SHA256"
  v3_auth_key = "authPassAgent1"
  v3_priv_protocol = "AES"
  v3_priv_key = "privPassAgent1"

  # 一定不要开如下两个配置，目前测试下来，开了之后就没法采集（超时、认证等错误）
  #v3_context_engine_id = "***"
  #v3_context_name = "myContext"
```
