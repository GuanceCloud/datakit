---
title     : 'DDTrace Python'
summary   : 'Tracing Python applications with DDTrace'
tags      :
  - 'DDTRACE'
  - 'PYTHON'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/ddtrace'
---

## Install Dependencies {#dependence}

Install the DDTrace SDK. The example below also uses Flask and requests:

```shell
python -m pip install ddtrace flask requests
```

After installation, run `ddtrace-run --info` to check the configuration available at process startup. Its output does not include configuration changed later in application code.

## Running the Application {#instrument}

<!-- markdownlint-disable MD046 -->
=== "Host Application"

    > Prefix the Python entry point with `ddtrace-run` and set service identity and the DataKit destination before the process starts. The common upstream trace-port default is `8126`; explicitly use `9529` with DataKit.
    
    ```shell linenums="1"
    DD_SERVICE="<YOUR-SERVICE-NAME>" \
      DD_ENV="<YOUR-ENV-NAME>" \
      DD_VERSION="<YOUR-APP-VERSION>" \
      DD_AGENT_HOST="<YOUR-DATAKIT-HOST>" \
      DD_TRACE_AGENT_PORT="9529" \
      DD_LOGS_INJECTION=true \
      ddtrace-run python my_app.py
    ```

=== "Kubernetes"

    ```yaml hl_lines="10-19" linenums="1"
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec:
          containers:
            - name: <CONTAINER_NAME>
              image: <CONTAINER_IMAGE>/<TAG>
              env:
                - name: DD_AGENT_HOST
                  value: "datakit-service.datakit.svc"
                - name: DD_TRACE_AGENT_PORT
                  value: "9529"
                - name: DD_ENV
                  value: <YOUR-ENV-NAME>
                - name: DD_SERVICE
                  value: <YOUR-SERVICE-NAME>
                - name: DD_VERSION
                  value: <YOUR-APP-VERSION>
                - name: DD_LOGS_INJECTION
                  value: "true"
    ```
<!-- markdownlint-enable MD046 -->

After startup, call an instrumented endpoint and confirm trace traffic in the DataKit monitor. For troubleshooting, temporarily add `DD_TRACE_DEBUG=true`, then disable it to avoid excessive diagnostic logs.

The following common options are also available.

### Profiling {#instrument-profile}

```shell linenums="1"
DD_PROFILING_ENABLED=true \
  ddtrace-run python my_app.py
```

### Sampling Rate {#instrument-sampling}

Set a sampling rate of 0.8, so only 80% of the traces will be retained.

```shell linenums="1"
DD_TRACE_SAMPLE_RATE="0.8" \
  ddtrace-run python my_app.py
```

### Enable Python Runtime Metrics Collection {#instrument-py-runtime-metrics}

> Runtime metrics are delivered through DogStatsD, not the trace port. Enable the [StatsD collector](statsd.md) and point `DD_DOGSTATSD_HOST` / `DD_DOGSTATSD_PORT` at the DataKit StatsD service (normally UDP `8125`). Do not set them to `9529`.

```shell linenums="1"
DD_RUNTIME_METRICS_ENABLED=true \
  ddtrace-run python my_app.py
```

## Code Example {#example}

```python title="service_a.py"
from flask import Flask
import requests

app = Flask(__name__)

@app.route('/a',  methods=['GET'])
def index():
    requests.get('http://127.0.0.1:54322/b', timeout=3)
    return 'OK', 200

# Start service A: HTTP service starts on port 54321
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54321, debug=False, use_reloader=False)
```

```python title="service_b.py"
from flask import Flask
import time

app = Flask(__name__)

@app.route('/b',  methods=['GET'])
def index():
    time.sleep(1)
    return 'OK', 200

# Start service B: HTTP service starts on port 54322
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54322, debug=False, use_reloader=False)
```

## Run {#run}

Here, we take the commonly used Python Web Server Flask application as an example. In the example, `SERVICE_A` provides an HTTP service and calls the `SERVICE_B` HTTP service.

- Run `SERVICE_A`

```shell
DD_SERVICE=service-a \
DD_ENV=test \
DD_VERSION=v1 \
DD_TAGS=project:your_project_name \
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
ddtrace-run python3 service_a.py >a.log 2>&1 &
SERVICE_A_PID=$!
```

- Run `SERVICE_B`

```shell
DD_SERVICE=service-b \
DD_ENV=test \
DD_VERSION=v1 \
DD_TAGS=project:your_project_name \
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
ddtrace-run python3 service_b.py >b.log 2>&1 &
SERVICE_B_PID=$!
```

Call service A to prompt it to call service B, which will generate corresponding trace data (this can be executed multiple times to trigger)

```shell
curl http://localhost:54321/a
```

Stop both services:

```shell
kill "$SERVICE_A_PID" "$SERVICE_B_PID"
```

## Environment Variable Support {#envs}

The following variables are common. Set them before the Python process starts. For the complete list, precedence, and version-specific behavior, see the [Datadog Python configuration guide](https://docs.datadoghq.com/tracing/trace_collection/library_config/python/){:target="_blank"}.

- `DD_ENV`: Sets the environment variable for the service.
- `DD_VERSION`: The version number of the APP.
- `DD_SERVICE`: Sets the application service name. Framework integrations usually use it; set it explicitly in production.
- `DD_SERVICE_MAPPING`: Defines dependency-service mappings to rename dependencies in traces.
- `DD_TAGS`: Adds default tags to each span in `key:val,key:val` form. Do not include user identifiers or sensitive content.
- `DD_AGENT_HOST`: The DataKit host name or IP. `DD_TRACE_AGENT_URL`, when set, normally takes precedence.
- `DD_TRACE_AGENT_PORT`: The trace receiver port. The common upstream default is `8126`; DataKit uses `9529`.
- `DD_TRACE_SAMPLE_RATE`: Sets SDK-side sampling from `0.0` (0%) to `1.0` (100%).
- `DD_TRACE_ENABLED`: Controls trace generation/delivery; during troubleshooting, make sure it is not `false`.
