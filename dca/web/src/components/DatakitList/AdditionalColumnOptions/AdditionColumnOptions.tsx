import { Button, Form, Popover, Space, Switch } from "antd"
import { useAdditionalColumnOptions } from "./useAdditionalColumnOptions";
import { useTranslation } from "react-i18next";
import { useEffect } from "react";

interface AdditionColumnOptionsProps {
  onValueChange: (columnInfo: Record<string, boolean>) => void;
}
export const AdditionColumnOptions = (props: AdditionColumnOptionsProps) => {
  const { onValueChange } = props;
  const { translatedColumns: columns, handleColumnToggle } = useAdditionalColumnOptions();
  const { t } = useTranslation();
  useEffect(() => {
    let columnInfo: Record<string, boolean> = {}
    for (let key in columns) {
      columnInfo[key] = columns[key].isVisible
    }
    onValueChange(columnInfo)
  }, [columns, onValueChange])
  return (
    <Popover
      placement="bottomRight"
      arrow={false}
      content={
        <Form
          style={{
            width: 180,
          }}
          labelCol={{ span: 18, offset: 1 }}
          labelAlign="left"
          layout="horizontal"
          size="middle"

        >
          <Space direction="vertical">
            {columns && Object.keys(columns).map(key => (
              <Form.Item key={key} label={t(columns[key].key)} style={{ marginBottom: 0 }}>
                <Switch
                  checked={columns[key].isVisible}
                  onChange={() => handleColumnToggle(key)}
                />
              </Form.Item>
            ))}
          </Space>
        </Form>
      }
      trigger="click"
    >

      <Button type="default" size="small" className="button">
        {t("display_column")}
      </Button>

    </Popover>
  )
}