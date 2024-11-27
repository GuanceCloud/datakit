import { Tooltip } from "antd";
import { IDatakit } from "src/store/type";
import styles from './DatakitStatus.module.scss'

export default function DatakitStatus({ datakit }: { datakit: IDatakit }) {
  let state = datakit.status ? datakit.status : "unknown"
  let textColor = {
    "running": "#6CBB87",
    "upgrading": "#CACACA",
    "offline": "#DE6357",
    "restarting": "#CACACA",
    "stopped": "#bfbfbf",
  }[state]

  let statusText = {
    "running": "在线",
    "offline": "离线状态",
    "upgrading": "升级中",
    "stopped": "已停止",
    "restarting": "重启中"
  }[state]

  return (
    <div className={styles.container}>
      <Tooltip overlayStyle={{ maxWidth: "400px" }} placement="right" title={(<>{statusText || "未知状态"}</>)}>
        <div className={styles.status} style={{ background: textColor || "#DE6357" }}>
          {state}
        </div>
      </Tooltip>
    </div>
  )
}