# Zap Logger Wrapper

本封装提供一个基本的 zap logger 封装，其目的在于统一部分项目之间的日志形式。

## 基本使用

### 在具体模块中使用

```golang
package abc

import (
	"github.com/GuanceCloud/cliutils/logger"
)

var (
	// 最好在模块中将 log 初始化一下，这样比较保险
	log = logger.DefaultSLogger("abc")
)

// 模块初始化
func Init() {
	log = logger.SLogger("abc")
}

func foo() {
	log.Debug("this is debug message")
	log.Infof("this is info message from %d", 1024)
}
```

### 在项目中使用 

一般而言，我在项目中使用，需要初始化一个 root-logger，该 root-logger 定义了全局的日志存放路径、日志等级等属性：

```golang
package main

import (
	"github.com/GuanceCloud/cliutils/logger"
)

func main() {
	r, err := logger.InitRoot(&logger.Option{
		Path: "/path/to/app/log", // 如果不填写路径，日志将出到 stdout
		Level: logger.DEBUG,      // 默认为 DEBUG
		Flags: logger.OPT_DEFAULT,// 开启了自动切割
	})
}
```

### 增加频率控制

如果某些日志比较高频但又不能完全移除，可以用带频率控制的日志函数。

```golang
import (
	"github.com/GuanceCloud/cliutils/logger"
)

logRate := 1.0
log = logger.SLogger("module-name", logger.WithRateLimiter(logRate, "")) // 每秒最多输出一条日志

// busy loop...
for {
    log.RLInfof(logRate, "this is high frequency log: %d", 1024)
    // 其它代码...
}
```

如果希望有多种频率，那么多加几个即可：

```golang
logRate1_0 := 1.0
logRate0_5 := .5
log = logger.SLogger("module-name",
        logger.WithRateLimiter(logRate1_0, ""),
        logger.WithRateLimiter(logRate0_5, ""))

for {
    log.RLInfof(logRate1_0, "this is 1 log/sec high frequency log: %d", 1024)
    log.RLInfof(logRate0_5, "this is 0.5 log/sec high frequency log: %d", 1024)
}
```

其它几个常规日志函数，只需加上 `RL` 前缀即可：

```golang
log.RLWarnf()
log.RLWarn()
log.RLDebugf()
log.RLDebug()
log.RLErrorf()
log.RLError()
```

注意事项：

- 如果 logger 没有设置 rate limit，或者指定频率的 limiter 不存在，那么调用 `RLXXX()` 函数等价于调用 `XXX()`
- 重复追加同样频率的控制，只有一个会生效

## 关于自动切割

默认情况下，会按照 32MB（大约） 一个文件来切割，最大保持 5 个切片，保存时长为 30 天。

## 提供环境变量来配置日志路径

调用 `InitRoot()` 时，如果传入的路径为空字符串，那么会尝试从 `LOGGER_PATH` 这个环境变量中获取有效的日志路径。某些情况下，可以将该路径设置成 `/dev/null`（UNIX） 或 `nul`（windows），用来屏蔽日志输出。
