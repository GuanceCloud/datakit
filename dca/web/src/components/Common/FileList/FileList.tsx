import { CaretDownOutlined, CaretRightOutlined, QuestionCircleOutlined } from "@ant-design/icons"
import { Tooltip, Typography } from "antd"
import { CSSProperties, MouseEvent, useState } from "react"
import './FileList.scss'

const {Text} = Typography

export type Item = {
  name: string
  content: string
  fileName?: string
  expand?: boolean
  hiddenArrow?: boolean
  hiddenName?:boolean
  hiddenChildListIcon?:boolean
  children?: Array<Item>
  style?: CSSProperties
  icon?: string
  path?: string
  tooltip?: string
  checked?: () => boolean
  onClick?: () => void
}

export type FileListProps = {
  list: Array<Item>
  title: string
  icon?: JSX.Element
  titleStyle?: React.CSSProperties
  selected: Item | null
  colWidth?: number
  setSelected: (item: Item | null) => void
  onBeforeSelected?: () => Promise<boolean>
  onAfterSelected?: (item: Item) => Promise<void>
}

export default function FileList({list, icon, titleStyle, title, selected, setSelected, onBeforeSelected, onAfterSelected, colWidth}: FileListProps) {
  const [listState, setListState] = useState<Array<Item>>([])
  
  const toggle = (v) => {
    v.expand = !v.expand
    setListState([...listState])
  }

  const checkItem = (item: Item) => {
    return async (event: MouseEvent<HTMLDivElement>) => {
      event.stopPropagation()
      let isCheck = true
      if (onBeforeSelected) {
        isCheck = await onBeforeSelected()
      }
      if (isCheck) {
        setSelected(item)
        if (onAfterSelected) {
          await onAfterSelected(item)
        }
      }
    }
  }

  return <>
    <div className="container">
      <div className="title" style={titleStyle}>
        {icon}
        <span className="text">{title}</span>
      </div>
      <div className="list">
        {
          list.map((v, index) => {
            return (
              <div className="list-item" key={index} style={v.style}>
                {
                  v.hiddenName 
                    || 
                  <div className="title" onClick={() => toggle(v)}>
                    {
                      v.hiddenArrow || <div className="icon">
                      {v.expand ? <CaretDownOutlined /> : <CaretRightOutlined />}
                      </div>
                    }
                    <div className="name" >{v.name}</div>
                    {
                      v.tooltip &&
                        <div className="tooltip">
                          <Tooltip title={v.tooltip} placement="right">
                            <QuestionCircleOutlined />
                          </Tooltip>
                        </div>
                    }
                  </div>
                }
                
                <div className={!v.expand ? "content hidden" : "content"}>
                  {v.children?.map((child, index) => {
                    let style: CSSProperties = child.style || {}
                    if (child.icon) {
                      Object.assign(style, {
                        backgroundImage: `url(${child.icon})`,
                        backgroundPosition: "right",
                        backgroundRepeat: "no-repeat",
                        backgroundSize: "12px"
                      })
                    }
                    return (
                      <Tooltip key={child.name} title={child.name} placement="right" mouseEnterDelay={0.5}>
                        <div 
                          className={selected?.name === child.name ? 'content-item selected' : "content-item"} 
                          key={index} 
                          onClick={checkItem(child)}
                          style={style}
                        >
                            <Text style={{width: colWidth ? (colWidth - 50) + 'px' : "100px"}} ellipsis={true}>
                              {v.hiddenChildListIcon ? child.name : `- ${child.name}`}
                            </Text>
                        </div>
                      </Tooltip>
                    )
                  })}
                </div>
                
              </div>
            )
          })
        }
      </div>
    </div>
  </>
}