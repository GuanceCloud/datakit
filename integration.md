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
