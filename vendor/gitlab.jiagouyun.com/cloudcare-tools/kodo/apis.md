 Kodo API 说明

 ## `/v1/datakit/online` |`POST`

 datakit online 通知

 示例：

	POST http://localhost:8080/v1/datakit/online?token=xxx&uuid=xxx
	Content-Type: application/json
	X-Token: tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Version: 0.1-389-gb778f18
	X-Agent-Uid: agnt_xxxxxxxxxxxxxxxxxxxxxxxxxxxx

	[
		"dkUUID_xxxx1", "dkUUID_xxxx2"
	]

	HTTP/1.1 200 OK
	Date: Wed, 26 Dec 2018 06:12:29 GMT
	Content-Length: 79
	Content-Type: application/json

	{
		"code": 200,
		"content": null
	}


## `/v1/cq/create` | `POST`

创建自动聚合规则

| 参数			      | 描述				  | 类型	 | required?  | 
|----------------:|:---------------------:|---------:|:----------:|
| measurement     | 指标名 | str | Yes | 
| aggr_period            |  聚合周期| str eg: "1w","1d","1h","1m","1s","1ms","1ns"| Yes | 
| aggr_every        | 聚合频率 | str eg: "1w","1d","1h","1m","1s","1ms","1ns" | Yes |
| aggr_func | 聚合函数| str | Yes |

示例： 

	POST http://localhost:8080/v1/cq/create
	Content-Type: application/json

	{
		"db_uuid": xxxxx, 
		"workspace_uuid": xxxxx, 
		"measurements": [{
		"measurement": "cpu",
		"aggr_func": "mean",
		"aggr_period": "1m",
		"aggr_every": "5m"}]
	}

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_c497cbbb-70ee-4b70-8121-9e49b8fafe40
	Date: Fri, 10 Apr 2020 13:02:07 GMT
	{
	    "code": 200, 
	    "content": [{
				"aggr_every":"5m",
				"workspace_uuid":"wksp_0bc86b4627c111ea8dd02e762ce1615d",
				"cq_uuid":"cq_ad8ca5b88ad746079508d771c74333b4",
				"db_uuid":"ifdb_5708c9c73eb04c79a31861d4b5990ba4",
				"db":"tanb_mig_test",
				"rp":"rp2",
				"cqrp":"autogen",
				"measurement":"kodo_slow_query",
				"aggr_period":"1m",
				"aggr_func":"sum"
			}],
		"errorCode": ""
	}


## `/v1/cq/modify` | `POST`

更新自动聚合规则， 参数详情见 cq/create

| 参数			      | 描述				  | 类型	 | required?  | 
|----------------:|:---------------------:|---------:|:----------:|
| cq_uuid     | cq 唯一标识 | str | Yes | 
| workspace_uuid | 工作空间UUID | str| Yes|
| aggr_period            |  聚合周期| str eg: "1w","1d","1h","1m","1s","1ms","1ns"| No | 
| aggr_every        | 聚合频率 | str eg: "1w","1d","1h","1m","1s","1ms","1ns" | No |
| aggr_func | 聚合函数| str | No |

示例：

	POST http://localhost:8080/v1/cq/modify
	Content-Type: application/json
	{
		"cq_uuid": xxxxx,
		"workspace_uuid": xxxxx, 
		"aggr_func": "mean",
		"aggr_period": "1m",
		"aggr_every": "5m"
	}


	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_c497cbbb-70ee-4b70-8121-9e49b8fafe40
	Date: Fri, 10 Apr 2020 13:02:07 GMT
	{
	    "code": 200, 
	    "content":
			{
				"aggr_every":"5m",
				"workspace_uuid":"wksp_0bc86b4627c111ea8dd02e762ce1615d",
				"cq_uuid":"cq_ad8ca5b88ad746079508d771c74333b4",
				"db_uuid":"ifdb_5708c9c73eb04c79a31861d4b5990ba4",
				"db":"tanb_mig_test",
				"rp":"rp2",
				"cqrp":"autogen",
				"measurement":"kodo_slow_query",
				"aggr_period":"1m",
				"aggr_func":"sum"
			},
		"errorCode": ""
	}

