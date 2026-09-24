import { render } from '@testing-library/react';
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
    <a href={path} data-slot="navbar-element" data-label={label} />
  ),
}));

describe('NavBar', () => {
  it('renders all four navbar links', () => {
    const { container } = render(<NavBar />);
    expect(container.querySelectorAll('[data-slot="navbar-element"]')).toHaveLength(4);
  });

  it('passes correct translated labels to links', () => {
    const { container } = render(<NavBar />);
    const elements = container.querySelectorAll('[data-slot="navbar-element"]');
    expect(elements[0]?.getAttribute('data-label')).toBe('DEPLOYMENTS.DEPLOYMENTS');
    expect(elements[1]?.getAttribute('data-label')).toBe('STATUS.STATUS');
    expect(elements[2]?.getAttribute('data-label')).toBe('LOGS.LOGS');
    expect(elements[3]?.getAttribute('data-label')).toBe('CONFIGURATION.CONFIGURATION');
  });

  it('passes correct paths to links', () => {
    const { container } = render(<NavBar />);
    const elements = container.querySelectorAll('[data-slot="navbar-element"]');
    expect(elements[0]?.getAttribute('href')).toBe('/deployments');
    expect(elements[1]?.getAttribute('href')).toBe('/status');
    expect(elements[2]?.getAttribute('href')).toBe('/logs');
    expect(elements[3]?.getAttribute('href')).toBe('/config');
  });
});
