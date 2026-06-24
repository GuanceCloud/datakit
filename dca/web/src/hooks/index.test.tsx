import { useAppDispatch, useAppSelector } from './index';
import { renderHook } from '@testing-library/react';

const mockDispatch = jest.fn();
const mockUseSelector = jest.fn((selector: any) => selector({ value: 1 }));

jest.mock('react-redux', () => ({
  useDispatch: () => mockDispatch,
  useSelector: (selector: any) => mockUseSelector(selector),
}));

describe('typed hooks exports', () => {
  it('proxies to react-redux hooks', () => {
    const { result: dispatchResult } = renderHook(() => useAppDispatch());

    expect(dispatchResult.current).toBe(mockDispatch);
    expect(typeof useAppSelector).toBe('function');
  });
});