## `/v1/cq/delete` | `POST`

删除自动聚合规则

示例：

	POST http://localhost:8080/v1/cq/delete
	Content-Type: application/json
	{
		"cq_uuid": xxxxx,
		"workspace_uuid": xxxxx, 
	}


	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_c497cbbb-70ee-4b70-8121-9e49b8fafe40
	Date: Fri, 10 Apr 2020 13:02:07 GMT
	{
	    "code": 200, 
	    "content": nil,
		"errorCode": ""
	}


## `/v1/cq/syncupdate` | `POST`

手动更新CQ，更新某个空间的白名单所有配置的CQ

示例：

	POST http://localhost:8080/v1/cq/syncupdate?workspace_uuid=xxxxxx

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_c497cbbb-70ee-4b70-8121-9e49b8fafCQe40
	Date: Fri, 10 Apr 2020 13:02:07 GMT
	{
	    "code": 200, 
	    "content": nil,
		"errorCode": ""
	}


## `/v1/ck/status` | `GET`

获取ck服务状态

示例：

	GET http://localhost:8080/v1/ck/status
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	
	{
		"code": 200,
		"errorCode": ""
	}

## `/v1/ck/create_db` | `POST`

ck创建db

| 参数			      | 描述				  | 类型	 | required?  | default |
|----------------:|:---------------------:|---------:|:----------:|:------:|
| db_uuid     | 数据库名 | str | Yes | |

示例：

	POST http://localhost:8080/v1/ck/create_db?db_uuid=ifdb_xxx
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	
	{
		"code": 200,
		"errorCode": ""
	}


## `/v1/ck/tablelist` | `GET`

| 参数			      | 描述				  | 类型	 | required?  | default |
|----------------:|:---------------------:|---------:|:----------:|:------:|
| db_uuid     | 数据库名 | str | Yes | |
| type   | 表类型| enum "table,view,all" | no|"all"|

示例：

	GET http://localhost:8080/v1/ck/tablelist?db_uuid=ifdb_xxx&type=all
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 594
	Content-Type: application/json

	{
		"code": 200,
		"content": {
			"Series": [
				{
					"columns": [
						"tableName",
						"type"
					],
					"tags": {},
					"values": [
						[
							"view_m",
							"MaterializedView"
						],
						[
							"test",
							"ReplicatedMergeTree"
						],
						[
							"test_view",
							"View"
						]
					]
				}
			],
			"Message": ""
		},
		"errorCode": ""
	}


## `/v1/ck/tableinfo` | `GET`

| 参数			      | 描述				  | 类型	 | required?  | 
|----------------:|:---------------------:|---------:|:----------:|
| db_uuid     | 数据库名 | str | Yes | 
| table_name   | 表名| str | Yes|

示例：

	GET http://localhost:8080/v1/ck/tableinfo?db_uuid=ifdb_xxx&table_name=xxxx
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 594
	Content-Type: application/json


	{
		"code": 200,
		"content": {
			"Series": [
				{
					"columns": [
						"name",
						"type",
						"default_type",
						"default_expression",
						"comment",
						"codec_expression",
						"ttl_expression"
					],
					"tags": {},
					"values": [
						[
							"tag",
							"String",
							"",
							"",
							"",
							"",
							""
						],
						[
							"time",
							"Int64",
							"",
							"",
							"",
							"",
							""
						]
					]
				}
			],
			"Message": ""
		},
		"errorCode": ""
	}

## `/v1/ck/drop/metrics` | `POST`

    clickhouse 表删除
    
