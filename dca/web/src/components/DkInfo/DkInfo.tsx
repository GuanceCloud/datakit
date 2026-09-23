import { createContext, ReactNode, SetStateAction, useCallback, useContext, useEffect, useState } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { App, Avatar, Button, message, Select, SelectProps, Space, Spin, Tooltip } from 'antd'
import { LoadingOutlined, ReloadOutlined, SyncOutlined } from '@ant-design/icons';
import { connect } from 'react-redux';

import styles from './DkInfo.module.scss'
import './Tabs.scss'
import NetworkErrorImg from '../../assets/network-error.png'
import { RootState } from 'src/store';
import { DatakitTab, setDatakitTab } from 'src/store/history/history';
import { IDatakit, IDatakitStat } from 'src/store/type';
import { useLazyGetDatakitStatQuery, useLazyReloadDatakitQuery, useLazyUpgradeDatakitQuery } from 'src/store/datakitApi';
import { DatakitInfoNav } from '../DatakitInfoNav/DatakitInfoNav';
import { useAppSelector } from 'src/hooks';
import { alertError, getLatestDatakitVersion, isContainerMode, isDatakitManagement, isDatakitSessionLost, isNewerDatakitVersionAvailable } from 'src/helper/helper';
import { DashboardContext, getOSIcon } from 'src/pages/Dashboard/Dashboard';
import { useTranslation } from 'react-i18next';

type DkInfoContextType = {
  datakit?: IDatakit
  datakitStat?: IDatakitStat
}

const defaultDkInfoContextValue: DkInfoContextType = {
  datakit: undefined,
  datakitStat: undefined,
}

export const DkInfoContext = createContext<DkInfoContextType>(defaultDkInfoContextValue)

export type DatakitProps = {
  datakit: IDatakit
  datakitStat: IDatakitStat | null
  loading: boolean
  error: boolean
  isAdmin?: boolean
  reload: () => void
  refresh: (showLoading?: boolean) => void
  goHelp?: (inputName?: string) => void
}

export function Nodata({ loading, isError = false, refresh }) {
  const { t } = useTranslation()
  return <div style={{ textAlign: "center", paddingTop: "50px", flex: 1 }}>
    {loading ?
      <Spin indicator={<LoadingOutlined style={{ fontSize: 40, color: "#FF6600" }} spin />} />
      :
      isError ?
        <div>
          <div>
            <img src={NetworkErrorImg} alt="" />
          </div>
          <div>
            <span>{t("network_error")}</span>
            <span onClick={refresh} style={{ paddingLeft: '5px', color: '#537CD5', cursor: 'pointer' }}>{t("refresh")}</span>
          </div>
        </div>
        : t("no_data")
    }
  </div>
}

