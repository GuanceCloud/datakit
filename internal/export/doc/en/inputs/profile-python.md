# Python profiling

DataKit Python profiling supports [dd-trace-py](https://github.com/DataDog/dd-trace-py){:target="_blank"} and [py-spy](https://github.com/benfred/py-spy){:target="_blank"}.

## Requirements {#py-spy-requirement}

Install [DataKit](https://www.guance.com){:target="_blank"} and enable [profile](profile.md#config) input.

## Use dd-trace-py {#ddtrace}

- Install dd-trace-py library

<!-- markdownlint-disable MD046 -->
???+ note

    Datakit is now compatible with dd-trace-py 1.14.x and below, higher versions are not tested.
<!-- markdownlint-enable -->

```shell
pip3 install ddtrace
```

- Profiling by attaching into the target process

```shell
DD_PROFILING_ENABLED=true \
DD_ENV=dev \
DD_SERVICE=my-web-app \
DD_VERSION=1.0.3 \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
ddtrace-run python app.py
```

- Profiling by writing code

```python
import time
import ddtrace
from ddtrace.profiling import Profiler

ddtrace.tracer.configure(
     https=False,
     hostname="localhost",
     port="9529",
)

prof = Profiler()
prof.start(True, True)

# your code here ...
# while True:
#     time.sleep(1)
```

There is no need to add `ddtrace-run` command

```shell
DD_ENV=testing DD_SERVICE=python-profiling-manual DD_VERSION=1.2.3 python3 app.py
```

### View Profile {#view}

After a minute or two, you can visualize your profiles on the [APM -> Profile](https://console.guance.com/tracing/profile){:target="_blank"} .

## Use `py-spy` {#py-spy}

`py-spy`is a non-invasive Python performance metric sampling tool provided by the open source community, which has the advantages of running independently and having low impact on target program load By default, `py-spy` will output sampling data in different formats to a local file based on the specified parameters. To simplify the integration of `py-spy` and DataKit, Observation Cloud provides a branch version [`py-spy-for-datakit`]（<https://github.com/GuanceCloud/py-spy-for-datakit>）{: target="_Blank"}, with little modifications made to the original version, supporting automatic profiling send data to DataKit.

- Installation

`pip install` is recommend way.

```shell
pip3 install py-spy-for-datakit
```

besides, [Github Release](https://github.com/GuanceCloud/py-spy-for-datakit/releases){:target="_blank"} page provides pre compiled versions of some mainstream platforms, which you can also download and install using PIP. Below is Linux x86_64 platform as an example (other platforms is similar), let's introduce the installation steps of the pre compiled version:

```shell
# download binary
curl -SL https://github.com/GuanceCloud/py-spy-for-datakit/releases/download/v0.3.15/py_spy_for_datakit-0.3.15-py2.py3-none-manylinux_2_5_x86_64.manylinux1_x86_64.whl -O

# use pip to install
pip3 install --force-reinstall --no-index --find-links . py-spy-for-datakit

# confirm successful installation
py-spy-for-datakit help
```

if your machine has `rust` and `cargo` installed, you can use `cargo` to install it.

```shell
cargo install py-spy-for-datakit
```

- Usage

`py-spy-for-datakit` has added the `datakit` command to the original subcommand of `py-spy`, specifically used to send sampling data to DataKit. You can type `py-spy-for-datakit help datakit` for usage help:

| Option             | describe                          | default                             |
|--------------------|-----------------------------------|-------------------------------------|
| -H, --host         | Datakit listening host            | 127.0.0.1                           |
| -P, --port         | Datakit listening port            | 9529                                |
| -S, --service      | Your service name                 | unnamed-service                     |
| -E, --env          | Your app deploy environment       | unnamed-env                         |
| -V, --version      | Your app version                  | unnamed-version                     |
| -p, --pid          | Target process PID                | You must set this option or command |
| -d, --duration     | Profiling duration                | 60                                  |
| -r, --rate         | Profiling rate                    | 100                                 |
| -s, --subprocesses | Whether profiling sub process     | false                               |
| -i, --idle         | Whether profiling inactive thread | false                               |

`py-spy-for-datakit` can analyze the currently running program by using the `--pid <PID>` or `-p <PID>` parameters to pass the process PID of the running Python program to `py-spy-for-datakit`.

Imaging your target process PID is 12345, and Datakit is listening at 127.0.0.1:9529:

```shell
py-spy-for-datakit datakit \
  --host 127.0.0.1 \
  --port 9529 \
  --service <your-service-name> \
  --env testing \
  --version v0.1 \
  --duration 60 \
  --pid 12345
```

If needed, please add `sudo` prefix.

`py-spy-for-datakit` also supports direct startup commands with Python projects, so there is no need to specify a process PID. At the same time, data sampling will be performed when the program starts, and the running commands are similar:

```shell
py-spy-for-datakit datakit \
  --host 127.0.0.1 \
  --port 9529 \
  --service your-service-name \
  --env testing \
  --version v0.1 \
  -d 60 \
  -- python3 server.py  # There is a blank in front of python3
```

After a minute or two, you can visualize your profiles on the [profile](https://console.guance.com/tracing/profile){:target="_blank"}.