示例：

    POST http://0.0.0.0:9527/v1/ck/drop/metrics
    {
        "db_uuid": "telegraf",   # db-uuid
        "table_name": ""
    }   

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_45a47ea0-10bf-4691-b037-6fa5a75e67ec
	Date: Tue, 07 Apr 2020 09:25:11 GMT
	Content-Length: 40

	{
		"code":200,
		"errorCode":"",
		"message":""
	}

## `/v1/ck/read/metrics` | `POST`

ck 读数据

| 参数			      | 描述				  | 类型	 | required?  | 默认值   |
|----------------:|:---------------------:|---------:|:----------:|:--------:|
| db_uuid      | 工作空间对应 DBUUID| str | Yes | |
| table_name   | 表名 | str | no | |
| command         | sql命令行,没有指定命令，则查询全表 | str | no | |

示例：

	POST http://localhost:8080/ck/read/metrics
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 97
	Content-Type: application/json

	[{
		"db_uuid": "xxx",
		"table_name": "xxx",
		"command": "xxx",
		"group_by": ["xxx"]
	}]

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_45a47ea0-10bf-4691-b037-6fa5a75e67ec
	Date: Tue, 07 Apr 2020 09:25:11 GMT
	Content-Length: 40

	{
		"code": 200,
		"content": {
			"Series": [
				{
					"columns": [
						"name",
						"type",
						"default_type",
						"default_expression",
						"comment",
						"codec_expression",
						"ttl_expression"
					],
					"tags": {},
					"values": [
						[
							"tag",
							"String",
							"",
							"",
							"",
							"",
							""
						],
						[
							"time",
							"Int64",
							"",
							"",
							"",
							"",
							""
						]
					]
				}
			],
			"Message": ""
		},
		"errorCode": ""
	}


## `/v1/rewrite` | `POST`

重写 SQL

`fill_opt` 的几种可能：

- `null`: 填充空值
- `none`: 不填充
- `number`: 需带上具体值
- `previous`: 填充前一个值
- `linear`: 线性填充

示例：

	POST http://localhost:8080/v1/rewrite
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 594
	Content-Type: application/json
	
	"db_uuid": "ifdb_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	"sqls": [
		{	
			"dimensions": ["time(5m)"],
			"measurements": ["a"],
			"max_point": 360,
			"time_range":[],
			"sql": """
		        select a from b where 1>0 group by time(1m); -- bad select
		        select a from b where 1>0 limit 10;
		        select MEAN(a) from b where 1>0 group by tag_a;
		        select MEAN(a) from b where 1>0 group by tag_a, time(1m);

		        select LAST(a) from (select a from b where 1>0) where 1>0;
		        select x from (select a from b where 1>0) where 1>0 limit 10;

		        select x /* this is comment */ from (select a from b where 1>0) where 1>0; """,
			"default_aggr_func": "LAST",
			"conditions": "a>0 and b<0",
			"fill_opt": "number_fill",
			"fill_val": 0，
			"limit": 5,
			"offset": 10,
			"slimit": 1,
			"soffset": 2
		}
	]

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_c497cbbb-70ee-4b70-8121-9e49b8fafe40
	Date: Fri, 10 Apr 2020 13:02:07 GMT
	Content-Length: 963
	
	{
	    "code": 200, 
	    "content": [
	        [
	            {
	                "rewrite_results": {}, 
	                "sql": "SELECT a FROM b WHERE (1 > 0) AND (a > 0 AND b < 0) GROUP BY time(1m)"
	            }, 
	            {
	                "rewrite_results": {}, 
	                "sql": "SELECT a FROM b WHERE (1 > 0) AND (a > 0 AND b < 0) LIMIT 10"
	            }, 
	            {
	                "rewrite_results": {}, 
	                "sql": "SELECT mean(a) FROM b WHERE (1 > 0) AND (a > 0 AND b < 0) GROUP BY tag_a"
	            }, 
	            {
	                "rewrite_results": {
	                    "downsampled_to": 300000000000
	                }, 
	                "sql": "SELECT mean(a) FROM b WHERE (1 > 0) AND (a > 0 AND b < 0) GROUP BY tag_a, time(5m)"
	            }, 
	            {
	                "rewrite_results": {}, 
	                "sql": "SELECT last(a) FROM (SELECT LAST(a) FROM b WHERE (1 > 0) AND (a > 0 AND b < 0) GROUP BY time(5m)) WHERE 1 > 0"
	            }, 
	            {
	                "rewrite_results": {}, 
	                "sql": "SELECT x FROM (SELECT a FROM b WHERE 1 > 0) WHERE 1 > 0 LIMIT 10"
	            }, 
	            {
	                "rewrite_results": {}, 
	                "sql": "SELECT x FROM (SELECT LAST(a) FROM b WHERE (1 > 0) AND (a > 0 AND b < 0) GROUP BY time(5m)) WHERE 1 > 0"
	            }
	        ]
	    ], 
	    "errorCode": ""
	}

