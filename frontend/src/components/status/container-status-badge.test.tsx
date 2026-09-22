import { render, screen } from '@testing-library/react';
import { ContainerStatusBadge } from './container-status-badge';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

describe('ContainerStatusBadge', () => {
  it('renders healthy badge with correct text and icon', () => {
    const { container } = render(<ContainerStatusBadge health="healthy" />);
    expect(screen.getByText('translated:healthy')).toBeTruthy();
    expect(container.querySelector('svg')).not.toBeNull();
    expect(
      container.querySelector('[data-slot="badge"]')?.classList.contains('border-green-400'),
    ).toBe(true);
  });

  it('renders unhealthy badge with red border', () => {
    const { container } = render(<ContainerStatusBadge health="unhealthy" />);
    expect(screen.getByText('translated:unhealthy')).toBeTruthy();
    expect(
      container.querySelector('[data-slot="badge"]')?.classList.contains('border-red-400'),
    ).toBe(true);
  });

  it('renders starting badge with blue border', () => {
    const { container } = render(<ContainerStatusBadge health="starting" />);
    expect(screen.getByText('translated:starting')).toBeTruthy();
    expect(
      container.querySelector('[data-slot="badge"]')?.classList.contains('border-blue-400'),
    ).toBe(true);
  });

  it('renders unknown badge when health is not provided', () => {
    render(<ContainerStatusBadge />);
    expect(screen.getByText('translated:unknown')).toBeTruthy();
  });

  it('renders iconOnly badge without text', () => {
    const { container } = render(<ContainerStatusBadge health="healthy" iconOnly />);
    expect(screen.queryByText('translated:healthy')).toBeNull();
    expect(container.querySelector('svg')).not.toBeNull();
  });

  it('uses custom label instead of health value', () => {
    render(<ContainerStatusBadge health="healthy" label="my-label" />);
    expect(screen.getByText('translated:my-label')).toBeTruthy();
  });

  it('applies custom className to Badge', () => {
    const { container } = render(
      <ContainerStatusBadge health="healthy" className="custom-class" />,
    );
    expect(container.querySelector('[data-slot="badge"]')?.classList.contains('custom-class')).toBe(
      true,
    );
  });

  it('uses state for border when both state and health are provided', () => {
    const { container } = render(<ContainerStatusBadge health="healthy" state="dead" />);
    expect(
      container.querySelector('[data-slot="badge"]')?.classList.contains('border-red-400'),
    ).toBe(true);
  });
});
