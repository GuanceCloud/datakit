import { renderHook } from '@testing-library/react';
import useRouter from './useRouter';

const mockNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

describe('useRouter', () => {
  it('returns the navigate function from react-router', () => {
    const { result } = renderHook(() => useRouter());

    result.current.navigate('/dashboard');

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });
});
