# 重新梳理pp的结构体和span父子结构

目前看到的情况：

- spanChunk(统称chunk) 是 span 的补充，chunk先发送，所有的chunk发送结束之后，最后 span 才发送到dk
- span 中 parent id 如果不为 -1 ，则证明是一个子span，后面一定有一个 span 为 -1的情况。
- chunk 中的 event 有先后 但不是父子级关系，区别在于，有没有depth。
- 关于 depth， 加入chunk中 从depth:4 开始。那么 123 是不知道的。由第一条得知。 
- 每一个 chunk 中的 都是一个span。没有depth的 统一归与上一个有depth的event。
- 接收信息的方法处都可以通过 `md, ok := metadata.FromIncomingContext(ctx)` 获取 meta 信息。

## 解决办法
event列表中 每一个event 的 span_id 是随机。

先遍历第一遍，赋值好每一个 span_id 。---> 当前函数缓存

再遍历第二遍，如当前depth=4 从缓存中找3 ，有 --> 父子关系。 没有--> 内存缓存（等待最后的span到dk，找到depth=3 赋值span_id）

缓存 span parentID 不为-1 。

缓存event

---

### 缓存 event
在接收 chunk 时只做一件事，放到内存缓存（key： agentid^startTime^sequence）中。

接收到 span 之后，按照缓存 key 将缓存的 spanEvent 全部取出并进行排序（由小及大）

缓存：`map[string][]string{event}` 其中 map 的 key:agentid^startTime^sequence val:event 数组。

必须设置缓存超时，否则出现 OOM 情况。


### 缓存 span
多服务情况下，出现关联错误问题。需要将 span 的 parent id 不为 -1 的情况缓存。这样在 parentID==-1 的情况下，将缓存中的span 取出，根据 nextSpan中的ID。

如果缓存时间到没有等待到root span 则应该作为顶层span处理。（出现这种情况，可以理解为顶层的span已经发送到了别的dk）

### 缓存 agentInfo
agentInfo 中包括 hostname,ip,version 等信息。将 agentID 作为 key 缓存在本地，填充到 span 和 event 中。


### 缓存 meta
meta 示例：
```text
map[:authority:[10.200.14.188:9991] agentid:[666] applicationname:[tmall] content-type:[application/grpc] grpc-accept-encoding:[gzip] servicetype:[1210] starttime:[1699946460396] user-agent:[grpc-java-netty/1.36.2]]
```

`agentid` 或者 `starttime` 可以作为缓存的 key。当前使用的是 `agentid`

