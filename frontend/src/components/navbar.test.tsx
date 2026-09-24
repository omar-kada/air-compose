import { render, screen } from '@testing-library/react';
import { NavBar } from './navbar';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
  }),
}));

vi.mock('@/lib', () => ({
  ROUTES: {
    DEPLOYMENTS: '/deployments',
    STATUS: '/status',
    LOGS: '/logs',
    CONFIG: '/config',
  },
}));

vi.mock('./view', () => ({
  NavbarElement: ({ label, path }: { label: string; path: string }) => (
    <a href={path} data-slot="navbar-element">
      {label}
    </a>
  ),
}));

describe('NavBar', () => {
  it('renders nav links with correct labels and paths', () => {
    render(<NavBar />);
    expect(screen.getAllByRole('link')).toHaveLength(4);
    expect(
      screen.getByRole('link', { name: 'DEPLOYMENTS.DEPLOYMENTS' }).getAttribute('href'),
    ).toBe('/deployments');
    expect(
      screen.getByRole('link', { name: 'STATUS.STATUS' }).getAttribute('href'),
    ).toBe('/status');
    expect(
      screen.getByRole('link', { name: 'LOGS.LOGS' }).getAttribute('href'),
    ).toBe('/logs');
    expect(
      screen.getByRole('link', { name: 'CONFIGURATION.CONFIGURATION' }).getAttribute('href'),
    ).toBe('/config');
  });
});
