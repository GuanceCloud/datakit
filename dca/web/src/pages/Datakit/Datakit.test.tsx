import { render, screen } from '@testing-library/react';
import Datakit from './Datakit';

describe('Datakit page', () => {
  it('renders the placeholder page text', () => {
    render(<Datakit />);
    expect(screen.getByText('datakit page')).toBeInTheDocument();
  });
});
