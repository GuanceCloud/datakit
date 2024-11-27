import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { PURGE } from 'redux-persist'
import { IWorkspace } from '../type'


interface WorkspaceState {
  currentWorkspace: IWorkspace | null,
  workspaces: Array<IWorkspace>
}

const initialState: WorkspaceState = {
  currentWorkspace: null,
  workspaces: []
}

export const workspaceSlice = createSlice({
  name: 'workspace',
  initialState,
  reducers: {
    setCurrentWorkspace: (state, action: PayloadAction<IWorkspace|null>) => {
      state.currentWorkspace = action.payload
    },
    update: (state, action: PayloadAction<Array<IWorkspace>>) => {
      state.workspaces = action.payload
    }
  },
  extraReducers(builder) {
    builder.addCase(PURGE, () => initialState)
},
})

export const { setCurrentWorkspace, update } = workspaceSlice.actions

export default workspaceSlice.reducer

// export type WorkspaceAction = {
//   type: string
//   workspace: IWorkspace
//   workspaces: Array<IWorkspace>
// }

// export function setWorkspace(workspace: IWorkspace) {
//   return {
//     type: SET_WORKSPACE,
//     workspace
//   }
// }

// export function updateWorkspaces(workspaces: Array<IWorkspace>) {
//   return {
//     type: UPDATE_WORKSPACE_LIST,
//     workspaces
//   }
// }

// const defaultWorkspacesState: WorkspaceState = {
//   currentWorkspace: null,
//   workspaces: []
// }

// function workspaces(state: WorkspaceState = defaultWorkspacesState, action: WorkspaceAction) {
//   let newState: WorkspaceState
//   switch (action.type) {
//     case SET_WORKSPACE:
//       newState = { ...state }
//       newState.currentWorkspace = action.workspace
//       return newState
//     case UPDATE_WORKSPACE_LIST:
//       newState = { ...state }
//       newState.workspaces = [...action.workspaces]
//       return newState
//     default:
//       return state
//   }
// }

// export default workspaces