function DkInfo() {
  const { t } = useTranslation()
  const { modal } = App.useApp()
  const location = useLocation()
  const { state } = location
  const routesWithoutDatakitStat = ['/log']
  const requiresDatakitStat = !routesWithoutDatakitStat.some((suffix) => location.pathname.endsWith(suffix))
  const [datakit, setDatakit] = useState<IDatakit>(state?.datakit)
  const [datakitStat, setDatakitStat] = useState<IDatakitStat>()
  const [datakitsOptions, setDatakitsOptions] = useState<SelectProps["options"]>([])

  const datakits = useAppSelector((state) => state.datakit.value)
  const { latestDatakitVersions } = useContext(DashboardContext)

  const [getDatakitStat, {
    data: datakitStatResponse,
    isFetching: isFetchingDatakitStat,
    isError: isErrorDatakitStat }] = useLazyGetDatakitStatQuery()

  const [reloadDatakit,
    {
      isFetching: isFetchingReloadDatakit,
      isError: isErrorReloadDatakit }] = useLazyReloadDatakitQuery()

  const [upgradeDatakit,
    {
      isFetching: isFetchingUpgradeDatakit,
      isError: isErrorUpgradeDatakit }] = useLazyUpgradeDatakitQuery()

  const fetchDatakitStat = useCallback(() => {
    if (!datakit) {
      return
    }

    getDatakitStat(datakit)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [datakit])

  useEffect(() => {
    if (isErrorDatakitStat) {
      setDatakitStat(undefined)
      return
    }

    if (!datakitStatResponse) {
      return
    }
    let { content } = datakitStatResponse
    content && setDatakitStat(content)
  }, [datakitStatResponse, isErrorDatakitStat])

  useEffect(() => {
    fetchDatakitStat()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [datakit])

  useEffect(() => {
    if (datakits?.length > 0) {
      let options: SetStateAction<SelectProps["options"]> = datakits.map((v: IDatakit) => {
        return {
          label: <Space size={1}><Avatar size={16} src={getOSIcon(v.os)} />{v.host_name}</Space>,
          value: v.host_name,
          datakit: v,
          // a running row without a live session cannot be managed either
          disabled: !isDatakitManagement(v)
        }
      })

      setDatakitsOptions(options)
    }
  }, [datakits])

  const reload = async () => {
    if (!datakit) {
      return
    }
    modal.confirm({
      title: t("reload"),
      content: t("confirm_reload_datakit"),
      onOk: async () => {
        return reloadDatakit(datakit).unwrap().then((res) => {
          if (res.success) {
            message.success(t("reload_datakit_success"))
          }
        }).catch((err) => {
          alertError(err)
        })
      }
    })
  }

  const upgrade = () => {
    if (!datakit) {
      return
    }
    modal.confirm({
      title: t("upgrade_datakit"),
        content: t("confirm_upgrade_datakit"),
        onOk: async () => {
          return upgradeDatakit(datakit).unwrap().then((res) => {
          if (res.success) {
            message.success(t("upgrade_datakit_success"))
          }
        }).catch((err) => {
          alertError(err)
        })
      }
    })
  }

  const refresh = () => {
    fetchDatakitStat()
  }

  const getReloadDisabledReason = useCallback((dk: IDatakit) => {
    if (isDatakitSessionLost(dk)) {
      return t("datakit.operation_disabled.no_session")
    }
    if (!isDatakitManagement(dk)) {
      return t("datakit.operation_disabled.not_running")
    }
    if (isContainerMode(dk)) {
      return t("datakit.operation_disabled.container_reload")
    }
    return ""
  }, [t])

  const getUpgradeDisabledReason = useCallback((dk: IDatakit) => {
    if (isDatakitSessionLost(dk)) {
      return t("datakit.operation_disabled.no_session")
    }
    if (!isDatakitManagement(dk)) {
      return t("datakit.operation_disabled.not_running")
    }
    if (isContainerMode(dk)) {
      return t("datakit.operation_disabled.container_upgrade")
    }
    if (!getLatestDatakitVersion(dk.version, latestDatakitVersions)) {
      return t("datakit.operation_disabled.version_unavailable")
    }
    if (!isNewerDatakitVersionAvailable(dk.version, latestDatakitVersions)) {
      return t("datakit.operation_disabled.latest_version")
    }
    return ""
  }, [latestDatakitVersions, t])

  const renderActionButton = (button: ReactNode, disabledReason: string) => {
    if (!disabledReason) {
      return button
    }

    return (
      <Tooltip title={disabledReason}>
        <span>{button}</span>
      </Tooltip>
    )
  }

  // change datakit
  const changeDatakit = (value, option) => {
    if (option?.datakit) {
      setDatakit(option.datakit)
    }
  }

  const searchDatakit = () => { }

  if (!datakit) {
    return <div className={styles.nodata}>
      <div className={styles.img}></div>
      <div className={styles.text}>{t("select_datakit_view")}</div>
    </div>
  }

  const isBlockingLoading = isFetchingReloadDatakit || isFetchingUpgradeDatakit || (requiresDatakitStat && isFetchingDatakitStat)
  const isBlockingError = isErrorReloadDatakit || isErrorUpgradeDatakit || (requiresDatakitStat && isErrorDatakitStat)
  const canRenderContent = datakit && !isBlockingLoading && (!requiresDatakitStat || !!datakitStat)
  const reloadDisabledReason = getReloadDisabledReason(datakit)
  const upgradeDisabledReason = getUpgradeDisabledReason(datakit)

  return (
    <div className={styles.dkinfo}>
      <div className={styles.nav}>
        <Select
          showSearch={true}
          placeholder={t("select_datakit")}
          style={{ width: 200 }}
          // labelRender={labelRender}
          defaultValue={datakit?.host_name}
          onChange={changeDatakit}
          onSearch={searchDatakit}
          options={datakitsOptions}
        // notFoundContent={fetching ? <Spin size="small" /> : null}
        />
        <div className={styles["tabs"]}>
          <DatakitInfoNav datakit={datakit} />
        </div>
        <Space className={styles["buttons"]}>
          {renderActionButton(
            <Button type="default" size={'small'} disabled={!!upgradeDisabledReason} onClick={() => upgrade()}>
              <span className="fth-iconfont-Update size-14"> </span>
              <span className={styles.text}>{t("upgrade")}</span>
            </Button>,
            upgradeDisabledReason
          )}
          {renderActionButton(
            <Button type="default" size={'small'} disabled={!!reloadDisabledReason} onClick={reload}>
              <ReloadOutlined className={styles.icon} />
              <span className={styles.text}>{t("reload")}</span>
            </Button>,
            reloadDisabledReason
          )}
          <Button type="default" size={'small'} onClick={() => refresh()}>
            <SyncOutlined className={styles.icon} />
            <span className={styles.text}>{t("refresh")}</span>
          </Button>
        </Space>
      </div>
      <div className={styles.content} >
        {
          canRenderContent ?
            <DkInfoContext.Provider value={{
              datakit,
              datakitStat,
            }}>
              <Outlet />
            </DkInfoContext.Provider>
            :
            <Nodata
              loading={isBlockingLoading}
              isError={isBlockingError}
              refresh={refresh} />
        }
      </div>
    </div>
  )
}


export default connect((state: RootState) => {
  return { datakitTab: state.history.datakitTab }
}, {
  updateDatakitTab: (datakitTab: DatakitTab) => setDatakitTab(datakitTab),
})(DkInfo)
