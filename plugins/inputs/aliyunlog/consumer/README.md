# Aliyun LOG Go Consumer Library



Aliyun LOG Go Consumer Library 是一个易于使用且高度可配置的golang 类库，专门为大数据高并发场景下的多个消费者协同消费同一个logstore而编写的纯go语言的类库。

## 功能特点
1. 线程安全 - consumer 内所有的方法以及暴露的接口都是线程安全的。
2. 异步拉取 - 调用consumer的拉取日志接口，会把当前拉取任务新开一个groutine中去执行，不会阻塞主groutine的执行。
3. 自动重试 - 对程序运行当中出现的可重试的异常，consumer会自动重试，重试过程不会导致数据的重复消费。
4. 优雅关闭 - 调用关闭程序接口，consumer会等待当前所有已开出的groutine任务结束后在结束主程序，保证下次开始不会重复消费数据。
5. 本地调试 - 可通过配置支持将日志内容输出到本地或控制台，并支持轮转、日志数、轮转大小设置。
6. 高性能 - 基于go语言的特性，go的goroutine在并发多任务处理能力上有着与生俱来的优势。所以consumer 对每一个获得的可消费分区都会开启一个单独的groutine去执行消费任务，相对比直接使用cpu线程处理，对系统性能消耗更小，效率更高。
7. 使用简单 - 在整个使用过程中，不会产生数据丢失，以及重复，用户只需要配置简单配置文件，创建消费者实例，写处理日志的代码就可以，用户只需要把重心放到自己的消费逻辑上面即可，不需关心消费断点保存，以及错误重试等问题。

## 功能优势

使用consumer library 相对于直接通过 API 或 SDK 从 LogStore 拉取数据进行消费会有如下优势。

- 用户可以创建多个消费者对同一Logstore中的数据进行消费，而且不用关心消费者之间的负载均衡，consumer library 会进行自动处理，并且保证数据不会被重复消费。在cpu等资源有限情况下可以尽最大能力去消费logstore中的数据，并且会自动为用户保存消费断点到服务端。
- 当网络不稳定出现网络震荡时，consumer library可以在网络恢复时继续消费并且保证数据不会丢失及重复消费。
- 提供了更多高阶用法，使用户可以通过多种方法去调控运行中的consumer library，具体事例请参考[aliyun log go consumer library 高阶用法](https://yq.aliyun.com/articles/693820)



## 安装

请先克隆代码到自己的GOPATH路径下(源码地址：[aliyun-go-consumer-library](https://github.com/aliyun/aliyun-log-go-sdk))，项目使用了vendor工具管理第三方依赖包，所以克隆下来项目以后无需安装任何第三方工具包。

```shell
git clone git@github.com:aliyun/aliyun-log-go-sdk.git
```



##原理剖析及快速入门

参考教程: [ALiyun LOG Go Consumer Library 快速入门及原理剖析](https://yq.aliyun.com/articles/693820)



## 使用步骤

1.**配置LogHubConfig**

LogHubConfig是提供给用户的配置类，用于配置消费策略，您可以根据不同的需求设定不同的值，各参数含义如其中所示

2.**覆写消费逻辑**

```
func process(shardId int, logGroupList *sls.LogGroupList) string {
    for _, logGroup := range logGroupList.LogGroups {
        err := client.PutLogs(option.Project, "copy-logstore", logGroup)
        if err != nil {
            fmt.Println(err)
        }
    }
    fmt.Println("shardId %v processing works sucess", shardId)
    return "" // 不需要重置检查点情况下，请返回空字符串，如需要重置检查点，请返回需要重置的检查点游标。
}
```

在实际消费当中，您只需要根据自己的需要重新覆写消费函数process 即可，上图只是一个简单的demo,将consumer获取到的日志进行了打印处理，注意，该函数参数和返回值不可改变，否则会导致消费失败。

3.**创建消费者并开始消费**

```
// option是LogHubConfig的实例
consumerWorker := consumerLibrary.InitConsumerWorker(option, process)
// 调用Start方法开始消费
consumerWorker.Start()
```

调用InitConsumerWorkwer方法，将配置实例对象和消费函数传递到参数中生成消费者实例对象,调用Start方法进行消费。

4.**关闭消费者**

```
ch:=make(chan os.Signal) //将os信号值作为信道
signal.Notify(ch)
consumerWorker.Start() 
if _, ok := <-ch; ok { // 当获取到os停止信号以后，例如ctrl+c触发的os信号，会调用消费者退出方法进行退出。
    consumerWorker.StopAndWait() 
}
```

上图中的例子通过go的信道做了os信号的监听，当监听到用户触发了os退出信号以后，调用StopAndWait()方法进行退出，用户可以根据自己的需要设计自己的退出逻辑，只需要调用StopAndWait()即可。



## 简单样例

为了方便用户可以更快速的上手consumer library 我们提供了两个简单的通过代码操作consumer library的简单样例，请参考[consumer library example](https://github.com/aliyun/aliyun-log-go-sdk/tree/master/example/consumer)

## 问题反馈
如果您在使用过程中遇到了问题，可以创建 [GitHub Issue](https://github.com/aliyun/aliyun-log-go-sdk/issues) 或者前往阿里云支持中心[提交工单](https://workorder.console.aliyun.com/#/ticket/createIndex)。
