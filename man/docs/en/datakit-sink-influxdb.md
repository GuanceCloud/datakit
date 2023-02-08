# InfluxDB
---

InfuxDB only supports writing Metric-type data.

## Step 1: Build Back-end Storage {#backend-storage}

Build your own InfuxDB server environment.

## Step 2: Add Configuration {#config}

Add the following fragment to `datakit.conf`:

```conf
...
[sinks]
  [[sinks.sink]]
    categories = ["M"]
    target = "influxdb"
    host = "localhost:8087"
    protocol = "http"
    database = "db1"
    precision = "ns"
    timeout = "15s"
...
```

In addition to the fact that the Sink must be configured with the [generic parameter](datakit-sink-guide.md), the Sink instance of InfuxDB currently supports the following parameters:

- `host`(required): HTTP/UDP host should be of the form `host:port` or `[ipv6-host%zone]:port`.
- `protocol`(required): `http` or `udp`.
- `database`(required): Database is the database to write points to.
- `precision`: Precision is the write precision of the points, defaults to "ns".
- `username`: Username is the influxdb username, optional.
- `password`: Password is the influxdb password, optional.
- `timeout`: Timeout for influxdb writes, defaults to 10 seconds.
- `user_agent`: UserAgent is the http User Agent, defaults to "InfluxDBClient".
- `retention_policy`: RetentionPolicy is the retention policy of the points.
- `write_consistency`: Write consistency is the number of servers required to confirm write.
- `write_encoding`: WriteEncoding specifies the encoding of write request
- `payload_size`(UDP protocol specific): PayloadSize is the maximum size of a UDP client message, optional. Tune this based on your network. Defaults to 512.

## Step 3: Restart DataKit {#restart-dk}

`$ sudo datakit --restart`

## Specify the InfuxDB Sink Setting During Installation {#setup-influxdb}

InfuxDB supports the way environment variables are turned on during installation.

```shell
DK_SINK_M="influxdb://localhost:8087?protocol=http&database=db1&precision=ns&timeout=15s" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```
