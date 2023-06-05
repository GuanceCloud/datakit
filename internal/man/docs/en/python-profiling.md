# Python profiling

Currently, DataKit Python profiling supports [ddtrace](https://github.com/DataDog/dd-trace-py){:target="_blank"} and  [py-spy](https://github.com/benfred/py-spy){:target="_blank"}.

## Preconditions {#py-spy-requirement}

[DataKit](https://www.guance.com){:target="_blank"} is installed and the [profile](profile.md#config) is turned on.

## ddtrace Access {#ddtrace}

`ddtrace` is an open source library for link tracing and performance analysis introduced by DataDog, which can collect metrics such as `CPU`, `memory` and `lock contention`, and has comprehensive functions.

- Install Python ddtrace Library

```shell
pip3 install ddtrace
```

- Automatic profiling

```shell
DD_PROFILING_ENABLED=true \
DD_ENV=dev \
DD_SERVICE=my-web-app \
DD_VERSION=1.0.3 \
DD_TRACE_AGENT_URL=http://127.0.0.1:9529 \
ddtrace-run python app.py
```

You can also turn on profiling by injecting code manually:

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

There is no need to add the `ddtrace-run` command to start the project at this time:

```shell
DD_ENV=testing DD_SERVICE=python-profiling-manual DD_VERSION=1.2.3 python3 app.py
```

### View Profile {#view}

After the program starts, ddtrace will collect data regularly (once every minute by default) and report it to DataKit. After a few minutes, you can view the corresponding data in Guance Cloud hosting [Application Performance Monitoring -> Profile](https://console.guance.com/tracing/profile){:target="_blank"}.

## py-spy Access {#py-spy}

`py-spy` is a non-invasive Python performance metrics sampling tool provided by the open source community that runs alone and has low impact on target program load.

`Py-spy` will output sample data in different formats to local files according to specified parameters by default. To simplify the integration of `py-spy` and DataKit, observation cloud provides a branch version [py-spy-for-datakit](https://github.com/GuanceCloud/py-spy-for-datakit){:target="_blank"}, which is based on the original version with a few modifications and supports automatic profiling.
Data is sent to DataKit.

- Installation

Pip installation is recommended

```shell
pip3 install py-spy-for-datakit
```

In addition, precompiled versions of some major platforms are available on the [Github Release](https://github.com/GuanceCloud/py-spy-for-datakit/releases){:target="_blank"} page, which you can also download later.
Install with pip, let's take Linux x86_64 platform as an example (other platforms are similar), and introduce the installation steps of precompiled version.

```shell
# Download the precompiled package for the corresponding platform
curl -SL https://github.com/GuanceCloud/py-spy-for-datakit/releases/download/v0.3.15/py_spy_for_datakit-0.3.15-py2.py3-none-manylinux_2_5_x86_64.manylinux1_x86_64.whl -O

# Install with pip
pip3 install --force-reinstall --no-index --find-links . py-spy-for-datakit

# Verify whether the installation was successful
py-spy-for-datakit help
```

If your system has rust and cargo installed, you can also use cargo to install it.

```shell
cargo install py-spy-for-datakit
```

- Use

py-spy-for-datakit adds the `datakit` command to the original `py-spy` subcommand, which is specially used to send sampled data to DataKit. You can enter `py-spy-for-datakit help datakit` to view the usage help:

| Parameter                 | Description                                                                           | Default Value                            |
| -------------------- | ----------------------------------------------                                 | --------------------              |
| -H, --host           | The address where Datakit listens to send data                                                  | 127.0.0.1                         |
| -P, --port           | Datakit listening port to which data send.                                                  | 9529                              |
| -S, --service        | Project name, which can be used to distinguish different projects in the background, and can be used for filtering and querying. It is recommended to set up.            | unnamed-service                   |
| -E, --env            | The environment in which the project is deployed can be used to distinguish between development, test and production environments, and can also be used for filtering. It is recommended to set up.   | unnamed-env                       |
| -V, --version        | Project version, which can be used for background query and filtering. It is recommended to set up.                                     | unnamed-version                   |
| -p, --pid            | Process PID of Python program to be analyzed.                                                  | The process PID and the project startup command must specify one. |
| -d, --duration       | The duration of the sample, which is sent to Datakit every time interval, in seconds, and can be set to a minimum of 10. | 60                                |
| -r, --rate           | Sampling frequency, sampling times per second                                                         | 100                               |
| -s, --subprocesses   | Whether to analyze child processes                                                                 | false                             |
| -i, --idle           | Whether to sample non-running threads                                                       | false                             |

py-spy-for-datakit can analyze the currently running program by passing the process PID of the running Python program to py-spy-for-datakit using the `--pid <PID>` or `-p <PID>` parameter.

Assuming your Python application is currently running a process PID of 12345 and DataKit listens at 127.0. 0.1: 9529, use commands similar to this:

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

If you are prompted for `sudo` permission, prefix the command with `sudo`.

py-spy-for-datakit also supports startup commands directly to Python projects, so that you don't need to specify process PID, and data sampling will be carried out when the program starts. The running command at this time is similar:
```shell
py-spy-for-datakit datakit \
  --host 127.0.0.1 \
  --port 9529 \
  --service your-service-name \
  --env testing \
  --version v0.1 \
  -d 60 \
  -- python3 server.py  # Note that you need to add an extra space before python3 here
```

If there is no error, wait a minute or two to view the specific performance index data on the observation cloud platform [Application Performance Monitoring -> Profile](https://console.guance.com/tracing/profile){:target="_blank"} page.
