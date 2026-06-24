import { Button, Dropdown, Input, Menu, MenuProps, Select, Switch, Tag } from 'antd'
import { useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { DkInfoContext } from '../DkInfo'
import styles from './Log.module.scss'

import { ClearOutlined, DownloadOutlined, DownOutlined, PauseOutlined, PlayCircleOutlined, ReloadOutlined } from '@ant-design/icons'
import { downloadLogFile, getQueryPath } from "src/api/api"
import { alertError } from "src/helper/helper"
import { useIsAdmin } from 'src/hooks/useIsAdmin'
import { useTranslation } from 'react-i18next'

const MAX_LINE_COUNT = 1000 // max lines of log to show

type ConnectionStatus = "connecting" | "connected" | "disconnected"
type LevelFilter = "ALL" | "ERROR" | "WARN" | "DEBUG" | "INFO"

const getLogLineClass = (line: string) => {
    if (/\bERROR\b/.test(line)) {
        return styles["log-line-error"]
    }
    if (/\bWARN\b/.test(line)) {
        return styles["log-line-warn"]
    }
    return ""
}

const normalizeLogLines = (logs: string[]) => {
    return logs.flatMap((log) => log.split(/\r?\n/)).filter((line) => line.length > 0)
}

const maskSensitiveLogLine = (line: string) => {
    return line
        .replace(/([?&](?:token|key|kv|secret|password|passwd|auth|credential)=)([^&\s]+)/gi, "$1******")
        .replace(/((?:token|key|kv|secret|password|passwd|auth|credential)\s*[:=]\s*)(["']?)([^"',\s]+)/gi, "$1$2******")
}

const matchesLevelFilter = (line: string, levelFilter: LevelFilter) => {
    return levelFilter === "ALL" || new RegExp(`\\b${levelFilter}\\b`).test(line)
}

const getConnectionStatusColor = (status: ConnectionStatus) => {
    switch (status) {
        case "connected":
            return "success"
        case "connecting":
            return "processing"
        case "disconnected":
            return "error"
    }
}

const renderLogLineContent = (line: string, keyword: string) => {
    if (!keyword) {
        return line
    }

    const lowerLine = line.toLowerCase()
    const parts: React.ReactNode[] = []
    let cursor = 0
    let matchIndex = lowerLine.indexOf(keyword)
    while (matchIndex >= 0) {
        if (matchIndex > cursor) {
            parts.push(line.slice(cursor, matchIndex))
        }
        const matchEnd = matchIndex + keyword.length
        parts.push(
            <mark className={styles["log-keyword-match"]} key={`${matchIndex}-${matchEnd}`}>
                {line.slice(matchIndex, matchEnd)}
            </mark>
        )
        cursor = matchEnd
        matchIndex = lowerLine.indexOf(keyword, cursor)
    }

    if (cursor < line.length) {
        parts.push(line.slice(cursor))
    }

    return parts
}

export default function Log() {
    const { t } = useTranslation()
    const { datakit } = useContext(DkInfoContext)
    const isAdmin = useIsAdmin()
    const logTypes = useMemo(() => ({
        "1": { type: "log", description: t("log.description.datakit") },
        "2": { type: "gin.log", description: t("log.description.gin") }
    }), [t])
    const [selectedLogTypeKey, setSelectedLogTypeKey] = useState("1")

    const [logList, setLogList] = useState<string[]>([])
    const [isPaused, setIsPaused] = useState(false)
    const [isFollowing, setIsFollowing] = useState(true)
    const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>("connecting")
    const [reconnectToken, setReconnectToken] = useState(0)
    const [searchText, setSearchText] = useState("")
    const [levelFilter, setLevelFilter] = useState<LevelFilter>("ALL")
    const contentRef = useRef<HTMLDivElement | null>(null)
    const pendingLogsRef = useRef<string[]>([])
    const isPausedRef = useRef(false)
    const connectionStatusRef = useRef<ConnectionStatus>("connecting")
    const activeConnectionIDRef = useRef(0)

    const updateConnectionStatus = useCallback((status: ConnectionStatus) => {
        connectionStatusRef.current = status
        setConnectionStatus(status)
    }, [])

    const logLines = useMemo(() => normalizeLogLines(logList), [logList])
    const normalizedSearchText = searchText.trim().toLowerCase()
    const filteredLogLines = useMemo(() => {
        return logLines.filter((line) => {
            if (!matchesLevelFilter(line, levelFilter)) {
                return false
            }
            if (!normalizedSearchText) {
                return true
            }
            return line.toLowerCase().includes(normalizedSearchText)
        })
    }, [levelFilter, logLines, normalizedSearchText])
    const matchCount = normalizedSearchText ? filteredLogLines.length : 0

    const appendLogs = useCallback((logs: string[]) => {
        setLogList((current) => {
            return normalizeLogLines([...current, ...logs]).map(maskSensitiveLogLine).slice(-MAX_LINE_COUNT)
        })
    }, [])

    const logType = logTypes[selectedLogTypeKey] || logTypes["1"]

    const selectLogType: MenuProps['onClick'] = e => {
        setSelectedLogTypeKey(e.key)
    }

    useEffect(() => {
        if (!datakit) return
        setLogList([])
        pendingLogsRef.current = []
        const connectionID = activeConnectionIDRef.current + 1
        activeConnectionIDRef.current = connectionID
        const isActiveConnection = () => activeConnectionIDRef.current === connectionID
        updateConnectionStatus("connecting")
        const conn = new WebSocket(getQueryPath("/api/datakit/ws/log", {
            datakit_id: datakit.id,
            type: logType.type,
        }))

        conn.addEventListener("open", function () {
            if (!isActiveConnection()) return
            updateConnectionStatus("connected")
        })
        conn.addEventListener("close", function () {
            if (!isActiveConnection()) return
            updateConnectionStatus("disconnected")
        })
        conn.addEventListener("error", function () {
            if (!isActiveConnection()) return
            updateConnectionStatus("disconnected")
        })
        conn.addEventListener("message", function (event) {
            if (!isActiveConnection()) return
            if (connectionStatusRef.current !== "connected") {
                updateConnectionStatus("connected")
            }
            if (isPausedRef.current) {
                pendingLogsRef.current = [...pendingLogsRef.current, event.data].slice(-MAX_LINE_COUNT)
                return
            }
            appendLogs([event.data])
        })
        return () => {
            if (isActiveConnection()) {
                activeConnectionIDRef.current += 1
            }
            conn.close()
        }
    }, [logType.type, datakit, appendLogs, reconnectToken, updateConnectionStatus])

    useEffect(() => {
        const content = contentRef.current
        if (!content || !isFollowing) {
            return
        }

        content.scrollTop = content.scrollHeight
        }, [filteredLogLines, isFollowing])

    if (!datakit) {
        return <div>no datakit</div>
    }

    const togglePause = () => {
        setIsPaused((paused) => {
            const next = !paused
            isPausedRef.current = next
            if (!next && pendingLogsRef.current.length > 0) {
                appendLogs(pendingLogsRef.current)
                pendingLogsRef.current = []
            }
            if (!next && connectionStatusRef.current === "disconnected") {
                setReconnectToken((token) => token + 1)
            }
            return next
        })
    }

    const clearLogs = () => {
        setLogList([])
        pendingLogsRef.current = []
    }

    const reconnectLogStream = () => {
        setReconnectToken((token) => token + 1)
    }

    const downloadLog = async (type: string) => {
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
            <div className={styles["toolbar"]}>
                <div className={styles["toolbar-left"]}>
                    <Dropdown overlay={menu}>
                        <div className={styles["menu"]}>
                            <Button size={"middle"}>
                                {logType.type}
                            </Button>
                            <DownOutlined />
                        </div>
                    </Dropdown>
                    <div className={styles["description"]}>{logType.description}</div>
                    <Tag className={styles["status"]} color={getConnectionStatusColor(connectionStatus)}>{t(`log.connection.${connectionStatus}`)}</Tag>
                    {connectionStatus === "disconnected" && (
                        <Button
                            type="text"
                            size="small"
                            title={t("log.action.reconnect")}
                            icon={<ReloadOutlined />}
                            onClick={reconnectLogStream}
                        />
                    )}
                    <span className={styles["meta"]}>{t("log.meta.showing", { shown: filteredLogLines.length, total: logLines.length })}</span>
                    {normalizedSearchText && <span className={styles["meta"]}>{t("log.meta.matches", { count: matchCount })}</span>}
                </div>
                <div className={styles["toolbar-right"]}>
                    <Input
                        className={styles["search"]}
                        allowClear
                        size="small"
                        placeholder={t("log.search.placeholder")}
                        value={searchText}
                        onChange={(event) => setSearchText(event.target.value)}
                    />
                    <Select
                        className={styles["level-filter"]}
                        size="small"
                        value={levelFilter}
                        onChange={setLevelFilter}
                        options={[
                            { label: t("all"), value: "ALL" },
                            { label: "ERROR", value: "ERROR" },
                            { label: "WARN", value: "WARN" },
                            { label: "DEBUG", value: "DEBUG" },
                            { label: "INFO", value: "INFO" },
                        ]}
                    />
                    <span className={styles["meta"]}>{t("log.meta.follow")}</span>
                    <Switch size="small" checked={isFollowing} onChange={setIsFollowing} />
                    <Button
                        type="default"
                        size="small"
                        title={isPaused ? t("log.action.resume") : t("log.action.pause")}
                        icon={isPaused ? <PlayCircleOutlined /> : <PauseOutlined />}
                        onClick={togglePause}
                    />
                    <Button
                        type="default"
                        size="small"
                        title={t("log.action.clear")}
                        icon={<ClearOutlined />}
                        onClick={clearLogs}
                    />
                {
                    isAdmin &&
                        <Button type="default" size="small" onClick={() => downloadLog(logType.type)}>
                            <DownloadOutlined />
                            <span>{t("export")}</span>
                        </Button>
                }
                </div>
            </div>
            <div className={styles["content"]} ref={contentRef}>
                {filteredLogLines.map((line, index) => (
                        <div className={`${styles["log-line"]} ${getLogLineClass(line)}`} key={`${index}-${line}`}>
                            <span className={styles["log-line-number"]}>{index + 1}</span>
                            <pre className={styles["log-line-content"]}>{renderLogLineContent(line, normalizedSearchText)}</pre>
                        </div>
                    ))}
            </div>
        </div>
    )
}
