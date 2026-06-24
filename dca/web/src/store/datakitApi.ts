import { baseApi } from "./baseApi"
import { IDatakitResponse, IDatakit, IDatakitStat, IFilter, PageQuery, PageInfo, ISearchValue } from "./type"

export const buildDatakitURL = (path: string, params?: Record<string, unknown>) => {
  if (!params) {
    return path
  }

  const query = Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== null && value !== "")
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)

  if (query.length === 0) {
    return path
  }

  return `${path}?${query.join("&")}`
}

const getURL = (datakit: IDatakit, url: string) => {
  return buildDatakitURL(`/api/datakit/${url}`, {
    datakit_id: datakit.id,
  })
}

type DatakitPageResponse<T> = IDatakitResponse<
  {
    data: T
    pageInfo?: PageInfo
  }
>
interface DatakitListParams extends PageQuery {
  search?: string
  version?: string
  isOnline?: string
  minLastUpdateTime?: number
  filter?: string
}

const datakitApi = baseApi.injectEndpoints({
  overrideExisting: false,
  endpoints(builder) {
    return {
      getDatakitStat: builder.query<IDatakitResponse<IDatakitStat>, IDatakit>({
        query: (datakit) => {
          return {
            url: getURL(datakit, "stats"),
          }
        }
      }),
      getDatakitList: builder.query<DatakitPageResponse<IDatakit[]>, DatakitListParams>({
        query: ({ pageIndex = 1, pageSize = 10, search = "", filter = "", minLastUpdateTime }) => {
          const url = buildDatakitURL("/api/datakit/list", {
            pageIndex,
            pageSize,
            search,
            filter,
            minLastUpdateTime,
          })
          return {
            url,
            method: "get",
          }
        },
        keepUnusedDataFor: 0,
      }),
      getDatakitListByID: builder.query<IDatakitResponse<IDatakit[]>, { ids: string }>({
        query: ({ ids }) => {
          const url = buildDatakitURL("/api/datakit/listByID", {
            ids,
          })
          return {
            url,
            method: "get",
          }
        },
        keepUnusedDataFor: 0,
      }),
      getSearchValue: builder.query<IDatakitResponse<ISearchValue>, void>({
        query: () => {
          let url = `/api/datakit/searchValue`
          return {
            url,
            method: "get",
          }
        },
        keepUnusedDataFor: 0,
      }),
      getDatakitByID: builder.query<any, any>({
        query: (id) => {
          return {
            url: `datakits/${id}`,
          }
        }
      }),
      getFilter: builder.query<IDatakitResponse<IFilter>, IDatakit>({
        query: (datakit) => {
          return {
            url: getURL(datakit, "filter")
          }
        }
      }),
      reloadDatakit: builder.query<IDatakitResponse<any>, IDatakit>({
        query: (datakit) => {
          return {
            url: getURL(datakit, "restart"),
            method: "PUT"
          }
        },
      }),
      upgradeDatakit: builder.query<IDatakitResponse<any>, IDatakit>({
        query: (datakit) => {
          return {
            url: getURL(datakit, "upgrade"),
            method: "POST"
          }
        },
      }),
      operateDatakit: builder.query<IDatakitResponse<any>, { ids: string, type: string }>({
        query: ({ ids, type }) => {
          return {
            url: buildDatakitURL(`/api/datakit/operation/${encodeURIComponent(type)}`, {
              ids,
            }),
            method: "POST"
          }
        },
      }),
    }
  }
})

export const {
  useLazyGetDatakitStatQuery,
  useLazyGetFilterQuery,
  useLazyReloadDatakitQuery,
  useLazyUpgradeDatakitQuery,
  useLazyGetDatakitListQuery,
  useLazyGetDatakitListByIDQuery,
  useLazyOperateDatakitQuery,
  useLazyGetSearchValueQuery,
} = datakitApi

export default datakitApi