## `/v1/manage/metrics` | `POST`

	后台influxdb 时序数据操作

示例：

	POST http://0.0.0.0:9527/v1/manage/metrics
	{
        "db_uuid":"ifdb_2b3644c0b7494963bc59e0e45af9e95d",
        "command":"show measurements"
	}

## `/v1/influx/add_db` | `POST`

| 参数			      | 描述				  | 类型	 | required?  | 默认值   |
|----------------:|:---------------------:|---------:|:----------:|:--------:|
| default_rp      | 工作空间对应 DB 的默认 RP | str | Yes | |
| cqrp            | 工作空间对应 DB 上的 CQ RP(如果不需要设置 CQRP，传空字符串) | str | Yes | |
| rp_list         | 工作空间对应 DB 上的所有 RP 列表 | str | Yes | |

示例：

	POST http://0.0.0.0:9527/v1/influx/add_db?default_rp=rp3&cqrp=autogen&rp_list=rp0,rp1,rp2,rp3
    
	{
		"code":200,
			"content":{
				"db":"biz_ebc9ca8175294bf7ab549beb920f2f9b",
				"rp":"rp3",
				"uuid":"iflx_9aae00c3640a4e42922489664bd524f7",
				"dbUUID": "ifdb_xxxx"
			},
			"errorCode":""
	}

## `/v1/modify/dbrp` | `POST`

	修改db默认rp

示例：

	POST http://0.0.0.0:9527/v1/modify/dbrp
	{
        "db_uuid": "ifdb_070aa92fa6e54e9d863994a59c834733",
        "default_rp": "rp2",    #db 指定的默认rp
        "rp_list": ["rp1","rp2"]   #新增的rp list
	}

## `/v1/measurement/get_rp` | `POST`

获取某个 DB-UUID 上指定 measurements 的 RP、DB 以及保存时长信息。

示例：

	POST http://localhost:8080/v1/measurement/get_rp
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 123
	Content-Type: application/json
	
	{
		"measurements": [
			"meas-6",
			"mock_cpu",
			"dataway_self",
			"meas-not-exists"
		],
		"db": "ifdb_5708c9c73eb04c79a31861d4b5990ba4"
	}


	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_49463533-08fa-4e96-8434-15feb4ab6fb9
	Date: Tue, 07 Apr 2020 06:44:21 GMT
	Content-Length: 369
	
	{
	  "code": 200,
	  "content": {
	    "dataway_self": {
	      "db": "tanb_mig_test",
	      "rp": "",
	      "duration": "2160h"
	    },
	    "meas-not-exists": {   # 对于(暂时)不存在的 measurement (一般不可能)，后端仍然返回默认的 RP 信息
	      "db": "tanb_mig_test",
	      "rp": "",
	      "duration": "2160h"
	    },
	    "meas-6": {
	      "db": "tanb_mig_test",
	      "rp": "rp6",
	      "duration": "25920h"
	    },
	    "mock_cpu": {
	      "db": "tanb_mig_test",
	      "rp": "rp6",
	      "duration": "25920h"
	    }
	  },
	  "errorCode": ""
	}

