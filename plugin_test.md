###datakit本地测试插件    
1. 编译完后，先`[sudo] ./datakit --init`生成默认配置文件 
2. 修改`datakit.conf`，配置`ftdataway`，配置在conf.d目录中要测试的插件，然后再运行`[sudo] ./datakit` 

相关文件:  
`agent`: 即telegraf可执行文件  
`agent.conf`: telegraf的配置文件，由datakit在启动时根据conf.d目录下的配置文件生成  
`agent.log`: telegraf的运行日志  
`conf.d`: 各个插件的配置目录，包括了telegraf的以及自研的插件配置文件，如果配置文件不存在/为空/全注释，则该插件不会起来  

如果只想看指标是否正确(不打到dataway)，可以在datakit.conf中配置output_file将指标打到本地文件中，eg:  
`output_file='~/test/metrics.txt'`

