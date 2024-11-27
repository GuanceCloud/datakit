import { CSSProperties, MouseEvent, useContext, useEffect, useState } from 'react'
import { CaretRightOutlined, CaretDownOutlined, ExclamationCircleOutlined, EditOutlined } from '@ant-design/icons'
import { Button, message, Modal, Space, Tooltip, Typography } from 'antd'
import toml from 'toml'
import { useBlocker } from "react-router-dom"

import './Config.scss'
import { DkInfoContext } from '../DkInfo'
import { getDatakitConfig, saveDatakitConfig, deleteDatakitConfig } from '../../../api/api'
import loadSuccessIcon from '../../../assets/load-success.png'
import loadFailedIcon from '../../../assets/load-failed.png'
import loadModifiedIcon from '../../../assets/load-modified.png'
import ResizeBar from 'src/components/Common/ResizeBar/ResizeBar'
import { alertError, isContainerMode } from 'src/helper/helper'
import { IDatakitStat } from "src/store/type"
import { useIsAdmin } from 'src/hooks/useIsAdmin'
import DCAEditor, { DCAEditorConfiguration } from 'src/components/DCAEditor/DCAEditor'
import { Editor } from 'codemirror'


const { Text } = Typography

type ConfigInfo = {
  name?: string
  inputName: string
  sampleConfig: string
  realConfig?: string
  path?: string
  config?: string
  children?: ConfigInfo[]
  loaded?: number // 0: failed 1: success 2: modified
  configDir?: string
  catalog?: string
  isNew?: boolean
  expand?: boolean
  isMainConf?: boolean
}