## 时间问题
```text
2023-11-03T15:13:00.228+0800	INFO	pinpointV2	pinpoint/span.go:59	ConvertPSpanToDKTrace x=version:1 transactionId:{agentStartTime:1698995586865 sequence:1} spanId:1884270327942814846 parentSpanId:5211995155288026382 startTime:1698995604377 elapsed:75 apiId:55 serviceType:1010 acceptEvent:{rpc:"/resource" endPoint:"localhost:8080" remoteAddr:"127.0.0.1" parentInfo:{parentApplicationName:"server" parentApplicationType:1210 acceptorHost:"localhost:8080"}} annotation:{key:46 value:{intValue:200}} 
spanEvent:{depth:1 startElapsed:1 endElapsed:72 serviceType:1011 apiId:54} 
spanEvent:{sequence:1 depth:2 endElapsed:72 serviceType:5051 apiId:43} 
spanEvent:{sequence:2 depth:3 endElapsed:1 serviceType:5071 apiId:47} 
spanEvent:{sequence:3 startElapsed:63 endElapsed:8 serviceType:5011 apiId:95} applicationServiceType:1210

2023-11-03T15:13:00.273+0800	INFO	pinpointV2	pinpoint/span.go:59	ConvertPSpanToDKTrace x=version:1 transactionId:{agentStartTime:1698995586865 sequence:1} spanId:1955224219112524780 parentSpanId:5211995155288026382 startTime:1698995604496 elapsed:9 apiId:55 serviceType:1010 acceptEvent:{rpc:"/auth" endPoint:"localhost:8080" remoteAddr:"127.0.0.1" parentInfo:{parentApplicationName:"server" parentApplicationType:1210 acceptorHost:"localhost:8080"}} annotation:{key:46 value:{intValue:200}} 
spanEvent:{depth:1 endElapsed:9 serviceType:1011 apiId:54} 
spanEvent:{sequence:1 depth:2 endElapsed:9 serviceType:5051 apiId:43} 
spanEvent:{sequence:2 depth:3 startElapsed:1 endElapsed:1 serviceType:5071 apiId:48} 
spanEvent:{sequence:3 startElapsed:1 endElapsed:6 serviceType:5011 apiId:95} applicationServiceType:1210

2023-11-03T15:13:00.607+0800	INFO	pinpointV2	pinpoint/span.go:59	ConvertPSpanToDKTrace x=version:1 transactionId:{agentStartTime:1698995586865 sequence:1} spanId:-5802727676775256750 parentSpanId:5211995155288026382 startTime:1698995604720 elapsed:122 apiId:47 serviceType:1010 acceptEvent:{rpc:"/client" endPoint:"localhost:8081" remoteAddr:"127.0.0.1" parentInfo:{parentApplicationName:"server" parentApplicationType:1210 acceptorHost:"localhost:8081"}} annotation:{key:46 value:{intValue:200}} 
spanEvent:{depth:1 startElapsed:17 endElapsed:104 serviceType:1011 apiId:46} 
spanEvent:{sequence:1 depth:2 startElapsed:48 endElapsed:56 serviceType:5051 apiId:43} 
spanEvent:{sequence:2 depth:3 startElapsed:30 endElapsed:1 serviceType:5071 apiId:45} applicationServiceType:1210

2023-11-03T15:13:00.629+0800	INFO	pinpointV2	pinpoint/span.go:59	ConvertPSpanToDKTrace x=version:1 transactionId:{agentStartTime:1698995586865 sequence:1} spanId:-8291885501612733658 parentSpanId:5211995155288026382 startTime:1698995604882 elapsed:7 apiId:55 serviceType:1010 acceptEvent:{rpc:"/billing" endPoint:"localhost:8080" remoteAddr:"127.0.0.1" parentInfo:{parentApplicationName:"server" parentApplicationType:1210 acceptorHost:"localhost:8080"}} annotation:{key:46 value:{intValue:200}} 
spanEvent:{depth:1 endElapsed:7 serviceType:1011 annotation:{key:41 value:{stringValue:"tag=null"}} apiId:54} 
spanEvent:{sequence:1 depth:2 endElapsed:5 serviceType:5051 apiId:43} 
spanEvent:{sequence:2 depth:3 startElapsed:1 endElapsed:2 serviceType:5071 apiId:49} 
spanEvent:{sequence:3 startElapsed:3 endElapsed:1 serviceType:5011 apiId:95} applicationServiceType:1210

2023-11-03T15:13:00.660+0800	INFO	pinpointV2	pinpoint/span.go:59	ConvertPSpanToDKTrace x=version:1 transactionId:{agentStartTime:1698995586865 sequence:1} spanId:5211995155288026382 parentSpanId:-1 startTime:1698995604209 elapsed:710 apiId:55 serviceType:1010 acceptEvent:{rpc:"/gateway" endPoint:"10.200.14.226:8080" remoteAddr:"10.200.6.21"} annotation:{key:46 value:{intValue:200}} 
spanEvent:{depth:1 startElapsed:10 endElapsed:700 serviceType:1011 apiId:54} 
spanEvent:{sequence:1 depth:2 startElapsed:28 endElapsed:672 serviceType:5051 apiId:43} 
spanEvent:{sequence:2 depth:3 startElapsed:36 endElapsed:633 serviceType:5071 apiId:46} 
spanEvent:{sequence:3 depth:4 startElapsed:81 endElapsed:8 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8080/resource"}} apiId:103 nextEvent:{messageEvent:{nextSpanId:1884270327942814846 destinationId:"localhost:8080"}}} 
spanEvent:{sequence:4 startElapsed:8 endElapsed:85 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8080/resource"}} apiId:104 nextEvent:{messageEvent:{destinationId:"localhost:8080"}}} 
spanEvent:{sequence:5 startElapsed:95 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8080/auth"}} apiId:103 nextEvent:{messageEvent:{nextSpanId:1955224219112524780 destinationId:"localhost:8080"}}} 
spanEvent:{sequence:6 endElapsed:39 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8080/auth"}} apiId:104 nextEvent:{messageEvent:{destinationId:"localhost:8080"}}} 
spanEvent:{sequence:7 startElapsed:41 endElapsed:10 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8081/client"}} apiId:103 nextEvent:{messageEvent:{nextSpanId:-5802727676775256750 destinationId:"localhost:8081"}}} 
spanEvent:{sequence:8 startElapsed:10 endElapsed:331 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8081/client"}} apiId:104 nextEvent:{messageEvent:{destinationId:"localhost:8081"}}} 
spanEvent:{sequence:9 startElapsed:356 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8080/billing?tag=null"}} apiId:103 nextEvent:{messageEvent:{nextSpanId:-8291885501612733658 destinationId:"localhost:8080"}}} 
spanEvent:{sequence:10 endElapsed:14 serviceType:9055 annotation:{key:40 value:{stringValue:"http://localhost:8080/billing?tag=null"}} apiId:104 nextEvent:{messageEvent:{destinationId:"localhost:8080"}}} 
spanEvent:{sequence:11 startElapsed:16 endElapsed:25 serviceType:5011 apiId:30} 
spanEvent:{sequence:12 depth:3 startElapsed:27 endElapsed:1 serviceType:5011 apiId:95} applicationServiceType:1210
```

从日志上看来，startElapsed 应当累计之前所有的时间，例如 `sequence:4` 的event  startElapsed:8，则证明当前event的起始时间为 8+81+36+28+10，后面的起始时间都应该按照这个逻辑来。

另一份带有 keyTime 的日志下载路径 https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/PinPoint/starttime.log

如果有key time，时间计算应该是，startElapsed += (keyTime - span.start)

## 总结
综上所述，dataKit 在接收并处理 PinPoint 发来的数据具有关联性，与ddtrace 和 OTEL 有这很大的区别。所以 在实际的测试环境和生产环境中，应当***有且只有一台dk***来处理PinPoint，才能使链路完美的串联起来。