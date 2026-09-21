import { render } from '@testing-library/react';
import type { DeploymentStatus } from '@/api/api';
import { DeploymentStatusBadge } from './deployment-status-badge';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

describe('DeploymentStatusBadge', () => {
  it('renders the translated status text as fallback label', () => {
    const status = 'success' as const;
    const { container } = render(<DeploymentStatusBadge status={status} />);
    const badge = container.querySelector('[data-slot="badge"]');
    expect(badge?.textContent).toContain(`translated:DEPLOYMENT_STATUS.${status}`);
  });

  it('does not render a text label when iconOnly is true', () => {
    const { container } = render(<DeploymentStatusBadge status="success" iconOnly />);
    expect(container.textContent?.trim()).toBe('');
  });

  it('renders an SVG icon', () => {
    const { container } = render(<DeploymentStatusBadge status="success" iconOnly />);
    expect(container.querySelector('svg')).not.toBeNull();
  });

  it('renders a custom label when provided', () => {
    const label = 'Deployed';
    const { container } = render(<DeploymentStatusBadge status="success" label={label} />);
    const badge = container.querySelector('[data-slot="badge"]');
    expect(badge?.textContent).toContain(label);
  });

  it.each<[DeploymentStatus, string]>([
    ['success', 'bg-green-400'],
    ['error', 'bg-red-400'],
    ['running', 'bg-blue-400'],
    ['planned', 'bg-slate-400'],
  ])('applies %s color class %s', (status, colorClass) => {
    const { container } = render(<DeploymentStatusBadge status={status} iconOnly />);
    expect(container.querySelector(`[class*="${colorClass}"]`)).not.toBeNull();
  });

  it('applies animate-spin to the icon when status is running', () => {
    const { container } = render(<DeploymentStatusBadge status="running" iconOnly />);
    const icon = container.querySelector('svg');
    expect(icon?.getAttribute('class')).toContain('animate-spin');
  });

  it('does not apply animate-spin when status is not running', () => {
    const { container } = render(<DeploymentStatusBadge status="success" iconOnly />);
    const icon = container.querySelector('svg');
    expect(icon?.getAttribute('class')).not.toContain('animate-spin');
  });

  it('applies custom className alongside the color class', () => {
    const { container } = render(
      <DeploymentStatusBadge status="success" iconOnly className="custom-class" />,
    );
    const badge = container.firstChild as HTMLElement;
    expect(badge.className).toContain('custom-class');
    expect(badge.className).toContain('bg-green-400');
  });

  it('falls back to "unknown" when status is undefined', () => {
    const { container } = render(
      <DeploymentStatusBadge status={undefined as unknown as 'success'} label={undefined} />,
    );
    const badge = container.querySelector('[data-slot="badge"]');
    expect(badge?.textContent).toContain('unknown');
  });
});
