采集器都放在目录: `plugin/inputs`  
新的采集器只要在该目录下创建对应的包。

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
 

 `Input`接口适用于固定每隔一段时间上报指标的插件，参考`plugin/inputs/mock`  
 `ServiceInput`接口较灵活，可自行定义上报的行为，参考`plugin/inputs/binlog`    

注:  
 集成新的采集器时，注意创建一个新分支。  
 目前datakit的依赖包还没有整理出来放到vendor下，需要自己get下。。。


###本地测试插件    
1. 编译完后，先`[sudo] ./datakit --init`生成默认配置文件 
2. 修改`datakit.conf`，配置`ftdataway`，配置在conf.d目录中要测试的插件，然后再运行`[sudo] ./datakit` 

相关文件:  
`agent`: 即telegraf可执行文件  
`agent.conf`: telegraf的配置文件，由datakit在启动时根据conf.d目录下的配置文件生成  
`agent.log`: telegraf的运行日志  
`conf.d`: 各个插件的配置目录，包括了telegraf的以及自研的插件配置文件，如果配置文件不存在/为空/全注释，则该插件不会起来  

如果只想看指标是否正确(不打到dataway)，可以在datakit.conf中配置output_file将指标打到本地文件中，eg:  
`output_file='~/test/metrics.txt'`