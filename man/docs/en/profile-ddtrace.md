# DDTrace profiling

## Install and Run Profiling Agent {#install}

Currently, java and python languages are supported using profiling libraries for the corresponding languages.

### Java {#java}

Download the latest ddtrace agent dd-java-agent.jar

```shell
# java Version Requirements: java 8 version needs to be higher than 8u262 +, or use java 11 and above
wget -O dd-java-agent.jar 'https://github.com/DataDog/dd-trace-java/releases/download/v0.107.0/dd-java-agent-0.107.0.jar'
```

Run Java Code

```shell
java -javaagent:/<your-path>/dd-java-agent.jar \
-Ddd.service=profiling-demo \
-Ddd.env=dev \
-Ddd.version=1.2.3  \
-Ddd.profiling.enabled=true  \
-XX:FlightRecorderOptions=stackdepth=256 \
-Ddd.trace.agent.port=9529 \
-jar your-app.jar 
```

### Python {#python}

Install DDTrace Python Function Library

```shell
pip install ddtrace
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

- Open profiling in code

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
#while True:
#    time.sleep(1)

```

You don't need to use the ddtrace-run command to run at this time.

```shell
DD_ENV=testing DD_SERVICE=python-profiling-manual DD_VERSION=7.8.9 python3 app.py
```

## View Profile {#view}

After the above program is started, profile data will be collected regularly (once every minute by default) and reported to DataKit. After a while, profile data can be seen in Guance Cloud workspace.
