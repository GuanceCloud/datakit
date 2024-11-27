import { Button, Dropdown, Menu, MenuProps } from 'antd'
import { useContext, useEffect, useState } from 'react'
import { DkInfoContext } from '../DkInfo'
import styles from './Log.module.scss'

import { DownloadOutlined, DownOutlined } from '@ant-design/icons'
import { downloadLogFile } from "src/api/api"
import { alertError } from "src/helper/helper"
import DCAEditor from 'src/components/DCAEditor/DCAEditor'
import { useIsAdmin } from 'src/hooks/useIsAdmin'

const logTypes = {
    "1": { type: "log", description: "DataKit 运行日志" },
    "2": { type: "gin.log", description: "gin 运行日志" }
}

const MAX_LINE_COUNT = 1000 // max lines of log to show

export default function Log() {
    const { datakit } = useContext(DkInfoContext)
    const isAdmin = useIsAdmin()
    const [logType, setLogType] = useState(logTypes["1"])

    const [logList, setLogList] = useState<string[]>([])

    const selectLogType: MenuProps['onClick'] = e => {
        setLogType(logTypes[e.key])
    }

    useEffect(() => {
        if (!datakit) return
        setLogList([])
        const conn = new WebSocket(`/api/datakit/ws/log?datakit_id=${datakit.id}&&type=${logType.type}`)

        conn.addEventListener("error", function (event) {
            console.log("websocket error:", event)
        })
        conn.addEventListener("message", function (event) {
            console.log("websocket message:", event.data)
            setLogList((logs) => {
                logs.push(event.data)
                if (logs.length > MAX_LINE_COUNT) {
                    logs = logs.slice(-MAX_LINE_COUNT)
                }
                return [...logs]
            })
        })
        return () => conn.close()
    }, [logType, datakit])

    if (!datakit) {
        return <div>no datakit</div>
    }
    const downloadLog = async (type: string) => {
        console.log("download log", type)
        let err = await downloadLogFile(datakit, type)

        if (err) {
            alertError(err)
        }
    }

    const menu = (
        <Menu
            onClick={selectLogType}
            items={
                Object.entries(logTypes).map(([k, v]) => {
                    return { label: v.type, key: k }
                })
            }
        />
    )

    return (
        <div className={styles["container"]}>
            <div className={styles["log"]}>
                <div className={styles["log-type"]}>
                    <Dropdown overlay={menu}>
                        <div className={styles["menu"]}>
                            <Button size={"middle"}>
                                {logType.type}
                            </Button>
                            <DownOutlined />
                        </div>
                    </Dropdown>
                    <div className={styles["description"]}>{logType.description}</div>
                </div>
                {
                    isAdmin &&
                    <div className={styles["log-export"]}>
                        <Button type="default" size="small" onClick={() => downloadLog(logType.type)}>
                            <DownloadOutlined />
                            <span>导出</span>
                        </Button>
                    </div>
                }
            </div>
            <div className={styles["content"]}>
                <DCAEditor value={logList.join("\n")} editorOptions={{ readOnly: true }} />
            </div>
        </div>
    )
}