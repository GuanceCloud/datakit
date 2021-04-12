{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# 简介

该采集器是网络拨测结果数据采集，所有拨测产生的数据，都以行协议方式，通过 `/v1/write/logging` 接口,上报DataFlux平台

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```python
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## json 文件示例 :  http 任务  

```python
{
	* "url": "http://example.com/some/api",
	* "method": "POST",
	* "external_id": "外部系统中给该任务定义的 ID",

	# 拨测数据的存放地址，对 SAAS 而言，是 openway.dataflux.cn
	# 对 PAAS 而言，需要一个单独的公网可访问的 Dataway。这里的 token
  # 对 Dataflux SAAS/PASSS 而言，实际上隐含了工作空间信息
	* "post_url": "https://dataway.cn?token=tkn_xxx",

	# 任务状态(OK/stopped)
	"status": "OK", 

	"name": "任务命名",
	"tags": {
		"tag1": "val1",
		"tag2": "val2"
	},

	* "frequency": "1m",   # 1min ~ 1 week

	# 区域：冗余字段，便于调试
	* "regions": "beijing",
	
	"advance_options": [{
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
		"request_body": { 
			"body_type": "text/plain|application/json|text/xml|None", # 以下几个类型只能单选 或者为空
			"body": ""
		},
		"certificate": {
			"ignore_server_certificate_error": false,
			"private_key": "",
			"certificate": "",
			"ca": ""
		},
		"proxy": {
			"url": "",
			"headers": {
				"header1": "value1"
			}
		}
	}],

	* "success_when":  [ 
		{

			# body|header|response_time|status_code 都是单个判定条件，它们各自出现的时候，表示 AND 的关系
			"body":{},
			"header": {
				"header-name":{ # 以下几个条件只能单选
					"contains":"",
					"not_contains": "",
					"is": "",
					"is_not": "",
					"match_regex": "",
					"not_match_regex": ""
				}
			},
			"response_time":  "100ms",
			"status_code": { # 以下几个条件只能单选
				"is": "200",
				"is_not": "400",
				"match_regex": "ok*",
				"not_match_regex": "*bad"
			}
		}
	]
}

```

##  HTTP 拨测结果数据结构定义

```python
{
    "class": "http_dial_testing",
   
    "name": "",
    "url": "",
    "用户额外指定的各种": "tags",

    # 每个具体的 datakit 只会在一个 region，故这里只有单个值
    "region": "",

    "status": "OK", # 只有 OK/FAIL 两种状态，便于分组以及过滤查找

    # HTTP 协议版本，HTTP/1.0 HTTP/1.1 等
    
    "response_dns":  1000 #DNS解析时间,单位 us
    
    "response_connection": 1000#连接时间（TCP连接）,单位 us
    
    "response_ssl":  1000#SSL连接时间,单位 us
    
    "response_ttfb": 1000 #首次回包时间（请求响应时间）,单位 us
    
    "response_download": 1000 #下载时间,单位 us
    
    "status_code_class": "2xx",
    "status_code_string": "OK"
    "status_code": "200"

    # HTTP 协议版本，HTTP/1.0 HTTP/1.1 等
    "proto": "HTTP/1.0"
   
    # 如失败，如实描述。如成功，无此指标
    "fail_reason": "字符串描述失败原因"

    # HTTP 相应时间, 单位 us
    "response_time": 300,

    # 返回 body 长度，单位字节。如无 body，则无此指标，或者填 0
    "response_body_size": 1024,

    # 只有 1/-1 两种状态, 1 表示成功, -1 表示失败, 便于 UI 绘制状态跃迁图（TODO: 现在前端图标支持负数么）
    "success": 1, 

    ### 在指标数据的基础上，增加如下数据
    "message": "便于做全文: 包括请求头(request_header)/请求体(request_body)/返回头(response_header)/返回体(response_body)/fail_reason 冗余一份",
    
    "fail_reason": "拨测失败原因(此处无需做全文)"

    # 其它指标可再增加...

    "date": time.Now()
}
```
