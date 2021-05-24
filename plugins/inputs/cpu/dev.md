# CPU 采集器

配置示例

```toml
[[inputs.cpu]]
  # Collect interval, default is 10 seconds(optional)
  interval = '10s'

  # Extra tags (optional)
  [inputs.cpu.tags]
    # tag1 = "a"
```

## CPU 指标采集

* cpu
    |指标|描述|数据类型|单位|
    |:--| - | -| -|
    |core_temperature|cpu core temperature|float|°C|
    |usage_guest|% CPU spent running a virtual CPU for guest operating systems.|float|%|
    |usage_guest_nice|% CPU spent running a niced guest(virtual CPU for guest operating systems).|float|%|
    |usage_idle|% CPU in the idle task.|float|%|
    |usage_iowait|% CPU waiting for I/O to complete.|float|%|
    |usage_irq|% CPU servicing hardware interrupts.|float|%|
    |usage_nice|% CPU in user mode with low priority (nice).|float|%|
    |usage_softirq|% CPU servicing soft interrupts.|float|%|
    |usage_steal|% CPU spent in other operating systems when running in a virtualized environment.|float|%|
    |usage_system|% CPU in system mode.|float|%|
    |usage_total|% CPU in total active usage, as well as (100 - usage_idle).|float|%|
    |usage_user|% CPU in user mode.|float|%|
