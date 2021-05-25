
## 前置条件

- 已安装 DataKit（[DataKit 安装文档](../../../02-datakit采集器/index.md)）
- python3 环境 且在 datakti 安装目录下的 `externals/qyt_all` 运行  `pip install -r requirement.txt` 安装python 脚本所依赖包文件

## 配置

进入 DataKit 安装目录下的 conf.d/qyt_all 目录，复制 qyt_all.conf.sample 并命名为 qyt_all.conf 示例如下：

**设置：**

```
[[inputs.external]]

   name = 'qyt_all'  # required

   # 是否以后台方式运行外部采集器
   daemon = false
   # 如果以非 daemon 方式运行外部采集器，则以该间隔多次运行外部采集器
   interval = '30s'

   # 外部采集器可执行程序路径(尽可能写绝对路径)
   cmd = "/usr/local/bin/python3" # python3路径

   args = [
     "/usr/local/cloudcare/dataflux/datakit/externals/qyt_all/main.py",      # python 脚本 入口函数
     "/usr/local/cloudcare/dataflux/datakit/externals/qyt_all/config.conf"   # python 脚本 依赖配置文件
   ]

   envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]  
```

## python 脚本配置说明如下
配置路径 : datakit 安装目录下面 `externals/qyt_all/config.conf` 如 `/usr/local/cloudcare/dataflux/datakit/externals/qyt_all/config.conf`

```
    [input.quanyuantang]
      log_level = "info"                  ## python 脚本日志等级     
      log_path = "/usr/local/cloudcare/dataflux/datakit/externals/qyt_all/log"    # python 日志路径
    [[input.quanyuantang.es_index]]       ## es_index 指标相关配置
      host = ""                           ## es host example "http://ip:port"
      user = ""                    ## 用户名
      password = ""             ## 密码
    [[input.quanyuantang.mysql]]
      host = ""                  ## 主机地址
      port = 3306                ## 主机端口 默认：3306 此处填写数字
      user = ""                  ## 账号
      password = ""              ## 密码
      hostname = ""              ## 主机名称
    [[input.quanyuantang.oracle]]
      connect_string = ""       ## 连接字符串 例：dbmonitor/dbmonitor@localhost/db
      oracle_server = ""        ## oracle_server
      oracle_port = ""          ## oracle_port
      host = ""                 ## host
      service_name = ""         ## service_name
      instance_id = ""          ## instance_id
      instance_desc = ""        ## instance_desc
```

## python 脚本所依赖的 环境介绍
### 1. 安装 Oracle Instant Client (linux)
```
$ wget -q https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/dataflux-func/resource/oracle-instantclient-basic-linux.x64-19.6.0.0.0dbru.zip  && unzip oracle-instantclient-basic-linux.x64-19.6.0.0.0dbru.zip -d /opt/oracle && sh -c "echo /opt/oracle/instantclient_19_6 > /etc/ld.so.conf.d/oracle-instantclient.conf" && ldconfig
```



  



