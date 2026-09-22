import { render, screen } from '@testing-library/react';
import type { ContainerStatus } from '@/api/api';
import { ServiceStatus, ServiceStatusSkeleton } from './service-status';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('../view', () => ({
  HumanTime: ({ time }: { time?: Date | string }) => (
    <span data-slot="human-time">{time ? String(time) : ''}</span>
  ),
}));

describe('ServiceStatus', () => {
  const mockContainers: Record<string, ContainerStatus> = {
    'container-1': {
      containerId: 'c1',
      state: 'running',
      name: 'web',
      health: 'healthy',
      startedAt: '2024-01-01T10:00:00Z',
    },
    'container-2': {
      containerId: 'c2',
      state: 'exited',
      name: 'worker',
      health: 'unhealthy',
      startedAt: '2024-01-01T11:00:00Z',
    },
  };

  it('renders service name', () => {
    render(<ServiceStatus serviceName="nginx" serviceContainers={mockContainers} />);
    expect(screen.getByText('nginx')).toBeTruthy();
  });

  it('renders ContainerStatusBadge for each container', () => {
    render(<ServiceStatus serviceName="nginx" serviceContainers={mockContainers} />);
    expect(screen.getByText('translated:web')).toBeTruthy();
    expect(screen.getByText('translated:worker')).toBeTruthy();
  });

  it('renders HumanTime with the latest container start time', () => {
    render(<ServiceStatus serviceName="nginx" serviceContainers={mockContainers} />);
    const humanTime = document.querySelector('[data-slot="human-time"]');
    expect(humanTime).not.toBeNull();
    expect(humanTime?.textContent).toBe(String(new Date('2024-01-01T11:00:00Z')));
  });

  it('applies custom className to the Item', () => {
    const { container } = render(
      <ServiceStatus
        serviceName="nginx"
        serviceContainers={mockContainers}
        className="custom-class"
      />,
    );
    expect(container.querySelector('[data-slot="item"]')?.classList.contains('custom-class')).toBe(
      true,
    );
  });

  it('renders no ContainerStatusBadges when serviceContainers is empty', () => {
    const { container } = render(<ServiceStatus serviceName="nginx" serviceContainers={{}} />);
    expect(container.querySelectorAll('[data-slot="badge"]').length).toBe(0);
  });
});

describe('ServiceStatusSkeleton', () => {
  it('renders skeleton layout with multiple skeletons', () => {
    render(<ServiceStatusSkeleton />);
    expect(document.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0);
  });
});
