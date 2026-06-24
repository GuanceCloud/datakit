import reducer, { set } from './user';

describe('user slice', () => {
  it('updates the current user', () => {
    const state = reducer(undefined, set({
      email: 'demo@example.com',
      mobile: '13800138000',
      name: 'demo',
    }));

    expect(state.value.name).toBe('demo');
    expect(state.value.email).toBe('demo@example.com');
  });
});
