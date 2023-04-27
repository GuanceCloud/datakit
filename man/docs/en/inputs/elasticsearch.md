
# ElasticSearch
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:  · [:fontawesome-solid-flag-checkered:](index.md#legends "支持选举")

---

ElasticSearch collector mainly collects node operation, cluster health, JVM performance, metric performance, retrieval performance and so on.

## Preconditions {#requirements}

- ElasticSearch version >= 6.0.0
- ElasticSearch collects `Node Stats` metrics by default. If you need to collect `Cluster-Health` related metrics, you need to set `cluster_health = true`
- Setting `cluster_health = true` produces the following measurement
  - `elasticsearch_cluster_health`

- Setting `cluster_stats = true` produces the following measurement
  - `elasticsearch_cluster_stats`

## User Rights Configuration {#user-permission}

If the account password access is turned on, the corresponding permissions need to be configured, otherwise it will lead to the failure of obtaining monitoring information. Elasticsearch, Open District for Elasticsearch, and OpenSearch are currently supported.

### Elasticsearch {#perm-es}

- Create the role `monitor` and set the following permissions.

```javascript
  {
    "applications": [],
    "cluster": [
        "monitor"
    ],
    "global": [],
    "indices": [
        {
            "allow_restricted_indices": false,
            "names": [
                "all"
            ],
            "privileges": [
                "manage_ilm",
                "monitor"
            ]
        },
    ],
    "run_as": []
  }

```

- Create a custom user and assign the newly created `monitor` role.
- Please refer to the profile description for additional information.

### Open Distro for Elasticsearch {#perm-open-es}

- Create a user
- Create the role `monitor` and set the following permissions:

```
PUT _opendistro/_security/api/roles/monitor
{
  "description": "monitor es cluster",
  "cluster_permissions": [
    "cluster:admin/opendistro/ism/managedindex/explain",
    "cluster_monitor",
    "cluster_composite_ops_ro"
  ],
  "index_permissions": [
    {
      "index_patterns": [
        "*"
      ],
      "fls": [],
      "masked_fields": [],
      "allowed_actions": [
        "read",
        "indices_monitor"
      ]
    }
  ],
  "tenant_permissions": []
}
```

- Set the mapping relationship between roles and users

### OpenSearch {#perm-opensearch}

- Create a user
- Create the role `monitor`, and set the following permissions:

```
PUT _plugins/_security/api/roles/monitor
{
  "description": "monitor es cluster",
  "cluster_permissions": [
    "cluster:admin/opendistro/ism/managedindex/explain",
    "cluster_monitor",
    "cluster_composite_ops_ro"
  ],
  "index_permissions": [
    {
      "index_patterns": [
        "*"
      ],
      "fls": [],
      "masked_fields": [],
      "allowed_actions": [
        "read",
        "indices_monitor"
      ]
    }
  ],
  "tenant_permissions": []
}
```

- Set the mapping relationship between roles and users

=== "Host Installation"

    Go to the `conf.d/db` directory under the DataKit installation directory, copy `elasticsearch.conf.sample` and name it `elasticsearch.conf`. Examples are as follows:
    
    ```toml
        
    [[inputs.elasticsearch]]
      ## Elasticsearch Server configuration
      # Support Basic authentication:
      # servers = ["http://user:pass@localhost:9200"]
      servers = ["http://localhost:9200"]
    
      ## collection interval
      # Unit "ns", "us" (or "µs"), "ms", "s", "m", "h"
      interval = "10s"
    
      ## HTTP timeout settings
      http_timeout = "5s"
    
      ## Distribution version: elasticsearch, opendistro, opensearch
      distribution = "elasticsearch"
    
      ## The default local is turned on, and only the current Node's own indicators are collected. If all Nodes in the cluster need to be collected, local should be set to false.
      local = true
    
      ## Set to true to collect cluster health
      cluster_health = false
    
      ## cluster health level settings, indices (default), and cluster
      # cluster_health_level = "indices"
    
      ## cluster stats can be collected when set to true.
      cluster_stats = false
    
      ## Get cluster_stats only from master Node, provided that local = true is set
      cluster_stats_only_from_master = true
    
      ## Indices to be collected, default is _ all
      indices_include = ["_all"]
    
      ## indices level, desirable values: "shards", "cluster", "indices"
      indices_level = "shards"
    
      ## node_stats supports configuration options such as "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http", "breaker"
      # Default is all
      # node_stats = ["jvm", "http"]
    
      ## HTTP Basic Authentication User Name and Password
      # username = ""
      # password = ""
    
      ## TLS Config
      tls_open = false
      # tls_ca = "/etc/telegraf/ca.pem"
      # tls_cert = "/etc/telegraf/cert.pem"
      # tls_key = "/etc/telegraf/key.pem"
      ## Use TLS but skip chain & host verification
      # insecure_skip_verify = false
    
      ## Set true to enable election
      election = true
    
      # [inputs.elasticsearch.log]
      # files = []
      # #grok pipeline script path
      # pipeline = "elasticsearch.p"
    
      [inputs.elasticsearch.tags]
        # some_tag = "some_value"
        # more_tag = "some_other_value"
    
    ```

    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.elasticsearch.tags]`:

``` toml
[inputs.elasticsearch.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
```

``` toml
[inputs.{{.InputName}}.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }} 