export default function Config() {
  const [code, setCode] = useState<string>("")
  const [enabledConfig, setEnabledConfig] = useState<ConfigInfo[]>([])
  const [datakitConfig, setDatakitConfig] = useState<ConfigInfo | undefined>()
  const [configSelected, setConfigSelected] = useState<ConfigInfo | null>(null)
  const [newConfigSelected, setNewConfigSelected] = useState<ConfigInfo | null>(null)
  const [configurableList, setConfigurableList] = useState<ConfigInfo[]>([])
  const [isEdit, setIsEdit] = useState<boolean>(false)
  const [isSample, setIsSample] = useState<boolean>(false)
  const [colWidth, setColWidth] = useState<number>(180)
  const [editorOptions, setEditorOptions] = useState<DCAEditorConfiguration>({ readOnly: true, mode: "toml", cursorBlinkRate: 0 })
  const [editor, setEditor] = useState<Editor>()

  const isAdmin = useIsAdmin()
  const dkInfoContext = useContext(DkInfoContext)
  const { datakit, datakitStat } = dkInfoContext

  useEffect(() => {
    setEditorOptions({ ...editorOptions, readOnly: !isEdit })
    if (isEdit && editor) {
      editor.focus()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isEdit])

  let blocker = useBlocker(
    ({ currentLocation, nextLocation }) => {
      return currentLocation.pathname !== nextLocation.pathname && isEdit
    }
  )

  useEffect(() => {
    if (blocker.state === "blocked") {
      Modal.confirm({
        title: '确认',
        icon: <ExclamationCircleOutlined />,
        content: '当前处于修改状态，确定离开吗？',
        okText: '确认',
        cancelText: '取消',
        centered: true,
        onOk: () => {
          blocker.proceed()
        },
        onCancel() {
          blocker.reset()
        }
      })
    }
  }, [blocker])

  useEffect(() => {
    datakitStat && initDatakitInfo(datakitStat)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [datakitStat])

  // set selected config null when change datakit
  useEffect(() => {
    setConfigSelected(null)
  }, [datakit])

  const checkToml = (value: string) => {
    try {
      toml.parse(value)
      return [true, null]
    } catch (error: any) {
      return [false, error]
    }
  }

  const toggleIsSample = () => {
    setIsSample(!isSample)
  }
  const helpRedirect = () => {
    if (configSelected && configSelected.inputName) {
      let key = configSelected.inputName
      window.open(`https://docs.guance.com/integrations/${key}/`)
    }
  }

  const toggle = (config) => {
    return () => {
      config.expand = !config.expand
      enabledConfig && setEnabledConfig([...enabledConfig])
    }
  }

  const editConfig = () => {
    setIsEdit(true)
  }

  const cancelEdit = () => {
    Modal.confirm({
      title: '确认',
      icon: <ExclamationCircleOutlined />,
      content: '当前处于修改状态，确定取消吗？',
      okText: '确认',
      cancelText: '取消',
      centered: true,
      onOk: () => {
        configSelected?.realConfig && setCode(configSelected?.realConfig)
        setIsEdit(false)
      }
    })
  }

  const deleteConfig = async () => {
    if (!configSelected || !datakit) {
      return
    }

    const handleDelete = async () => {
      let path = configSelected?.path || ""
      if (!path) {
        return alertError("当前配置文件路径不存在")
      }

      const [err] = await deleteDatakitConfig(datakit, { path, inputName: configSelected.inputName })
      if (!err) {
        message.success("删除成功")
        setIsEdit(false)
        setConfigSelected(null)
      } else {
        alertError(err)
      }
    }

    Modal.confirm({
      title: "请确认",
      icon: <ExclamationCircleOutlined />,
      content: '确定要删除当前采集器配置文件吗?',
      okText: '确认',
      cancelText: '取消',
      onOk: handleDelete,
    })

  }

  const saveConfig = async () => {
    if (!configSelected || !datakit) {
      return
    }
    let [isCorrectToml, checkErr] = checkToml(code)
    if (!isCorrectToml) {
      Modal.info({
        title: "配置格式错误！",
        icon: <ExclamationCircleOutlined />,
        content: checkErr ? `Line ${checkErr.line}:${checkErr.column}: ${checkErr.message}` : "格式错误，请重新编辑",
        okText: '确定',
        centered: true,
      })
      return
    }
    let path = configSelected?.path || ""
    if (!path) {
      return alertError("当前配置文件路径不存在")
    }
    const [err] = await saveDatakitConfig(datakit, { path, config: code, inputName: configSelected.inputName, isNew: configSelected.isNew })
    if (!err) {
      message.success("保存成功")
      setIsEdit(false)
      setConfigSelected({
        inputName: configSelected.inputName,
        path: configSelected.path,
        sampleConfig: configSelected.sampleConfig,
        config: configSelected.sampleConfig,
        isNew: configSelected.isNew,
      })
      if (datakitConfig?.path === path) {
        datakitConfig.loaded = 2
        setDatakitConfig(datakitConfig)
      } else {
        enabledConfig.forEach((v) => {
          v.children?.forEach((c) => {
            if (c.path === path) {
              c.loaded = 2
            }
          })
        })

        setEnabledConfig([...enabledConfig])
      }
    } else {
      alertError(err)
    }
  }

  const checkConfig = async (config: ConfigInfo) => {
    if (!config.path) {
      alertError("当前文件路径不存在，请确认！")
      return
    }
    if (!datakit) {
      return
    }
    const [err, content] = await getDatakitConfig(datakit, config.path)
    if (err) {
      console.error(err)
      alertError("配置获取失败，请检查路径是否正确，或重启Datakit！")
      return
    }
    setIsEdit(false)
    config.realConfig = content as string
    setConfigSelected(config)
    setNewConfigSelected(null)
    setCode(config.realConfig)
  }

  const selectConfig = (config: ConfigInfo) => {
    return async (event: MouseEvent<HTMLDivElement>) => {
      event.stopPropagation()
      if (isEdit) {
        Modal.confirm({
          title: '确认',
          icon: <ExclamationCircleOutlined />,
          content: '当前处于修改状态，确定离开吗？',
          okText: '确认',
          cancelText: '取消',
          centered: true,
          onOk: () => {
            checkConfig(config)
          }
        })
      } else {
        checkConfig(config)
      }
    }
  }

  const addNewConfig = async (config: ConfigInfo) => {
    if (isEdit) {
      const isOk = await new Promise((resolve) => {
        Modal.confirm({
          title: '确认',
          icon: <ExclamationCircleOutlined />,
          content: '当前处于修改状态，确定离开吗？',
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
    }

    setIsEdit(false)
    setNewConfigSelected(config)
    setConfigSelected(null)
    setCode(config.sampleConfig)
    setConfigSelected({
      inputName: config.inputName,
      path: `${config.configDir}/${config.catalog ? config.catalog + '/' : ''}${config.inputName}.conf`,
      sampleConfig: config.sampleConfig,
      config: config.sampleConfig,
      isNew: true
    })
  }

  const initDatakitInfo = async (stat: IDatakitStat) => {
    const enabledConfig: ConfigInfo[] = []
    const configurableList: ConfigInfo[] = []
    const inputsConfig = stat?.config_info?.inputs
    if (!inputsConfig) {
      console.warn("datakit config info is empty")
      return
    }
    Object.keys(inputsConfig).forEach((inputName) => {
      if (inputName === 'self') {
        return
      }
      const info = inputsConfig[inputName]
      const configInfo: ConfigInfo = {
        inputName,
        expand: configSelected?.inputName === inputName,
        sampleConfig: info.sample_config,
        configDir: info.config_dir,
        catalog: info.catalog,
        children: info.config_paths.map((c) => {
          return {
            name: !stat.os_arch.includes("windows") ? c.path.split("/").pop() : c.path.split("\\").pop(),
            inputName,
            path: c.path,
            sampleConfig: info.sample_config,
            loaded: c.loaded,
            selected: configSelected?.name === inputName && c.path === configSelected.path
          }
        })
      }
      if (
        configInfo &&
        configInfo.children && configInfo.children.length > 0
      ) {
        enabledConfig.push(configInfo)
        configurableList.push(configInfo)
      } else {
        !["demo", "self"].includes(configInfo.inputName) && configurableList.push(configInfo)
      }
    })

    setEnabledConfig([...enabledConfig])
    setConfigurableList([...configurableList])
    let datakitConfig = stat?.config_info?.datakit
    if (datakitConfig && datakitConfig.config_paths?.length > 0) {
      setDatakitConfig({
        name: "datakit.conf",
        inputName: "",
        path: datakitConfig.config_paths[0].path,
        loaded: datakitConfig.config_paths[0].loaded,
        sampleConfig: datakitConfig.sample_config,
        isMainConf: true
      })
    }
  }

  const getLoadedStyle = (c: ConfigInfo): CSSProperties => {
    let pic = ""
    if (c.loaded === 0) {
      pic = loadFailedIcon
    } else if (c.loaded === 1) {
      pic = loadSuccessIcon
    } else if (c.loaded === 2) {
      pic = loadModifiedIcon
    }

    return {
      backgroundImage: `url(${pic})`,
      backgroundPosition: "right",
      backgroundRepeat: "no-repeat",
      backgroundSize: "12px"
    }
  }

  return (
    <div className="collect-config">
      <div className="config-menu" style={{ minWidth: colWidth }}>
        <div className="enable-container">
          <div className="title">
            <Space size={4}>
              <span className="fth-iconfont-configured"></span>
              <span>已配置</span>
            </Space>
          </div>
          <div className="list">
            {
              datakitConfig &&
              <div className="list-item" >
                <div className="content">
                  <div
                    className={!configSelected?.inputName && configSelected?.path === datakitConfig?.path ? 'content-item selected' : "content-item"}
                    onClick={selectConfig(datakitConfig)}
                    style={getLoadedStyle(datakitConfig)}
                  >
                    <Text style={{ width: colWidth ? (colWidth - 50) + 'px' : "100px" }} ellipsis={true}>{`${datakitConfig?.name}`}</Text>
                  </div>
                </div>
              </div>
            }
            {
              enabledConfig.map((v, index) => {
                return (
                  <div className="list-item" key={index} onClick={toggle(v)}>
                    <div className="title">
                      <div className="icon">
                        {v.expand ? <CaretDownOutlined /> : <CaretRightOutlined />}
                      </div>
                      <div className="name">{v.inputName}</div>
                    </div>
                    <div className={!v.expand ? "content hidden" : "content"}>
                      {v.children?.map((child, index) => {
                        return (
                          <div
                            className={configSelected?.inputName === v.inputName && configSelected?.path === child.path ? 'content-item selected' : "content-item"}
                            key={index}
                            onClick={selectConfig(child)}
                            style={getLoadedStyle(child)}
                          >
                            <Text style={{ width: colWidth ? (colWidth - 50) + 'px' : "100px" }} ellipsis={true}>{`- ${child.name}`}</Text>
                          </div>
                        )
                      })}
                    </div>

                  </div>
                )
              })
            }
          </div>
        </div>
        <div className="config-list-container">
          <div className="title">
            <Space size={4}>
              <span className="fth-iconfont-Sample"></span>
              <span>Sample 列表</span>
            </Space>
          </div>
          <ul className="list">
            {
              configurableList.map((v, index) => {
                return (
                  <div key={v.inputName} className={newConfigSelected?.inputName === v.inputName ? "item active" : "item"} onClick={() => addNewConfig(v)}>{v.inputName}</div>
                )
              })
            }
          </ul>
        </div>
      </div>
      <ResizeBar
        oriWidth={colWidth}
        setWidth={setColWidth}
        className="config-resize"
        minWidth={180}
      ></ResizeBar>
      <div className="config-detail">
        {configSelected ?
          <>
            <div className="setting">
              <div className="path">
                {configSelected.isNew ? "Sample 列表下都是官方提供的采集器配置文件示例，支持点击编辑另存为同名 conf 文件，并追加到已配置目录下" : `路径： ${configSelected?.path}`}
              </div>
              <div className="edit">
                <Space>

                  {
                    isAdmin && !isContainerMode(datakit)
                    &&
                    (!isEdit ?
                      <>
                        <Button type="primary" size="small" className="button" onClick={editConfig}>
                          <EditOutlined />
                          编辑
                        </Button>
                      </>
                      :
                      <>
                        <Button type="primary" size="small" className="button" onClick={saveConfig}>
                          {configSelected.isNew ?
                            <Tooltip title="另存为同名 conf 文件到已配置列表中" placement="bottom">
                              <span>另存为</span>
                            </Tooltip>
                            :
                            <span>保存</span>
                          }
                        </Button>
                        <Button size="small" className="button" onClick={cancelEdit}>
                          取消
                        </Button>
                      </>)
                  }
                  <Button type="default" size="small" onClick={() => toggleIsSample()}>
                    {isSample ? "关闭 Sample" : "Sample 配置"}
                  </Button>
                  <Button type="default" size="small" onClick={() => helpRedirect()}>
                    帮助
                  </Button>
                  {isAdmin && !isContainerMode(datakit) && !configSelected.isMainConf &&
                    (!configSelected.isNew && <Button type="default" size="small" onClick={() => deleteConfig()}>
                      <span className="fth-iconfont-trash size-14"></span>
                    </Button>)
                  }
                </Space>

              </div>
            </div>
            <div className="content">
              <div className="content-body">
                <div className={isSample ? "content-code sample" : "content-code"}>
                  <DCAEditor value={code} setValue={setCode} editorOptions={editorOptions} editorDidMount={(editor) => { setEditor(editor) }} />
                </div>
                {
                  isSample &&
                  <div className="content-sample sample">
                    <DCAEditor value={configSelected.sampleConfig} editorOptions={{ readOnly: true }} />
                  </div>
                }
              </div>
            </div>
          </>
          :
          <div></div>
        }
      </div>
    </div>
  )
}
