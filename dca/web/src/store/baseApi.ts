import { QueryReturnValue } from '@reduxjs/toolkit/query';
import {
  BaseQueryFn,
  createApi,
  FetchArgs,
  fetchBaseQuery,
  FetchBaseQueryError,
  FetchBaseQueryMeta,
} from '@reduxjs/toolkit/query/react';
import { alertError } from 'src/helper/helper';
import { CONSOLE_AUTH_TOKEN_FAILED, DCA_AUTH_FAILED, ResonseError } from './type';
import { clearStore } from '.';

import i18n from '../i18n';

const baseQuery = fetchBaseQuery({
  baseUrl: '',
});

function isResponseError(value: unknown): value is ResonseError {
  return !!value && typeof value === "object" && "errorCode" in value
}

function isFetchAbortError(error: FetchBaseQueryError): boolean {
  return "status" in error
    && error.status === "FETCH_ERROR"
    && "error" in error
    && typeof error.error === "string"
    && error.error.includes("AbortError")
}

export function getMsg(err: ResonseError, params?: Record<string, string>): string {
  if (!err) {
    return i18n.t("api.unknown_error")
  }

  let msg = i18n.t("api." + err.errorCode, params)
  if (err.message) {
    msg += ": " + formatErrorMessage(err.message)
  }

  return msg
}

function formatErrorMessage(message: unknown): string {
  if (typeof message === "string") {
    return message
  }

  if (message instanceof Error) {
    return message.message
  }

  if (typeof message === "object") {
    try {
      return JSON.stringify(message)
    } catch {
      return String(message)
    }
  }

  return String(message)
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
  if (error) {
    if (isFetchAbortError(error)) {
      return Promise.reject(error);
    }

    const errorData = "data" in error ? error.data : undefined

    if (isResponseError(errorData)) {
      alertError(getMsg(errorData))
    } else if (typeof errorData === "string" && errorData) {
      alertError(errorData)
    } else {
      console.error(error)
      alertError("Unexpected server error")
    }

    return Promise.reject(errorData || error);
  }

  if (!data) {
    alertError("Unexpected server error")
    return Promise.reject(new Error("empty response data"));
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

export { fetchWithIntercept };

export const baseApi = createApi({
  baseQuery: fetchWithIntercept,
  reducerPath: 'baseApi',
  keepUnusedDataFor: 60, //cache time seconds 
  tagTypes: ['Datakit', 'Pipeline', 'Config', 'Workspace', 'CurrentWorkspace'],
  refetchOnMountOrArgChange: 30,
  endpoints: () => ({}),
});
