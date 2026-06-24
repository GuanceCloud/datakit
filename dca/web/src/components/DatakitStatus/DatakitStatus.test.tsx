import { render, screen } from '@testing-library/react';
import DatakitStatus from './DatakitStatus';

jest.mock('antd', () => ({
  Tooltip: ({ children }: any) => <div>{children}</div>,
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('DatakitStatus', () => {
  it('renders current datakit status text', () => {
    render(<DatakitStatus datakit={{ status: 'running' } as any} />);
    expect(screen.getByText('running')).toBeInTheDocument();
  });

  it('falls back to unknown when status is missing', () => {
    render(<DatakitStatus datakit={{} as any} />);
    expect(screen.getByText('unknown')).toBeInTheDocument();
  });
});
