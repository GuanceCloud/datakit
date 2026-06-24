import { fireEvent, render, screen } from '@testing-library/react';
import { AdditionColumnOptions } from './AdditionColumnOptions';

const mockHandleColumnToggle = jest.fn();

jest.mock('./useAdditionalColumnOptions', () => ({
  useAdditionalColumnOptions: () => ({
    translatedColumns: {
      host_name: { key: 'host_name', title: 'host_name', isVisible: true },
      ip: { key: 'ip', title: 'IP', isVisible: false },
    },
    handleColumnToggle: (...args: any[]) => mockHandleColumnToggle(...args),
  }),
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

jest.mock('antd', () => ({
  Button: ({ children }: any) => <button>{children}</button>,
  Form: Object.assign(({ children }: any) => <div>{children}</div>, {
    Item: ({ children, label }: any) => <label>{label}{children}</label>,
  }),
  Popover: ({ children, content }: any) => <div>{children}{content}</div>,
  Space: ({ children }: any) => <div>{children}</div>,
  Switch: ({ checked, onChange }: any) => <input aria-label={`switch-${checked}`} type="checkbox" checked={checked} onChange={() => onChange()} />,
}));

describe('AdditionColumnOptions', () => {
  it('reports visible columns and toggles options', () => {
    const onValueChange = jest.fn();
    render(<AdditionColumnOptions onValueChange={onValueChange} />);

    expect(onValueChange).toHaveBeenCalledWith({
      host_name: true,
      ip: false,
    });

    fireEvent.click(screen.getAllByRole('checkbox')[0]);
    expect(mockHandleColumnToggle).toHaveBeenCalled();
  });
});
