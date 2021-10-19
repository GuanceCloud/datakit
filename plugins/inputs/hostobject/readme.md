### 主机Object

| Key | 类型 | 描述 | 必选 |
| --- | ---- | ---- | ---- |
| host | string | 主机名 | Y |
| message | [message](#message) | 主机详细信息 | Y |

### message

| Key | 类型 | 描述 |
| --- | ---- | ---- |
| host | [host](#host) | 主机信息 |
| collectors | json(字符串数组) | 开启的采集器名称列表 |

---

### host  
| Key | 类型 | 描述 |
| --- | ---- | ---- |
| meta | [meta](#meta) | 主机基础信息 |
| cpu | [cpu](#cpu)数组 | cpu信息 |
| mem | [mem](#mem) | 内存 |
| net | [net](#net)数组 | 网络接口 |
| disk | [disk](#disk)数组 | 磁盘 |

---

### meta
| Key | 类型 | 描述 |
| --- | ---- | ---- |
| host_name | string | 主机名 |
| boot_time | int | 主机启动时间(unix时间戳) |
| os | string | 操作系统类别，例: linux, freebsd, windows... |
| platform | string | 例: centos |
| platform_family | string | 例: rhel |
| platform_version | string | 例: 7.4.1708 |
| kernel_release | string | 系统内核代号 |
| arch | string | 系统架构，例：x86_64 |

---

### cpu
| Key | 类型 | 描述 |
| --- | ---- | ---- |
| vendor_id | string | 供应商id |
| module_name | string | 模块名 |
| cores | int | 核数 |
| mhz | float | 频率 |
| cache_size | int | 缓存大小 |

---

### mem
| Key | 类型 | 描述 |
| --- | ---- | ---- |
| memory_total | int | 内存总字节数 |
| swap_total | int | 交换内存总字节数 |

---

### net
| Key | 类型 | 描述 |
| --- | ---- | ---- |
| name | string | 接口名称 |
| mtu | int | MTU |
| ip4 | string | ip4地址 |
| ip6 | string | ip6地址 |
| ip4_all | []string | 所有ip4地址 |
| ip6_all | []string | 所有ip6地址 |
| mac | string | MAC地址 |
| flags | string | 接口属性 |

---

### disk
| Key | 类型 | 描述 |
| --- | ---- | ---- |
| device | string | 分区名 |
| total | int | 分区总字节数 |
| mountpoint | string | 挂载点 |
| fstype | string | 文件系统 |

---

**message示例：**  

    {
        "collectors": ["hostobject", "rum"]
        "host": {
            "meta": {
                "host_name": "izbp1dsyh39swucxotofd",
                "boot_time": 1548689369,
                "os": "linux",
                "platform": "centos",
                "platform_family": "rhel",
                "platform_version": "7.4.1708",
                "kernel_release": "3.10.0-693.2.2.el7.x86_64",
                "arch": "x86_64"
            },

            "cpu": [{
                "vendor_id": "GenuineIntel",
                "module_name": "Intel(R) Xeon(R) Platinum 8163 CPU @ 2.50GHz",
                "cores": 1,
                "mhz": 2499.998,
                "cache_size": 33792
            }],

            "mem": {
                "memory_total": 1040551936,
                "swap_total": 0
            },

            "net": [{
                "mtu": 1500,
                "name": "eth0",
                "mac": "00:16:3e:0e:4f:63",
                "flags": ["up", "broadcast", "multicast"],
                "ip4": "172.16.79.249/20",
                "ip6": "",
                "addrs": null
            }, {
                "mtu": 1500,
                "name": "docker0",
                "mac": "02:42:80:7d:d5:1c",
                "flags": ["up", "broadcast", "multicast"],
                "ip4": "172.17.0.1/16",
                "ip6": "",
                "addrs": null
            }],

            "disk": [{
                "device": "/dev/vda1",
                "total": 42139451392,
                "mountpoint": "/",
                "fstype": "ext4"
            }]
        },
    }
