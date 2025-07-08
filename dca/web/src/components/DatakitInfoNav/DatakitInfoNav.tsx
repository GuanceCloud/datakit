import { Tabs } from "antd"
import { useEffect, useMemo, useState } from "react"
import { useLocation, useNavigate } from "react-router-dom"

import './DatakitInfoNav.scss'
import { IDatakit } from "src/store/type"
import { useTranslation } from "react-i18next"

export function DatakitInfoNav({ datakit }: { datakit: IDatakit }) {
    const { t } = useTranslation()

    const defaultTabItems = useMemo(() => {
        return [
            {
                label: t("label.index"),
                key: "0",
                path: "",
                types: ["host", "container"]
            },
            {
                label: t("label.run_info"),
                key: "1",
                path: "/runinfo",
                types: ["host", "container"]
            },
            {
                label: t("label.collector_config"),
                key: "2",
                path: "/config",
                types: ["host", "container"]
            },
            {
                label: t("label.pipeline"),
                key: "3",
                path: "/pipeline",
                types: ["host", "container"]
            },
            {
                label: t("label.blacklist"),
                key: "4",
                path: "/blacklist",
                types: ["host", "container"]
            },
            {
                label: t("label.log"),
                key: "5",
                path: "/log",
                types: ["host", "container"]
            },
        ]

    }, [t])
    const [key, setKey] = useState<string>("1")
    const [navMap, setNavMap] = useState<Record<string, string>>({})
    const [tabItems, setTabItems] = useState(defaultTabItems)
    const navigate = useNavigate()
    const location = useLocation()
    const pathname = location.pathname

    useEffect(() => {
        let type = datakit?.run_in_container ? "container" : "host"
        let items = defaultTabItems.filter(item => item.types.includes(type))
        setTabItems(items)
        let navItemsMap: Record<string, string> = {}
        items.forEach(item => {
            navItemsMap[item.key] = `/dashboard${item.path}`
        })

        setNavMap(navItemsMap)
    }, [datakit, t, defaultTabItems])

    useEffect(() => {
        for (let key in navMap) {
            if (pathname === navMap[key]) {
                setKey(key)
                break
            }
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [location.pathname, navMap])

    const changeKey = (key: string) => {
        if (navMap[key]) {
            navigate(navMap[key])
        } else {
            navigate("/dashboard")
        }
    }

    let navBarStyle: React.CSSProperties = {
        color: "#A7B1BD",
        fontSize: "14px",
        fontWeight: 500,
        marginBottom: "0"
    }

    return (
        <div className="nav-container">
            <Tabs
                activeKey={key}
                onChange={changeKey}
                tabBarStyle={navBarStyle}
                items={tabItems}
            >
            </Tabs>
        </div>
    )
}