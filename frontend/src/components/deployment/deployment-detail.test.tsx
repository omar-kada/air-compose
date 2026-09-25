import { render, screen } from '@testing-library/react';
import { DeploymentDetail } from './deployment-detail';
import type { DeploymentWithDetails, DeploymentStatus } from '@/api/api';

const { mockUseFilteredQuery, mockRefetch, mockRefetchQueries } = vi.hoisted(() => ({
  mockUseFilteredQuery: vi.fn(),
  mockRefetch: vi.fn(),
  mockRefetchQueries: vi.fn(),
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getDeploymentOptions: (id: string) => ({ queryKey: ['deployment', id] }),
  getDeploymentsQueryOptions: () => ({ queryKey: ['deployments'] }),
  useFilteredQuery: (...args: unknown[]) => mockUseFilteredQuery(...args),
  useRelativeTime: (time: string) => time,
}));

vi.mock('@tanstack/react-query', () => ({
  useQueryClient: () => ({
    refetchQueries: mockRefetchQueries,
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

const mockDeployment: DeploymentWithDetails = {
  id: 'dep1',
  title: 'My Deployment',
  author: 'John Doe',
  time: '2024-01-01T00:00:00Z',
  endTime: '',
  diff: '',
  status: 'done' as DeploymentStatus,
  files: [],
  events: [],
  commit: 'abc123',
  repo: 'owner/repo',
  branch: 'main',
};

describe('DeploymentDetail', () => {
  beforeEach(() => {
    mockUseFilteredQuery.mockClear();
    mockRefetch.mockClear();
    mockRefetchQueries.mockClear();
  });

  afterEach(() => {
    vi.useRealTimers();
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
    expect(container.querySelector('[data-slot="alert"]')).not.toBeNull();
  });

  it('renders spinner when fetching and deployment is running', () => {
    vi.useFakeTimers();
    const runningDeployment = { ...mockDeployment, status: 'running' as DeploymentStatus };
    mockUseFilteredQuery.mockReturnValue({
      data: runningDeployment,
      error: null,
      isPending: false,
      isFetching: true,
      refetch: mockRefetch,
    });
    render(<DeploymentDetail id="dep1" />);
    expect(screen.getByRole('status')).toBeTruthy();
  });

  it('refetches when deployment status is running', () => {
    vi.useFakeTimers();
    const runningDeployment = { ...mockDeployment, status: 'running' as DeploymentStatus };
    mockUseFilteredQuery.mockReturnValue({
      data: runningDeployment,
      error: null,
      isPending: false,
      isFetching: false,
      refetch: mockRefetch,
    });
    render(<DeploymentDetail id="dep1" />);
    vi.advanceTimersByTime(1000);
    expect(mockRefetch).toHaveBeenCalled();
    expect(mockRefetchQueries).toHaveBeenCalled();
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
    render(<DeploymentDetail id="dep1" />);
    expect(screen.getByText('2024-01-01T00:00:00Z')).toBeTruthy();
  });
});
