## `datakit --api-restart`
### 简介
本命令主要是用于通过 DCA API方式来重启DataKit，不同平台实现方式不同。
### 实现
当DCA API需要重启 DataKit 时，会通过命令来执行`datakit --api-restart`，实现 DataKit 重启的目的。

```go
  cmd := exec.Command("/path/to/datakit", "--api-restart")
  cmd.Start()
```

**linux**
linux实现起来相对简单，直接调用service restart 方法
```go
import (
  dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)
svc, err := dkservice.NewService()
service.Control(svc, "restart")
```

**windows**
先调用停止命令`datakit.exe --stop`来停止 DataKit，停止成功后，再调用启动命令 `datakit.exe --start` 来启动 DataKit。

**darwin**
通过`signal.Nofify`捕获父进程的信号，目的是在主进程退出的时候，避免当前进程被杀死，导致 DataKit 启动失败。
具体流程：
`signal.Notify` -> `stop Datakit` -> `wait signal` -> `start datakit` -> `exit`