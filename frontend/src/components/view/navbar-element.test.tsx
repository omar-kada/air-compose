import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Home } from 'lucide-react';
import { NavbarElement } from './navbar-element';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

describe('NavbarElement', () => {
  it('renders translated label and Icon inside a link', () => {
    render(
      <MemoryRouter>
        <NavbarElement label="NAV.HOME" path="/non-active" Icon={Home} />
      </MemoryRouter>,
    );
    expect(screen.getByText('translated:NAV.HOME')).toBeTruthy();
    expect(screen.getByRole('link')).toBeTruthy();
  });

  it('applies opacity-50 when path does not match', () => {
    render(
      <MemoryRouter>
        <NavbarElement label="NAV.HOME" path="/non-active" Icon={Home} />
      </MemoryRouter>,
    );
    const link = screen.getByRole('link');
    expect(link.className).toContain('opacity-50');
  });

  it('does not apply opacity-50 when path matches', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <NavbarElement label="NAV.HOME" path="/" Icon={Home} />
      </MemoryRouter>,
    );
    const link = screen.getByRole('link');
    expect(link.className).not.toContain('opacity-50');
  });
});
