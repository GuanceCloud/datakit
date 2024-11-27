import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit'
import { PURGE } from 'redux-persist'
import { sleep } from 'src/helper/helper'
import { RootState } from '..'
import { IDatakit } from '../type'

interface DatakitState {
  currentDatakit: IDatakit | null
  value: Array<IDatakit>
}

const initialState: DatakitState = {
  currentDatakit: null,
  value: []
}

export const asyncAdd = createAsyncThunk(
  "datakit/asyncAdd",
  async (datakit: IDatakit) => {
    await sleep(1000)
    return datakit as IDatakit
  }
)

export const datakitSlice = createSlice({
  name: 'datakit',
  initialState,
  reducers: {
    add: (state, action: PayloadAction<IDatakit>) => {
      state.value.push(action.payload)
    },
    update: (state, action: PayloadAction<Array<IDatakit>>) => {
      state.value = action.payload
    },
    setCurrentDatakit: (state, action: PayloadAction<IDatakit | null>) => {
      state.currentDatakit = action.payload
    }
  },
  extraReducers: (builder) => { // async action
    builder.addCase(asyncAdd.fulfilled, (state, action: PayloadAction<IDatakit>) => {
      state.value.push(action.payload)
    })
    builder.addCase(PURGE, () => initialState)
  }
})

// actions
export const { add, update, setCurrentDatakit } = datakitSlice.actions

export const selectDatakits = (state: RootState) => state.datakit.value

export default datakitSlice.reducer