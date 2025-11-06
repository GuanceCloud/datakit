# LLDP 配置示例

> 本示例基于华为 S5735S-S48T4S-A。

虽然设别上可能启用了 LLDP，但是远程请求 LLDP 数据的时候，还是可能发现不了任何设备。

** 切换到 system view **

```shell
<HUAWEI> system-view # <> 表示处于 user view
[HUAWEI]             # [] 表示处于 system view

# 执行如下命令
[HUAWEI] display current-configuration | include snmp-agent community

# 会得到类似如下结果：
# snmp-agent community read <my-community-string>

# create a permissive MIB view
# This command creates a new view named "iso-view" that includes everything from the root "iso" OID downwards.
[HUAWEI] snmp-agent mib-view included iso-view iso

# Replace <my-community-string> with your actual community string
[HUAWEI] snmp-agent community read <MY-COMMUNITY-STRING> mib-view iso-view

# 退出并保存设置

[HUAWEI] quit
<HUAWEI> save
Warning: The current configuration will be written to the device. Continue? [Y/N]:y

# 验证是否能拿到 neighbor 信息
snmpwalk -v2c -c <MY-COMMUNITY-STRING> <IP> 1.0.8802.1.1.2.1.4.1
```

华为设备 lldp 扫描到的数据示例

```txt
----------------------------------
Date: Fri Sep 19 17:28:47 CST 2025
Host: tanbiaos-MacBook-Pro.local
----------------------------------
=== RUN   TestLLDP
=== RUN   TestLLDP/basic
    lldp_test.go:103: discovered 64 local interfaces on 10.200.10.253
    lldp_test.go:105: 9: Vlanif999
    lldp_test.go:105: 17: GigabitEthernet0/0/5
    lldp_test.go:105: 29: GigabitEthernet0/0/17
    lldp_test.go:105: 41: GigabitEthernet0/0/29
    lldp_test.go:105: 49: GigabitEthernet0/0/37
    lldp_test.go:105: 1: InLoopBack0
    lldp_test.go:105: 14: GigabitEthernet0/0/2
    lldp_test.go:105: 23: GigabitEthernet0/0/11
    lldp_test.go:105: 33: GigabitEthernet0/0/21
    lldp_test.go:105: 46: GigabitEthernet0/0/34
    lldp_test.go:105: 52: GigabitEthernet0/0/40
    lldp_test.go:105: 8: Vlanif306
    lldp_test.go:105: 31: GigabitEthernet0/0/19
    lldp_test.go:105: 35: GigabitEthernet0/0/23
    lldp_test.go:105: 58: GigabitEthernet0/0/46
    lldp_test.go:105: 18: GigabitEthernet0/0/6
    lldp_test.go:105: 19: GigabitEthernet0/0/7
    lldp_test.go:105: 40: GigabitEthernet0/0/28
    lldp_test.go:105: 61: GigabitEthernet0/0/49
    lldp_test.go:105: 63: GigabitEthernet0/0/51
    lldp_test.go:105: 38: GigabitEthernet0/0/26
    lldp_test.go:105: 45: GigabitEthernet0/0/33
    lldp_test.go:105: 54: GigabitEthernet0/0/42
    lldp_test.go:105: 56: GigabitEthernet0/0/44
    lldp_test.go:105: 16: GigabitEthernet0/0/4
    lldp_test.go:105: 39: GigabitEthernet0/0/27
    lldp_test.go:105: 50: GigabitEthernet0/0/38
    lldp_test.go:105: 59: GigabitEthernet0/0/47
    lldp_test.go:105: 2: NULL0
    lldp_test.go:105: 21: GigabitEthernet0/0/9
    lldp_test.go:105: 30: GigabitEthernet0/0/18
    lldp_test.go:105: 37: GigabitEthernet0/0/25
    lldp_test.go:105: 47: GigabitEthernet0/0/35
    lldp_test.go:105: 65: Eth-Trunk100
    lldp_test.go:105: 27: GigabitEthernet0/0/15
    lldp_test.go:105: 55: GigabitEthernet0/0/43
    lldp_test.go:105: 7: Vlanif305
    lldp_test.go:105: 11: Eth-Trunk2
    lldp_test.go:105: 44: GigabitEthernet0/0/32
    lldp_test.go:105: 62: GigabitEthernet0/0/50
    lldp_test.go:105: 4: MEth0/0/1
    lldp_test.go:105: 15: GigabitEthernet0/0/3
    lldp_test.go:105: 22: GigabitEthernet0/0/10
    lldp_test.go:105: 51: GigabitEthernet0/0/39
    lldp_test.go:105: 6: Vlanif60
    lldp_test.go:105: 25: GigabitEthernet0/0/13
    lldp_test.go:105: 28: GigabitEthernet0/0/16
    lldp_test.go:105: 32: GigabitEthernet0/0/20
    lldp_test.go:105: 60: GigabitEthernet0/0/48
    lldp_test.go:105: 64: GigabitEthernet0/0/52
    lldp_test.go:105: 34: GigabitEthernet0/0/22
    lldp_test.go:105: 43: GigabitEthernet0/0/31
    lldp_test.go:105: 57: GigabitEthernet0/0/45
    lldp_test.go:105: 10: Eth-Trunk1
    lldp_test.go:105: 13: GigabitEthernet0/0/1
    lldp_test.go:105: 26: GigabitEthernet0/0/14
    lldp_test.go:105: 42: GigabitEthernet0/0/30
    lldp_test.go:105: 53: GigabitEthernet0/0/41
    lldp_test.go:105: 36: GigabitEthernet0/0/24
    lldp_test.go:105: 48: GigabitEthernet0/0/36
    lldp_test.go:105: 3: Console9/0/0
    lldp_test.go:105: 12: Eth-Trunk5
    lldp_test.go:105: 24: GigabitEthernet0/0/12
    lldp_test.go:105: 20: GigabitEthernet0/0/8
    lldp_test.go:111: get 1th neighbor
    lldp_test.go:111: get 2th neighbor
    lldp_test.go:111: get 3th neighbor
    lldp_test.go:111: get 4th neighbor
    lldp_test.go:111: get 5th neighbor
    lldp_test.go:111: get 6th neighbor
    lldp_test.go:111: get 7th neighbor
    lldp_test.go:111: get 8th neighbor
    lldp_test.go:111: get 9th neighbor
    lldp_test.go:111: get 10th neighbor
    lldp_test.go:111: get 11th neighbor
    lldp_test.go:111: get 12th neighbor
    lldp_test.go:111: get 13th neighbor
    lldp_test.go:111: get 14th neighbor
    lldp_test.go:111: get 15th neighbor
    lldp_test.go:111: get 16th neighbor
    lldp_test.go:111: get 17th neighbor
    lldp_test.go:111: get 18th neighbor
    lldp_test.go:111: get 19th neighbor
    lldp_test.go:111: get 20th neighbor
    lldp_test.go:111: get 21th neighbor
    lldp_test.go:111: get 22th neighbor
    lldp_test.go:111: get 23th neighbor
    lldp_test.go:111: get 24th neighbor
    lldp_test.go:111: get 25th neighbor
    lldp_test.go:111: get 26th neighbor
    lldp_test.go:111: get 27th neighbor
    lldp_test.go:111: get 28th neighbor
    lldp_test.go:111: get 29th neighbor
    lldp_test.go:111: get 30th neighbor
    lldp_test.go:111: get 31th neighbor
    lldp_test.go:111: get 32th neighbor
    lldp_test.go:111: get 33th neighbor
    lldp_test.go:111: get 34th neighbor
    lldp_test.go:111: get 35th neighbor
    lldp_test.go:111: get 36th neighbor
    lldp_test.go:111: get 37th neighbor
    lldp_test.go:111: get 38th neighbor
    lldp_test.go:111: get 39th neighbor
    lldp_test.go:111: get 40th neighbor
    lldp_test.go:111: get 41th neighbor
    lldp_test.go:111: get 42th neighbor
    lldp_test.go:111: get 43th neighbor
    lldp_test.go:111: get 44th neighbor
    lldp_test.go:111: get 45th neighbor
    lldp_test.go:111: get 46th neighbor
    lldp_test.go:111: get 47th neighbor
    lldp_test.go:111: get 48th neighbor
    lldp_test.go:111: get 49th neighbor
    lldp_test.go:111: get 50th neighbor
    lldp_test.go:111: get 51th neighbor
    lldp_test.go:111: get 52th neighbor
    lldp_test.go:111: get 53th neighbor
    lldp_test.go:111: get 54th neighbor
    lldp_test.go:117: got 54 neighbor pdus
2025/09/19 17:28:49 col: 4, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.4.0.32.3|val: 7
2025/09/19 17:28:49 col: 4, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.4.0.32.4|val: 7
2025/09/19 17:28:49 col: 4, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.4.0.33.1|val: 4
2025/09/19 17:28:49 col: 4, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.4.0.34.1|val: 4
2025/09/19 17:28:49 col: 4, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.4.0.45.1|val: 4
2025/09/19 17:28:49 col: 4, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.4.0.46.1|val: 4
2025/09/19 17:28:49 col: 5, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.5.0.32.3|val: "DESKTOP-ONGRI03"
2025/09/19 17:28:49 col: 5, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.5.0.32.4|val: "ZHUYUNPC"
2025/09/19 17:28:49 col: 5, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.5.0.33.1|val: "h\x86\xa7td3"
2025/09/19 17:28:49 col: 5, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.5.0.34.1|val: "h\x86\xa7tf\x97"
2025/09/19 17:28:49 col: 5, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.5.0.45.1|val: "(\xfb\xae\x9f\xea\x97"
2025/09/19 17:28:49 col: 5, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.5.0.46.1|val: "(\xfb\xae\x9f\xea\x97"
2025/09/19 17:28:49 col: 6, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.6.0.32.3|val: 3
2025/09/19 17:28:49 col: 6, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.6.0.32.4|val: 3
2025/09/19 17:28:49 col: 6, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.6.0.33.1|val: 5
2025/09/19 17:28:49 col: 6, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.6.0.34.1|val: 5
2025/09/19 17:28:49 col: 6, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.6.0.45.1|val: 5
2025/09/19 17:28:49 col: 6, pdu: type:2|name:.1.0.8802.1.1.2.1.4.1.1.6.0.46.1|val: 5
2025/09/19 17:28:49 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.32.3|val: "\x04|\x16\xf7F\xfb"
2025/09/19 17:28:49 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.32.4|val: "\x88\x88\x88\x88\x87\x88"
2025/09/19 17:28:49 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.33.1|val: "gi24"
2025/09/19 17:28:49 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.34.1|val: "gi24"
2025/09/19 17:28:49 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.45.1|val: "GigabitEthernet0/0/7"
2025/09/19 17:28:49 col: 7, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.7.0.46.1|val: "GigabitEthernet0/0/6"
2025/09/19 17:28:49 col: 8, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.8.0.32.3|val: ""
2025/09/19 17:28:49 col: 8, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.8.0.32.4|val: ""
2025/09/19 17:28:49 col: 8, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.8.0.33.1|val: ""
2025/09/19 17:28:49 col: 8, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.8.0.34.1|val: ""
2025/09/19 17:28:49 col: 8, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.8.0.45.1|val: ""
2025/09/19 17:28:49 col: 8, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.8.0.46.1|val: ""
2025/09/19 17:28:49 col: 9, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.9.0.32.3|val: ""
2025/09/19 17:28:49 col: 9, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.9.0.32.4|val: ""
2025/09/19 17:28:49 col: 9, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.9.0.33.1|val: ""
2025/09/19 17:28:49 col: 9, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.9.0.34.1|val: ""
2025/09/19 17:28:49 col: 9, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.9.0.45.1|val: "firewall-usg6325E"
2025/09/19 17:28:49 col: 9, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.9.0.46.1|val: "firewall-usg6325E"
2025/09/19 17:28:49 col: 10, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.10.0.32.3|val: ""
2025/09/19 17:28:49 col: 10, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.10.0.32.4|val: ""
2025/09/19 17:28:49 col: 10, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.10.0.33.1|val: ""
2025/09/19 17:28:49 col: 10, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.10.0.34.1|val: ""
2025/09/19 17:28:49 col: 10, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.10.0.45.1|val: "firewall-usg6325E, Huawei Versatile Routing Platform Software, Software Version : USG6300E V600R007C20SPC603 (VRP (R) Software, Version 5.170), Copyright (C) 2014-2023 Huawei Technologies Co., Ltd."
2025/09/19 17:28:49 col: 10, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.10.0.46.1|val: "firewall-usg6325E, Huawei Versatile Routing Platform Software, Software Version : USG6300E V600R007C20SPC603 (VRP (R) Software, Version 5.170), Copyright (C) 2014-2023 Huawei Technologies Co., Ltd."
2025/09/19 17:28:49 col: 11, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.11.0.32.3|val: "\x00"
2025/09/19 17:28:49 col: 11, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.11.0.32.4|val: "\x00"
2025/09/19 17:28:49 col: 11, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.11.0.33.1|val: " "
2025/09/19 17:28:49 col: 11, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.11.0.34.1|val: " "
2025/09/19 17:28:49 col: 11, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.11.0.45.1|val: "\b"
2025/09/19 17:28:49 col: 11, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.11.0.46.1|val: "\b"
2025/09/19 17:28:49 col: 12, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.12.0.32.3|val: "\x00"
2025/09/19 17:28:49 col: 12, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.12.0.32.4|val: "\x00"
2025/09/19 17:28:49 col: 12, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.12.0.33.1|val: " "
2025/09/19 17:28:49 col: 12, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.12.0.34.1|val: " "
2025/09/19 17:28:49 col: 12, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.12.0.45.1|val: "\b"
2025/09/19 17:28:49 col: 12, pdu: type:4|name:.1.0.8802.1.1.2.1.4.1.1.12.0.46.1|val: "\b"
    lldp_test.go:123: LLDP neighbors: [
          {
            "local_port": "GigabitEthernet0/0/34",
            "chassis_id": "(\ufffd\ufffd\ufffd\ufffd\ufffd",
            "port_id": "GigabitEthernet0/0/6",
            "system_name": "firewall-usg6325E",
            "system_desc": "firewall-usg6325E, Huawei Versatile Routing Platform Software, Software Version : USG6300E V600R007C20SPC603 (VRP (R) Software, Version 5.170), Copyright (C) 2014-2023 Huawei Technologies Co., Ltd."
          },
          {
            "local_port": "GigabitEthernet0/0/20",
            "chassis_id": "DESKTOP-ONGRI03",
            "port_id": "\u0004|\u0016\ufffdF\ufffd",
            "system_name": "",
            "system_desc": ""
          },
          {
            "local_port": "GigabitEthernet0/0/20",
            "chassis_id": "ZHUYUNPC",
            "port_id": "\ufffd\ufffd\ufffd\ufffd\ufffd\ufffd",
            "system_name": "",
            "system_desc": ""
          },
          {
            "local_port": "GigabitEthernet0/0/21",
            "chassis_id": "h\ufffd\ufffdtd3",
            "port_id": "gi24",
            "system_name": "",
            "system_desc": ""
          },
          {
            "local_port": "GigabitEthernet0/0/22",
            "chassis_id": "h\ufffd\ufffdtf\ufffd",
            "port_id": "gi24",
            "system_name": "",
            "system_desc": ""
          },
          {
            "local_port": "GigabitEthernet0/0/33",
            "chassis_id": "(\ufffd\ufffd\ufffd\ufffd\ufffd",
            "port_id": "GigabitEthernet0/0/7",
            "system_name": "firewall-usg6325E",
            "system_desc": "firewall-usg6325E, Huawei Versatile Routing Platform Software, Software Version : USG6300E V600R007C20SPC603 (VRP (R) Software, Version 5.170), Copyright (C) 2014-2023 Huawei Technologies Co., Ltd."
          }
        ]
--- PASS: TestLLDP (1.60s)
    --- PASS: TestLLDP/basic (1.60s)
PASS
coverage: 83.3% of statements
ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/snmp/lldp	1.624s
```
