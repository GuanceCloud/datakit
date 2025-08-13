import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

export type IColumn = {
  key: string;
  title: string;
  isVisible: boolean;
}
interface ColumnState {
  host_name: IColumn;
  ip: IColumn;
  os_arch: IColumn;
  status_text: IColumn;
  uptime: IColumn;
  environment: IColumn;
  last_update: IColumn;
  is_container_running: IColumn;
  datakit_version: IColumn;
  operation: IColumn;
}

const initialState: ColumnState = {
  host_name: {
    key: "host_name",
    title: "host_name",
    isVisible: true,
  },
  ip: {
    key: "ip",
    title: "IP",
    isVisible: true,
  },
  os_arch: {
    key: "os_arch",
    title: "os_arch",
    isVisible: true,
  },
  status_text: {
    key: "status_text",
    title: "status_text",
    isVisible: true,
  },
  uptime: {
    key: "uptime",
    title: "uptime",
    isVisible: true,
  },
  environment: {
    key: "environment",
    title: "environment",
    isVisible: false,
  },
  last_update: {
    key: "last_update",
    title: "last_update",
    isVisible: true,
  },
  is_container_running: {
    key: "is_container_running",
    title: "is_container_running",
    isVisible: true,
  },
  datakit_version: {
    key: "datakit_version",
    title: "datakit_version",
    isVisible: true,
  },
  operation: {
    key: "operation",
    title: "operation",
    isVisible: true,
  },
}

const storageColumnsKey = "datakit_columns";

export const useAdditionalColumnOptions = () => {
  let columnsString = localStorage.getItem(storageColumnsKey);
  let columns = initialState

  if (columnsString) {
    try {
      columns = JSON.parse(columnsString);
    } catch (error) {
      console.log(error)
    }
  }
  const [newColumns, setNewColumns] = useState<ColumnState>(columns);
  const [translatedColumns, setTranslatedColumns] = useState<ColumnState>(columns);
  const handleColumnToggle = (key: string) => {
    let columns = { ...newColumns };
    columns[key].isVisible = !columns[key].isVisible;
    setNewColumns(columns);

    localStorage.setItem(storageColumnsKey, JSON.stringify(columns));
  }
  const { t } = useTranslation();
  useEffect(() => {
    if (newColumns) {
      let columns = { ...newColumns };
      for (const key in columns) {
        columns[key] = {
          ...columns[key],
          title: t(columns[key].title),
        }
      }
      setTranslatedColumns({ ...columns });
    }
  }, [t, newColumns]);
  return {
    translatedColumns,
    handleColumnToggle,
  }
}


