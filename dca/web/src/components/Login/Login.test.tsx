import { render, screen } from '@testing-library/react';
import Login from './Login';

jest.mock('antd', () => ({
  Button: ({ children }: any) => <button>{children}</button>,
  Flex: ({ children }: any) => <div>{children}</div>,
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('Login', () => {
  it('renders the login redirect CTA', () => {
    render(<Login />);

    expect(screen.getByText('login_dca_redirect')).toBeInTheDocument();
    expect(screen.getByText('go_forward')).toBeInTheDocument();
    expect(screen.getByRole('link')).toHaveAttribute('href', '/console/dca');
  });
});
