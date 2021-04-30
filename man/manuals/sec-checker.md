{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# 简介

一般在运维过程中有非常重要的工作就是对系统，软件，包括日志等一系列的状态进行巡检，传统方案往往是通过工程师编写shell（bash）脚本进行类似的工作，通过一些远程的脚本管理工具实现集群的管理。但这种方法实际上非常危险，由于系统巡检操作存在权限过高的问题，往往使用root方式运行，一旦恶意脚本执行，后果不堪设想。  
实际情况中存在两种恶意脚本，一种是恶意命令，如 `rm -rf`，另外一种是进行数据偷窃，如将数据通过网络 IO 泄露给外部。 
因此 Security Checker 希望提供一种新型的安全的脚本方式（限制命令执行，限制本地IO，限制网络IO）来保证所有的行为安全可控，并且 Security Checker 将以日志方式通过统一的网络模型进行巡检事件的收集。同时 Security Checker 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。

具体的使用方法请参考[Security Checker帮助]()


# 接入 DataKit

可以将检测日志发送到 DataKit，再由 DataKit 上报到 DataFlux。

将配置文件 `checker.conf` 的 `output` 设置为 DataKit 的 server 地址：  

```toml
...

output = 'http://localhost:9529/v1/write/security' # datakit 1.1.6(含)以上版本才支持

...
```
