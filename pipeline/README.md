# 文本处理器 定义

以下是 文本处理器 定义。随着不同语法的逐步支持，该文档会做不同程度的调整和增删。

全局约束如下：

- 非关键字（如指标名、标签名等）大小写敏感，**关键字及函数名大小写不敏感**

- 以 `#` 为行注释字符，不支持行内注释

- 支持的操作符：

	- `+`  - 加法
	- `-`  - 减法
	- `*`  - 乘法
	- `/`  - 除法
	- `%`  - 取模
	- `=` - 等于
	- `!=` - 不等于
	- `<=` - 大于等于
	- `<` - 小于
	- `>=` - 大于等于
	- `>` -  大于
	- `^` - 指数运算
	- `&&` - 逻辑与
	- `||` - 逻辑或


- 标识符：标识符有几种形式，便于兼容各种变量命名形式

	- 正常变量名中只能出现 `[_a-zA-Z0-9]` 这些字符，且首字符不能是数字。如 `_abc, _abc123, _123ab`
	- 其它形式的变量名处理方式：
		- `this+is-a*xx/yy^zz?variable`，`by` 需写成 `` `this+is-a*xx/yy^zz?variable` ``，`` `by` ``，前者变量中带运算符，后者的 `by` 是 DQL 关键字
		- `这是一个中文或其它非英文字符变量` 需写成 `` `这是一个中文或其它非英文字符变量` ``
		- 变量中就带了一个反引号，`` this`is-a-vairalbe `` 需写成 `` `identifier("this`is-a-vairalbe")` `` 来修饰

- 字符串值可用双引号和单引号： `"this is a string"` 和 `'this is a string'` 是等价的

- 数据类型：支持浮点（`123.4`, `5.67E3`）、整形（`123`, `-1`）、字符串（`'张三'`, `"hello world"`）、Boolean（`true`, `false`）四种类型


## sdk使用

用法：
- 加载函数处理脚本
```
p := NewPipeline(script)  //script为string类型
```
传入函数处理表达式脚本文件或脚本字符串, 注意在编写函数表示式时，通过分号分割，支持多行

- 文本处理
```
p.Run(data) //data为string类型
```

- 获取结果
```
p.Result()  //返回类型为(map[string]interface{}, error)
```

## 脚本函数

支持的处理函数

函数: grok([grok_parttern], [json_path])

参数:
- `grok_parttern`: grok表达式 (必选参数)
- `json_path`: 待处理数据的json_path (选填参数)

说明: 将对应json_path的字符串执行Grok，并成为json_data子结构

示例:
```
json_data = `{"content": "127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207", "app": "dev"}`
grok("%{COMMONAPACHELOG}", content);
```

函数: rename(json_path], [new_key])

参数:
- `json_path`: 待改名的json_path (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将结构中的原有key改名为新的new_key, 并成为json_data子结构

示例:
```
json_data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
rename(a.second, bb);
```

函数: lowercase([json_path], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将对应处理数据转化为小写，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"a":{"first":2.3,"second":2,"thrid":"ABC","forth":true},"age":47}`
lowercase(a.thrid, bb);
```

函数: uppercase([json_path], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将对应处理数据转化为大写，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
uppercase(a.thrid, bb);
```

函数: nullif([json_path], [range])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将对应处理数据转化为大写，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
nullif(a.second, bb);
```

函数: user_agent([json_path], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 提取user-agent中的信息，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"userAgent":"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/535.2 (KHTML, like Gecko) Ubuntu/11.10 Chromium/15.0.874.106 Chrome/15.0.874.106 Safari/535.2"}`
user_agent(userAgent, user_agent);
```

函数: urldecode([json_path], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将对应字符串执行urldecode，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"url":"http://www.example.org/default.html?ct=32&op=92&item=98"}`
urldecode(url, url_dic);
```

函数: geoip([json_path], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将对应IP字符串获取响应地理信息，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"ip":"172.168.0.3"}`
geoip(ip, ip_addr);
```

函数: datetime([json_path], [datetime_parttern]，[new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `datetime_parttern`: 时间格式化字符串(格式化标准，待完善) (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将对应日期字段格式化为新的日期显示方式，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"date":"23/Apr/2014:22:58:32 +0200"}`
datetime(date, "yyyy-mm-dd HH:MM:SS");
```

函数: expr([json_path], [expr], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `expr`: 计算表达式 (必选参数)
- `new_key`: 新的key name  (必选参数)

说明: 将对应字段进行计算后，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
expr(a.second*10+(2+3)*5, bb);
```

函数: stringf([new_key], [format_pattern], [json_path]....)

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `format_pattern`: 格式化字符串 (必选参数)
- `new_key`:   新的key name      (必选参数)

说明: 通过printf格式自定义字符串

示例:
```
json_data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
stringf(bb, "%d %s %v", a.second, a.thrid, a.forth);
```

函数: cast([json_path], [type]，[new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `type`: 转化的类型, 以下枚举值int, bool, str, float (必选参数)
- `new_key`:   新的key name  (必选参数)

说明: 将对应的字段进行类型转化，插入到new_key下,  并成为json_data子结构

示例:
```
json_data = `{"a":{"first":2.3,"second":2,"thrid":"abc","forth":true},"age":47}`
cast(bb, a.second, "float");
```

函数: group([json_path], [range], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `range`: 指定范围 (必选参数)
- `new_key`:   新的key name  (必选参数)

说明: 将对应的字段进行按条件分组赋值

示例:
```
json_data = `{"age":10, "name": "张三"}`
group(age, [0-16], children);
```


函数: group_in([json_path], [set], [new_key])

参数:
- `json_path`: 待处理数据的json_path (必选参数)
- `set`: 指定集合 (必选参数)
- `new_key`:   新的key name  (必选参数)

说明: 将对应的字段进行按条件分组赋值

示例:
```
json_data = `{"age":10, "name": "张三"}`
group_in(name, ["张三"], zhangsan);
```