# Dataway LuaScirpt Usage Doc

---

#### 目录

- [前言](#前言)
	- [global_lua](#global_lua)
	- [route_lua](#route_lua)
- [使用方式](#使用方式)
- [内置功能函数](#内置功能函数)
	- [sql_connect](#sql_connect)
	- [redis_connect](#redis_connect)
	- [mongo_connect](#mongo_connect)
	- [http_request](#http_requestt)
	- [json_decode](#json_decode)
	- [xml_decode](#xml_decode)
	- [csv_decode](#csv_decode)
	- [cache](#cache)
	- [regex](#regex)
	- [crypto](#crypto)
	- [log](#log)
- [实用示例](#实用示例)

---

## 前言

为方便用户对采集到的数据进行控制，引入执行 lua scprit 的功能。lua 程序的调用方式有两种—— global_lua 和 route_lua。

**注意，lua 代码中仅支持引入标准库，和提供的内置函数，暂不支持使用第三方库。**

#### global_lua

  - 全局 lua 程序，独立于所有功能之外，定时执行该 script 的内容，多用来更新全局 cache

  - 配置文件 global_lua 项，path 为 lua 文件路径，circle 为当前文件的执行周期，例如：

        ``` yaml
        global_lua:
                - path: XXX.lua
                  circle: 0 * * * 0,6

                - path: XXX.lua
                  circle: 0 * * * 0,6
        ```

   - circle 使用**Unix cron**格式，支持“秒、分、时、天、月”，格式简述如下表，详细语法可自行查阅：
	
| Field name   | Mandatory? | Allowed values  | Example to     | 
| ------------ | ---------- | --------------- | -------------- |
| Seconds      | Yes        | 0-59            | 31  *  *  *  * |
| Minutes      | Yes        | 0-59            |  0 25  *  *  * |
| Hours        | Yes        | 0-23            |  0  0 17  *  * |
| Day of month | Yes        | 1-31            |  0  0  0 12  * |
| Month        | Yes        | 1-12 or JAN-DEC |  0  0  0  0  9 |

> 例如：`*/1 * * * *`为每秒执行一次。

#### route_lua

  - 路由级别的 lua 程序，每个路由配置 lua script 文件列表，对数据依次对其执行所有 script，将最终结果上报

  - 任意一组执行出错，将不再执行下去，并丢弃此次数据内容不再上报，对执行错误写入日志

  - script 指定路径： dataway.yaml -- routes_config，例如

	``` yaml
	# route config
	routes_config:
		- name: default
		  lua :
			- path: default01.lua
			- path: default02.lua
			- path: default03.lua
	
		- name: test
		  lua :
			- path: test01.lua
			- path: test02.lua
	```

---

## 使用方式

在配置文件中指定 lua script 文件路径后，该文件内容格式为：

``` lua
function handle(points)

	-- lua code

	return points
end
```

其中：

- `function handle(points)`为整个 script 的执行入口，所有的 lua 代码都在该函数内部，`points`传入的数据参数

- `points`类型为 table，类似于c语言系的`struct[]`，其中 strcut 的各个成员变量分别是是`name, tags, fields, time`，详细介绍如下：

	- `name` string，本组数据的存储表名，一般不会改变
	- `tags` table，标签组，该 table 中只允许存放 string
	- `fields` table，数据组，可以存放 string/number/bool 类型的数据
	- `time` 本条记录的时间戳，类型 number，一般不会改变

- `points`类型为数据所用类型，handle 函数传入 points，函数结束返回 points，points 格式固定，但支持用户自定义其中的字段内容，这也是 lua script 的核心功能

- `points`示例

``` lua
points
{
	{
		name = "t_name",
		tags = {
			t1 = "tags_01",
			t2 = "tags_02"
		},
		fields = {
			f1 = "fields_01",
			f2 = 12345,
			f3 = true
		},
		time = 1575970133841355008 
	}，

	{
		name = "t_name",
		tags = {
			t1 = "tags_03",
			t2 = "tags_04"
		},
		fields = {
			f1 = "fields_02",
			f2 = 66666,
			f3 = false
		},
		time = 1575970133841355008 
	}
}
```

- 在函数末尾依照 lua 语法添加`end`

---

## 内置功能函数

内置了常用函数，包括 HTTP、sql、cache，以及各种数据的解析等。

#### http_request

函数签名：
``` lua
response_table, error_string = http_request(method_string, URL_string, headers_table {
									query_string,
									heaers_table,
									body_string
								})
```

*query, headers, body如果不存在可以不写*

示例：

``` lua
function handle(points)

	print("http test:")

	response, error = http_request("GET", "http://www.baidu.com", {
		query="",
		headers={
			Accept="*/*"
		},
		body="test123"
	})
	
	if err == nil then
		print(response.body)
		print(response.status_code)
	else
		print(err)
	end

	return points
end
```

函数有2个返回值，response 和 error，如果函数执行错误，则 response 为 nil，error 为string 类型的错误说明（后续内置函数基本都沿用此方案）。

response 是 tablele 类型，各个字段名和类型分别为：
- headers		(table)
- cookies		(table)
- status_code	(number)
- url			(string)
- body			(string)
- body_size		(number)


#### sql_connect

sql 采用“数据库+连接信息”的方式进行连接，现支持的数据库和对应的信息模板如下：

| database name | connect infomation format                                             | 
| --------------| --------------------------------------------------------------------- |
| mysql         | username:password@protocol(address)/dbname?param=value                |
| postgres      | postgres://user:password@localhost/dbname?sslmode=disable&param=value |
| mssql         | sqlserver://username:password@host:port?database=dbname&param=value   |

sql_connect 接收的类型为连接数据库所用的 table，返回一条连接和错误信息。该连接含有一个成员函数 query，参数为查询语句，支持占位符操作。

**注：该连接使用结束后，需调用成员函数`close()`关闭该连接，否则会造成资源泄露**

示例：

``` lua
function handle(points)
	print("mysql test:")
	conn, err = sql_connect("mysql", "root:123456@tcp(192.168.0.2:3306)/test_mysql?charset=utf8")
	if err ~= "" then
		res, err = conn:query('SELECT * FROM students where year>?;', 1986)
		if res ~= nil then
			for _, row in pairs(res) do
				for k, v in pairs(row) do
					print(k, v)
				end
			end
		else
			print(err)
		end
	else
		print(err)
	end
	conn:close()

	return points
end
```

#### redis_connect

redis_connect 接收的类型为连接数据库所用的 table，参数分别为：

```
- host		服务所在 IP
- port		服务监听的端口
- passwd	密码
- index		redis数据库索引，dbindex
```

redis_connect 仅返回一条连接句柄，不返回错误信息。

该连接含有一个成员函数 docmd，第一个参数为命令符，如`get`、`set`、`keys`等，后续参数为命令符所需的参数，详情看示例。

redis 具体命令可查看[官方文档](https://redis.io/commands)。

**注：该连接使用结束后，需调用成员函数`close()`关闭该连接，否则会造成资源泄露**

示例：
``` lua
function handle(points)
	print("redis test")

	conn = redis_connect({host="10.100.64.106", port=26379, passwd="", index=0})
	print(conn:docmd("set", "luakey", "luavalue"))
	res, err = conn:docmd("keys", "wallet*")
	if err == nil then
		for k, v in ipairs(res) do
			print(k, v)
		end
	else 
		print(err)			
	end
	print(conn:docmd("get", "luakey"))
	conn:close()

	return points
end
```

#### mongo_connect

mongo_connect 使用形参为[标准URI连接语法](https://docs.mongodb.com/manual/reference/connection-string/)，

格式：`mongodb://[username:password@]host1[:port1][,host2[:port2],...[,hostN[:portN]]][/[database][?options]]`

注意:
> 如果 MongoDB 是运行在 Docker 中，并且开启了 Replica Set 副本集，需要在 URI 中添加参数 `connect=direct`，例如连接到对应数据库服务器：`mongodb://user123:password123@192.168.0.100:27071/?connect=direct`
> usename 和 password 中如果含有`@`等字符，需要进行**百分比编码**，编码表查看[文档](https://developer.mozilla.org/zh-CN/docs/Glossary/percent-encoding)

mongo_connect 如果连接失败，则返回的第一个参数为`nil`，第二个参数为错误信息。

连接成功返回一条连接句柄，包含成员函数`query`和`close`，其中`query`为查询函数，函数签名为`result_table, err = query(dbname_string, collection_string, select_table)`

函数的三个参数依次为查询的数据库，数据集合，查询条件。查询条件为 table 类型，key 为 string，value 支持常用的整型、浮点型、布尔类型、字符串，**暂不支持以二进制数据为查询条件**。

query() 查询结果是 record 风格的 table 和 err 错误信息。table 可以使用访问成员变量的方式进行操作，例如`tb.name`，不能使用 for 遍历，具体见示例代码。

**注：该连接使用结束后，需调用成员函数`close()`关闭该连接，以免资源泄露**

以上仅支持 MongoDB 2.6 及更高版本。

``` lua
function handle(points)
	print("mongo test")

	conn, err  = mongo_connect("mongodb://10.100.64.106:27017")
	if err == nil then 
		res, err = conn:query("testdb","testcollect", { name="cloud" })
		if err == nil then
			print(res._id)
			print(res.name)
		else 
			print(err)
		end
		conn:close()
	else
		print(err)
	end

	return points
end
```



#### json_decode

json_decode 解析字符串为 table。
json_encode 将 table 或 string 转为 json 格式的字符串。

示例：

``` lua
function handle(points)
	json_str = '{ "hostname":"ubuntu18.04LTS", "date":"2019年12月10日 星期二 11时14分47秒 CST", "ip":["127.0.0.1","192.168.0.1","172.16.0.1"] }'

	print("json_str:", json_str)
	json_table = json_decode(json_str)

	print(json_table)

	for k, v in pairs(json_table) do
		print(k, v)
	end
	for _, v in pairs(json_table["ip"]) do
		print(, v)
	
	return points
end
```

#### xml_decode

xml_decode 将 xml 格式字符串解析为 table。

示例：

``` lua
function handle(points)
	xml_str ="<booklist><book>100</book><book>100.5</book><book>200</book></booklist>"
	print("xml test:", xml_str)
	xml_table = xml_decode(xml_str)

	for _, row in pairs(xml_table) do
		for k, v in pairs(row) do
			for _, vv in pairs(v) do
				print(vv)
				print("number add 1", vv+1)
			end
		end
		print("----------------")
	end

	return points
end
```
数值类型的解析优先级高于字符型，示例代码中 100、100.5、200 都解析成 Number 类型，可以进行数值操作（例如加1）。


#### csv_decode

csv_decode 将 csv 格式字符串解析为 N 维 table，N 为 csv 行数减1。table 的 key 为首行 header 对应。

示例：

``` lua
function handle(points)
	csv_str = "name,year,address\nAA,22, NewYork\nBB, 21, Seattle"
	print("csv test:", csv_str)
	csv_table = csv_decode(csv_str)

	for _, row in pairs(csv_table) do
		for k, v in pairs(row) do
			print(k, v)
		end
		print("----------------")
	end

	return points
end
```
标准 csv 格式，分隔符为英文状态下逗号“,”


#### cache

KV 缓存，key 为 string 类型，value 为 string/number/table 任意类型

- cache_set 缓存写入
- cache_get 缓存读取，如果没有找到对应的 key，则返回 nil
- cache_list 缓存队列中所有的 key 值，类型为 table，可使用`#list`或`table.getn(list)`得到该 list 的长度

示例：

``` lua
function handle(points)

	cache_set("AAA", "hello,world")
	cache_set("BBB", 123456)
	cache_set("CCC", true)
	cache_set("DDD", { host = '10.100.64.106', port = 13306, database = 'golua_mysql', user = 'root', password = '123456' })

	list = cache_list()
	print("cache_list:")
	for k,v in pairs(list) do 
		print(k, v)
	end
	print("cache key list: ", #list)
	print("cache key list: ", table.getn(list))

	print("AAA: ", cache_get("AAA"))
	print("BBB: ", cache_get("BBB"))
	print("CCC: ", cache_get("CCC"))

	print("DDD --------")
	dd = cache_get("DDD")
	for k,v in pairs(dd) do
		print(k, v)
	end

	return points
end
```

#### regex

- re_quote  转义
- re_find   查找
- re_gsub   替换
- re_match  匹配

示例：

``` lua
function handle(points)
	-- quote ----------

	print(re_quote("^$.?a"))
	-- "\^\$\.\?a"

	-- find ----------

	print(re_find('', ''))
	-- 1   0

	print(re_find("abcd efgh ijk", "cd e", 1, true))
	-- 3   6

	print(re_find("abcd efgh ijk", "cd e", -1, true))
	-- nil

	print(re_find("abcd efgh ijk", "i([jk])"))
	-- 11  12  "j"

	-- gsub ----------

	print(re_gsub("hello world", [[(\w+)]], "${1} ${1}"))
	-- "hello hello world world"  2

	print(re_gsub("name version", [[\w+]], {name="lua", version="5.1"}))
	-- "lua-5.1.tar.gz"  2

	print(re_gsub("name version", [[\w+]], {name="lua", version="5.1"}))
	-- "lua 5.1"  2

	print(re_gsub("$ world", "\\w+", string.upper))
	-- "$ WORLD"  1
	
	print(re_gsub("4+5 = $return 4+5$", "\\$(.*)\\$", function (s)
			return loadstring(s)()
		end))
	-- "4+5 = 9"  1

	-- match ----------

	print(re_match("$$$ hello", "z"))
	-- nil

	print(re_match("$$$ hello", "\\w+"))
	-- "hello"

	print(re_match("hello world", "\\w+", 6))
	-- "world"

	print(re_match("hello world", "\\w+", -5))
	-- "world"

	print(re_match("from=world", "(\\w+)=(\\w+)"))
	-- "from" "world"

	return points
end
```

#### crypto

加密解密以及base64转换

fucntion signature:

- `base64_encode(data_str)`
- `base64_decode(data_str)`
- `hex_encode(data_str)`
- `hmac(mothod_str, data_str, key_str, notHexString_bool)`

> mothod 为方法枚举 { "md5", "sha1", "sha256", "sha512" }
> 最后一个参数`notHexString`是bool类型，含义是**不进行编码为字符串操作**，默认是`flase`，即`对输入或输出进行编码为字符串的操作`。

- `encrypt(data_str, mothod_str, key_str, iv_str, notHexString)`

> mothod 为方法枚举 { "aes_cbc", "des_cbc", "des_ecb" }
> 当方法为`aes_cbc`时，key 的长度可以是 16,24,32，iv 向量长度为 16
> 当方法为`des_cbc`或`des_ecb`时，key 的长度可以是 8，iv 向量长度为 8
> 第三个形参和`hmac()`的 notHexString 意义相同。

- `decrypt(data_str, mothod_str, key_str, iv_str, notHexString) `
> 同上


示例：

``` lua
function handle(points)
	print("----------- base64 ------------")
	b64 = base64_encode("hello,world!")
	print(b64)
	print(base64_decode(b64))

	print("---------- hex encode -----------")
	print(hex_encode("Hello"))

	print("----------- crc32 ------------")
	print(crc32("hello,world!"))

	print("----------- hmac ------------")
	print(hmac("md5", "hello", "world"))
	print(hmac("sha1", "hello", "world", true))
	print(hmac("sha256", "hello", "world"))
	print(hmac("sha512", "hello", "world", true))

	print("----------- encrypt/decrypt ------------")

	-- aes-cbc  key length 16,24,32, iv length 16
	aescbc = encrypt("hello", "aes-cbc", "1234abcd1234abcd", "iviv12345678abcd", true)
	print(aescbc)
	print(decrypt(aescbc, "aes-cbc", "1234abcd1234abcd", "iviv12345678abcd", true))

	-- des-cbc key length 8, iv length 8
	descbc = encrypt("hello", "des-cbc", "1234abcd", "iviv1234")
	print(descbc)
	print(decrypt(descbc, "des-cbc", "1234abcd", "iviv1234"))

	-- des-cbc key length 8, iv length 8
	desecb = encrypt("hello", "des-ecb", "1234abcd", "iviv1234")
	print(desecb)
	print(decrypt(descbc, "des-ecb", "1234abcd", "iviv1234"))

	return points
end

```

---

## 实用示例

使用 lua script 下载阿里云 OSS 数据，[官方 OSS 验证文档](https://help.aliyun.com/document_detail/31951.html?spm=a2c4g.11174283.6.1486.61aa7da2V4RObj)

代码：

``` lua
function handle(points)
	ak, sk = "AccessKeyXXXXXXX", "SecretKeyXXXXXXXX"

	-- /BucketName/ObjectName/ObjectName..
	resource = "/leecha-test/test-directory/main.rs"

	date = os.date("!%a, %d %b %Y %X GMT")

	info = string.format("GET\n\n\n%s\n%s", date, resource)

	// hmac(mothod_str, data_str, key_str, true)
	sign = base64_encode(hmac("sha1", info, sk, true))

	auth = "OSS " .. ak .. ":" .. sign

	// http://OSS_EndPoint/ObejctName/ObejctName..
	response, err = http_request("GET", "http://leecha-test.oss-cn-shanghai.aliyuncs.com/test-directory/main.rs", {
		headers={
			["Date"]=date, 
			["Authorization"]=auth
		}
	})

	if err == nil then
		print(response.status_code)
		print(response.body)
	else
		print(err)
        end

	return points
end
```

**各种参数需要自行替换**

## 注意事项
暂无


#### log

在 lua 中提供四种级别的 log 输出到 dataway 日志文件中。

- log_info
- log_debug
- log_warn
- log_errorn

支持`bool`、`number`、`string`、`table`和`nil`类型，不支持`function`和`userdata`。

示例：

``` lua
function handle(points)

        tb = {"Hello","World",a=1,b=2,z=3,x=10,y=20,"Good","Bye"}

        log_info("this is info message", 11, tb)
        log_debug("this is debug message", 22)
        log_warn("this is warn message", 33)
        log_error("this is error message", 44)

	return points
end
```

输出：

```
2020-04-23T16:40:36+08:00 [info] this is info message 111 map[1:Hello 2:World 3:Good 4:Bye a:1 b:2 x:10 y:20 z:3]
2020-04-23T16:40:36+08:00 [debug] this is debug message 222
2020-04-23T16:40:36+08:00 [warn] this is warn message 333
2020-04-23T16:40:36+08:00 [error] this is error message 444
```
