# 网络拨测功能定义

全局定义：

- 所有拨测产生的数据，都以行协议方式，通过 `/v1/write/logging` 接口存为日志数据

## HTTP 拨测任务定义

```python
{
	"url": "http://example.com/some/api", 

	# 拨测数据的存放地址，对 SAAS 而言，是 openway.dataflux.cn
	# 对 PAAS 而言，需要一个单独的公网可访问的 Dataway
	"post_url": "https://dataway.cn",

	"name": "give your test a name",
	"tags": {
		"tag1": "val1",
		"tag2": "val2"
	},

	"frequency": "1m",   # 1min ~ 1 week

	# 区域可多选
	"locations": ["hangzhou", "beijing", "chengdu"],
	
	"advance_options": {
		"request_options": {
			"follow_redirect": true,
			"headers": {
				"header1": "value1",
				"header2": "value2"
			},
			"cookies": "",
			"auth": {
				"username": "",
				"password": ""
			}
		},
		"request_body": { # 以下几个类型只能单选
			"body_type": "text/plain|application/json|text/xml|None",
			"body": ""
		},
		"certificate": {
			"ignore_server_certificate_error": false,
			"private_key": "",
			"certificate": ""
		},
		"proxy": {
			"url": "",
			"headers": {
				"header1": "value1"
			}
		}
	},

	"success_when":  [ 
		{

			# body|header|response_time|status_code 都是单个判定条件，它们各自出现的时候，表示 AND 的关系

			"body":{},
			"header": {
				"header-name":{ # 以下几个条件只能单选
					"contains":"",
					"does_not_contains": "",
					"is": "",
					"is_not": "",
					"match_regex": "",
					"does_not_match_regex": ""
				},
				"another-header-name": "...",
			},
			"response_time": {
				"less_than": "100ms"
			},
			"status_code": { # 以下几个条件只能单选
				"is": 200,
				"is_not": 400,
				"match_regex": "ok*",
				"not_match_regex": "*bad"
			}
		},
		
		{
			"AND_another_assert": "..."
		}
	]
}
```

### HTTP 拨测行协议定义

```python
{
	"measurement": "http_dial_testing",
	"tags": {
		"name": "",
		"url": "",
		"用户额外指定的各种": "tags",

		# 每个具体的 datakit 只会在一个 location，故这里只有单个值
		"location": "", 
	},

	"fields": {
		"success": true/false,
		# 如果 success == false
		"fail_reason": "字符串描述失败原因"
	},
	"time": time.Now()
}
```

注意，这里的 `fail_reason` 要描述 `body/header/response_time/status_code` 各自失败的原因。如果可以，所有原因都描述一遍，如 `response time larger than 100ms; status code match regex 4*`

## TCP 拨测任务定义

```python
{
	"host": "www.example.com",
	"port": "443",
	"name": "give your test a name",

	# 拨测数据的存放地址，对 SAAS 而言，是 openway.dataflux.cn
	# 对 PAAS 而言，需要一个单独的公网可访问的 Dataway
	"post_url": "https://dataway.cn",

	"tags": {
		"tag1": "val1",
		"tag2": "val2"
	},

	"frequency": "1m",   # 1min ~ 1 week
	"locations": ["hangzhou", "beijing", "chengdu"],

	"success_when":  [
		{
			"response_time": {
				"less_than": "100ms"
			}
		}
	]
}
```

### TCP 拨测行协议定义

```python
{
	"measurement": "tcp_dial_testing",
	"tags": {
		"name": "",
		"host": "",
		"port": "",
		"用户额外指定的各种": "tags",

		# 每个具体的 datakit 只会在一个 location，故这里只有单个值
		"location": "", 
	},

	"fields": {
		"success": true/false,
		# 如果 success == false
		"fail_reason": "字符串描述失败原因"
	},
	"time": time.Now()
}
```

## DNS 拨测任务定义

```python
{
	"domain": "www.example.com",
	"dns_server": "",
	"name": "give your test a name",

	# 拨测数据的存放地址，对 SAAS 而言，是 openway.dataflux.cn
	# 对 PAAS 而言，需要一个单独的公网可访问的 Dataway
	"post_url": "https://dataway.cn",

	"tags": {
		"tag1": "val1",
		"tag2": "val2"
	},

	"frequency": "1m",   # 1min ~ 1 week
	"locations": ["hangzhou", "beijing", "chengdu"],

	"success_when":  [
		{
			"response_time": {
				"less_than": "100ms"
			},
			"at_least_one_record": {
				"of_type_a": {
					"is": "",
					"contains": "",
					"match_regex": "",
					"not_match_regex": ""
				},
				"of_type_aaaa": {},
				"of_type_cname": {},
				"of_type_mx": {},
				"of_type_txt": {}
			},
			"every_record": {}
		},
		{
			"AND_another_assert": "..."
		}
	]
}
```

关于 DNS 的各种 [`of_type_xxx`](https://support.dnsimple.com/categories/dns/)

### DNS 拨测行协议定义

```python
{
	"measurement": "dns_dial_testing",
	"tags": {
		"name": "",
		"domain": "",
		"dns_server": "",
		"用户额外指定的各种": "tags",

		# 每个具体的 datakit 只会在一个 location，故这里只有单个值
		"location": "", 
	},

	"fields": {
		"success": true/false,
		# 如果 success == false
		"fail_reason": "字符串描述失败原因"
	},
	"time": time.Now()
}
```

## 架构设计

用户在 Studio 中添加拨测任务后，具体的拨测任务以一条记录的方式存入 MySQL 中，MySQL 基础字段

```sql
CREATE TABLE IF NOT EXITS net_dial_testing (
		`id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
		`uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 dialt_',

		`location` varchar(48) NOT NULL COMMENT '拨测任务部署区域, 多个 location 之间逗号分割'

		`workspaceUUID` varchar(48) NOT NULL DEFAULT '' COMMENT '工作空间UUID',

		`status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',

		`creator` varchar(64) NOT NULL DEFAULT '' COMMENT '创建者 account-id',
		`updator` varchar(64) NOT NULL DEFAULT '' COMMENT '更新者 account-id',

		`createAt` int(11) NOT NULL DEFAULT '-1',
		`deleteAt` int(11) NOT NULL DEFAULT '-1',
		`updateAt` int(11) NOT NULL DEFAULT '-1',

		`task` text NOT NULL COMMENT '任务的 json 描述'
		PRIMARY KEY (`id`),
		UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
		KEY `k_ws_uuid` (`workspaceUUID`),
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 整体架构设计

![拨测任务整体架构](net-dial-testing-arch.png)

### DataKit 实现

1. 开启对应拨测采集器，其 conf 如下

```python
[[inputs.net_dial_testing]]
	location = "hangzhou"
	server = "dial.dataflux.cn"

	[[inputs.net_dial_testing.tags]]
	# 各种可能的 tag
```
<!--
2. 调用 server 上的 `/tasks?location=<xxx>` 接口，拿本区域所有的拨测任务
	- 循环开启拿到的所有任务，单个任务常驻运行
		- 如果某个任务已经在运行，跳过
		- 如果某个任务在运行，但任务有更新（本地缓存的任务跟最新拿到的任务，更新之间不一致），杀掉之前的任务（任务 ID 相同），重新运行

3. Sleep 一段时间，回到步骤 2 -->

由于但各个任务都比较简单，经初步测试，单个 DataKit 上运行上万个任务，基本没问题。
