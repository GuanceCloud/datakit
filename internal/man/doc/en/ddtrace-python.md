
# Python Example
---

## Install Dependency {#dependence}

Installing Python Flask

```shell
pip install flask
```

Installing DDTrace Python Function Library

```shell
pip install ddtrace
```

**Note:** This command requires pip version 18.0. 0 or higher. For Ubuntu, Debian, or other package management tools, upgrade the pip version using the following command:

```shell
pip install --upgrade pip
```

## Code Example {#example}

**service_a.py**

```python
from flask import Flask, request
import requests, os
from ddtrace import tracer

app = Flask(__name__)

def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()

@app.route('/a',  methods=['GET'])
def index():
    requests.get('http://127.0.0.1:54322/b')
    return 'OK', 200

@app.route('/stop',  methods=['GET'])
def stop():
    shutdown_server()
    return 'Server shutting down...\n'

# 启动 service A: HTTP 服务启动在 54321 端口上
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54321, debug=True)
```

**service_b.py**

```python
from flask import Flask, request
import os, time, requests
from ddtrace import tracer

app = Flask(__name__)

def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()

@app.route('/b',  methods=['GET'])
def index():
    time.sleep(1)
    return 'OK', 200

@app.route('/stop',  methods=['GET'])
def stop():
    shutdown_server()
    return 'Server shutting down...\n'

# Start service B: The HTTP service starts on port 54322
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54322, debug=True)
```

## Run {#run}

Take the Webserver Flask application commonly used in Python as an example. In the example, `SERVICE_A` provides the HTTP service and calls the `SERVICE_B` HTTP service.

**Run SERVICE_A**

```shell
DD_SERVICE=SERVICE_A \
DD_SERVICE_MAPPING=postgres:postgresql,defaultdb:postgresql \
DD_TAGS=project:your_project_name,env=test,version=v1 \
DD_AGENT_HOST=localhost \
DD_AGENT_PORT=9529 \
ddtrace-run python3 service_a.py &> a.log &
```

**Run SERVICE_B**

```shell
DD_SERVICE=SERVICE_B \
DD_SERVICE_MAPPING=postgres:postgresql,defaultdb:postgresql \
DD_TAGS=project:your_project_name,env=test,version=v1 \
DD_AGENT_HOST=localhost \
DD_AGENT_PORT=9529 \
ddtrace-run python3 service_b.py &> b.log &
```

Invoke the A service, causing it to invoke the B service, which produces the corresponding trace data (here you can perform multiple triggers)

```shell
curl http://localhost:54321/a
```

Stop both services separately

```shell
curl http://localhost:54321/stop
curl http://localhost:54322/stop
```

## Environment Variable Support {#envs} 

- DD_ENV: Set environment variables for the service.
- DD_VERSION: APP version number.
- DD_SERVICE: The service name used to set the application, which is passed when setting up middleware for Web framework integration such as Pylons, Flask, or Django. For Tracing without Web integration, it is recommended that you set the service name in your code.
- DD_SERVICE_MAPPING: Define service name mappings for renaming services in Tracing.
- DD_TAGS: Add default Tags for each Span.
- DD_AGENT_HOST: The name of the address where Datakit listens, default localhost.
- DD_AGENT_PORT: The port number on which Datakit listens, the default is 9529.
- DD_TRACE_SAMPLE_RATE: Set the sampling rate from 0.0 (0%) to 1.0 (100%).
