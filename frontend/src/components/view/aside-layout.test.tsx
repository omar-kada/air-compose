import { render, screen } from '@testing-library/react';
import { AsideLayout } from './aside-layout';

describe('AsideLayout', () => {
  it('renders header, aside, and main content', () => {
    render(
      <AsideLayout header={<span>Header</span>} aside={<span>Aside</span>}>
        <span>Main</span>
      </AsideLayout>,
    );
    expect(screen.getByText('Header')).toBeTruthy();
    expect(screen.getByText('Aside')).toBeTruthy();
    expect(screen.getByText('Main')).toBeTruthy();
  });

  it('does not render aside when not provided', () => {
    render(
      <AsideLayout header={<span>Header</span>}>
        <span>Main</span>
      </AsideLayout>,
    );
    expect(screen.queryByText('Aside')).toBeNull();
    expect(screen.getByText('Main')).toBeTruthy();
  });

  it('hides aside on mobile when focusMain is true', () => {
    const { container } = render(
      <AsideLayout header={<span>Header</span>} aside={<span>Aside</span>} focusMain>
        <span>Main</span>
      </AsideLayout>,
    );
    const aside = container.querySelector('aside');
    expect(aside?.className).toContain('hidden sm:flex');
  });

  it('shows aside on mobile when focusMain is false', () => {
    const { container } = render(
      <AsideLayout header={<span>Header</span>} aside={<span>Aside</span>} focusMain={false}>
        <span>Main</span>
      </AsideLayout>,
    );
    const aside = container.querySelector('aside');
    expect(aside?.className).not.toContain('hidden sm:flex');
  });
});
