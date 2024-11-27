import { QueryReturnValue } from '@reduxjs/toolkit/dist/query/baseQueryTypes';
import {
  BaseQueryFn,
  createApi,
  FetchArgs,
  fetchBaseQuery,
  FetchBaseQueryError,
  FetchBaseQueryMeta,
} from '@reduxjs/toolkit/query/react';
import { alertError } from 'src/helper/helper';
import { CONSOLE_AUTH_TOKEN_FAILED, DCA_AUTH_FAILED, ERRMSG, ResonseError } from './type';
import { clearStore } from '.';

const baseQuery = fetchBaseQuery({
  baseUrl: '',
});

export function getMsg(err: ResonseError): string {
  if (!err) {
    return "未知错误"
  }

  const errFormat = ERRMSG[err.errorCode]
  const msg = err.message || "未知错误"
  if (errFormat) {
    return errFormat.replace("{{msg}}", msg)
  }
  return msg
}

const fetchWithIntercept: BaseQueryFn<
  string | FetchArgs,
  unknown,
  FetchBaseQueryError
> = async (args, api, extraOptions) => {
  const result: QueryReturnValue<
    any,
    FetchBaseQueryError,
    FetchBaseQueryMeta
  > = await baseQuery(args, api, extraOptions);

  const { data, error } = result;
  console.log(data, error)
  if (!data) {
    console.error(error)
    alertError("Unexpected server error")
    return Promise.reject(error);
  }

  if (error) {
    alertError(error.data)
    return Promise.reject(error);
  }

  if (data?.code === 401) {
    if (data?.errorCode === CONSOLE_AUTH_TOKEN_FAILED) {
      // clear all cache items, or rtk query will be in pending status
      if (window.location.pathname !== "/login") {
        clearStore().then(() => {
          window.location.href = '/login'
        })
      }
    }

    if (data?.errorCode === DCA_AUTH_FAILED) {
      alertError(getMsg(data))
      return Promise.reject(data?.message)
    }
  }

  if (data?.code !== 200) {
    alertError(getMsg(data))
    return Promise.reject(data?.message)
  }

  return result

};

export const baseApi = createApi({
  baseQuery: fetchWithIntercept,
  reducerPath: 'baseApi',
  keepUnusedDataFor: 60, //cache time seconds 
  tagTypes: ['Datakit', 'Pipeline', 'Config', 'Workspace', 'CurrentWorkspace'],
  refetchOnMountOrArgChange: 30,
  endpoints: () => ({}),
});