import { act, renderHook } from '@testing-library/react';
import { useAdditionalColumnOptions } from './useAdditionalColumnOptions';

const mockT = (key: string) => `translated:${key}`;

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: mockT,
  }),
}));

describe('useAdditionalColumnOptions', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('loads defaults, translates titles and toggles visibility', () => {
    const { result } = renderHook(() => useAdditionalColumnOptions());

    expect(result.current.translatedColumns.host_name.title).toBe('translated:host_name');

    act(() => {
      result.current.handleColumnToggle('environment');
    });

    expect(result.current.translatedColumns.environment.isVisible).toBe(true);
    expect(localStorage.getItem('datakit_columns')).toContain('"environment"');
  });

  it('drops invalid local storage payloads', () => {
    localStorage.setItem('datakit_columns', 'not-json');
    renderHook(() => useAdditionalColumnOptions());

    expect(localStorage.getItem('datakit_columns')).toBeNull();
  });
});
