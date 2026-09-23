import { render, screen } from '@testing-library/react';
import { HeaderLayout } from './header-layout';

describe('HeaderLayout', () => {
  it('renders header, separator, and children', () => {
    render(<HeaderLayout header={<span>Header</span>}>Content</HeaderLayout>);
    expect(screen.getByText('Header')).toBeTruthy();
    expect(screen.getByText('Content')).toBeTruthy();
    expect(document.querySelector('[data-slot="separator"]')).not.toBeNull();
  });

  it('does not render header or separator when null', () => {
    render(<HeaderLayout header={null}>Content</HeaderLayout>);
    expect(screen.getByText('Content')).toBeTruthy();
    expect(document.querySelector('[data-slot="separator"]')).toBeNull();
  });
});