如果 DB-UUID 不存在：

	{
		"code":400,
		"errorCode":"kodo.dbNotFound",
		"message":"db not found"
	}

获取指定工作空间的时间线(大概)数量

| 参数			      | 描述				  | 类型	 | required?  | 默认值   |
|----------------:|:---------------------:|---------:|:----------:|:--------:|
|                 | 工作空间对应 DB 的 UUID 列表 | str[] | Yes | |

## `/v1/tscnt` | `GET`

获取指定工作空间的时间线(大概)数量

| 参数			      | 描述				  | 类型	 | required?  | 默认值   |
|----------------:|:---------------------:|---------:|:----------:|:--------:|
|                 | 工作空间对应 DB 的 UUID 列表 | str[] | Yes | |

> 注意：由于 influxdb 本身的特性，这里统计的是 DB 最近 5 分钟的时间线。如果某个 DB 已经被删除(后面又重建，但未打入数据)，
但其统计数据不会更新。所以这里选取最近 5 分钟内的时间线统计

示例：

	GET http://localhost:8080/v1/tscnt
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 410
	Content-Type: application/json
	
	[ # 以下是指定工作空间对应的 DB-UUID
		"ifdb_1337970c3b2148c0a56ae10208499ccd",
		"ifdb_1548eed998314bf9a3beff4d68aa1c63",
		"ifdb_1779891eb1364f96a0f058758342e604",
		"ifdb_179bc0a8854e49c7815a399f26ee066e",
		"ifdb_1a0439a2699c454f9ab1827cb3d32283",
		"ifdb_1bc7ef27031944eda40711a55c7d1c84",
		"ifdb_1be3970954ab40f4b50c0cd51409838d",
		"ifdb_1cd9d46572df4f0f91a4dda741017e9a",
		"ifdb_20c8df9827bd4237a19b2b3bbfb6d3d8",
		"ifdb_245bfb992a87471e8b7e271d9d3d3dec"
	]

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_460ac74e-99ac-47a9-b284-d2b2ae1a8819
	Date: Fri, 03 Apr 2020 06:49:39 GMT
	Content-Length: 62

	{
		"code":200,
		"content": [0,1000,0,0,0,0,0,0,0,0],
		"errorCode":""
	}

## `/v1/cache/dataway/clean` | `POST`

清除一个 workspace token。需事先删除该 workspace 在数据库中的记录。不然 cache 删除无效。

| 参数			| 描述				  | 类型	 | required?  | 默认值   |
|----------------:|:---------------------:|---------:|:----------:|:--------:|
| `token` | 工作空间 token | str | Yes | |

示例：

	POST http://localhost:8080/v1/cache/dataway/clean?token=tkn_xxxxxxxxxxxxxxx
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 0

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_1cc52eb1-c662-4f09-b47a-b08b34e66611
	Date: Fri, 03 Apr 2020 03:09:04 GMT
	Content-Length: 40

	{
		"code":200,
		"errorCode":"",
		"message":""
	}

## `/v1/read/metrics` | `POST`

