import { baseApi } from "./baseApi"
import { IDatakitResponse, IDatakit, IDatakitStat, IFilter, PageQuery, PageInfo } from "./type"

const getURL = (datakit: IDatakit, url: string) => {
  return `/api/datakit/${url}?datakit_id=${datakit.id}`
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
        query: ({ pageIndex = 1, pageSize = 10, search = "", minLastUpdateTime }) => {
          let url = `/api/datakit/list?pageIndex=${pageIndex}&pageSize=${pageSize}`
          if (search) {
            url += `&search=${search}`
          }
          if (minLastUpdateTime) {
            url += `&minLastUpdateTime=${minLastUpdateTime}`
          }
          return {
            url,
            method: "get",
          }
        },
        keepUnusedDataFor: 0,
      }),
      getDatakitListByID: builder.query<IDatakitResponse<IDatakit[]>, { ids: string }>({
        query: ({ ids }) => {
          let url = `/api/datakit/listByID?ids=${ids}`
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
} = datakitApi

export default datakitApi