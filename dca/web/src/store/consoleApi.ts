import { baseApi } from "./baseApi"
import type { IAccountInfo, IAccountPermission, IVersion, IWorkspace, PageInfo } from "./type"



type ConsoleResponse<T> = {
  code: number
  content: T
  errorCode: string
  message: string
  success: boolean
}

type ConsolePageResponse<T> = ConsoleResponse<
  {
    data: T
    pageInfo?: PageInfo
  }
>

const getURL = (url: string) => {
  return `/api/console/${url}`
}

const consoleApi = baseApi.injectEndpoints({
  overrideExisting: false,
  endpoints(builder) {
    return {
      getWorkspaceList: builder.query<ConsolePageResponse<IWorkspace[]>, void>({
        query: () => {
          return {
            url: getURL("workspaceList"),
          }
        },
        keepUnusedDataFor: 5, // cache time 
        // providesTags(result) {
        //   return result && result.map(({ id }) => ({ type: "Workspace", id }))
        // },
      }),
      getCurrentWorkspace: builder.query<IWorkspace, void>({
        query: () => {
          return {
            url: getURL("currentWorkspace"),
          }
        },
        transformResponse(response: { content: IWorkspace }) {
          return response.content
        },
      }),
      changeWorkspace: builder.mutation<ConsoleResponse<any>, string>(
        {
          query(workspaceID: string) {
            return {
              url: getURL("changeWorkspace"),
              method: "post",
              body: {
                workspaceUUID: workspaceID
              },
              headers: {
                "X-Workspace-Uuid": workspaceID,
              }
            }
          },
        }
      ),
      getAccountPermissions: builder.query<ConsoleResponse<IAccountPermission>, void>({
        query: () => {
          return {
            url: getURL("accountPermissions"),
          }
        },
      }),
      getCurrentAccount: builder.query<ConsoleResponse<IAccountInfo>, void>({
        query: () => {
          return {
            url: getURL("currentAccount"),
          }
        },
      }),
      getDatakitVersion: builder.query<ConsoleResponse<IVersion>, any>({
        query: () => {
          return {
            url: "/api/lastDatakitVersion"
          }
        }
      }),
      logout: builder.query<ConsoleResponse<any>, void>({
        query: () => {
          return {
            method: "post",
            url: getURL("logout"),
          }
        },
      }),
    }
  }
})

export const {
  useLazyGetWorkspaceListQuery,
  useGetCurrentWorkspaceQuery,
  useLazyGetCurrentWorkspaceQuery,
  useChangeWorkspaceMutation,
  useLazyGetAccountPermissionsQuery,
  useLazyGetCurrentAccountQuery,
  useLazyLogoutQuery,
  useLazyGetDatakitVersionQuery,
} = consoleApi

export default consoleApi