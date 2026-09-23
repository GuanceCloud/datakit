import { Tooltip } from "antd";
import { DCA_STATUS, IDatakit } from "src/store/type";
import styles from './DatakitStatus.module.scss'
import { useTranslation } from "react-i18next";

export default function DatakitStatus({ datakit }: { datakit: IDatakit }) {
  const { t } = useTranslation()
  // a row can look running while the DCA backend has no session for it: show
  // that instead of a green "running" the operator cannot act on.
  // NOTE: keep this check inline, importing the helper drags the api layer in.
  const sessionLost = datakit.status === DCA_STATUS.RUNNING && datakit.alive === false
  let state = sessionLost ? "session_lost" : (datakit.status ? datakit.status : "unknown")
  let textColor = {
    "running": "#6CBB87",
    "upgrading": "#CACACA",
    "offline": "#DE6357",
    "restarting": "#CACACA",
    "stopped": "#bfbfbf",
    "session_lost": "#FA8C16",
  }[state]

  let statusText = {
    "running": t("status.running"),
    "offline": t("status.offline"),
    "upgrading": t("status.upgrading"),
    "stopped": t("status.stopped"),
    "restarting": t("status.restarting"),
    "unknown": t("status.unknown"),
    "session_lost": t("status.session_lost")
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
