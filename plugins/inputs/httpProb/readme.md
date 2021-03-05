### 简介
httpProb采集用于接受其它http流量镜像工具复制的流量数据，对请求的流量数据通过pipeline脚本进行业务分析，支持以下流量镜像
- Nginx mirror (v1.13.4以上)
- goreplay

### mirror配置

#### nginx
通过nginx代理方式

- nginx mirror参考配置
```
upstream backend {
    server 127.0.0.1:8080;
}

upstream mirror {
    server 127.0.0.1:9090;
}

server {
    server_name www.abc.com;
    listen 80;

    location / {
        mirror /mirror;
        proxy_pass http://backend;
    }

    location = /mirror {
        internal;
        proxy_pass http://mirror/plob$request_uri;
    }
}
```

#### goreplay
![goreplay介绍](https://github.com/buger/goreplay)

```
sudo ./gor --input-raw :80 --output-http="http://127.0.0.1:9090"
```

### 采集器配置
```
[[inputs.httpProb]]
	drop_body = false
	# log source(required)
	source = "xxx-app"

    # global tags
    [inputs.httpProb.tags]
    # tag1 = val1
    # tag2 = val2

    [[inputs.httpProb.url]]
    # uri or uri_regex
    # uri = "/"         # regist all routes
    # uri_regex = "/*"
    # pipeline = "all_route.p" # datakit/pipeline/all_route.p

	[[inputs.httpProb.url]]
    # uri = "/user/info"
    # uri_regex = "/user/info/*"
    # pipeline = "user_info.p" # datakit/pipeline/user_info.p
```

### pipeline处理数据结构
```
{
	"method": "string",
	"uri": "string",
	"query": {
			"key": "value",
		},
	"header": {
			"key": "value",
		},
	"body": "string"
}
```

