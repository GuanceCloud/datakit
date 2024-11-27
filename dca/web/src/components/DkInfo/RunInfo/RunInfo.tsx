import { Col, Row, Space, Table } from 'antd'
import { ColumnsType } from 'antd/lib/table'
import humanformat from 'human-format'
import moment from 'moment'
import { useContext, useEffect, useState } from 'react'

import { DkInfoContext } from '../DkInfo'
import styles from './RunInfo.module.scss'
import { IDatakitStat } from 'src/store/type'
import { showDuration } from 'src/helper/helper'

interface InputStat {
  name: string
  avgCollectCost: number
  avgSize: number
  category: string
  pts_total: number
  first: string
  frequency: string
  last: string
  lastError: string
  lastErrorTime: string
  maxCollectCost: number
  feed_total: number
  instanceCount: number
  dataType: string
  crashCount: number
  p90_lat: string
  p90_pts: string
}

interface EnableInput {
  input: string
  instances: number
  panic: number
}

interface GoroutineInfo {
  finished_goroutines: number
  running_goroutines: number
  total_cost_time: string
  min_cost_time: string
  max_cost_time: string
  err_count: number
  name: string
}

interface HttpInfo {
  total_count: Number
  limited: Number
  limited_percent: Number
  "2xx": Number
  "3xx": Number
  "4xx": Number
  "5xx": Number
  max_letency: Number
  avg_latency: Number
}

interface FilterInfo {
  conditions: number
  cost: number
  cost_per_point: number
  filtered: number
  total: number
  category: string
}

interface PipelineInfo {
  Id: number
  Pt: number
  PtDrop: number
  PtError: number
  RunLastErrs: string[]
  TotalCost: number
  MetaTS: string
  Script: string
  FirstTS: string
  ScriptTS: string
  ScriptUpdateTimes: number
  Category: string
  NS: string
  Name: string
  Enable: boolean
  Deleted: boolean
  CompileError: string
}

interface IOInfo {
  drop_pts: number
  chan_usage_list: {
    cat: string
    chan_usage: string
    send_failed: string
  }[]
}
interface DKStat extends IDatakitStat {
  inputsStatus: InputStat[]
  enabledInputs: EnableInput[]
  goroutineStat: GoroutineInfo[]
  httpStat: HttpInfo[]
  filterStat: FilterInfo[]
  pipelineStat: PipelineInfo[]
  ioStat: IOInfo
}

function getDataType(category: string): string {
  if (!category) {
    return "-"
  }

  const metricMap = {
    "metric": "M",
    "custom_object": "CO",
    "object": "O",
    "logging": "L",
    "keyevent": "E",
    "tracing": "T",
    "rum": "R",
    "security": "S",
    "network": "N",
    "profiling": "P"
  }

  for (let key in metricMap) {
    if (category.indexOf(key) > -1) {
      return metricMap[key]
    }
  }
  return '-'
}

function humanFormat(value) {
  try {
    const v = humanformat(value)
    return v
  } catch (err) {
    return ""
  }
}

function showMemSize(mem: number) {
  if (!mem) {
    return ""
  }

  let unit = ['TB', 'GB', 'MB', 'KB', 'B']
  let index = unit.length - 1
  while (index > 0 && mem > 1024) {
    mem = mem / 1024
    index--
  }
  return mem.toFixed(2) + unit[index]
}

function showTimeString(time: string): string {
  if (!time) {
    return ""
  }

  let [prev, post] = time.split(".")
  if (!post || post.length <= 3) {
    return time
  }

  return prev + "." + post.slice(0, 2) + post.slice(-1)
}

