import reducer, { setCurrentWorkspace, update } from './workspace';
import { PURGE } from 'redux-persist';

describe('workspace slice', () => {
  const workspace = {
    uuid: 'ws-1',
    name: 'demo',
    wsName: 'demo',
    extend: { isAdmin: true, role: 'owner' },
  } as any;

  it('updates current workspace and list', () => {
    let state = reducer(undefined, setCurrentWorkspace(workspace));
    state = reducer(state, update([workspace]));

    expect(state.currentWorkspace?.uuid).toBe('ws-1');
    expect(state.workspaces).toHaveLength(1);
  });

  it('resets on purge', () => {
    const state = reducer({ currentWorkspace: workspace, workspaces: [workspace] }, { type: PURGE } as any);
    expect(state.currentWorkspace).toBeNull();
    expect(state.workspaces).toEqual([]);
  });
});
