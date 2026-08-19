import { Divider, Dropdown, message, Modal, Space, Tooltip } from 'antd'
import { Outlet, useNavigate } from 'react-router-dom';
import { CaretDownOutlined, ExclamationCircleOutlined, LogoutOutlined, TranslationOutlined, UserOutlined } from '@ant-design/icons';
import { connect, ConnectedProps } from 'react-redux';
import { createContext, useEffect, useMemo, useState } from 'react';
import { Typography } from 'antd';

import styles from './Dashboard.module.scss'
import { clearStore, RootState } from 'src/store';
import { alertError } from 'src/helper/helper';
import { useChangeWorkspaceMutation, useLazyGetCurrentAccountQuery, useLazyGetCurrentWorkspaceQuery, useLazyGetWorkspaceListQuery, useLazyLogoutQuery } from 'src/store/consoleApi';
import { ILatestDatakitVersions, IWorkspace } from 'src/store/type';
import { useLazyGetDatakitVersionQuery } from 'src/store/consoleApi';
import { set, User } from 'src/store/user/user';
import linuxIcon from "src/assets/linux.png"
import windowsIcon from "src/assets/windows.png"
import macIcon from "src/assets/mac.png"
import { useTranslation } from 'react-i18next';
import config from "src/config"
import { toggleLanguage } from 'src/i18n';

const isTrueWatch = config.brandName === "truewatch"
const iconClass = isTrueWatch ? "icon-truewatch" : "icon"

const { Text } = Typography
const osIcons = {
  "linux": linuxIcon,
  "windows": windowsIcon,
  "mac": macIcon
}

type DashboardContextType = {
  currentWorkspace: IWorkspace | undefined
  latestDatakitVersions: ILatestDatakitVersions
}
const defaultDashboardContext: DashboardContextType = {
  currentWorkspace: undefined,
  latestDatakitVersions: {},
}
export const DashboardContext = createContext<DashboardContextType>(defaultDashboardContext)

export function getOSIcon(os: string): string {
  return osIcons[os]
}

