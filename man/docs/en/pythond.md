{{.CSS}}
# Developing Custom Collector with Python
---

{{.AvailableArchs}}

---

{{.InputName}} is a complete set of scenes for firing user-defined python collection scripts at regular intervals.

## Preconditions {#reqirement}

### Python Environment {#req-python}

Currently in the alpha phase, **it is compatible with Python 3++ only**.

The following dependency libraries need to be installed:

- requests

The installation method is as follows:

```shell
# python3
python3 -m pip install requests
```

The above installation requires pip installation. If you don't have it, you can refer to the following method (from: [here](https://pip.pypa.io/en/stable/installation/){:target="_blank"}):

```shell
# Linux/MacOS
python -m ensurepip --upgrade

# Windows
py -m ensurepip --upgrade
```

### Write a User-defined Script {#add-script}

Create a folder named as Python package under the path `datakit/python.d`, then create a Python script(`.py`) inside it.

For example, as the Python package name is `Demo`, the path is like below. The `demo.py` is the Python script, and its name could change to whatever you want.

```
datakit
   └── python.d
       ├── Demo
       │   ├── demo.py
```

You need the user to inherit the `DataKitFramework` class and then override the `run` method.

>The `DataKitFramework` class source code file path is `datakit_framework.py` at `datakit/python.d/core/datakit_framework.py`.

??? note "Python script example"
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

Python SDK API definition (see `datakit_framework.py`):

- Reporting metrics data: `feed_metric(self, input=None, measurement=None, tags=None, fields=None, time=None, **kwargs)`;
- Reporting metrics data: `feed_logging(self, input=None, source=None, tags=None, message=None, time=None, **kwargs)`;
- Reporting metrics data: `feed_object(self, input=None, cls=None, name=None, tags=None, fields=None, time=None, **kwargs)`; (`cls` is `class`. Since `class` is a Python keyword, `class` is abbreviated to `cls`.)

### Write Python to Report Events {#report-event}

You can use the following three built-in functions to report event events:

- Events reporting `df_source = user`: `feed_user_event(self, df_user_id=None, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs)`
- Events reporting `df_source = monitor`: `feed_monitor_event(self, df_dimension_tags=None, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs)`
- Events reporting `df_source = system`: `feed_system_event(self, tags=None, df_date_range=10, df_status=None, df_event_id=None, df_title=None, df_message=None, **kwargs)`

General event field description:

|  Field Name   | Type  | Required or not  | Description  |
|  ----  | ----  | ----  | ----  |
| df_date_range  | Integer | Required | Time range. Unit s |
| df_source  | String | Required | Data source, value `system` , `monitor` , `user` |
| df_status  | Enum | Required | Status, value `ok` , `info` , `warning` , `error` , `critical` , `nodata` |
| df_event_id  | String | Required | event ID |
| df_title  | String | Required | Title |
| df_message  | String |  | Description |
| {other field}  | `kwargs`, such as `k1=5, k2=6` |  | Other extra field |

- When `df_source = monitor`:

Represent an event generated by Guance Cloud detection function, with the following additional fields:

|  Extra Field Name   | Type  | Required or not  | Description  |
|  ----  | ----  | ----  | ----  |
| df_dimension_tags  | String(JSON format) | Required | Detect latitude labels, such as `{"host":"web01"}` |

- When `df_source = user`:

Represent an event created directly by the user, with the following additional fields:

|  Extra Field Name   | Type  | Required or not | Description  |
|  ----  | ----  | ----  | ----  |
| df_user_id  | String | Required | 用户 ID |

- When `df_source = system`:

Represent an event generated by the system, and no additional fields exist.

Sample:

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

## Configuration {#config}

Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

```toml
{{.InputSample}}
```

## Git Support {#git}

Support the use of git repo. Once git repo is enabled, the path filled in args in conf is relative to the path of `gitrepos` . For example, args will fill in `mytest` in the following case:

```
├── datakit
└── gitrepos
    └── myconf
        ├── conf.d
        │   └── pythond.conf
        └── python.d
            └── mytest
                └── mytest.py
```

## Complete Example {#example}

Step 1: Write a class that inherits `DataKitFramework`:

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

Step 2: We don't turn on git repo here. Put `test.py` under the `mytest` folder of `python.d`:

```
└── python.d
    ├── mytest
    │   ├── test.py
```

Step 3: Configure {{.InputName}}.conf:

```toml
[[inputs.pythond]]

  # Python collector name
  name = 'some-python-inputs'  # required

  # Environment variables required to run Python collector
  #envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

  # Python collector executable program path (write absolute path wherever possible)
  cmd = "python3" # required. python3 is recommended.

  # The relative path of the user script (fill in the folder, after which the modules and py files in the next directory of the folder will be applied)
  dirs = ["mytest"]
```

Step 3: Restart DataKit:

```shell
sudo datakit --restart
```

## FAQ {#faq}

### How to Troubleshoot Errors {#log}

If the results are not as expected, you can view the following log files:

- `~/_datakit_pythond_cli.log`
- `_datakit_pythond_framework_[pythond name]_.log`
