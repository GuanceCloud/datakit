import { configureStore } from '@reduxjs/toolkit'
import { combineReducers } from 'redux'
import storageSession from 'redux-persist/lib/storage/session'
import autoMergeLevel2 from 'redux-persist/lib/stateReconciler/autoMergeLevel2';
import { persistStore, persistReducer } from 'redux-persist';

import datakit from './datakit/datakit'
import user from './user/user'
import workspace from './workspace/workspace'
import history from './history/history'
import { baseApi } from 'src/store/baseApi';
import { setupListeners } from '@reduxjs/toolkit/query';

const reducers = combineReducers({
  datakit,
  workspace,
  user,
  history,
  [baseApi.reducerPath]: baseApi.reducer
})

const persistConfig = {
  key: 'root',
  storage: storageSession,//storage,
  stateReconciler: autoMergeLevel2
};

const persistReducers = persistReducer<any, any>(persistConfig, reducers)

const store = configureStore({
  reducer: persistReducers,
  middleware: (getDefaultMiddleware) => {
    return getDefaultMiddleware({
      serializableCheck: false
    }).concat(baseApi.middleware)
  }
})

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export default store

export const persistor = persistStore(store);

setupListeners(store.dispatch)

export function clearStore(): Promise<any> {
  return persistor.purge()
}