# Logstash
---

Logstash only supports writing data of the Logging type.

## Step 1: Build Back-end Storage {#backend-storage}

To build your own Logstash environment, you need to open the [HTTP module](https://www.elastic.co/cn/blog/introducing-logstash-input-http-plugin){:target="_blank"}, and the opening method is also very simple, which can be configured directly in the pipeline file of Logstash.

### New Logstash Pipeline File {#new-pipeline}

Create a new Logstash pipeline file `pipeline-http.conf` as follows:

```conf
input {
    http {
        host => "0.0.0.0" # default: 0.0.0.0
        port => 8080 # default: 8080
    }
}

output {
    elasticsearch {
        hosts => [ "localhost:9200"] # What I configured here is to write data to elasticsearch
    }
}
```

This file can be put anywhere, and I put it under `/opt/elastic/logstash` , that is, the full path is `/opt/elastic/logstash/pipeline-http.conf`.

### Configure Logstash to Use the Pipeline File above {#setup-pipeline}

There are two ways: configuration file mode and command line mode. Just choose one.

- Profile mode

Add a line to the configuration file `logstash/config/logstash.yml`:

```yml
path.config: /opt/elastic/logstash/pipeline-http.conf
```

- Command line mode

Specify the pipeline file on the command line:

```shell
$ logstash/bin/logstash -f /opt/elastic/logstash/pipeline-http.conf
```

## Step 2: Add Configuration {#config-sink}

Add the following fragment to `datakit.conf`:

```conf
...
[sinks]

  [[sinks.sink]]
    categories = ["L"]
    host = "1.1.1.1:8080"
    protocol = "http"
    request_path = "/index/type/id"
    target = "logstash"
    timeout = "5s"
    write_type="json"
...
```

In addition to the fact that the Sink must be configured with the [generic parameter](datakit-sink-guide.md), the Sink instance of Logstash currently supports the following parameters:

- `host`(required): HTTP host should be of the form `host:port` or `[ipv6-host%zone]:port`.
- `protocol`(required): `http` or `https`.
- `write_type`(required): The format type of the source data being written: `json` or `plain`。
- `request_path`: The path of the request URL.
- `timeout`: Timeout for influxdb writes, defaults to 10 seconds.

## Step 3: Restart DataKit {#restart-dk}

`$ sudo datakit --restart`

## Specifying the LogStash Sink Setting in Installation Phase {#logstash-on-installer}

LogStash supports the way environment variables are turned on during installation.

```shell
DK_SINK_L="logstash://localhost:8080?protocol=http&request_path=/index/type/id&timeout=5s&write_type=json" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```
