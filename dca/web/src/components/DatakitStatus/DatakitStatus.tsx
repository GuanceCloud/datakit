import { Tooltip } from "antd";
import { IDatakit } from "src/store/type";
import styles from './DatakitStatus.module.scss'
import { useTranslation } from "react-i18next";

export default function DatakitStatus({ datakit }: { datakit: IDatakit }) {
  const { t } = useTranslation()
  let state = datakit.status ? datakit.status : "unknown"
  let textColor = {
    "running": "#6CBB87",
    "upgrading": "#CACACA",
    "offline": "#DE6357",
    "restarting": "#CACACA",
    "stopped": "#bfbfbf",
  }[state]

  let statusText = {
    "running": t("status.running"),
    "offline": t("status.offline"),
    "upgrading": t("status.upgrading"),
    "stopped": t("status.stopped"),
    "restarting": t("status.restarting"),
    "unknown": t("status.unknown")
  }[state]

  return (
    <div className={styles.container}>
      <Tooltip overlayStyle={{ maxWidth: "400px" }} placement="right" title={(<>{statusText || t("status.unknown")}</>)}>
        <div className={styles.status} style={{ background: textColor || "#DE6357" }}>
          {state}
        </div>
      </Tooltip>
    </div>
  )
}