## Log Collection {#logging}

???+ attention

    Log collection only supports log collection on installed DataKit hosts

To collect ElasticSearch logs, open `files` in ElasticSearch.conf and write to the absolute path of the ElasticSearch log file. For example:

```toml
[[inputs.elasticsearch]]
  ...
[inputs.elasticsearch.log]
files = ["/path/to/your/file.log"]
```

When log collection is turned on, a log with a log `source` of `elasticsearch` is generated by default.

## Log pipeline Feature Cut Field Description {#pipeline}

- ElasticSearch Universal Log Cutting
  
Example of common log text:

```
[2021-06-01T11:45:15,927][WARN ][o.e.c.r.a.DiskThresholdMonitor] [master] high disk watermark [90%] exceeded on [A2kEFgMLQ1-vhMdZMJV3Iw][master][/tmp/elasticsearch-cluster/nodes/0] free: 17.1gb[7.3%], shards will be relocated away from this node; currently relocating away shards totalling [0] bytes; the node is expected to continue to exceed the high disk watermark when these relocations are complete
```

The list of cut fields is as follows:

| Field Name | Field Value                         | Description         |
| ---    | ---                            | ---          |
| time   | 1622519115927000000            | Log generation time |
| name   | o.e.c.r.a.DiskThresholdMonitor | Component name     |
| status | WARN                           | Log level     |
| nodeId | master                         | Node name     |

- ElastiSearch Search for Slow Log Cutting
  
Example of Searching for Slow Log Text: 

```
[2021-06-01T11:56:06,712][WARN ][i.s.s.query              ] [master] [shopping][0] took[36.3ms], took_millis[36], total_hits[5 hits], types[], stats[], search_type[QUERY_THEN_FETCH], total_shards[1], source[{"query":{"match":{"name":{"query":"Nariko","operator":"OR","prefix_length":0,"max_expansions":50,"fuzzy_transpositions":true,"lenient":false,"zero_terms_query":"NONE","auto_generate_synonyms_phrase_query":true,"boost":1.0}}},"sort":[{"price":{"order":"desc"}}]}], id[], 
```

The list of cut fields is as follows:

| Field Name   | Field Value              | Description             |
| ---      | ---                 | ---              |
| time     | 1622519766712000000 | Log generation time     |
| name     | i.s.s.query         | Component name         |
| status   | WARN                | Log level         |
| nodeId   | master              | Node name         |
| index    | shopping            | Index name         |
| duration | 36000000            | Request time, in ns |

- ElasticSearch Index Slow Log Cutting

Example of indexing slow log text:

```
[2021-06-01T11:56:19,084][WARN ][i.i.s.index              ] [master] [shopping/X17jbNZ4SoS65zKTU9ZAJg] took[34.1ms], took_millis[34], type[_doc], id[LgC3xXkBLT9WrDT1Dovp], routing[], source[{"price":222,"name":"hello"}]
```

The list of cut fields is as follows:

| Field Name   | Field Value              | Description             |
| ---      | ---                 | ---              |
| time     | 1622519779084000000 | Log generation time     |
| name     | i.i.s.index         | Component name         |
| status   | WARN                | Log level         |
| nodeId   | master              | Node name         |
| index    | shopping            | Index name         |
| duration | 34000000            | Request time, in ns |

