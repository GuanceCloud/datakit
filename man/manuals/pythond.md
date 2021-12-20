{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

{{.InputName}} 是定时触发用户自定义 python 采集脚本的一整套方案。

## 前置条件

### Python 环境

目前处于 alpha 阶段，<strong>同时兼容 Python 2.7+ 和 Python 3+<strong>。

需要安装以下依赖库:

- requests

安装方法如下:

```shell
# python2
python -m pip install requests

# python3
python3 -m pip install requests
```

上述的安装需要安装 pip，如果你没有，可以参考以下方法(源自: [这里](https://pip.pypa.io/en/stable/installation/)):

```shell
# Linux/MacOS
python -m ensurepip --upgrade

# Windows
py -m ensurepip --upgrade
```

### 编写用户自定义脚本

需要用户继承 `DataKitFramework` 类，然后对 `run` 方法进行改写。DataKitFramework 类源代码文件路径是 `datakit_framework.py` 在 `datakit/python.d/core/datakit_framework.py`。

具体的使用可以参见源代码文件 `datakit/python.d/core/demo.py`:

```python
from datakit_framework import DataKitFramework

class Demo(DataKitFramework):
    __name = 'Demo'
    interval = 10 # triggered interval seconds.

    # if your datakit ip is 127.0.0.1 and port is 9529, you won't need use this,
    # just comment it.
    # def __init__(self, **kwargs):
    #     super().__init__(ip = '127.0.0.1', port = 9529)

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
            'M':data,
            'input': "datakitpy"
        }

        return self.report(in_data) # you must call self.report here
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

## 其它

支持使用 git repo，一旦开启 git repo 功能，则 conf 里面的 args 里面填写的路径是相对于 `gitrepos` 的路径。比如下面这种情况，args 就填写 `myconf/mytest.py`:

```
├── datakit
├── gitrepos
│   └── myconf
│       ├── mytest.py
│       └── pythond.conf
```

## 完整示例

第一步：写一个类，继承 `DataKitFramework`:

```python
from datakit_framework import DataKitFramework

class MyTest(DataKitFramework):
    __name = 'MyTest'
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

第二步：我们这里不开启 git repo 功能。将 `test.py` 放到 `python.d` 的 `mytest` 文件夹下:

```
└── python.d
    ├── mytest
    │   ├── test.py
```

第三步：配置 {{.InputName}}.conf:

```conf
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

第四步: 重启 DataKit:

```shell
sudo datakit --restart
```
