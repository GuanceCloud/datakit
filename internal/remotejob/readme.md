# 远程任务

接收观测云中心发下来的任务之后，使用python脚本去执行。最后将结果返回到中心。

## jvm dump
执行命令 
```shell
jmap -dump:live,format=b,file=heap.bin <pid>
```

如果在宿主机环境下需要使用绝对路径执行 `$JAVA_HOME/bin/jmap -dump:live,format=b,file=heap.bin <pid>` 所以在宿主机环境下需要通过env获取java_home


## 问题
启动脚本落盘：方便运行中调试，但是有个问题 重启之后 就会将修改的文件覆盖掉。

1. 脚本是否分开（宿主机，k8s），脚本名字如何定义？  
2. 用户是否可以修改脚本。
3. 下发到DK的参数都有哪些：宿主机：（pid,oss路径）, k8s: (pod_name,pid,oss路径)，文件名是否传递还是自动生成的（jvmdump-2024-10-25-10-43）
4. command 不应该是脚本，应该是 python3 或者 python，、usr/bin/python3.10 ... 或者别的python版本。

脚本：

脚本放到固定目录（template/service-task/）下，脚本是否内嵌到dk中，是否允许用户修改。不同的环境 脚本是否同的？

目前的做法是，内嵌到dk 每次启动的时候将内容写到文件中，以后每次收到任务 都会执行该文件，这样 用户就可修改文件，方便调试。缺点是重启dk之后就复原了。

还可以：不内嵌放到官方文档中，用户开启这个功能 需要下载文件到固定目录，随意修改。

host环境是：jvm_dump_host.py  在k8s环境上是jvm_dump_k8s.py
