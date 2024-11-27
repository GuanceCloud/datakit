import { useContext, useEffect, useState } from 'react'
import { Button } from "antd";

import DCAEditor from "src/components/DCAEditor/DCAEditor";
import { alertError } from 'src/helper/helper';
import { DkInfoContext, Nodata } from "../DkInfo";
import './BlackList.scss'
import { useLazyGetFilterQuery } from 'src/store/datakitApi';

const editorOptions = { readOnly: true, mode: "javascript", cursorBlinkRate: 0 }

export default function BlackList() {
  const [code, setCode] = useState<string>("")
  const [filterPath, setFilterPath] = useState<string>("")
  const { datakit } = useContext(DkInfoContext)

  const helpRedirect = () => {
    window.open("https://docs.guance.com/management/overall-blacklist/")
  }

  const [getFilter, { data: filterResponse, isLoading, isError }] = useLazyGetFilterQuery()

  useEffect(() => {
    if (filterResponse?.code !== 200) {
      alertError(filterResponse)
      return
    }

    let { content, filePath } = filterResponse?.content
    try {
      let parsedContent = JSON.parse(content)
      content = JSON.stringify(parsedContent, null, '  ')
    } catch (err) {
      console.error(err)
    }

    setCode(content)
    setFilterPath(filePath)
  }, [filterResponse])

  useEffect(() => {
    datakit && getFilter(datakit)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="blacklist-container">
      {
        filterPath === ""
          ?
          <Nodata loading={isLoading} isError={isError} refresh={() => { datakit && getFilter(datakit) }} />
          :
          <>
            <div className="setting">
              <div className="path">
                {`路径： ${filterPath}`}
              </div>
              <div className="edit">
                <Button type="default" size="small" onClick={() => helpRedirect()}>
                  帮助
                </Button>
              </div>
            </div>
            <div className="content">
              <DCAEditor value={code} setValue={setCode} editorOptions={editorOptions} />
            </div>
          </>
      }
    </div>

  )
}