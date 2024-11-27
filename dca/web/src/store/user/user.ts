import { createSlice, PayloadAction } from "@reduxjs/toolkit"

export interface User {
  email: string
  mobile?: string
  name: string
}

interface UserState {
  value: User
}


const initialState: UserState = {
  value: {
    email: "",
    mobile: "",
    name: "",
  }
}

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    set: (state, action: PayloadAction<User>) => {
      state.value = action.payload
    }
  }
})

export const { set } = userSlice.actions

export default userSlice.reducer