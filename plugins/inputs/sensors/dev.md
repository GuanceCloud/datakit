# summary

计算机芯片温度状态监测。使用 `lm-sensors` 命令抓取温度数据。目前只支持 `Linux` 操作系统。

# prerequisite

`apt install lm-sensors -y`

# raw data sample

## sensors -u

```
coretemp-isa-0000
Adapter: ISA adapter
Package id 0:
  temp1_input: 36.000
  temp1_max: 80.000
  temp1_crit: 100.000
  temp1_crit_alarm: 0.000
Core 0:
  temp2_input: 35.000
  temp2_max: 80.000
  temp2_crit: 100.000
  temp2_crit_alarm: 0.000
Core 1:
  temp3_input: 36.000
  temp3_max: 80.000
  temp3_crit: 100.000
  temp3_crit_alarm: 0.000
Core 2:
  temp4_input: 34.000
  temp4_max: 80.000
  temp4_crit: 100.000
  temp4_crit_alarm: 0.000
Core 3:
  temp5_input: 35.000
  temp5_max: 80.000
  temp5_crit: 100.000
  temp5_crit_alarm: 0.000

acpitz-acpi-0
Adapter: ACPI interface
temp1:
  temp1_input: 27.800
  temp1_crit: 119.000
temp2:
  temp2_input: 29.800
  temp2_crit: 119.000
```

# config sample

```
[[inputs.sensors]] ## Command path of 'senssor' usually under /usr/bin/sensors # path = "/usr/bin/senssors"

    ## Gathering interval
    # interval = "10s"

    ## Command timeout
    # timeout = "3s"

    ## Customer tags, if set will be seen with every metric.
    [inputs.sensors.tags]
    	# "key1" = "value1"
    	# "key2" = "value2"

```

# metrics

## sensors

| 标签名   | 描述             |
| -------- | ---------------- |
| hostname | host name        |
| adapter  | device adapter   |
| chip     | chip id          |
| feature  | gathering target |

| 指标               | 类型       | 指标源       | 单位          | 描述                                                                               |
| ------------------ | ---------- | ------------ | ------------- | ---------------------------------------------------------------------------------- |
| tmep\*\_crit       | inputs.Int | inputs.Gauge | inputs.NCount | critical temperature of this chip, '\*' is the order number in the chip list.      |
| temp\*\_crit_alarm | inputs.Int | inputs.Gauge | inputs.NCount | alarm count, '\*' is the order number in the chip list.                            |
| temp\*\_input      | inputs.Int | inputs.Gauge | inputs.NCount | current input temperature of this chip, '\*' is the order number in the chip list. |
| tmep\*\_max        | inputs.Int | inputs.Gauge | inputs.NCount | max temperature of this chip, '\*' is the order number in the chip list.           |
