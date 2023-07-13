---
title     : 'Pythond'
summary   : '通过 Python 扩展采集数据'
__int_icon      : 'icon/pythond'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# 用 Python 开发自定义采集器
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

PythonD 是定时触发用户自定义 Python 采集脚本的一整套方案。

## 配置 {#config}

进入 DataKit 安装目录下的 *conf.d/{{.Catalog}}* 目录，复制 *{{.InputName}}.conf.sample* 并命名为 *{{.InputName}}.conf*。示例如下：

```toml
{{.InputSample}}
```

### Python 环境 {#req-python}

目前处于 alpha 阶段，**只兼容 Python 3+**。已测试的版本：

- [x] 3.10.1

需要安装以下依赖库：

- requests

安装方法如下：

```shell
# python3
python3 -m pip install requests
```

上述的安装需要安装 pip，如果你没有，可以参考以下方法（[源自这里](https://pip.pypa.io/en/stable/installation/){:target="_blank"}）：

```shell
# Linux/MacOS
python -m ensurepip --upgrade

# Windows
py -m ensurepip --upgrade
```

### 编写用户自定义脚本 {#add-script}

在 `datakit/python.d` 目录下创建以 "Python 包名" 命名的目录，然后在该目录下创建 Python 脚本(`.py`)。

以包名 `Demo` 为例，其路径结构如下。其中 `demo.py` 为 Python 脚本，Python 脚本的文件名可以自定义：

```shell
datakit
   └── python.d
       ├── Demo
       │   ├── demo.py
```

Python 脚本需要用户继承 `DataKitFramework` 类，然后对 `run` 方法进行改写。

> `DataKitFramework` 类的源代码文件路径是 `datakit_framework.py` 在 `datakit/python.d/core/datakit_framework.py`。

<!-- markdownlint-disable MD046 -->
???- note "Python 脚本源码参考示例"

    ```python
    #encoding: utf-8

    from datakit_framework import DataKitFramework

    class Demo(DataKitFramework):
        name = 'Demo'
        interval = 10 # triggered interval seconds.

        # if your datakit ip is 127.0.0.1 and port is 9529, you won't need use this,
        # just comment it.
        # def __init__(self, **kwargs):
        #     super().__init__(ip = '127.0.0.1', port = 9529)

        # General report example.
        def run(self):
            print("Demo")
            data = [
                    {
                        "measurement": "abc",
                        "tags": {
                        "t1": "b",
                        "t2": "d"
                        },
                        "fields": {
                        "f1": 123,
                        "f2": 3.4,
                        "f3": "strval"
                        },
                        # "time": 1624550216 # you don't need this
                    },

                    {
                        "measurement": "def",
                        "tags": {
                        "t1": "b",
                        "t2": "d"
                        },
                        "fields": {
                        "f1": 123,
                        "f2": 3.4,
                        "f3": "strval"
                        },
                        # "time": 1624550216 # you don't need this
                    }
                ]

            in_data = {
                'M':data, # 'M' for metrics, 'L' for logging, 'R' for rum, 'O' for object, 'CO' for custom object, 'E' for event.
                'input': "datakitpy"
            }

            return self.report(in_data) # you must call self.report here

        # # KeyEvent report example.
        # def run(self):
        #     print("Demo")

        #     tags = {"tag1": "val1", "tag2": "val2"}
        #     date_range = 10
        #     status = 'info'
        #     event_id = 'event_id'
        #     title = 'title'
        #     message = 'message'
        #     kwargs = {"custom_key1":"custom_value1", "custom_key2": "custom_value2", "custom_key3": "custom_value3"}

        #     # Feed df_source=user event.
        #     user_id="user_id"
        #     return self.feed_user_event(
        #         user_id,
        #         tags, date_range, status, event_id, title, message, **kwargs
        #         )

        #     # Feed df_source=monitor event.
        #     dimension_tags='{"host":"web01"}' # dimension_tags must be the String(JSON format).
        #     return self.feed_monitor_event(
        #         dimension_tags,
        #         tags, date_range, status, event_id, title, message, **kwargs
        #         )

        #     # Feed df_source=system event.
        #     return self.feed_system_event(
        #         tags, date_range, status, event_id, title, message, **kwargs
        #         )

        # # metrics, logging, object example.
        # def run(self):
        #     print("Demo")

        #     measurement = "mydata"
        #     tags = {"tag1": "val1", "tag2": "val2"}
        #     fields = {"custom_field1": "val1","custom_field2": 1000}
        #     kwargs = {"custom_key1":"custom_value1", "custom_key2": "custom_value2", "custom_key3": "custom_value3"}

        #     # Feed metrics example.
        #     return self.feed_metric(
        #         measurement=measurement,
        #         tags=tags,
        #         fields=fields,
        #         **kwargs
        #         )

        #     # Feed logging example.
        #     message = "This is the message for testing"
        #     return self.feed_logging(
        #         source=measurement,
        #         tags=tags,
        #         message=message,
        #         **kwargs
        #         )

        #     # Feed object example.
        #     name = "name"
        #     return self.feed_object(
        #         cls=measurement,
        #         name=name,
        #         tags=tags,
        #         fields=fields,
        #         **kwargs
        #         )
    ```
<!-- markdownlint-enable -->

Python SDK API 定义(详情参见 `datakit_framework.py`)：

- 上报 metrics 数据：`feed_metric(self, input=None, measurement=None, tags=None, fields=None, time=None, **kwargs)`;
- 上报 logging 数据：`feed_logging(self, input=None, source=None, tags=None, message=None, time=None, **kwargs)`;
- 上报 object 数据：`feed_object(self, input=None, cls=None, name=None, tags=None, fields=None, time=None, **kwargs)`; （`cls` 就是 `class`。因为 `class` 是 Python 的关键字，所以里把 `class` 缩写为 `cls`）

### 编写 Pythond 上报 event 事件 {#report-event}

可以使用以下三个内置函数来上报 event 事件：

- 上报 `df_source = user` 的事件：`feed_user_event(self, df_user_id=None, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs)`
- 上报 `df_source = monitor` 的事件：`feed_monitor_event(self, df_dimension_tags=None, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs)`
- 上报 `df_source = system` 的事件：`feed_system_event(self, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs)`

通用 event 字段说明：

| 字段名        | 类型                        | 是否必须 | 说明                                                                   |
| ----          | ----                        | ----     | ----                                                                   |
| df_date_range | Integer                     | 必须     | 时间范围。单位 s                                                       |
| df_source     | String                      | 必须     | 数据来源。取值 `system` , `monitor` , `user`                           |
| df_status     | Enum                        | 必须     | 状态。取值 `ok` , `info` , `warning` , `error` , `critical` , `nodata` |
| df_event_id   | String                      | 必须     | event ID                                                               |
| df_title      | String                      | 必须     | 标题                                                                   |
| df_message    | String                      |          | 详细描述                                                               |
| {其他字段}    | `kwargs`, 例如 `k1=5, k2=6` |          | 其他额外字段                                                           |

- 当 `df_source = monitor` 时：

表示由观测云检测功能产生的事件，额外存在以下字段：

| 额外字段名        | 类型                | 是否必须 | 说明                                |
| ----              | ----                | ----     | ----                                |
| df_dimension_tags | String(JSON format) | 必须     | 检测纬度标签，如 `{"host":"web01"}` |

- 当 `df_source = user` 时：

表示由用户直接创建的事件，额外存在以下字段：

| 额外字段名 | 类型   | 是否必须 | 说明    |
| ----       | ----   | ----     | ----    |
| df_user_id | String | 必须     | 用户 ID |

- 当 `df_source = system` 时：

表示为系统生成的事件，不存在额外字段。

使用示例：

```py
#encoding: utf-8

from datakit_framework import DataKitFramework

class Demo(DataKitFramework):
    name = 'Demo'
    interval = 10 # triggered interval seconds.

    # if your datakit ip is 127.0.0.1 and port is 9529, you won't need use this,
    # just comment it.
    # def __init__(self, **kwargs):
    #     super().__init__(ip = '127.0.0.1', port = 9529)

    # KeyEvent report example.
    def run(self):
        print("Demo")

        tags = {"tag1": "val1", "tag2": "val2"}
        date_range = 10
        status = 'info'
        event_id = 'event_id'
        title = 'title'
        message = 'message'
        kwargs = {"custom_key1":"custom_value1", "custom_key2": "custom_value2", "custom_key3": "custom_value3"}

        # Feed df_source=user event.
        user_id="user_id"
        return self.feed_user_event(
            df_user_id=user_id,
            tags=tags, df_date_range=date_range, df_status=status, df_event_id=event_id, df_title=title, df_message=message, **kwargs
            )

        # Feed df_source=monitor event.
        dimension_tags='{"host":"web01"}' # dimension_tags must be the String(JSON format).
        return self.feed_monitor_event(
            df_dimension_tags=dimension_tags,
            tags=tags, df_date_range=date_range, df_status=status, df_event_id=event_id, df_title=title, df_message=message, **kwargs
            )

        # Feed df_source=system event.
        return self.feed_system_event(
            tags=tags, df_date_range=date_range, df_status=status, df_event_id=event_id, df_title=title, df_message=message, **kwargs
            )
```

### Git 支持 {#git}

支持使用 git repo，一旦开启 git repo 功能，则 conf 里面的 args 里面填写的路径是相对于 `gitrepos` 的路径。比如下面这种情况，args 就填写 `mytest`：

```shell
├── datakit
└── gitrepos
    └── myconf
        ├── conf.d
        │   └── pythond.conf
        └── python.d
            └── mytest
                └── mytest.py
```

## 完整示例 {#example}

第一步：写一个类，继承 `DataKitFramework`：

```python
from datakit_framework import DataKitFramework

class MyTest(DataKitFramework):
    name = 'MyTest'
    interval = 10 # triggered interval seconds.

    # if your datakit ip is 127.0.0.1 and port is 9529, you won't need use this,
    # just comment it.
    # def __init__(self, **kwargs):
    #     super().__init__(ip = '127.0.0.1', port = 9529)

    def run(self):
        print("MyTest")
        data = [
                {
                    "measurement": "abc",
                    "tags": {
                      "t1": "b",
                      "t2": "d"
                    },
                    "fields": {
                      "f1": 123,
                      "f2": 3.4,
                      "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                },

                {
                    "measurement": "def",
                    "tags": {
                      "t1": "b",
                      "t2": "d"
                    },
                    "fields": {
                      "f1": 123,
                      "f2": 3.4,
                      "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                }
            ]

        in_data = {
            'M':data,
            'input': "datakitpy"
        }

        return self.report(in_data) # you must call self.report here
```

第二步：我们这里不开启 git repo 功能。将 `test.py` 放到 `python.d` 的 `mytest` 文件夹下：

```shell
└── python.d
    ├── mytest
    │   ├── test.py
```

第三步：配置 *{{.InputName}}.conf*:

```toml
[[inputs.pythond]]

  # Python 采集器名称
  name = 'some-python-inputs'  # required

  # 运行 Python 采集器所需的环境变量
  #envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

  # Python 采集器可执行程序路径(尽可能写绝对路径)
  cmd = "python3" # required. python3 is recommended.

  # 用户脚本的相对路径(填写文件夹，填好后该文件夹下一级目录的模块和 py 文件都将得到应用)
  dirs = ["mytest"]
```

第四步：重启 DataKit:

```shell
sudo datakit service -R
```

## FAQ {#faq}

### :material-chat-question: 如何排查错误 {#log}

如果结果不及预期，可以查看以下日志文件：

- `~/_datakit_pythond_cli.log`
- `_datakit_pythond_framework_[pythond name]_.log`