function Dashboard({ user, setUserInfo }: Props) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [visible, setVisible] = useState<Boolean>(false)
  const [latestDatakitVersions, setLatestDatakitVersions] = useState<ILatestDatakitVersions>({})
  const [workspaces, setWorkspaces] = useState<IWorkspace[]>([])
  const [workspaceKeyword, setWorkspaceKeyword] = useState("")
  const [selectedWorkspace, setSelectedWorkspace] = useState<IWorkspace | undefined>()

  // rtk query
  const [getWorkSpaceList, { data: workspaceListData }] = useLazyGetWorkspaceListQuery()
  const [getCurrentWorkspace, { data: currentWorkspace }] = useLazyGetCurrentWorkspaceQuery()
  const [getDatakitVersion, { data: dataDatakitVersion }] = useLazyGetDatakitVersionQuery()
  const [getCurrentAccount] = useLazyGetCurrentAccountQuery()
  const [changeWorkspace] = useChangeWorkspaceMutation()
  const [userLogout] = useLazyLogoutQuery()

  useEffect(() => {
    if (dataDatakitVersion?.code === 200) {
      setLatestDatakitVersions({
        v1: dataDatakitVersion.content?.v1?.version,
        v2: dataDatakitVersion.content?.v2?.version,
      })
    }
  }, [dataDatakitVersion])

  // workspace list menu
  useEffect(() => {
    if (!workspaceListData) {
      return
    }
    let { code, content: { data } } = workspaceListData
    if (code !== 200) {
      setWorkspaces([])
      return
    }
    setWorkspaces(data || [])
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [workspaceListData])

  const workspaceMenu = useMemo(() => {
    const keyword = workspaceKeyword.trim().toLowerCase()
    const filteredWorkspaces = keyword
      ? workspaces.filter((workspace) => {
        const searchText = [
          workspace.name,
          workspace.wsName,
          workspace.uuid,
        ].filter(Boolean).join(" ").toLowerCase()
        return searchText.includes(keyword)
      })
      : workspaces

    if (filteredWorkspaces.length === 0) {
      return {
        items: [
          {
            key: "__empty",
            disabled: true,
            label: (
              <div className={styles.noWorkspace}>{t("no_data")}</div>
            )
          }
        ]
      }
    }

    return {
      items: filteredWorkspaces.map((workspace) => {
        return {
          key: workspace.uuid,
          label: (
            <Text
              className={styles.workspaceItem}
              ellipsis={true}>
              {workspace.name || workspace.wsName}
            </Text>
          )
        }
      }),
      onClick: async ({ key }) => {
        const workspace = workspaces.find((item) => item.uuid === key)
        if (!workspace) {
          return
        }

	        try {
	          await changeWorkspace(key).unwrap()
	          setVisible(false)
	          setWorkspaceKeyword("")
	          setSelectedWorkspace(workspace)
	          getWorkSpaceList(undefined, false)
	          getCurrentWorkspace(undefined, false)
	        } catch (error) {
	          alertError(error)
	        }
      }
    }
	  }, [changeWorkspace, getCurrentWorkspace, getWorkSpaceList, t, workspaceKeyword, workspaces])

  const logout = async () => {
    const isOk = await new Promise((resolve) => {
      Modal.confirm({
        title: t("confirm"),
        icon: <ExclamationCircleOutlined />,
        content: t("is_logout"),
        okText: t("confirm"),
        cancelText: t("cancel"),
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
        message.success(t("logout_success"))
        navigate("/login", { replace: true })
      })
    }).catch((err) => {
      console.error(err)
      alertError(t("logout_fail"))
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
              <span>{t("logout")}</span>
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
    return currentWorkspace ? (currentWorkspace.name || currentWorkspace.wsName) : t("workspace_list")
  }

  useEffect(() => {
    if (currentWorkspace) {
      navigate("/dashboard")
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentWorkspace])

	  useEffect(() => {
	    setSelectedWorkspace(currentWorkspace)
	  }, [currentWorkspace])

  useEffect(() => {
    init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div
          className={styles[iconClass]}
        ></div>
        <div className={styles.list}>
          <Dropdown
            menu={workspaceMenu}
            className={styles.menu}
            open={Boolean(visible)}
            overlayStyle={{ width: "220px" }}
            dropdownRender={(originNode) => (
              <div className={styles.workspaceDropdown}>
                <input
                  className={styles.workspaceSearch}
                  placeholder={t("search_workspace")}
                  value={workspaceKeyword}
                  onChange={(event) => setWorkspaceKeyword(event.target.value)}
                  onClick={(event) => event.stopPropagation()}
                  onKeyDown={(event) => {
                    event.stopPropagation()
                    if (["Enter", "ArrowUp", "ArrowDown"].includes(event.key)) {
                      event.preventDefault()
                    }
                  }}
                />
                <div className={styles.workspaceMenuList}>
                  {originNode}
                </div>
              </div>
            )}
            onOpenChange={(flag) => {
              setVisible(flag)
              if (flag) {
                setWorkspaceKeyword("")
              }
            }}
          >
            <div>
              <div className={styles.name}>
                <Text className={styles.text} ellipsis={true}>
                  {getWorkSpaceName(selectedWorkspace)}
                </Text>
              </div>
              <div className={styles.arrow}>
                <CaretDownOutlined size={12} />
              </div>
            </div>
          </Dropdown>
        </div>
        <div className={styles.right}>
          {
            !isTrueWatch && (
              <div className={styles.lang}>
                <Tooltip title="中文 / English">
                  <Space onClick={() => {
                    toggleLanguage()
                  }}>
                    <TranslationOutlined size={16} />
                    <span>{t("language")}</span>
                  </Space>
                </Tooltip>
                <Divider type="vertical" />
              </div>
            )
          }
          <div className={styles.help}>
            <a target="_blank" rel="noreferrer" href={`${config.docURL}/datakit/dca`}>
              <Space>
                <span className='fth-iconfont-help2'></span>
                <span>{t("help")}</span>
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
            currentWorkspace: selectedWorkspace,
            latestDatakitVersions,
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
