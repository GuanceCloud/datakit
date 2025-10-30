import { CSSProperties, MouseEvent, useContext, useEffect, useState } from 'react'
import { CaretRightOutlined, CaretDownOutlined, ExclamationCircleOutlined, EditOutlined, QuestionCircleFilled } from '@ant-design/icons'
import { Button, Input, message, Modal, Space, Tooltip, Typography } from 'antd'
import toml from 'toml'
import path from 'path-browserify'
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
import { useTranslation } from 'react-i18next'
import config from "src/config"


const { Text } = Typography

type ConfigInfo = {
  name?: string
  inputName: string
  sampleConfig: string
  realConfig?: string
  path?: string
  dir?: string
  config?: string
  children?: ConfigInfo[]
  loaded?: number // 0: failed 1: success 2: modified
  configDir?: string
  catalog?: string
  isNew?: boolean
  expand?: boolean
  isMainConf?: boolean
}

function getConfigNameFromPath(path: string, os: string): string {
  if (!path) {
    return ""
  }
  const separator = getSeparator(os)

  return path.split(separator).pop() || ""
}

function getSeparator(os: string): string {
  return !os.includes("windows") ? "/" : "\\"
}

export default function Config() {
  const { t } = useTranslation()
  const [code, setCode] = useState<string>("")
  const [createConfigVisible, setCreateConfigVisible] = useState<boolean>(false)
  const [configSaveName, setConfigSaveName] = useState<string>("")
  const [configSavePath, setConfigSavePath] = useState<string>("")
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
        title: t('confirm'),
        icon: <ExclamationCircleOutlined />,
        content: t("edit_state_away"),
        okText: t("confirm"),
        cancelText: t("cancel"),
        centered: true,
        onOk: () => {
          blocker.proceed()
        },
        onCancel() {
          blocker.reset()
        }
      })
    }
  }, [blocker, t])

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
      window.open(`${config.docURL}/integrations/${key}/`)
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
      title: t('confirm'),
      icon: <ExclamationCircleOutlined />,
      content: t('edit_state_cancel'),
      okText: t("confirm"),
      cancelText: t("cancel"),
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
        return alertError(t("file_path_not_exists"))
      }

      const [err] = await deleteDatakitConfig(datakit, { path, inputName: configSelected.inputName })
      if (!err) {
        message.success(t("delete_success"))
        setIsEdit(false)
        setConfigSelected(null)
        enabledConfig.forEach((v) => {
          v.children = v.children?.filter((c) => c.path !== path)
        })
        setEnabledConfig([...enabledConfig])
      } else {
        alertError(err)
      }
    }

    Modal.confirm({
      title: t("confirm"),
      icon: <ExclamationCircleOutlined />,
      content: t("confirm_delete_config"),
      okText: t("confirm"),
      cancelText: t("cancel"),
      onOk: handleDelete,
    })

  }

  const saveConfig = async () => {
    if (configSelected?.isNew) {
      setCreateConfigVisible(true)
      return
    }

    doSaveConfig()
  }

  const doSaveConfig = async (isForce: boolean = false): Promise<boolean> => {
    if (!configSelected || !datakit) {
      return false
    }

    let [isCorrectToml, checkErr] = checkToml(code)
    if (!isCorrectToml) {
      Modal.info({
        title: t("config.format_error"),
        icon: <ExclamationCircleOutlined />,
        content: checkErr ? `Line ${checkErr.line}:${checkErr.column}: ${checkErr.message}` : t("config.format_error.message"),
        okText: t("confirm"),
        centered: true,
      })
      return false
    }

    let path = configSelected?.path || ""
    let isNew = configSelected?.isNew || false

    if (!isNew) {
      isForce = false
    } else {
      path = configSavePath
    }

    if (!path) {
      alertError(t("config.file_not_exists"))
      return false
    }

    const [err, data] = await saveDatakitConfig(datakit, { path, config: code, inputName: configSelected.inputName, isNew: configSelected.isNew, isForce })
    if (!err) {
      message.success(t("save_success"))
      setIsEdit(false)
      setConfigSelected({
        inputName: configSelected.inputName,
        path,
        sampleConfig: configSelected.sampleConfig,
        config: configSelected.sampleConfig,
        isNew: configSelected.isNew,
        dir: configSelected.dir,
      })
      if (datakitConfig?.path === path) {
        datakitConfig.loaded = 2
        setDatakitConfig(datakitConfig)
      } else {
        let isFound = false
        let configIndex = -1
        let selectedConfig: ConfigInfo | null = null
        enabledConfig.forEach((v, index) => {
          if (v.inputName === configSelected.inputName) {
            configIndex = index
          }

          v.children?.forEach((c) => {
            if (c.path === path) {
              isFound = true
              selectedConfig = c
              c.loaded = 2
            }
          })
        })

        if (isNew && !isFound) {
          // add to enabled list
          selectedConfig = {
            name: getConfigNameFromPath(path, datakitStat?.os_arch || ""),
            inputName: configSelected.inputName,
            path,
            sampleConfig: configSelected.sampleConfig,
            loaded: 2,
            dir: configSelected.dir,
          }

          if (configIndex > -1) {
            if (!enabledConfig[configIndex].children) {
              enabledConfig[configIndex].children = []
            }
            enabledConfig[configIndex].children?.push(selectedConfig)
          } else {
            // not found, add new input
            enabledConfig.push({
              inputName: configSelected.inputName,
              sampleConfig: configSelected.sampleConfig,
              expand: true,
              children: [selectedConfig]
            })
          }
        }

        if (selectedConfig) {
          checkConfig(selectedConfig)
        }
        setEnabledConfig([...enabledConfig])
        setCreateConfigVisible(false)
      }
    } else {
      // overwrite confirm
      if (!isForce && data === "file.path.exists") {
        Modal.confirm({
          title: t("confirm"),
          icon: <ExclamationCircleOutlined />,
          content: t("config.confirm_file_cover"),
          okText: t("confirm"),
          cancelText: t("cancel"),
          centered: true,
          onOk: async () => {
            isForce = true
            return doSaveConfig(true)
          }
        })
        return false
      }
      alertError(err)
      return false
    }

    return true
  }

  const onConfigSaveNameChange = (e) => {
    let name = e.target.value
    setConfigSaveName(name)
  }

  useEffect(() => {
    const separator = getSeparator(datakitStat?.os_arch || "")
    if (configSelected && configSelected.dir && configSaveName) {
      setConfigSavePath(`${configSelected.dir}${separator}${configSaveName}`)
    } else {
      setConfigSavePath("")
    }
  }, [configSaveName, configSelected, datakitStat])

  const checkConfig = async (config: ConfigInfo) => {
    if (!config.path) {
      alertError(t("config.confirm_file_not_exists"))
      return
    }
    if (!datakit) {
      return
    }
    const [err, content] = await getDatakitConfig(datakit, config.path)
    if (err) {
      console.error(err)
      alertError(t("config.get_fail"))
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
          title: t("confirm"),
          icon: <ExclamationCircleOutlined />,
          content: t('edit_state_away'),
          okText: t("confirm"),
          cancelText: t("cancel"),
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
          title: t("confirm"),
          icon: <ExclamationCircleOutlined />,
          content: t('edit_state_away'),
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
    }

    setIsEdit(false)
    setNewConfigSelected(config)
    setConfigSelected(null)
    setCode(config.sampleConfig)
    const separator = getSeparator(datakitStat?.os_arch || "")
    let dir = config.configDir
    setConfigSelected({
      inputName: config.inputName,
      path: `${config.configDir}${separator}${config.catalog ? config.catalog + separator : ''}${config.inputName}.conf`,
      sampleConfig: config.sampleConfig,
      config: config.sampleConfig,
      dir,
      isNew: true
    })

    setConfigSaveName(config.inputName + ".conf")
  }

  const onSaveConfig = async () => {
    doSaveConfig().then((isSuccess) => {
      isSuccess && setCreateConfigVisible(false)
    }).catch((err) => {
      alertError(err)
      return false
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
            name: getConfigNameFromPath(c.path, stat.os_arch),
            inputName,
            path: c.path,
            sampleConfig: info.sample_config,
            loaded: c.loaded,
            dir: path.dirname(c.path),
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
              <span>{t("config.configured")}</span>
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
              <span>Sample {t("list")}</span>
              <Tooltip title={t("config.sample_desc")} style={{ maxWidth: '300px' }} placement="right">
                <QuestionCircleFilled style={{ cursor: 'pointer', fontSize: '14px' }} />
              </Tooltip>
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
          <Modal
            title={t("save_as")}
            open={createConfigVisible}
            onOk={onSaveConfig}
            onCancel={() => setCreateConfigVisible(false)}
            destroyOnHidden={true}
            okText={t("confirm")}
            cancelText={t("cancel")}
          >
            <div style={{ padding: '10px 0' }}>
              <div style={{ marginBottom: '10px' }} >{t("config.new_config_name")}：</div>
              <div>
                <Input
                  value={configSaveName}
                  onChange={onConfigSaveNameChange}
                />
              </div>
              <div style={{ marginTop: '10px', color: '#888' }}>
                <Text  >{t("config.file_path")}: {configSavePath}</Text>
              </div>
            </div>
          </Modal >

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
                {configSelected.isNew ? t("config.new_config_message") : `${t("file_path")}： ${configSelected?.path}`}
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
                          {t("edit")}
                        </Button>
                      </>
                      :
                      <>
                        <Button type="primary" size="small" className="button" onClick={saveConfig}>
                          {configSelected.isNew ?
                            <Tooltip title={t("config.save_as_message")} placement="bottom">
                              <span>{t("save_as")}</span>
                            </Tooltip>
                            :
                            <span>{t("save")}</span>
                          }
                        </Button>
                        <Button size="small" className="button" onClick={cancelEdit}>
                          {t("cancel")}
                        </Button>
                      </>)
                  }
                  <Button type="default" size="small" onClick={() => toggleIsSample()}>
                    {isSample ? t("config.close_sample") : t("config.configure_sample")}
                  </Button>
                  <Button type="default" size="small" onClick={() => helpRedirect()}>
                    {t("help")}
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
