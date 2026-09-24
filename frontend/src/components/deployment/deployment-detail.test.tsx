import { render } from '@testing-library/react';
import { DeploymentDetail } from './deployment-detail';
import type { Deployment, DeploymentStatus } from '@/api/api';

const { mockUseFilteredQuery } = vi.hoisted(() => ({
  mockUseFilteredQuery: vi.fn(),
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getDeploymentOptions: (id: string) => ({ queryKey: ['deployment', id] }),
  getDeploymentsQueryOptions: () => ({ queryKey: ['deployments'] }),
  useFilteredQuery: (...args: unknown[]) => mockUseFilteredQuery(...args),
}));

vi.mock('@/lib', () => ({
  cn: (...args: unknown[]) => args.filter(Boolean).join(' '),
  ROUTES: { DEPLOYMENT: (id: string) => `/deployments/${id}` },
}));

vi.mock('@tanstack/react-query', () => ({
  useQueryClient: () => ({
    refetchQueries: vi.fn(),
  }),
}));

vi.mock('react-router-dom', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to} data-slot="link">
      {children}
    </a>
  ),
}));

vi.mock('./deployment-status-badge', () => ({
  DeploymentStatusBadge: ({ status }: { status: string }) => (
    <span data-slot="status-badge" data-status={status} />
  ),
}));

vi.mock('./deployment-diff', () => ({
  DeploymentDiff: ({ fileDiffs }: { fileDiffs: unknown[] }) => (
    <div data-slot="deployment-diff" data-files-count={fileDiffs.length} />
  ),
}));

vi.mock('./deployment-event-log', () => ({
  DeploymentEventLog: ({ events }: { events: unknown[] }) => (
    <div data-slot="deployment-event-log" data-events-count={events.length} />
  ),
}));

vi.mock('../ui/scroll-area', () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="scroll-area">{children}</div>
  ),
}));

vi.mock('../ui/skeleton', () => ({
  Skeleton: () => <div data-slot="skeleton" />,
}));

vi.mock('../ui/spinner', () => ({
  Spinner: () => <div data-slot="spinner" />,
}));

vi.mock('../view', () => ({
  ErrorAlert: ({ error }: { error?: unknown }) => (error ? <div data-slot="error-alert" /> : null),
  HumanTime: ({ time }: { time: string }) => <div data-slot="human-time" data-time={time} />,
}));

const mockDeployment: Deployment = {
  id: 'dep1',
  title: 'My Deployment',
  stack: 'main',
  status: 'done' as DeploymentStatus,
  author: 'John Doe',
  time: '2024-01-01T00:00:00Z',
  repo: 'owner/repo',
  branch: 'feature-branch',
  files: [],
  events: [],
};

describe('DeploymentDetail', () => {
  beforeEach(() => {
    mockUseFilteredQuery.mockClear();
  });

  it('renders skeleton when loading', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: undefined,
      error: null,
      isPending: true,
      isFetching: false,
      refetch: vi.fn(),
    });
    const { container } = render(<DeploymentDetail id="dep1" />);
    expect(container.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(5);
  });

  it('renders deployment title and author when loaded', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: mockDeployment,
      error: null,
      isPending: false,
      isFetching: false,
      refetch: vi.fn(),
    });
    const { container } = render(<DeploymentDetail id="dep1" />);
    expect(container.textContent).toContain('My Deployment');
    expect(container.textContent).toContain('John Doe');
  });

  it('renders deployment link with correct route', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: mockDeployment,
      error: null,
      isPending: false,
      isFetching: false,
      refetch: vi.fn(),
    });
    const { container } = render(<DeploymentDetail id="dep1" />);
    const link = container.querySelector('[data-slot="link"]');
    expect(link).not.toBeNull();
    expect(link?.getAttribute('href')).toBe('/deployments/dep1');
    expect(link?.textContent).toContain('#dep1 -');
  });

  it('renders error alert when there is an error', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: undefined,
      error: new Error('Test error'),
      isPending: false,
      isFetching: false,
      refetch: vi.fn(),
    });
    const { container } = render(<DeploymentDetail id="dep1" />);
    expect(container.querySelector('[data-slot="error-alert"]')).not.toBeNull();
  });

  it('renders spinner when fetching and deployment is running', () => {
    const runningDeployment = { ...mockDeployment, status: 'running' as DeploymentStatus };
    mockUseFilteredQuery.mockReturnValue({
      data: runningDeployment,
      error: null,
      isPending: false,
      isFetching: true,
      refetch: vi.fn(),
    });
    const { container } = render(<DeploymentDetail id="dep1" />);
    expect(container.querySelector('[data-slot="spinner"]')).not.toBeNull();
  });

  it('renders deployment diff and event log', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: mockDeployment,
      error: null,
      isPending: false,
      isFetching: false,
      refetch: vi.fn(),
    });
    const { container } = render(<DeploymentDetail id="dep1" />);
    expect(container.querySelector('[data-slot="deployment-diff"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="deployment-event-log"]')).not.toBeNull();
  });

  it('renders human time', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: mockDeployment,
      error: null,
      isPending: false,
      isFetching: false,
      refetch: vi.fn(),
    });
    const { container } = render(<DeploymentDetail id="dep1" />);
    const humanTime = container.querySelector('[data-slot="human-time"]');
    expect(humanTime).not.toBeNull();
    expect(humanTime?.getAttribute('data-time')).toBe('2024-01-01T00:00:00Z');
  });
});
