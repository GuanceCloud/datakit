
## confd 开发工作计划
1. ### confd 目标:
   1. 通过 confd ，运维可以在线修改 datakit 采集器的配置参数，对相应采集器执行增、删、改、重启。
   2. 通过 confd ，运维可以在线修改 pipeline 脚本。
   3. confd 配置信息的引入，除了修改 datakit.conf 之外，还可以通过 K8S 的环境变量引入。
   4. confd 配置信息的引入，还可以通过安装脚本的参数引入，引入的参数会出现在 datakit.conf 中。
      ```shell
      # Linux/Mac
      DK_CONFD_BACKEND="etcdv3" DK_CONFD_BACKEND_NODES="[127.0.0.1:2379]" DK_DATAWAY="https://openway.guance.com?      token=<TOKEN>" bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
      
      # Windows
      $env:DK_CONFD_BACKEND="etcdv3";$env:DK_CONFD_BACKEND_NODES="[127.0.0.1:2379]"; $env:DK_DATAWAY="https://openway.      guance.com?token=<TOKEN>"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer;       start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .      install.ps1;
   3. #### 附加要求:
      1. 后台源数据重复的，滤除。
      2. cpu、disk 等采集器是默认启动的，confd 可以修改，删除无效。客户如果想删除，需要删除 conf.d 子目录下的对应配置文件，重启  datakit。
      3. datakit 内有的采集器只准启动单例（如cpu、netstat等），confd 的来源数据只保留1条。
      4. 只有配置信息变化的采集器才重启，不可以重启全部采集器。
      5. 最新的 confd 更新后，在 /datakit/data/remote.conf 留有备份。
      6. 最新的 pipeline 更新后，在 /datakit/pipeline_confd/ 子目录留有备份。

2. ### 业务实现以及 confd 模块功能:
   1. 根键KEY为 /datakit/
   2. confd 完整键KEY为 /datakit/confd/XXX , XXX 随意写，建议是采集器名字。系统认可的采集器名字来自于 `[[inputs.XXX]]`
      1. VELUE值是完整采集器的配置（就是当前cond.d目录下每个文件的完整内容），用反引号 `内容` 框起来，格式如下：
         ```toml
         `[[inputs.netstat]]
         ##(optional) collect interval, default is 10 seconds
         interval = '10s'

         [inputs.netstat.tags]
         # some_tag = "some_value"
         # more_tag = "some_other_value"`
         ```
      4. etcd 可以用 ./etcdctl put /datakit/XXX "$(cat 全路径配置文件名)" 注入。
      5. file 文件直接放入上述内容。
      6. redis 目前是用go单独写了一个项目注入的。
      7. 已经启动的采集器 .config 信息来自于inputs.InputsInfo。
      8. 检测到数据变化后，懒加载2秒，GetValues 全量数据。（懒加载2秒，是防止运维跑数据的过程中，采集器频繁重启）。
      9. 拿到 confd 最新全量配置数据之后，解析成完整的对象，congf 数据和 inputs.InputsInfo 正向、反向分别比对，查出来增、删、改信息，  有变化的，进行必要的增加、删除、重启采集器动作。
   3. pipeline 完整键KEY为 /datakit/pipeline/YYY/XXX.p
      1. YYY 表示指标集名。例如 `metric` `logging` 。
      2. XXX 表示采集器名。例如 `cpu` `mem` 。
      3. value 值就是 pipeline 脚本本身。  
3. ### 流程:
   1. 在 cmd/datakit/main.go 判断 DK是否有confd需求， config/导向confd.go。
      2. config/confd.go 中监测、读取 后台cond源数据。然后导向 plugins/inputs/confdCompare.go。
      3. confd.go 读取数据，是有可能阻塞的。加 context timeout 60秒，并且 timeout 加 log 日志告警。
      4. confdCompare.go 进行采集器比对，采集器增删改、启动。
      5. config/confd.go 中监测、读取 后台pipeline源数据。注入pipeline进程。然后落盘备份。

4. ### goroutine 使用情况:
   1. 使用 sync.Once 启动 一个 goroutine ，确保单例运行。
   2. 每使用一个不同种类的后台源，开一个 goroutine ，监测本种类后台源有变化的信号。

5. ### K8S环境变量引入 congfd配置开发过程:
      1. config/env.go 加入系列语句
   ```
   	// k8s 环境变量配置 confd 后台源
   	if backend := datakit.GetEnv("ENV_CONFD_BACKEND"); backend != "" {
   		authToken := datakit.GetEnv("ENV_CONFD_AUTH_TOKEN")
   		authType := datakit.GetEnv("ENV_CONFD_AUTH_TYPE")
   		basicAuthBool := datakit.GetEnv("ENV_CONFD_BASIC_AUTH")    // 可选
         ......
   ```

6.  ### datakit 安装过程引入 congfd配置开发过程:
      1. install.sh.template 加入系列语句，for linux 系统
      ```
            confd_backend=""
             if [ -n "$CONFD_BACKEND" ]; then
             	confd_backend=$CONFD_BACKEND
             fi
            --confd-backend ="${confd_backend}" \ 
      ```
      2. install.ps1.template 加入系列语句，for Windows 系统
      ```
            $confd_backend=""
            $x = [Environment]::GetEnvironmentVariable   ("CONFD_BACKEND")
            if ($x -ne $null) {
            	$confd_backend = $x
            	Write-COutput yellow "* set   confd_backend"
            }
            --confd-backend ="${confd_backend}"  
      ```
      3. cmd/installer/installer/dkconf.go 加入系列语句
      ```
            var ConfdBackend string 
      ```
      4. cmd/installer/main.go 加入系列语句
      ```
            flag.StringVar(&installer.ConfdBackend, "confd-backend", "", "confd backend")
      ```
      5. cmd/installer/installer/install.go 加入系列语句
      ```
	   if ConfdBackend != "" {
	   	mc.Confds = []*config.ConfdCfg{{
	   		Enable:         true,
	   		Backend:        ConfdBackend,
            ...   
	   	}}
	   }
      ```
      6. 最后，cmd/installer/main.go 执行的时候，会把 mc.Confd 跟随 mc 序列化 进 datakit.conf里

7. ### 测试用例（TODO）:

8. ### confd 测试前期准备:
   1. 下载 docker 支持
   2. 下载 go 语言支持，建议1.18。
   3. 下载vscode

9. ### 测试时候 datakit.conf 附加内容
   
   测试的时候，把 ip 改成 127.0.0.1 或者 你自己的 ip

   测试的时候，把 file 改成 你自己的 文件位置

   测试的时候，enable = false 表示不测这一项

   测试的时候，client_key = "654123" 改成 你的 redis 密码，也可以没有这一行
 ```
[[confds]]
  enable = true
  backend = "file"
  file = ["/home/zhangsr/GolandProjects/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/config/confdtest/testconf.conf","/home/zhangsr/GolandProjects/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/config/confdtest/testconf02.conf"]
[[confds]]
  enable = true
  backend = "zookeeper"
  nodes = ["10.100.65.54:2181"]
[[confds]]
  enable = true
  backend = "etcdv3"
  nodes = ["10.100.65.54:2379"]
[[confds]]
  enable = true
  backend = "redis"
  nodes = ["10.100.65.54:6379"]
  client_key = "654123"
[[confds]]
  enable = true
  backend = "consul"
  nodes = ["10.100.65.54:8500"]

 ```
10. ### 测试准备
 ```
file 测试 cpu
zookeeper 测试 netstat，起多例程只能成功一个
etcdv3 测试 cpu
redis 测试 net
consul 测试 nvidia_smi，多例程。
 ```

测试准备
```
docker 开启 zookeeper 服务， 端口：2181
docker pull zookeeper
docker run -d -e TZ="Asia/Shanghai" -p 2181:2181 -v /data/zookeeper:/data --name zookeeper --restart always zookeeper
客户端：docker run -it --rm --link zookeeper:zookeeper zookeeper zkCli.sh -server zookeeper

docker 开启 etcd 服务， 端口：2379
docker pull quay.io/coreos/etcd:v3.3.0
docker run -it --rm -p 2379:2379 quay.io/coreos/etcd:v3.3.0 etcd --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2379
客户端：./etcdctl put ...

安装 开启 redis 服务， 端口：6379，密码 654123
客户端：redis-cli  然后 auth 654123

安装 开启 consul 服务，端口：8500
consul agent -dev -client 0.0.0.0 -ui
客户端： localhost:8500
```

 10. ### 测试方案 :
      1. confd 测试安排
       ```
      file 测试 cpu
      zookeeper 测试 netstat，起多例程只能成功一个
      etcdv3 测试 cpu
      redis 测试 net
      consul 测试 nvidia_smi，多例程。多例程观察通过 pipeline 中注入的 tconsul 指标，是上来 1个还是多个。 
       ```
      1. 通过datakit.conf 引入confd 配置。
      2. 观察方法：[快捷入口]->[DQL查询]->输入类似指令 `M::`nvidia_smi`:((`tconsul`)) limit 20` 
      3. confd 修改、重启 input 采集器。
         通过`etcdv3`或`zookeeper`或`redis`或`consul`，更改`/datakit/confd/XXX`对应的 input 配置，观察前端页面对应input采集器变化。
      4. confd 修改 pipeline。 
         通过`etcdv3`或`zookeeper`或`redis`或`consul`，更改`/datakit/pipeline/YYY/XXX.p`对应的 YYY 类指标的 XXX 指标集，观察前端页面对应指标的变化。
             1. 磁盘 datakit/pipeline/metric/disk.p , 内容 add_key(tconfd, 4) ，发表磁盘来源。
             2. zookeeper 后台源，key=/datakit/pipeline/metric/disk.p 进行 add_key(tconfd, 1~2)变化，或删除键值。在前端观察 disk 下 tconfd 的变化，采信了谁的脚本。
             3. etcdv3 后台源，key=/datakit/pipeline/metric/cpu.p 进行 add_key(tetcdv3, 1~2)变化，或删除键值。在前端观察 cpu 下 tetcdv3 的变化，或消失。
             4. redis 后台源，key=/datakit/pipeline/metric/net.p 进行 add_key(tredis, 1~2)变化，或删除键值。在前端观察 net 下 tredis 的变化，或消失。
             5. cinsul 后台源，key=/datakit/pipeline/metric/nvidia_smi.p 进行 add_key(tconsul, 0~2)变化。在前端观察 nvidia_smi 下 tconsul 的变化。
      5. 多订阅者测试，多台运行DK的机器，订阅同一个数据源，数据源变化，多台机器的DK跟随变化。
      6. 疲劳测试，通过软件，每70秒修改一次数据源，10个小时，所有订阅者的DK依然工作正常。
      7. (TODO，做不了)K8s ENV 引入 confd 配置。 
      8. DK安装的时候，通过参数引入 confd 配置。测试 mac、linux、windoes，成功使 参数值进入到datakit.conf 文件里。
      ```

1.   ### 文档 :
      1. (DONE)文档修改 man/manuals/datakit-daemonset-deploy.md。
      2. (DONE)文档修改 man/manuals/datakit-install.md。 
      3. (DONE)文档修改 man/manuals/confd.md
      4. (DONE)修改 man/manuals/datakit.pages man/manuals/confd.md 挂载到网页上
      5. (DONE)修改 mkdocs.sh 加上 confd.md
      6. (不用做，./mkdocs.sh自动做)三个文档，还要复制到，~/git/dataflux-doc/docs/datakit
2.    ### 待解决问题:
      1. 目前调试成功ctedv3/file/redis/zookeeper ，etcdv2的服务暂时启动不好。应该支持 etcd、consul、vault、environment variables、file、redis、zookeeper、dynamodb、rancher、ssm 等11种。



