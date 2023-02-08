
# DataKit Self-metric
---

{{.AvailableArchs}}

---

Self collector is used to collect the basic information of DataKit itself, including running environment information, CPU, memory occupation and so on.

## Preconditions {#reqirement}

None

## Configuration {#config}

The self collector runs automatically without configuration and cannot be shut down.

## Metrics {#measurements}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}


## More Readings {#more-reading}

- [Host collector](hostobject.md)
