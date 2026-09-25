import { render, screen } from '@testing-library/react';
import { NavBar } from './navbar';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
  }),
}));

vi.mock('react-router-dom', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to} data-slot="link">
      {children}
    </a>
  ),
  useMatch: () => null,
}));

describe('NavBar', () => {
  it('renders nav links with correct labels and paths', () => {
    render(<NavBar />);
    expect(screen.getAllByRole('link')).toHaveLength(4);
    expect(
      screen.getByRole('link', { name: 't:DEPLOYMENTS.DEPLOYMENTS' }).getAttribute('href'),
    ).toBe('/deployments');
    expect(screen.getByRole('link', { name: 't:STATUS.STATUS' }).getAttribute('href')).toBe(
      '/status',
    );
    expect(screen.getByRole('link', { name: 't:LOGS.LOGS' }).getAttribute('href')).toBe('/logs');
    expect(
      screen.getByRole('link', { name: 't:CONFIGURATION.CONFIGURATION' }).getAttribute('href'),
    ).toBe('/config');
  });
});
