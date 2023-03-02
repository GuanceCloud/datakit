
# IIS
---

{{.AvailableArchs}}

---

Microsoft IIS collector

## Preconditions {#requirements}

Operating system requirements::

* Windows Vista and above (excluding Windows Vista)
* Windows Server 2008 R2 and above

## Configuration {#config}

Go to the `conf.d/iis` directory under the DataKit installation directory, copy `iis.conf.sample` and name it `iis.conf`. Examples are as follows:

```toml

[[inputs.iis]]
  ## (optional) collect interval, default is 15 seconds
  interval = '15s'
  ##

  [inputs.iis.log]
    files = []
    ## grok pipeline script path
    pipeline = "iis.p" 
```

After configuration, restart DataKit.

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.iis.tags]`:

``` toml
  [inputs.iis.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    # ...
```

## Measurements {#measurements}

{{ range $i, $m := .Measurements }}

{{if or (eq $m.Type "metric") (eq $m.Type "")}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}

## Log {#logging}

If you need to collect IIS logs, open the log-related configuration in the configuration, such as:

```toml
[inputs.iis.log]
    # Fill in the absolute path
    files = ["C:/inetpub/logs/LogFiles/W3SVC1/*"] 
```