export default function RunInfo() {
  const dkInfoContext = useContext(DkInfoContext)

  const { datakitStat, datakit } = dkInfoContext

  const enabledInputsColumns: ColumnsType<any> = [
    {
      title: "Input",
      dataIndex: "input",
      ellipsis: true,
      width: "180px"
    },
    {
      title: "Instances",
      dataIndex: "instances",
      ellipsis: true,
      width: "150px"
    },
    {
      title: "Crashed",
      dataIndex: "panic",
      ellipsis: true,
    }
  ]
  const inputsInfoColumns: ColumnsType<any> = [
    {
      title: 'Input',
      dataIndex: 'name',
      key: 'name',
      width: "150px",
      fixed: 'left'
    },
    {
      title: 'Cat',
      dataIndex: 'dataType',
      key: 'dataType',
      width: '65px'
    },
    {
      title: 'Feeds',
      dataIndex: 'feed_total',
      key: 'feed_total',
      width: '80px',
      render(v) {
        return humanFormat(Number(v))
      }
    },
    {
      title: 'P90Lat',
      dataIndex: 'p90_lat',
      key: 'p90_lat',
      width: '80px',
    },
    {
      title: 'P90Pts',
      dataIndex: 'p90_pts',
      key: 'p90_pts',
      width: '80px',
    },
    {
      title: 'LastFeed',
      dataIndex: 'last',
      key: 'last',
      width: '100px',
      ellipsis: true
    },
    {
      title: 'AvgCost',
      dataIndex: 'avgCollectCost',
      key: 'avgCollectCost',
      render: (text, record, index) => {
        return text ? showDuration(text) : "-"
      },
      width: '100px'
    },
  ]

  const [dkStat, setDkStat] = useState<DKStat>()

  const initDatakitInfo = async () => {
    if (!datakitStat) {
      return
    }
    const inputStats: InputStat[] = []

    datakitStat.inputs_status && Object.keys(datakitStat.inputs_status).forEach((key) => {
      let instanceCount = 1 // TODO: default 1. 
      let crashCount = 0
      let input: {
        input: string;
        instances: number;
        panic: number;
      } | undefined

      if (datakitStat.enabled_inputs) { // for old datakit api
        input = datakitStat.enabled_inputs.find((i) => i.input === key)
      } else if (datakitStat.enabled_input_list) {
        input = datakitStat.enabled_input_list[key]
      }

      if (input) {
        instanceCount = input.instances
        crashCount = input.panic
      }

      const info = datakitStat.inputs_status[key]
      inputStats.push({
        name: key,
        avgCollectCost: info.avg_collect_cost,
        avgSize: info.avg_size,
        category: info.category,
        pts_total: info.pts_total,
        first: moment(info.first).fromNow(),
        frequency: info.frequency || '-',
        last: moment(info.last).fromNow(),
        lastError: info.last_error,
        lastErrorTime: info.last_error_ts,
        maxCollectCost: info.max_collect_cost,
        feed_total: info.feed_total,
        instanceCount,
        dataType: getDataType(info.category),
        crashCount,
        p90_lat: info.p90_lat,
        p90_pts: info.p90_pts,
      })
    })

    const goroutineStat: GoroutineInfo[] = []
    if (datakitStat.goroutine_stats?.Items) {
      Object.keys(datakitStat.goroutine_stats.Items).forEach((v) => {
        goroutineStat.push({ ...datakitStat.goroutine_stats.Items[v], name: v })
      })
    }

    const httpStat: HttpInfo[] = []
    if (datakitStat.http_metrics) {
      Object.keys(datakitStat.http_metrics).forEach((v) => {
        httpStat.push({ ...datakitStat.http_metrics[v], path: v })
      })
    }

    const filterStat: FilterInfo[] = []
    if (datakitStat.filter_stats && datakitStat.filter_stats.rule_stats) {
      Object.keys(datakitStat.filter_stats.rule_stats).forEach((v) => {
        filterStat.push({ ...datakitStat.filter_stats.rule_stats[v], category: v })
      })
    }

    const pipelineStat: PipelineInfo[] = []
    datakitStat.pl_stats && datakitStat.pl_stats.forEach((v, index) => {
      pipelineStat.push({ ...v, Id: index })
    })

    const ioStat: IOInfo = {
      drop_pts: 0,
      chan_usage_list: []
    }

    if (datakitStat.io_stats && datakitStat.io_stats.chan_usage) {
      ioStat.drop_pts = datakitStat.io_stats.drop_pts
      Object.keys(datakitStat.io_stats.chan_usage).forEach((k) => {
        let [x, y] = datakitStat.io_stats.chan_usage[k]
        let cat = getDataType(k)
        let failPts = datakitStat.io_stats[`${cat}_fail_pts`] || 0
        let sendPts = datakitStat.io_stats[`${cat}_send_pts`] || 0
        let info = {
          cat,
          chan_usage: `${x} / ${humanFormat(y)}`,
          send_failed: `${humanFormat(sendPts)} / ${humanFormat(failPts)}`
        }
        cat !== "-" && ioStat.chan_usage_list.push(info)
      })
    }

    setDkStat({
      ...datakitStat,
      inputsStatus: inputStats,
      uptime: showTimeString(datakitStat.uptime),
      enabledInputs: Object.values(datakitStat.enabled_input_list || {}),
      goroutineStat: goroutineStat,
      httpStat,
      filterStat,
      pipelineStat,
      ioStat
    })
  }

  useEffect(() => {
    datakitStat && initDatakitInfo()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [datakitStat])

  return (
    <div className={styles.info}>
      {datakitStat && dkStat ?
        <>
          <div className={styles.detail}>
            <div className={styles.row}>
              <div className={styles.basic} style={{ flex: 1 }}>
                <div className={styles.title}>
                  <span>Basic Info</span>
                </div>
                <div className={styles.content}>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Hostname</Col>
                    <Col span={18}>{dkStat.hostname}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> OS/ARCH</Col>
                    <Col span={18}>{dkStat.os_arch}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> DataKit Version</Col>
                    <Col span={18}>{datakit?.version}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> IP Address</Col>
                    <Col span={18}>{datakit?.ip || "-"}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Runtime ID</Col>
                    <Col span={18}>{datakit?.runtime_id}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Run Mode</Col>
                    <Col span={18}>{datakitStat?.usage_trace?.run_mode || "-"}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Run in Container</Col>
                    <Col span={18}>{datakit?.run_in_container ? "yes" : "no"}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Usage Cores</Col>
                    <Col span={18}>{datakitStat?.usage_trace?.usage_cores}</Col>
                  </Row>
                </div>
              </div>
              <div className={styles.basic} style={{ flex: 1 }}>
                <div className={styles.title}>
                  <Space>
                    <span></span>
                  </Space>
                </div>
                <div className={styles.content}>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Uptime</Col>
                    <Col span={18}>{dkStat.uptime}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Goroutines</Col>
                    <Col span={18}>{dkStat.golang_runtime?.goroutines}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Resource Limit</Col>
                    <Col span={18}>{dkStat.resource_limit}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> CPU(%)</Col>
                    <Col span={18}>{dkStat.datakit_runtime_info?.cpu_usage ? dkStat.datakit_runtime_info?.cpu_usage : 0}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> SysMem</Col>
                    <Col span={18}>{showMemSize(dkStat.golang_runtime?.total_sys)}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Mem </Col>
                    <Col span={18}>{showMemSize(dkStat.golang_runtime?.heap_alloc)}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> OpenFiles</Col>
                    <Col span={18}>{dkStat.open_files}</Col>
                  </Row>
                  <Row gutter={16}>
                    <Col span={6} className={styles.label}> Elected</Col>
                    <Col span={18}>{dkStat.elected}</Col>
                  </Row>
                </div>
              </div>

            </div>
            <div className={styles.row}>
              <div className={styles.table} style={{ width: "30%" }}>
                <div className={styles.title}>
                  <Space>
                    {/* <span className="fth-iconfont-Operation"></span> */}
                    <span>Enabled Inputs({dkStat.enabledInputs.length} inputs)</span>
                  </Space>
                </div>
                <div className={styles.content}>
                  <Table
                    size={'small'}
                    // style={{ "width": "500px" }}
                    columns={enabledInputsColumns}
                    scroll={{ x: 100, y: 500 }}
                    dataSource={dkStat.enabledInputs}
                    className="run-info-table"
                    rowKey={'input'}
                    pagination={{ hideOnSinglePage: true, pageSize: dkStat.enabledInputs.length }}
                  />
                </div>
              </div>
              <div className={styles.table} style={{ width: "calc(70% - 3calc(70% - 20px)0px)" }}>
                <div className={styles.title}>
                  <Space>
                    <span>Inputs Info({dkStat.inputsStatus.length} inputs)</span>
                  </Space>
                </div>
                <div className={styles.content}>
                  <Table
                    size={'small'}
                    columns={inputsInfoColumns}
                    scroll={{ y: 500 }}
                    dataSource={dkStat.inputsStatus}
                    className="run-info-table"
                    rowKey={'name'}
                    pagination={{ hideOnSinglePage: true, pageSize: dkStat.inputsStatus.length }}
                  />
                </div>
              </div>
            </div>
          </div>
        </>
        :
        <div>no data</div>
      }
    </div>
  )
}
