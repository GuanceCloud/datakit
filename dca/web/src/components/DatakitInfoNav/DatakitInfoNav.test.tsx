import { fireEvent, render, screen } from '@testing-library/react';
import { DatakitInfoNav } from './DatakitInfoNav';

const mockNavigate = jest.fn();
const mockLocation = { pathname: '/dashboard/runinfo' };
const mockT = (key: string) => key;

jest.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useLocation: () => mockLocation,
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: mockT,
  }),
}));

jest.mock('antd', () => ({
  Tabs: ({ items, onChange }: any) => (
    <div>
      {items.map((item: any) => (
        <button key={item.key} onClick={() => onChange(item.key)}>{item.label}</button>
      ))}
    </div>
  ),
}));

describe('DatakitInfoNav', () => {
  beforeEach(() => {
    mockNavigate.mockReset();
  });

  it('renders nav items and navigates on tab change', () => {
    render(<DatakitInfoNav datakit={{ run_in_container: false } as any} />);

    fireEvent.click(screen.getByText('label.pipeline'));

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/pipeline');
  });
});
