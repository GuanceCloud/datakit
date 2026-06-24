import { render, screen } from '@testing-library/react';
import { EllipsisMiddle } from './Ellipsis';

jest.mock('antd', () => ({
  Typography: {
    Text: ({ children }: any) => <span>{children}</span>,
  },
}));

describe('EllipsisMiddle', () => {
  it('renders the text without the suffix segment', () => {
    render(<EllipsisMiddle suffixCount={4} maxWidth={120}>abcdef1234</EllipsisMiddle>);

    expect(screen.getByText('abcdef')).toBeInTheDocument();
  });
});