查询 influxDB 数据

	POST http://localhost:8080/v1/read/metrics
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 99
	Content-Type: application/json
	
	{
		"command": "select * from processes order by time limit 3",
		"db": "ifdb_75d2c0a6018a492ba5e7ec0e31a3eabe"    # DB UUID
	}

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json
	X-Trace-Id: trace_02941b6c-81c0-4953-b4ad-5d3d12adcde6
	Date: Tue, 07 Apr 2020 09:16:35 GMT
	Content-Length: 445

	{
	  "code": 200,
	  "content": {
	    "error": "",
	    "results": [
	      {
	        "statement_id": 0,
	        "Series": [       # 注意：这里是大写的 `S'
	          {
	            "name": "processes",
	            "columns": [
	              "time", "blocked", "dead", "host", "idle",
	              "paging", "running", "sleeping", "stopped", "total",
	              "total_threads", "unknown", "zombies"
								],
	            "values": [
	              [ 1586014431000, 0, 0, "ubuntu-lwc", 46, 0, 0, 69, 0, 115, 212, 0, 0 ],
	              [ 1586014441000, 0, 0, "ubuntu-lwc", 51, 0, 0, 63, 0, 114, 205, 0, 0 ],
	              [ 1586014451000, 0, 0, "ubuntu-lwc", 51, 0, 0, 63, 0, 114, 205, 0, 0 ]
	            ]
	          }
	        ],
	        "Messages": null  # 注意：这里是大写的 `M'
	      }
	    ]
	  },
	  "errorCode": ""
	}


## `/v1/drop/metrics` | `POST`

移除一个指标集

	POST http://localhost:8080/v1/drop/metrics
	Connection: keep-alive
	Accept-Encoding: gzip, deflate
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 97
	Content-Type: application/json
	
	{
		"measurement": "xxxx",
		"db_uuid": "ifdb_5708c9c73eb04c79a31861d4b5990ba4"
	}

	HTTP/1.1 200 OK
	Access-Control-Allow-Credentials: true
	Access-Control-Allow-Headers: Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Token, X-DatakitUUID, X-RP, X-Precision
	Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT
	Access-Control-Allow-Origin: *
	Content-Type: application/json

	{
		"code":200,
		"errorCode":"",
		"message":""
	}

## `/v1/write` | `POST` 或 `/v1/write/metrics` | `POST`

行协议数据上传至 Kodo，需要有 API 签名

示例:

	POST /v1/write/metrics HTTP/1.1
	Host: kodo.cloudcare.com:9527
	User-Agent: Go-http-client/1.1
	Authorization: DATAFLUX tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxx:7olWKdgPuD78HF5p3pgCxoN+Ces=
	Content-Length: 1838
	Content-Encoding: gzip
	Date: Wed, 12 Dec 2018 06:33:11 GMT
	Content-MD5: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Token: tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Version: 0.1-389-gb778f18
	X-Agent-Uid: agnt_xxxxxxxxxxxxxxxxxxxxxxxxxxxx

	<行协议 body>

	HTTP/1.1 200 OK
	Date: Wed, 26 Dec 2018 06:12:29 GMT
	Content-Length: 79
	Content-Type: application/json

	{
		"code": 200,
		"content": null
	}

字段说明:

- `Authorization`: 签名
- `X-Version`: dataway 版本
- `X-Agent-Uid`: dataway UUID
- `X-Token`: dataway Token，决定将数据打到哪个工作空间

其中, 如下几个字段是必须的:

- `Authorization`
- `Date`
- `X-Token`

## `/v1/ck/read/metrics` | `POST`

TODO

## `/v1/write/object` | `POST`

单独的 object 上传接口，需要有 API 签名

object 数据通过 NSQ 写入ES, 不写influxdb

示例：
	
	POST /v1/write/object HTTP/1.1
	Host: kodo.cloudcare.com:9527
	User-Agent: Go-http-client/1.1
	Authorization: DATAFLUX tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxx:7olWKdgPuD78HF5p3pgCxoN+Ces=
	Content-Length: 1838
	Content-Encoding: gzip
	Date: Wed, 12 Dec 2018 06:33:11 GMT
	Content-MD5: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Token: tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Version: 0.1-389-gb778f18
	X-Agent-Uid: agnt_xxxxxxxxxxxxxxxxxxxxxxxxxxxx

	[{
		"__name" : "必填",
		"__tags": {
			"a": "b",
			"__class": "必填",
			"c": "d",        
			"any-user-defined-tags": "..."    
		},    
		"__content": "..."
	}]

	HTTP/1.1 200 OK
	Date: Wed, 26 Dec 2018 06:12:29 GMT
	Content-Length: 79
	Content-Type: application/json

	{
		"code": 200,
		"content": null
	}

