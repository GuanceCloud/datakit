import { createContext, SetStateAction, useCallback, useContext, useEffect, useState } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { App, Avatar, Button, message, Select, SelectProps, Space, Spin } from 'antd'
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
import { alertError, isContainerMode, isDatakitManagement, isDatakitUpgradeable } from 'src/helper/helper';
import { DashboardContext, getOSIcon } from 'src/pages/Dashboard/Dashboard';

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
            <span>网络不给力，请点击</span>
            <span onClick={refresh} style={{ paddingLeft: '5px', color: '#537CD5', cursor: 'pointer' }}>刷新</span>
          </div>
        </div>
        : "暂无数据"
    }
  </div>
}

function DkInfo() {
  const { modal } = App.useApp()
  const location = useLocation()
  const { state } = location
  const [datakit, setDatakit] = useState<IDatakit>(state?.datakit)
  const [datakitStat, setDatakitStat] = useState<IDatakitStat>()
  const [isLoading, setIsLoading] = useState(false)
  const [isError, setIsError] = useState(false)
  const [datakitsOptions, setDatakitsOptions] = useState<SelectProps["options"]>([])

  const datakits = useAppSelector((state) => state.datakit.value)
  const { latestDatakitVersion } = useContext(DashboardContext)

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
    setIsLoading(isFetchingDatakitStat || isFetchingReloadDatakit || isFetchingUpgradeDatakit)
  }, [isFetchingDatakitStat, isFetchingReloadDatakit, isFetchingUpgradeDatakit])

  useEffect(() => {
    setIsError(isErrorDatakitStat || isErrorReloadDatakit || isErrorUpgradeDatakit)
  }, [isErrorDatakitStat, isErrorReloadDatakit, isErrorUpgradeDatakit])

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
          disabled: v.status !== "running"
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
      title: "重启",
      content: "确定要重启吗？",
      onOk: async () => {
        return reloadDatakit(datakit).unwrap().then((res) => {
          if (res.success) {
            message.success("Datakit 重启命令已发送成功")
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
      title: "升级 DataKit",
      content: "确定要升级 DataKit 吗？",
      onOk: async () => {
        return upgradeDatakit(datakit).unwrap().then((res) => {
          console.log(res)
          if (res.success) {
            message.success("Datakit 升级命令已发送成功")
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

  // change datakit
  const changeDatakit = (value, option) => {
    if (option?.datakit) {
      setDatakit(option.datakit)
    }
  }

  const searchDatakit = (value) => {
    console.log("search", value)
  }

  if (!datakit) {
    return <div className={styles.nodata}>
      <div className={styles.img}></div>
      <div className={styles.text}>请选择 DataKit 进行查看</div>
    </div>
  }

  return (
    <div className={styles.dkinfo}>
      <div className={styles.nav}>
        <Select
          showSearch={true}
          placeholder="请选择 DataKit"
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
          <Button type="default" size={'small'} disabled={!isDatakitUpgradeable(datakit, latestDatakitVersion)} onClick={() => upgrade()}>
            <span className="fth-iconfont-Update size-14"> </span>
            <span className={styles.text}>升级</span>
          </Button>
          <Button type="default" size={'small'} disabled={!isDatakitManagement(datakit) || isContainerMode(datakit)} onClick={reload}>
            <ReloadOutlined className={styles.icon} />
            <span className={styles.text}>重启</span>
          </Button>
          <Button type="default" size={'small'} onClick={() => refresh()}>
            <SyncOutlined className={styles.icon} />
            <span className={styles.text}>刷新</span>
          </Button>
        </Space>
      </div>
      <div className={styles.content} >
        {
          datakit && datakitStat && !isLoading ?
            <DkInfoContext.Provider value={{
              datakit,
              datakitStat,
            }}>
              <Outlet />
            </DkInfoContext.Provider>
            :
            <Nodata
              loading={isLoading}
              isError={isError}
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