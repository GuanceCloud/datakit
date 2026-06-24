jest.mock('src/helper/helper', () => ({
  sleep: jest.fn(() => Promise.resolve()),
}));

import reducer, { add, asyncAdd, selectDatakits, setCurrentDatakit, update } from './datakit';
import { configureStore } from '@reduxjs/toolkit';
import { PURGE } from 'redux-persist';

describe('datakit slice', () => {
  const datakit = {
    id: 'dk-1',
    status: 'running',
  } as any;

  it('adds and updates datakits', () => {
    let state = reducer(undefined, add(datakit));
    state = reducer(state, setCurrentDatakit(datakit));
    state = reducer(state, update([datakit]));

    expect(state.value).toHaveLength(1);
    expect(state.currentDatakit?.id).toBe('dk-1');
  });

  it('handles asyncAdd fulfilled and purge', () => {
    let state = reducer(undefined, asyncAdd.fulfilled(datakit, 'req', datakit));
    expect(state.value).toHaveLength(1);

    state = reducer(state, { type: PURGE } as any);
    expect(state.value).toEqual([]);
  });

  it('selects datakits from root state', () => {
    expect(selectDatakits({ datakit: { value: [datakit] } } as any)).toEqual([datakit]);
  });

  it('dispatches asyncAdd with the configured thunk', async () => {
    const store = configureStore({
      reducer: {
        datakit: reducer,
      },
    });

    await store.dispatch(asyncAdd(datakit) as any);

    expect(store.getState().datakit.value).toHaveLength(1);
  });
});
