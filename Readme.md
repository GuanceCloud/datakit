<h1>安装</h1>

`sudo bash -c "$(curl http://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/ftcollector/test/install.sh)"`


<h1>配置</h1>

<h3>[Mysql/Mariadb]</h3>

权限：`SUPER, REPLICATION CLIENT`

    log_bin = ON
    binlog_format = ROW  
    binlog_row_image = FULL #mysql


<h3>ftcollector中binlog配置规范</h3>

目前不支持全库抓取，即至少配置一个表

    ...

    binlog:
    disable: false #是否开启binlog抓取
    jobs:
    - ft_gateway: "http://localhost:9528/v1/write/metrics" #ftgateway地址
        addr: "192.168.56.20:3306" #mysql地址(包含端口)
        user: "" #mysql用户名
        password: "" #mysql密码

        inputs:
        - database: test #目标数据库
        tables: #至少配置一个表
        - table: table1 #表名
            measurement: table1 #指标名称，不设置则为表名
            columns: #设置表中低端哪些作为field，哪些作为tag，不支持blob类型
            username: field #至少配置一个field
            sex: tag #可为空，应使用值不常变或只有固定范围值的表字段
        #  exclude_events:#监听的事件，默认全部开启：insert, delete, update
        #  - delete
        #- table: table2
        #  measurement: table2
        #  columns:
        #    f1: field
        #    t1: tag
        #exclude_tables:#排除以下表名
        #- table3

