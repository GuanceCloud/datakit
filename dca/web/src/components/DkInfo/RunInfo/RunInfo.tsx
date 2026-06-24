import { Space, Table } from 'antd'
import { ColumnsType } from 'antd/lib/table'
import humanformat from 'human-format'
import moment from 'moment'
import { ReactNode, useContext, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

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
  if (typeof value === "number" && !Number.isFinite(value)) {
    return "-"
  }

  try {
    const v = humanformat(value)
    return String(v) === "NaN" ? "-" : v
  } catch (err) {
    return "-"
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

function showRelativeTime(time: string): string {
  if (!time) {
    return "-"
  }

  const m = moment(time)
  if (!m.isValid() || m.year() < 2000 || m.isAfter(moment().add(5, "minutes"))) {
    return "-"
  }

  return m.fromNow()
}

function displayValue(value: ReactNode): ReactNode {
  if (value === undefined || value === null || value === "" || value === "NaN") {
    return "-"
  }

  if (typeof value === "number" && !Number.isFinite(value)) {
    return "-"
  }

  if (typeof value === "boolean") {
    return value ? "true" : "false"
  }

  return value
}

export default function RunInfo() {
  const { t } = useTranslation()
  const dkInfoContext = useContext(DkInfoContext)

  const { datakitStat, datakit } = dkInfoContext

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
      ellipsis: true,
      render: (text) => displayValue(text)
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
        first: showRelativeTime(info.first),
        frequency: info.frequency || '-',
        last: showRelativeTime(info.last),
        lastError: info.last_error,
        lastErrorTime: info.last_error_ts,
        maxCollectCost: info.max_collect_cost,
        feed_total: info.feed_total,
        instanceCount,
        dataType: getDataType(info.category),
	        crashCount,
	        p90_lat: displayValue(info.p90_lat) as string,
	        p90_pts: displayValue(info.p90_pts) as string,
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

  const renderInfoItem = (label: string, value: ReactNode) => (
    <div className={styles.infoItem}>
      <div className={styles.infoLabel}>{label}</div>
      <div className={styles.infoValue}>
        <span>{displayValue(value)}</span>
      </div>
    </div>
  )

  const totalCrashCount = dkStat?.enabledInputs.reduce((count, input) => count + (input.panic || 0), 0) || 0
  const runningStatus = datakit?.status === "running"

  return (
    <div className={styles.info}>
      {datakitStat && dkStat ?
        <div className={styles.detail}>
          <section className={styles.overview}>
            <div className={styles.sectionHeader}>
              <span>{t("run_info.overview")}</span>
              <span className={styles.refreshTime}>{t("run_info.stats_realtime")}</span>
            </div>
            <div className={styles.metrics}>
              <div className={`${styles.metricItem} ${runningStatus ? styles.statusOk : styles.statusWarn}`}>
                <div className={styles.metricLabel}>{t("status_text")}</div>
                <div className={styles.metricValue}>{datakit?.status || "unknown"}</div>
              </div>
              <div className={styles.metricItem}>
                <div className={styles.metricLabel}>Uptime</div>
                <div className={styles.metricValue}>{displayValue(dkStat.uptime)}</div>
              </div>
              <div className={styles.metricItem}>
                <div className={styles.metricLabel}>CPU(%)</div>
                <div className={styles.metricValue}>{dkStat.datakit_runtime_info?.cpu_usage || 0}</div>
              </div>
              <div className={styles.metricItem}>
                <div className={styles.metricLabel}>Mem</div>
                <div className={styles.metricValue}>{displayValue(showMemSize(dkStat.golang_runtime?.heap_alloc))}</div>
              </div>
              <div className={styles.metricItem}>
                <div className={styles.metricLabel}>Goroutines</div>
                <div className={styles.metricValue}>{displayValue(dkStat.golang_runtime?.goroutines)}</div>
              </div>
              <div className={styles.metricItem}>
                <div className={styles.metricLabel}>OpenFiles</div>
                <div className={styles.metricValue}>{displayValue(dkStat.open_files)}</div>
              </div>
              <div className={styles.metricItem}>
                <div className={styles.metricLabel}>Inputs</div>
                <div className={styles.metricValue}>{dkStat.enabledInputs.length} / {dkStat.inputsStatus.length}</div>
              </div>
              <div className={styles.metricItem}>
                <div className={styles.metricLabel}>Election</div>
                <div className={styles.metricValue}>{displayValue(dkStat.elected)}</div>
              </div>
            </div>
          </section>

          <section className={styles.basicPanel}>
            <div className={styles.sectionHeader}>
              <span>{t("run_info.basic_info")}</span>
            </div>
            <div className={styles.infoGrid}>
              {renderInfoItem("Hostname", dkStat.hostname)}
              {renderInfoItem("OS/ARCH", dkStat.os_arch)}
              {renderInfoItem("DataKit Version", datakit?.version)}
              {renderInfoItem("IP Address", datakit?.ip)}
              {renderInfoItem("Runtime ID", datakit?.runtime_id)}
              {renderInfoItem("Run Mode", datakitStat?.usage_trace?.run_mode)}
              {renderInfoItem("Run in Container", datakit?.run_in_container ? t("yes") : t("no"))}
              {renderInfoItem("Usage Cores", datakitStat?.usage_trace?.usage_cores)}
              {renderInfoItem("Resource Limit", dkStat.resource_limit)}
              {renderInfoItem("SysMem", showMemSize(dkStat.golang_runtime?.total_sys))}
            </div>
          </section>

          <div className={styles.tablesLayout}>
            <section className={styles.enabledPanel}>
              <div className={styles.sectionHeader}>
                <span>{t("run_info.enabled_collectors")}</span>
                <span className={styles.countText}>{t("run_info.inputs_count", { count: dkStat.enabledInputs.length })}</span>
              </div>
              <div className={totalCrashCount > 0 ? styles.crashAlert : styles.crashOk}>
                {totalCrashCount > 0 ? t("run_info.crash_count", { count: totalCrashCount }) : t("run_info.no_crash_record")}
              </div>
              <div className={styles.inputList}>
                {dkStat.enabledInputs.length > 0 ? dkStat.enabledInputs.map((input) => (
                  <div className={styles.inputItem} key={input.input}>
                    <div>
                      <div className={styles.inputName}>{input.input}</div>
                      <div className={styles.inputMeta}>{t("run_info.instances_count", { count: input.instances || 0 })}</div>
                    </div>
                    <span className={input.panic > 0 ? styles.crashedTag : styles.normalTag}>
                      {input.panic > 0 ? t("run_info.crashed_count", { count: input.panic }) : t("run_info.normal")}
                    </span>
                  </div>
                )) : (
                  <div className={styles.emptyText}>{t("run_info.no_enabled_collectors")}</div>
                )}
              </div>
            </section>

            <section className={styles.inputsPanel}>
              <div className={styles.tableHeader}>
                <div className={styles.sectionHeader}>
                  <Space>
                    <span>{t("run_info.collector_metrics")}</span>
                    <span className={styles.countText}>{t("run_info.inputs_count", { count: dkStat.inputsStatus.length })}</span>
                  </Space>
                </div>
                <span className={styles.refreshTime}>{t("run_info.last_feed_zero_time_filtered")}</span>
              </div>
              <Table
                size={'small'}
                columns={inputsInfoColumns}
                scroll={{ y: 500 }}
                dataSource={dkStat.inputsStatus}
                className="run-info-table"
                rowKey={'name'}
                pagination={{ hideOnSinglePage: true, pageSize: dkStat.inputsStatus.length }}
              />
            </section>
          </div>
        </div>
        :
        <div>no data</div>
      }
    </div>
  )
}
