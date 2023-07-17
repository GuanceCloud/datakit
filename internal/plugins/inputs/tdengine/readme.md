## TDengine 采集器开发

配置文件：配置用户名、密码、库名等。
``` toml
[[inputs.tdengine]]
  ## adapter config (Required)
  adapter_endpoint = "<http://taosadapter.test.com>"
  user = "<userName>"
  password = "<pw>"
  
  ## Setting disable_temperature_collect to false will collect cpu temperature stats for linux.

  ## add tag (optional)
  [inputs.cpu.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

```

### 启动流程
1. 健康检查 CheckHealth() : 访问不通或者返回码错误时->报错。否则 下一步.
1. 定时执行所有指标集，映射到 measurement 中。最后放入管道等待发送到IO。


### 指标集

指标集详细说明文档位置: man/manuals/tdengine.md

TODO : 

1. 登录日志
1. 磁盘使用率：在每个盘上的使用率。
1. 类型检查完善
