import { Divider, Dropdown, message, Modal, Space } from 'antd'
import { Outlet, useNavigate } from 'react-router-dom';
import { CaretDownOutlined, ExclamationCircleOutlined, LogoutOutlined, UserOutlined } from '@ant-design/icons';
import { connect, ConnectedProps } from 'react-redux';
import { createContext, useEffect, useState } from 'react';
import { Typography } from 'antd';

import styles from './Dashboard.module.scss'
import { clearStore, RootState } from 'src/store';
import { alertError } from 'src/helper/helper';
import { useChangeWorkspaceMutation, useLazyGetCurrentAccountQuery, useLazyGetCurrentWorkspaceQuery, useLazyGetWorkspaceListQuery, useLazyLogoutQuery } from 'src/store/consoleApi';
import { IWorkspace } from 'src/store/type';
import { useLazyGetDatakitVersionQuery } from 'src/store/consoleApi';
import { set, User } from 'src/store/user/user';
import linuxIcon from "src/assets/linux.png"
import windowsIcon from "src/assets/windows.png"
import macIcon from "src/assets/mac.png"

const { Text } = Typography
const osIcons = {
  "linux": linuxIcon,
  "windows": windowsIcon,
  "mac": macIcon
}

const defaultMenu = {
  items:
    [
      {
        key: "1",
        label: (
          <div style={{ color: "#C6C6C6", textAlign: "center", height: "110px", lineHeight: "110px" }}>暂无数据</div>
        )
      }
    ]
}
type DashboardContextType = {
  currentWorkspace: IWorkspace | undefined
  latestDatakitVersion: string
}
const defaultDashboardContext: DashboardContextType = {
  currentWorkspace: undefined,
  latestDatakitVersion: "",
}
export const DashboardContext = createContext<DashboardContextType>(defaultDashboardContext)

export function getOSIcon(os: string): string {
  return osIcons[os]
}

function Dashboard({ user, setUserInfo }: Props) {
  const navigate = useNavigate()
  const [menu, setMenu] = useState(defaultMenu)
  const [visible, setVisible] = useState<Boolean>(false)
  const [latestDatakitVersion, setLatestDatakitVersion] = useState("")

  // rtk query
  const [getWorkSpaceList, { data: workspaceListData }] = useLazyGetWorkspaceListQuery()
  const [getCurrentWorkspace, { data: currentWorkspace }] = useLazyGetCurrentWorkspaceQuery()
  const [getDatakitVersion, { data: dataDatakitVersion }] = useLazyGetDatakitVersionQuery()
  const [getCurrentAccount] = useLazyGetCurrentAccountQuery()
  const [changeWorkspace] = useChangeWorkspaceMutation()
  const [userLogout] = useLazyLogoutQuery()

  useEffect(() => {
    if (dataDatakitVersion?.code === 200) {
      setLatestDatakitVersion(dataDatakitVersion?.content?.version)
    }
  }, [dataDatakitVersion])

  // workspace list menu
  useEffect(() => {
    if (!workspaceListData) {
      return
    }
    let { code, content: { data } } = workspaceListData
    if (code !== 200) {
      setMenu(defaultMenu)
      return
    }
    if (data && data.length > 0) {
      const menu = {
        items:
          data.map((w, index) => {
            return (
              {
                key: `${index}`,
                label: (
                  <Text
                    style={{ maxWidth: "130px" }}
                    ellipsis={true}>
                    {w.name}
                  </Text>
                )
              }

            )
          }),
        onClick: ({ key }) => {
          changeWorkspace(data[key].uuid).then(() => {
            return getCurrentWorkspace()
          }).catch((error) => {
            alertError(error)
          })
        }
      }
      setMenu(menu)

    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workspaceListData])

  const logout = async () => {
    const isOk = await new Promise((resolve) => {
      Modal.confirm({
        title: '确认',
        icon: <ExclamationCircleOutlined />,
        content: '是否退出当前账户',
        okText: '确认',
        cancelText: '取消',
        centered: true,
        onOk: () => {
          resolve(true)
        },
        onCancel: () => {
          resolve(false)
        }
      })
    })

    if (!isOk) {
      return
    }

    userLogout().unwrap().then(() => {
      clearStore().finally(() => {
        message.success("退出成功")
        navigate("/login", { replace: true })
      })
    }).catch((err) => {
      console.error(err)
      alertError("退出失败")
    })
  }

  const logoutMenu = {
    items: [
      {
        key: '1',
        label: (
          <div onClick={logout}>
            <Space>
              <LogoutOutlined />
              <span>退出</span>
            </Space>
          </div>
        ),
      },
    ]
  }

  const init = async () => {
    getWorkSpaceList()
    getDatakitVersion("")
    getCurrentWorkspace()

    if (!user?.name) {
      getCurrentAccount().unwrap().then((userData) => {
        if (userData?.code === 200) {
          const { content: user } = userData
          user && setUserInfo({
            name: user.name,
            email: user.email
          })
        }
      })
    }
  }

  const getWorkSpaceName = (currentWorkspace: IWorkspace | undefined): string => {
    return currentWorkspace ? (currentWorkspace.name || currentWorkspace.wsName) : '工作空间列表'
  }

  useEffect(() => {
    if (currentWorkspace) {
      navigate("/dashboard")
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentWorkspace])

  useEffect(() => {
    init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.icon}></div>
        <div className={styles.list}>
          <Dropdown
            menu={menu}
            className={styles.menu}
            open={Boolean(visible)}
            overlayStyle={{ maxHeight: "300px", overflow: "auto", }}
            onOpenChange={(flag) => setVisible(flag)}
          >
            <div>
              <div className={styles.name}>
                <Text className={styles.text} ellipsis={true}>
                  {getWorkSpaceName(currentWorkspace)}
                </Text>
              </div>
              <div className={styles.arrow}>
                <CaretDownOutlined size={12} />
              </div>
            </div>
          </Dropdown>
        </div>
        <div className={styles.right}>
          <div className={styles.help}>
            <a target="_blank" rel="noreferrer" href="https://docs.guance.com/datakit/dca">
              <Space>
                <span className='fth-iconfont-help2'></span>
                <span>帮助</span>
              </Space>

            </a>
            <Divider type="vertical" />
          </div>
          <Dropdown menu={logoutMenu} overlayStyle={{}}>
            <Space>
              <UserOutlined />
              <span>{user.name}</span>
              <CaretDownOutlined size={12} />
            </Space>
          </Dropdown>
        </div>
      </div>
      <div className={styles.body}>
        <div className={styles.info}>
          <DashboardContext.Provider value={{
            currentWorkspace: currentWorkspace,
            latestDatakitVersion,
          }}>
            <Outlet />
          </DashboardContext.Provider>
        </div>
      </div>
    </div>
  )
}

const connector = connect((state: RootState) => {
  return {
    user: state.user.value,
  }
}, {
  setUserInfo: (userInfo: User) => {
    return set(userInfo)
  },
})

type PropsFromRedux = ConnectedProps<typeof connector>
interface Props extends PropsFromRedux { }

export default connector(Dashboard)
