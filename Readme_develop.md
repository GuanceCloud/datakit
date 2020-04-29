# 采集器开发手册

## 基本原则

- 采集器都放在目录: `plugin/inputs`  
- 新的采集器只要在该目录下创建对应的包。

采集器需要实现telegraf的接口: `Input`或`ServiceInput`:  

    type Input interface {
        // SampleConfig returns the default configuration of the Input
        SampleConfig() string

        // Description returns a one-sentence description on the Input
        Description() string

        // Gather takes in an accumulator and adds the metrics that the Input
        // gathers. This is called every "interval"
        Gather(Accumulator) error
    }

    type ServiceInput interface {
        Input

        // Start the ServiceInput.  The Accumulator may be retained and used until
        // Stop returns.
        Start(Accumulator) error

        // Stop stops the services and closes any necessary channels and connections
        Stop()
    }
 

- `Input`接口适用于固定每隔一段时间上报指标的插件，参考`plugin/inputs/mock`  
- `ServiceInput`接口较灵活，可自行定义上报的行为，参考`plugin/inputs/binlog`    

注:  

- 集成新的采集器时，注意创建一个新分支，**不要在 `dev`/`release` 等分支开发**。  
- 大家可自行 `go get` 安装对应的包，然后将对应的包同步到 `vendor` 目录下，不然 CI 可能编译不过
	- 大家自己开发环境，建议用 `go mod` 做包管理。参见 [how-to](https://gitlab.jiagouyun.com/cloudcare-tools/go-howto)

## 本地测试插件
1. 编译完后，先`[sudo] ./datakit --init`生成默认配置文件 
2. 修改`datakit.conf`，配置`ftdataway`，配置在conf.d目录中要测试的插件，然后再运行`[sudo] ./datakit` 

相关文件:  
`agent`: 即telegraf可执行文件  
`agent.conf`: telegraf的配置文件，由datakit在启动时根据conf.d目录下的配置文件生成  
`agent.log`: telegraf的运行日志  
`conf.d`: 各个插件的配置目录，包括了telegraf的以及自研的插件配置文件，如果配置文件不存在/为空/全注释，则该插件不会起来  

如果只想看指标是否正确(不打到dataway)，可以在datakit.conf中配置output_file将指标打到本地文件中，eg:  
`output_file='~/test/metrics.txt'`

## 编译

建议都以 Makefile 方式编译，这样可以尽早发现采集器以及其依赖包中，是否有对应的平台或动态库依赖。
