import { CopyOutlined, LoadingOutlined, SearchOutlined } from '@ant-design/icons'
import { useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom';
import moment from 'moment';
import { App, Avatar, Button, Input, Space, Spin, Table, TableColumnsType, Typography, message } from 'antd'
import { connect, ConnectedProps } from 'react-redux';

import { alertError, isContainerMode, isDatakitManagement, isDatakitUpgradeable, isLoadingStatus, runJob } from 'src/helper/helper';
import styles from './DkList.module.scss'
import { IDatakit, IWorkspace, PageInfo, PageQuery } from 'src/store/type'
import { update } from '../../store/datakit/datakit';
import { RootState } from 'src/store';
import { useAppSelector } from 'src/hooks'
import { DashboardContext, getOSIcon } from 'src/pages/Dashboard/Dashboard';
import { useLazyGetDatakitListQuery, useLazyReloadDatakitQuery, useLazyUpgradeDatakitQuery, useLazyGetDatakitListByIDQuery } from 'src/store/datakitApi';
import DatakitStatus from '../DatakitStatus/DatakitStatus';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;
const maxRequestNumber = 10

interface DatakitListProps {
  workspace?: IWorkspace
}

type PropsFromRedux = ConnectedProps<typeof connector>
interface Props extends PropsFromRedux, DatakitListProps { }

interface DataKitDataType extends IDatakit {
  key: React.Key;
  name: string;
  state: string;
  is_container: boolean;
}

function getDatakitStatus(datakit: IDatakit): string {
  // let diffSeconds = moment().diff(moment(datakit.lastUpdateTime * 1000), 'seconds')
  // let status = diffSeconds < 600 ? "online" : "offline"


  return "online"
}

function DatakitList({ updateDatakits }: Props) {
  const { modal } = App.useApp()
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [timer, setTimer] = useState<NodeJS.Timeout>()
  const [searchName, setSearchName] = useState("")
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [enableSelection, setEnableSelection] = useState(false)
  const [loadingDatakits, setLoadingDatakits] = useState<Record<string, boolean>>({})
  const [pageQuery, setPageQuery] = useState<PageQuery>(
    {
      pageIndex: 1,
      pageSize: 20
    }
  )
  const [pageInfo, setPageInfo] = useState<PageInfo>({
    count: 0,
    pageSize: 10,
    pageIndex: 1,
    totalCount: 0
  })

  const datakits = useAppSelector((state) => state.datakit.value)
  const navigate = useNavigate()

  const { currentWorkspace, latestDatakitVersion } = useContext(DashboardContext)

  const [queryDatakitList, { currentData: datakitListResponse, isFetching: isFetchingDatakitList, isError: isErrorDatakitList }] = useLazyGetDatakitListQuery()
  const [reloadDatakit] = useLazyReloadDatakitQuery()

  const queryParams = useMemo(() => { return { pageIndex: pageQuery.pageIndex, pageSize: pageQuery.pageSize, search: searchName } }, [pageQuery, searchName])
  const initDatakitList = async () => {
    setSelectedRowKeys([])
    queryDatakitList({ ...queryParams, minLastUpdateTime: moment().subtract(1, 'days').unix() }) // use memo or will hang in fetching.
  }
  const [upgradeDatakit] = useLazyUpgradeDatakitQuery()
  const [getDatakitListByID] = useLazyGetDatakitListByIDQuery()

  useEffect(() => {
    if (isErrorDatakitList || !datakitListResponse?.success) {
      setLoading(false)
      updateDatakits([])
      setPageInfo({ ...pageInfo, count: 0, totalCount: 0 })
      return alertError(datakitListResponse?.message)
    }
    let datakits = datakitListResponse.content.data.map<IDatakit>((datakit) => {
      if (isLoadingStatus(datakit)) {
        setLoadingDatakits({ ...loadingDatakits, [datakit.id]: true })
      }

      return {
        ...datakit,
        state: getDatakitStatus(datakit),
      }
    })

    datakitListResponse.content.pageInfo && setPageInfo(datakitListResponse.content.pageInfo)
    updateDatakits(datakits)
    setLoading(false)

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [datakitListResponse, isErrorDatakitList])

  useEffect(() => {
    if (timer) {
      clearInterval(timer)
    }

    let ids = Object.keys(loadingDatakits)
    if (ids.length > 0) {
      const t = setInterval(() => {
        getDatakitListByID({ ids: ids.join(",") }).unwrap().then((res) => {
          if (res.success) {
            let dks = res.content
            let dkMap = new Map()
            for (let dk of dks) {
              dkMap.set(dk.id, dk)
            }

            let newDatakits = datakits.map((dk) => {
              if (dkMap.has(dk.id)) {
                let newDK = dkMap.get(dk.id)
                if (!isLoadingStatus(newDK)) {
                  delete loadingDatakits[dk.id]
                }
                return newDK
              }
              return dk
            })
            updateDatakits(newDatakits)
            setLoadingDatakits({ ...loadingDatakits })
          }
        })
      }, 10000)

      setTimer(t)
      return () => {
        clearInterval(t)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadingDatakits, datakits])

  const batchUpgrade = async () => {
    const upgradeDatakits: IDatakit[] = []

    for (let k of selectedRowKeys) {
      let dk = datakits.find((d) => d.id === k)
      if (dk && isDatakitUpgradeable(dk, latestDatakitVersion)) {
        upgradeDatakits.push(dk)
      }
    }

    if (upgradeDatakits.length === 0) {
      return
    }

    modal.confirm({
      title: t("datakit.upgrade"),
      content: t("datakit.upgrade_datakit_message", { count: upgradeDatakits.length }),
      onOk: async () => {
        return runJob(maxRequestNumber, upgradeDatakits, (dk) => {
          return upgradeSingleDatakit(dk)
        }).then((res) => {
          console.log("batch upgrade res: ", res)
        })
      }
    })
  }

  const upgradeSingleDatakit = async (dk: IDatakit) => {
    setLoadingDatakits((state) => {
      return { ...state, [dk.id]: true }
    })
    return upgradeDatakit(dk).unwrap().then((res) => {
      if (res.success) {
        message.success(t("upgrade_datakit_success"))
        updateDatakits(
          datakits.map((d) => {
            if (d.id === dk.id) {
              return { ...d, status: "upgrading" }
            }
            return d
          }))
      }
      return res
    }).catch((err) => {
      alertError(err)
      return err
    })
  }
  const upgrade = async (dk: IDatakit) => {
    if (!dk) {
      return alertError(t("select_datakit"))
    }
    const isLatest = dk.version === latestDatakitVersion
    modal.confirm({
      title: t("upgrade_datakit"),
      content: `${isLatest ? t("version_is_latest") + ", " : ""}${t("confirm_upgrade_datakit")}`,
      onOk: () => {
        upgradeSingleDatakit(dk)
      }
    })
  }

  const reloadSingleDatakit = async (dk: IDatakit) => {
    setLoadingDatakits((state) => {
      return { ...state, [dk.id]: true }
    })
    return reloadDatakit(dk).unwrap().then((res) => {
      if (res.success) {
        message.success(t("reload_datakit_success"))
        updateDatakits(
          datakits.map((d) => {
            if (d.id === dk.id) {
              return { ...d, status: "restarting" }
            }
            return d
          }))
      }
      return res
    }).catch((err) => {
      alertError(err)
      return err
    })
  }

  const batchReload = async () => {
    const reloadDatakits: IDatakit[] = []

    for (let k of selectedRowKeys) {
      let dk = datakits.find((d) => d.id === k)
      if (dk && isDatakitManagement(dk) && !isContainerMode(dk)) {
        reloadDatakits.push(dk)
      }
    }

    if (reloadDatakits.length === 0) {
      return
    }

    modal.confirm({
      title: t("reload_datakit"),
      content: t("reload_datakit_message", { count: reloadDatakits.length }),
      onOk: async () => {
        return runJob(maxRequestNumber, reloadDatakits, (dk) => {
          return reloadSingleDatakit(dk)
        }).then((res) => {
          console.log("batch reload res: ", res)
        })
      }
    })
  }

  const reload = async (dk: IDatakit) => {
    if (!dk) {
      return alertError(t("select_datakit"))
    }

    modal.confirm({
      title: t("reload"),
      content: t("confirm_reload_datakit"),
      onOk: async () => {
        return reloadSingleDatakit(dk)
      }
    })
  }

  const DatakitListColumns: TableColumnsType<DataKitDataType> = [
    {
      title: t("host_name"),
      dataIndex: 'host_name',
      render: (value, record) => {
        return <Space>
          <Avatar size={16} src={getOSIcon(record.os)} alt="-" />
          <span style={{ color: "#222222", fontWeight: 600 }}> {value} </span>
        </Space>
      }
    },
    {
      title: "IP",
      dataIndex: "ip",
      render(text, record) {
        return text || "-"
      }
    },
    {
      title: t("os_arch"),
      render(text, record) {
        return `${record.os}/${record.arch}`
      }
    },
    {
      title: t("status_text"),
      dataIndex: "status",
      render(text, record) {
        return <DatakitStatus datakit={record}></DatakitStatus>
      }
    },
    {
      title: t("uptime"),
      dataIndex: "start_time",
      render(value, record) {
        return moment.duration(moment(value).diff(record.updated_at), "millisecond").humanize()
      }
    },
    {
      title: t("last_update"),
      dataIndex: "updated_at",
      render(text) {
        return moment(text).format('YYYY-MM-DD HH:mm:ss')
      }
    },
    {
      title: t("is_container_running"),
      dataIndex: 'run_in_container',
      render(value) {
        return value ? t("yes") : t("no")
      }
    },
    {
      title: t("datakit_version"),
      dataIndex: 'version',
      render(text, record) {
        return (
          <span>
            {record.version}
            {record.version !== latestDatakitVersion &&
              <span style={{ color: "#19be6b" }} className="fth-iconfont-Update"></span>
            }
          </span>

        )
      }
    },
    {
      title: t("operation"),
      render(text, record) {
        return (
          loadingDatakits[record.id] ?
            <Spin size='small' /> // loading row
            :
            <Space>
              <Button size='small' type='link' disabled={!isDatakitManagement(record)} onClick={() => { navigate("/dashboard/runinfo", { state: { datakit: record } }) }}>{t("management")}</Button>
              <Button size='small' type='link' disabled={!isDatakitManagement(record) || isContainerMode(record)} onClick={() => { reload(record) }}>{t("reload")}</Button>
              <Button size='small' type='link' disabled={!isDatakitUpgradeable(record, latestDatakitVersion)} onClick={() => { upgrade(record) }}>{t("upgrade")}</Button>
            </Space>
        )
      }
    }
  ];

  const searchDatakitList = (e?: React.KeyboardEvent<HTMLInputElement>) => {
    setPageQuery({
      ...pageQuery,
      pageIndex: 1,
    })
    initDatakitList()
  }

  const onSelectChange = (newSelectedRowKeys: React.Key[]) => {
    console.log('selectedRowKeys changed: ', newSelectedRowKeys);
    setSelectedRowKeys(newSelectedRowKeys);
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectChange,
    getCheckboxProps: (record: DataKitDataType) => ({
      disabled: !isDatakitManagement(record), // Column configuration not to be checked
      name: record.name,
    }),
  };

  useEffect(() => {
    initDatakitList()

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentWorkspace, pageQuery])

  return (
    <div className={styles.container}>
      <div className={`${styles.search} ${styles.item}`}>
        <div className={styles['search-input']}>
          <Input
            placeholder={t("search_host_ip")}
            prefix={<SearchOutlined />}
            value={searchName}
            onChange={(e) => setSearchName(e.target.value)}
            onPressEnter={(e) => { searchDatakitList(e) }}
          />
        </div>
      </div>
      <div className={styles["operation"]}>
        <div className={styles['total']}>
          {t("total_datakit", { count: pageInfo.totalCount })}
        </div>
        <div className={styles["button"]}>
          <Space>
            <Button
              type="default"
              size="small"
              className="button"
              onClick={() => setEnableSelection(!enableSelection)} >
              <CopyOutlined />
              {t("batch_operation")}
            </Button>
            <Button
              type="default"
              size="small"
              className="button"
              onClick={() => initDatakitList()}>
              <span className="fth-iconfont-refresh1 size-14"> </span>
              <span style={{ paddingLeft: '5px' }}>{t("refresh")}</span>
            </Button>
          </Space>
        </div>
      </div>
      <div className={styles['list-container']}>
        {
          enableSelection &&
          <div className={styles['edit']}>
            <div className={styles["text"]}>
              {t("selected_num", { num: selectedRowKeys.length })}
            </div>
            <div className={styles["upgrade"]}>
              <Text disabled={rowSelection.selectedRowKeys.length === 0} onClick={() => batchUpgrade()}>
                <span className="fth-iconfont-Update size-14"></span>{t("upgrade")}
              </Text>
            </div>
            <div className={styles["reload"]}>
              <Text disabled={rowSelection.selectedRowKeys.length === 0} onClick={() => batchReload()}>
                <span className="fth-iconfont-Reload1 size-14"></span>{t("reload")}
              </Text>
            </div>
            <div className={styles["cancel"]} onClick={() => { setSelectedRowKeys([]) }}>
              <Text disabled={rowSelection.selectedRowKeys.length === 0}>{t("cancel")}</Text>
            </div>
          </div>
        }
        {loading ?
          <div className={styles.nodata}><Spin indicator={<LoadingOutlined style={{ fontSize: 24, color: "#FF6600" }} spin />} /></div>
          :
          <Table
            rowKey={"id"}
            rowSelection={enableSelection ? rowSelection : undefined}
            columns={DatakitListColumns}
            pagination={{
              pageSize: pageInfo.pageSize,
              total: pageInfo.totalCount,
              defaultCurrent: pageInfo.pageIndex,
              onChange: (page, pageSize) => {
                setPageQuery({
                  pageIndex: page,
                  pageSize: pageSize,
                })
              }
            }}
            loading={isFetchingDatakitList}
            dataSource={datakits}
          />
        }
      </div>
    </div >
  )
}

const connector = connect((state: RootState) => {
  return {}
}, {
  updateDatakits: (datakits: Array<IDatakit>) => update(datakits),
})

export default connector(DatakitList) 