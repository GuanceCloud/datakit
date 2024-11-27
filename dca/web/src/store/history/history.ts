import { createSlice, PayloadAction } from "@reduxjs/toolkit"
import { PURGE } from "redux-persist"

export interface DatakitTab {
    key: string
}

export interface HistoryState {
    datakitTab: DatakitTab
}

const initialState: HistoryState = {
    datakitTab: {key: "1"}
}

export const historySlice = createSlice({
    name: 'history',
    initialState,
    reducers: {
        setDatakitTab: (state, action: PayloadAction<DatakitTab>) => {
            state.datakitTab = action.payload
        }
    },
    extraReducers(builder) {
        builder.addCase(PURGE, () => initialState)
    },
})

export const {setDatakitTab} = historySlice.actions

export default historySlice.reducer