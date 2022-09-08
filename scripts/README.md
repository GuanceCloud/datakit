# 打数据工具

post-v1-write.go 是一个用来模拟 DK 客户端往 DK 发送数据的程序，它有两个功能：

- 产生较大量的数据，以行协议方式发送给 DK 的 /v1/write/ 接口
- 本身作为 dataway，DK 收到这些行协议数据后，将数据发送给该 dataway，以免给中心造成垃圾数据

其目的是测试 datakit 的 IO 以及数据发送模块。

可以启动两个 post-v1-write.go 程序，一个作为数据压测方，专门打数据给 DK，一个作为 dataway，接收 DK 上传的数据：

```
# 启动 0 个 worker，即仅仅作为 dataway 来运行
a.out -worer 0

# 启动一个打数据 worker，同时禁掉 dataway 功能。worker 每次发送数据的间隔为 1ms
a.out -worker 1 -dataway "" -worker-sleep 1ms
```

更多用法，参考 `a.out -h`。

数据类容分别在当前目录下的几个文件：

- logging.data
- metric.data
- object.data

可以适当修改这几个文件，然后重新编译 post-v1-write.go 即可。
