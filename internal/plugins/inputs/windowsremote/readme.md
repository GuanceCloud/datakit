# windows_remote 采集器

## wmi 配置和采集

配置说明：

```toml
[[inputs.windows_remote]]
  ## ip_list 和 cidrs 共同组成主机ip列表
  ip_list       = [ ]  # e.g. ["127.0.0.1"]
  cidrs         = [ ]  # e.g. ["10.100.1.0/24"]
  # ip 扫描时间间隔
  scan_interval = "10m"

  ## 关闭选举
  election = false
  ## Maximum number of workers. Default value is calculated as datakit.AvailableCPUs * 2 + 1.
  worker_num = 0

  ## Select mode 'wmi' or 'snmp'
  mode = "snmp"

  ## WMI Collection Module
  [inputs.windows_remote.wmi]
    port   = 135  # Port for WMI (DCOM 135 / WinRM 5985)
    log_enable = true
    ## Authentication configuration 使用统一的用户名和密码
    [inputs.windows_remote.wmi.auth]
      username  = "user"
      password  = "password"

  ## SNMP Collection Module (Independent configuration)
  [inputs.windows_remote.snmp]
    ports     = [ 161 ]  # SNMP ports (default 161)
    community = "datakit"

  [inputs.windows_remote.tags]
  # "some_key" = "some_value"
```

安装完成之后配置服务：打开服务列表 找到`datakit.ext`右键`Properties`,找到 `Log On` 选中 `This account` 并填写用户名（如果是admin应该是：.\administrator）和密码

配置完成，并赋予权限之后 重启datakit。