## `/v1/write/keyevent` | `POST`

单独的 keyevent 上传接口，需要有 API 签名

keyevent 数据通过 NSQ 写入ES, 不写influxdb

示例：

	POST /v1/write/keyevent HTTP/1.1
	Host: kodo.cloudcare.com:9527
	User-Agent: Go-http-client/1.1
	Authorization: DATAFLUX tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxx:7olWKdgPuD78HF5p3pgCxoN+Ces=
	Content-Length: 1838
	Content-Encoding: gzip
	Date: Wed, 12 Dec 2018 06:33:11 GMT
	Content-MD5: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Token: tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Version: 0.1-389-gb778f18
	X-Agent-Uid: agnt_xxxxxxxxxxxxxxxxxxxxxxxxxxxx

	<行协议 body>

	HTTP/1.1 200 OK
	Date: Wed, 26 Dec 2018 06:12:29 GMT
	Content-Length: 79
	Content-Type: application/json

	{
		"code": 200,
		"content": null
	}

## `/v1/write/logging` | `POST`

单独的 log 上传接口，需要有 API 签名

log 数据通过 NSQ 写入ES, 不写influxdb

示例：

	POST /v1/write/logging?source=xxxxx HTTP/1.1
	Host: kodo.cloudcare.com:9527
	User-Agent: Go-http-client/1.1
	Authorization: DATAFLUX tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxx:7olWKdgPuD78HF5p3pgCxoN+Ces=
	Content-Length: 1838
	Content-Encoding: gzip
	Date: Wed, 12 Dec 2018 06:33:11 GMT
	Content-MD5: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Token: tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	X-Version: 0.1-389-gb778f18
	X-Agent-UID: agnt_xxxxxxxxxxxxxxxxxxxxxxxxxxxx

	<行协议 body>

	HTTP/1.1 200 OK
	Date: Wed, 26 Dec 2018 06:12:29 GMT
	Content-Length: 79
	Content-Type: application/json

	{
		"code": 200,
		"content": null
	}


## `/v1/cache/logging/set` | `POST`

ck 读数据

| 参数			      | 描述				  | 类型	 | required?  | 默认值   |
|----------------:|:---------------------:|---------:|:----------:|:--------:|
| workspace_uuid      | 工作空间对应 DBUUID| str | Yes | |
| source   | 来源 | str | Yes | |
| url         | func的url路径 | str | Yes | |

示例：

	POST http://localhost:8080/v1/cache/logging/set
	Connection: keep-alive
	Accept: */*
	User-Agent: python-requests/2.22.0
	Content-Length: 97
	Content-Type: application/json

	{
		"workspace_uuid": "xxx",
		"source": "xxx",
		"url": "xxx"
	}

	HTTP/1.1 200 OK
	Date: Mon,01 June 2020 15:12:29 GMT
	Content-Length: 79
	Content-Type: application/json

	{
		"code": 200,
		"content": null
	}

## `/v1/check/token/:token` | `GET`

检查token是否有效

示例：

	GET http://localhost:8080/v1/check/token/tkn-xxxx

	HTTP/1.1 200 OK
	Date: Mon,01 June 2020 15:12:29 GMT
	Content-Length: 79
	Content-Type: application/json

	{
	    "code": 200,
		"errorCode": "",
    	"message": ""
	} 

	{
		"code": 403,
		"errorCode": "kodo.tokenNotFound",
		"message": "token not found"
	}



## `/v1/droppingmetrics/:db_uuid` | `GET`

查询指定DB正在删除的指标集

示例：

	GET http://localhost:8080/v1/droppingmetrics/ifdb_xxxxx

	HTTP/1.1 200 OK
	Date: Mon,01 June 2020 15:12:29 GMT

	{
		"code":200,
		"content": ["xxxx1","xxxx2"],
		"errorCode":""
	}	