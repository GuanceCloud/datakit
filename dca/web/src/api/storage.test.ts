import { delToken, delWorkspaceID, getToken, getWorkspaceID, saveToken, saveWorkspaceID } from './storage';

describe('storage helpers', () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it('stores and clears the auth token', () => {
    expect(getToken()).toBe('');

    saveToken('secret');
    expect(getToken()).toBe('secret');

    delToken();
    expect(getToken()).toBe('');
  });

  it('stores and clears the workspace id', () => {
    expect(getWorkspaceID()).toBe('');

    saveWorkspaceID('ws-1');
    expect(getWorkspaceID()).toBe('ws-1');

    delWorkspaceID();
    expect(getWorkspaceID()).toBe('');
  });
});
