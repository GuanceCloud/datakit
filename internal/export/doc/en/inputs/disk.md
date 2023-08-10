
# Disk
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

Disk collector is used to collect disk information, such as disk storage space, inodes usage, etc.

## Preconditions {#requirements}

None


## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:


​    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Supports modifying configuration parameters as environment variables:
    
    | Environment Variable Name                            | Corresponding Configuration Parameter Item       | Parameter Example                                                                                 |
    | ---                                   | ---                    | ---                                                                                      |
    | `ENV_INPUT_DISK_EXCLUDE_DEVICE`       | `exclude_device`       | `"/dev/loop0","/dev/loop1"`, separated by English commas                      |
    | `ENV_INPUT_DISK_EXTRA_DEVICE`         | `extra_device`         | `"/nfsdata"`， separated by English commas                        |
    | `ENV_INPUT_DISK_TAGS`                 | `tags`                 | `tag1=value1,tag2=value2`; If there is a tag with the same name in the configuration file, it will be overwritten                             |
    | `ENV_INPUT_DISK_ONLY_PHYSICAL_DEVICE` | `only_physical_device` | Ignore non-physical disks (such as network disk, NFS, etc., only collect local hard disk/CD ROM/USB disk, etc.) and give a string value at will|
    | `ENV_INPUT_DISK_INTERVAL`             | `interval`             | `10s`                                                                                    |

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.disk.tags]`:

``` toml
 [inputs.disk.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
