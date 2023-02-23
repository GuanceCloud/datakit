# Zap Logger Wrapper

本封装提供一个基本的 zap logger 封装，其目的在于统一部分项目之间的日志形式。

## 基本使用

### 在具体模块中使用

```golang
package abc

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func main() {
	r, err := logger.InitRoot(&logger.Option{
		Path: "/path/to/app/log", // 如果不填写路径，日志将出到 stdout
		Level: logger.DEBUG,      // 默认为 DEBUG
		Flags: logger.OPT_DEFAULT,// 开启了自动切割
	})
}
```

## 关于自动切割

默认情况下，会按照 32MB（大约） 一个文件来切割，最大保持 5 个切片，保存时长为 30 天。

## 提供环境变量来配置日志路径

调用 `InitRoot()` 时，如果传入的路径为空字符串，那么会尝试从 `LOGGER_PATH` 这个环境变量中获取有效的日志路径。某些情况下，可以将该路径设置成 `/dev/null`（UNIX） 或 `nul`（windows），用来屏蔽日志输出。
