import React from 'react';
import { render, screen } from '@testing-library/react';
import App from './App';

jest.mock('./router', () => ({
  __esModule: true,
  default: { routes: [] },
}));

jest.mock('react-router-dom', () => ({
  RouterProvider: () => <div>router mounted</div>,
}));

test('renders app router', () => {
  render(<App />);
  const linkElement = screen.getByText(/router mounted/i);
  expect(linkElement).toBeInTheDocument();
});
