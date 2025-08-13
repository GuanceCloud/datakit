import { CloseCircleOutlined, CopyOutlined, FilterFilled, LoadingOutlined, PlusOutlined, SearchOutlined, SelectOutlined } from '@ant-design/icons'
import { useCallback, useContext, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom';
import moment from 'moment';
import { App, Avatar, Button, Checkbox, Input, Modal, Select, Space, Spin, Table, TableColumnsType, Tooltip, Typography, message } from 'antd'
import { connect, ConnectedProps } from 'react-redux';

import { alertError, isContainerMode, isDatakitManagement, isDatakitUpgradeable, isLoadingStatus, runJob } from 'src/helper/helper';
import styles from './DkList.module.scss'
import { IDatakit, ISearchValue, IWorkspace, PageInfo, PageQuery } from 'src/store/type'
import { update } from '../../store/datakit/datakit';
import { RootState } from 'src/store';
import { useAppSelector } from 'src/hooks'
import { DashboardContext, getOSIcon } from 'src/pages/Dashboard/Dashboard';
import { useLazyGetDatakitListQuery, useLazyReloadDatakitQuery, useLazyUpgradeDatakitQuery, useLazyGetDatakitListByIDQuery, useLazyOperateDatakitQuery, useLazyGetSearchValueQuery } from 'src/store/datakitApi';
import DatakitStatus from '../DatakitStatus/DatakitStatus';
import { useTranslation } from 'react-i18next';
import { AdditionColumnOptions } from './AdditionalColumnOptions/AdditionColumnOptions';

const { Text } = Typography;
const maxRequestNumber = 10

interface DatakitListProps {
  workspace?: IWorkspace
}

type FilterItem = {
  id: number
  field: string
  operator: string
  value: string[]
}

type PropsFromRedux = ConnectedProps<typeof connector>
interface Props extends PropsFromRedux, DatakitListProps { }

interface DataKitDataType extends IDatakit {
  key: React.Key;
  name: string;
  state: string;
  is_container: boolean;
}

const isValidFilterItem = (item: FilterItem): boolean => {
  return !!(item.field && item.operator && item.value && item.value.length > 0 && (item.value[0] !== ""))
}

function DatakitList({ updateDatakits }: Props) {
  const { modal } = App.useApp()
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [modalFilterOpen, setModalFilterOpen] = useState(false)
  const [timer, setTimer] = useState<NodeJS.Timeout>()
  const [searchName, setSearchName] = useState("")
  const [isSelectAll, setIsSelectAll] = useState(false)
  const [filterRelation, setFilterRelation] = useState('and')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [enableSelection, setEnableSelection] = useState(false)
  const [loadingDatakits, setLoadingDatakits] = useState<Record<string, boolean>>({})
  const [pageQuery, setPageQuery] = useState<PageQuery>(
    {
      pageIndex: 1,
      pageSize: 20
    }
  )
  const [filterParams, setFilterParams] = useState<string>()
  const [datakitListColumns, setDatakitListColumns] = useState<TableColumnsType<DataKitDataType>>([])

  const [pageInfo, setPageInfo] = useState<PageInfo>({
    count: 0,
    pageSize: 10,
    pageIndex: 1,
    totalCount: 0
  })

  const [filterItems, setFilterItems] = useState<FilterItem[]>([
    { id: 1, field: '', operator: '', value: [] },
  ]);
  const [searchValues, setSearchValues] = useState<ISearchValue>()

  const datakits = useAppSelector((state) => state.datakit.value)
  const navigate = useNavigate()

  const { currentWorkspace, latestDatakitVersion } = useContext(DashboardContext)

  const [queryDatakitList, { currentData: datakitListResponse, isFetching: isFetchingDatakitList, isError: isErrorDatakitList }] = useLazyGetDatakitListQuery()
  const [reloadDatakit] = useLazyReloadDatakitQuery()
  const [operateDatakit] = useLazyOperateDatakitQuery()
  const [getSearchValue, { currentData: searchValueResponse, isError: isErrorSearchValue }] = useLazyGetSearchValueQuery()

  const queryParams = useMemo(() => {
    return {
      pageIndex: pageQuery.pageIndex,
      pageSize: pageQuery.pageSize,
      search: searchName,
      filter: filterParams,
    }
  }, [pageQuery, searchName, filterParams])
  const initDatakitList = useCallback(async () => {
    setSelectedRowKeys([])
    setIsSelectAll(false)
    queryDatakitList({ ...queryParams, minLastUpdateTime: moment().subtract(1, 'days').unix() }) // use memo or will hang in fetching.
  }, [queryDatakitList, queryParams])

  const refresh = useCallback(() => {
    initDatakitList()
    getSearchValue()

  }, [initDatakitList, getSearchValue])

  const [upgradeDatakit] = useLazyUpgradeDatakitQuery()
  const [getDatakitListByID] = useLazyGetDatakitListByIDQuery()

  useEffect(() => {
    if (isErrorDatakitList || !datakitListResponse?.success) {
      setLoading(false)
      updateDatakits([])
      setPageInfo(prev => ({ ...prev, count: 0, totalCount: 0 }))
      return alertError(datakitListResponse?.message)
    }

    let datakits = datakitListResponse.content.data.map<IDatakit>((datakit) => {
      if (isLoadingStatus(datakit)) {
        setLoadingDatakits(prev => ({ ...prev, [datakit.id]: true }))
      }

      let globalHostTags = {}
      try {
        globalHostTags = JSON.parse(datakit.global_host_tags_string)
      } catch (e) {
        console.error("parse global host tags failed", e)
      }

      return {
        ...datakit,
        global_host_tags: globalHostTags,
      }
    })

    datakitListResponse.content.pageInfo && setPageInfo(datakitListResponse.content.pageInfo)
    updateDatakits(datakits)
    setLoading(false)

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [datakitListResponse, isErrorDatakitList, updateDatakits])

  useEffect(() => {
    if (isErrorSearchValue || !searchValueResponse?.success) {
      return alertError(searchValueResponse?.message)
    }

    setSearchValues(searchValueResponse?.content)
  }, [searchValueResponse, isErrorSearchValue])

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

            const newLoadingDatakits = { ...loadingDatakits }

            let isChanged = false
            let isLoadingChanged = false
            let newDatakits = datakits.map((dk) => {
              if (dkMap.has(dk.id)) {
                let newDK = dkMap.get(dk.id)
                if (!isLoadingStatus(newDK)) {
                  isLoadingChanged = true
                  delete newLoadingDatakits[dk.id]
                }
                isChanged = true
                return newDK
              }
              return dk
            })
            if (isChanged) {
              updateDatakits(newDatakits)
            }
            if (isLoadingChanged) {
              setLoadingDatakits(newLoadingDatakits)
            }
          }
        })
      }, 10000)

      setTimer(t)
      return () => {
        clearInterval(t)
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loadingDatakits])

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

  const upgradeSingleDatakit = useCallback(async (dk: IDatakit) => {
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
  }, [t, upgradeDatakit, datakits, updateDatakits])

  const upgrade = useCallback(
    async (dk: IDatakit) => {
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
    }, [latestDatakitVersion, modal, t, upgradeSingleDatakit])

  const reloadSingleDatakit = useCallback(async (dk: IDatakit) => {
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
  }, [t, reloadDatakit, updateDatakits, datakits])

  const batchReload = async () => {
    const reloadDatakits: IDatakit[] = []

    let totalDatakits = 0
    if (isSelectAll) {
      totalDatakits = pageInfo.totalCount
    } else {
      for (let k of selectedRowKeys) {
        let dk = datakits.find((d) => d.id === k)
        if (dk && isDatakitManagement(dk) && !isContainerMode(dk)) {
          reloadDatakits.push(dk)
        }
      }

      totalDatakits = reloadDatakits.length
    }

    if (totalDatakits === 0) {
      return
    }

    modal.confirm({
      title: t("reload_datakit"),
      content: t("reload_datakit_message", { count: totalDatakits }),
      onOk: async () => {
        if (isSelectAll) {
          return operateDatakit({ ids: "all", type: "reload" }).unwrap().then((res) => {
            if (res.success) {
              message.success(t("reload_datakit_success"))
              initDatakitList()
            }
          }).catch((err) => {
            alertError(err)
          })
        }
        return runJob(maxRequestNumber, reloadDatakits, (dk) => {
          return reloadSingleDatakit(dk)
        }).then((res) => {
          console.log("batch reload res: ", res)
        })
      }
    })
  }

  const reload = useCallback(async (dk: IDatakit) => {
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
  }, [t, modal, reloadSingleDatakit])

  // useMemo
  const DefaultDatakitListColumns: TableColumnsType<DataKitDataType> = useMemo(() => [
    {
      title: t("host_name"),
      dataIndex: 'host_name',
      key: 'host_name',
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
      key: 'ip',
      render(text, record) {
        return text || "-"
      }
    },
    {
      title: t("os_arch"),
      key: 'os_arch',
      render(text, record) {
        return `${record.os}/${record.arch}`
      }
    },
    {
      title: t("status_text"),
      dataIndex: "status",
      key: 'status_text',
      render(text, record) {
        return <DatakitStatus datakit={record}></DatakitStatus>
      }
    },
    {
      title: t("uptime"),
      key: 'uptime',
      dataIndex: "start_time",
      render(value, record) {
        return moment.duration(moment(value).diff(record.updated_at), "millisecond").humanize()
      }
    },
    {
      title: t("environment"),
      key: "environment",
      hidden: true,
      render(text, record) {
        if (record.global_host_tags) {
          return record.global_host_tags["env"]
        }
        return "-"
      }
    },
    {
      title: t("last_update"),
      dataIndex: "updated_at",
      key: "last_update",
      render(text) {
        return moment(text).format('YYYY-MM-DD HH:mm:ss')
      }
    },
    {
      title: t("is_container_running"),
      dataIndex: 'run_in_container',
      key: "is_container_running",
      render(value) {
        return value ? t("yes") : t("no")
      }
    },
    {
      title: t("datakit_version"),
      dataIndex: 'version',
      key: "datakit_version",
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
      key: "operation",
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
  ], [t, latestDatakitVersion, loadingDatakits, navigate, reload, upgrade]);


  const handleRelationChange = (value) => {
    setFilterRelation(value);
  };

  const handleFilterChange = (id, field, value) => {
    const updatedItems = filterItems.map(item => {
      if (item.id === id) {
        item.value = []
        return { ...item, [field]: value };
      }
      return item;
    });

    setFilterItems(updatedItems);
  };

  const handleApplyFilters = () => {
    let items: Array<any> = []
    filterItems.forEach(item => {
      if (isValidFilterItem(item)) {
        items.push({
          field: item.field,
          operator: item.operator,
          value: item.value
        })
      }
    })

    if (items.length === 0) {
      setFilterParams("")
      return
    }

    let filter = {
      "relation": filterRelation,
      items
    }

    setFilterParams(encodeURIComponent(JSON.stringify(filter)))
  };


  const removeFilterItem = (id) => {
    const updatedItems = filterItems.filter(item => item.id !== id);
    setFilterItems(updatedItems);
  };
  const addFilterItem = () => {
    setFilterItems([...filterItems, { id: Date.now(), field: '', operator: '', value: [] }]);
  };

  const clearAllFilter = () => {
    setFilterItems([{ id: Date.now(), field: '', operator: '', value: [] }]);
    setFilterRelation('and');
  }
  const getFilterCountString = () => {
    return filterItems?.filter(item => item.field && item.operator && item.value && item.value.length > 0).length || ""
  };

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
    setIsSelectAll(false)
  };

  const selectAll = () => {
    if (isSelectAll) {
      setSelectedRowKeys([])
    } else {
      setSelectedRowKeys(datakits.map((d) => d.id))
    }
    setIsSelectAll(!isSelectAll)
  }

  const handleColumnToggle = useCallback((value: Record<string, boolean>) => {
    if (!value) {
      return
    }
    let newItems = DefaultDatakitListColumns.map((d) => ({
      ...d,
      hidden: !(value[d.key as string])
    }))
    setDatakitListColumns(newItems)
  }, [setDatakitListColumns, DefaultDatakitListColumns])

  const rowSelection = {
    selectedRowKeys,
    onChange: onSelectChange,
    columnTitle: (
      <Tooltip title={
        selectedRowKeys.length > 0 && selectedRowKeys.length === datakits.length ? t("deselect_current_page") : t("select_current_page")}
      >
        <Checkbox
          checked={selectedRowKeys.length > 0 && selectedRowKeys.length === datakits.length}
          indeterminate={selectedRowKeys.length > 0 && selectedRowKeys.length < datakits.length}
          onChange={(e) => {
            setSelectedRowKeys(e.target.checked ? datakits.map((d) => d.id) : [])
          }}
        />
      </Tooltip>),
    getCheckboxProps: (record: DataKitDataType) => ({
      disabled: !isDatakitManagement(record), // Column configuration not to be checked
      name: record.name,
    }),
  };

  useEffect(() => {
    initDatakitList()
    getSearchValue()

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentWorkspace, queryParams])

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
        <div className={styles['filter']}>
          <Button onClick={() => setModalFilterOpen(true)}>
            <FilterFilled /> {t("filter")} {getFilterCountString()}
          </Button>
        </div>
      </div>
      <Modal
        title={t("set_filter")}
        style={{ top: 100 }}
        open={modalFilterOpen}
        onOk={() => setModalFilterOpen(false)}
        onCancel={() => setModalFilterOpen(false)}
        afterClose={() => { handleApplyFilters() }}
        footer={null}
      >

        <div style={{ marginBottom: 16, marginTop: 16 }}>
          <span>{t("match_below")}</span>
          <Select
            defaultValue="and"
            value={filterRelation}
            style={{ width: 80, margin: '0 8px' }}
            onChange={handleRelationChange}
            options={[
              { value: "and", label: t("all") },
              { value: "or", label: t("any") },
            ]}
          >
          </Select>
          <span>{t("condition")}</span>
        </div>

        {filterItems.map((item) => {
          return (
            <Space key={item.id} style={{ display: 'flex', marginBottom: 16 }} align="baseline">
              <Select
                showSearch
                placeholder={t("filter_item")}
                style={{ width: 120 }}
                onChange={(value) => handleFilterChange(item.id, 'field', value)}
                options={((searchValues && Object.keys(searchValues)) || []).map(key => ({ value: key, label: key }))}
              >
              </Select>

              <Select
                placeholder={t("operator")}
                style={{ width: 100, margin: '0 8px' }}
                onChange={(value) => handleFilterChange(item.id, 'operator', value)}
                options={[
                  { value: 'in', label: 'in' },
                  { value: 'not_in', label: 'not in' },
                  { value: 'match', label: 'match' },
                  { value: 'not_match', label: 'not match' },
                ]}
              >
              </Select>

              {item.operator === 'in' || item.operator === 'not_in' ? (
                <Select
                  placeholder={t("select_value")}
                  style={{ width: 200 }}
                  mode="multiple"
                  value={item.value || []}
                  onChange={(value) => handleFilterChange(item.id, 'value', value)}
                  options={((searchValues && searchValues[item.field]) || []).map(key => ({ value: key, label: key }))}
                >
                </Select>
              ) : item.operator === 'match' || item.operator === 'not_match' ? (
                <Input
                  placeholder={t("input_regex")}
                  style={{ width: 200 }}
                  onChange={(e) => handleFilterChange(item.id, 'value', [e.target.value])}
                />
              ) : (
                <Input
                  placeholder={t("input_value")}
                  style={{ width: 200 }}
                  onChange={(e) => handleFilterChange(item.id, 'value', [e.target.value])}
                />
              )}

              <CloseCircleOutlined
                onClick={() => removeFilterItem(item.id)}
                style={{ color: '#999', marginLeft: 8, cursor: 'pointer' }}
              />
            </Space>
          )
        }
        )}


        <div style={{ display: 'flex', justifyContent: "space-between" }}>

          <Button type="text" icon={<PlusOutlined />} onClick={addFilterItem}>
            {t("add_filter")}
          </Button>

          <Button type="text" onClick={clearAllFilter}>
            {t("clear_all_filter")}
          </Button>

        </div>
      </Modal>
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
              onClick={refresh}>
              <span className="fth-iconfont-refresh1 size-14"> </span>
              <span style={{ paddingLeft: '5px' }}>{t("refresh")}</span>
            </Button>
            <AdditionColumnOptions onValueChange={handleColumnToggle} />
          </Space>
        </div>
      </div>
      <div className={styles['list-container']}>
        {
          enableSelection &&
          <div className={styles['edit']}>
            <div className={styles["text"]}>
              {t("selected_num", { num: isSelectAll ? pageInfo.totalCount : selectedRowKeys.length })}
            </div>
            <div className={styles["select-all"]}>
              <Text onClick={selectAll}>
                <SelectOutlined /> {isSelectAll ? t("deselect_all") : t("select_all")}
              </Text>
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
            <div className={styles["cancel"]} onClick={() => { setSelectedRowKeys([]); setIsSelectAll(false) }}>
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
            columns={datakitListColumns}
            pagination={{
              pageSize: pageInfo.pageSize,
              total: pageInfo.totalCount,
              defaultCurrent: pageInfo.pageIndex,
              onChange: (page, pageSize) => {
                setPageQuery({
                  pageIndex: page,
                  pageSize: pageSize,
                })
              },
              locale: {
                items_per_page: t("items_per_page"),
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