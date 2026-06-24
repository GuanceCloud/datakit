import reducer, { setDatakitTab } from './history';
import { PURGE } from 'redux-persist';

describe('history slice', () => {
  it('updates the selected tab', () => {
    const state = reducer(undefined, setDatakitTab({ key: '2' }));
    expect(state.datakitTab.key).toBe('2');
  });

  it('resets on purge', () => {
    const state = reducer({ datakitTab: { key: '9' } }, { type: PURGE } as any);
    expect(state.datakitTab.key).toBe('1');
  });
});
