import { renderHook, waitFor } from '@testing-library/react';

const mockTrigger = jest.fn();
const queryState: any = {};

jest.mock('src/store/consoleApi', () => ({
  useLazyGetAccountPermissionsQuery: () => [mockTrigger, queryState],
}));

import { useIsAdmin } from './useIsAdmin';

describe('useIsAdmin', () => {
  beforeEach(() => {
    mockTrigger.mockReset();
    Object.keys(queryState).forEach((key) => delete queryState[key]);
  });

  it('requests permissions on mount', () => {
    renderHook(() => useIsAdmin());

    expect(mockTrigger).toHaveBeenCalledTimes(1);
  });

  it('returns true for owner and wsAdmin roles', async () => {
    queryState.data = {
      code: 200,
      content: { roles: ['owner'] },
    };

    const { result, rerender } = renderHook(() => useIsAdmin());
    rerender();

    await waitFor(() => {
      expect(result.current).toBe(true);
    });
  });

  it('returns false when permissions are missing or not admin roles', async () => {
    queryState.data = {
      code: 200,
      content: { roles: ['readOnly'] },
    };

    const { result, rerender } = renderHook(() => useIsAdmin());
    rerender();

    await waitFor(() => {
      expect(result.current).toBe(false);
    });
  });
});
