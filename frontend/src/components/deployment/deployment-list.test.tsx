import { render } from '@testing-library/react';
import { DeploymentList } from './deployment-list';

const mockFetchNextPage = vi.fn();
const mockOnSelect = vi.fn();

const { mockUseInfiniteQuery } = vi.hoisted(() => ({
  mockUseInfiniteQuery: vi.fn(),
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getDeploymentsQueryOptions: vi.fn(() => ({ queryKey: 'deployments' })),
}));

vi.mock('@tanstack/react-query', () => ({
  useInfiniteQuery: (...args: unknown[]) => mockUseInfiniteQuery(...args),
}));

vi.mock('react-intersection-observer', () => ({
  useInView: () => ({ ref: vi.fn(), inView: false }),
}));

vi.mock('./deployment-list-item', () => ({
  DeploymentListItem: ({ deployment }: { deployment: { id: string } }) => (
    <div data-slot="deployment-item" data-id={deployment.id} />
  ),
  DeploymentItemSkeleton: () => <div data-slot="skeleton" />,
}));

vi.mock('@/components/ui/scroll-area', () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="scroll-area">{children}</div>
  ),
}));

vi.mock('@/components/ui/spinner', () => ({
  Spinner: () => <div data-slot="spinner" />,
}));

vi.mock('@/components/view', () => ({
  ErrorAlert: ({ title }: { title: string }) => <div data-slot="error-alert" data-title={title} />,
}));

describe('DeploymentList', () => {
  beforeEach(() => {
    mockUseInfiniteQuery.mockReturnValue({
      data: undefined,
      fetchNextPage: mockFetchNextPage,
      hasNextPage: false,
      isFetchingNextPage: false,
      isPending: false,
      error: null,
    });
    mockFetchNextPage.mockClear();
    mockOnSelect.mockClear();
  });

  it('renders skeleton when loading', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: undefined,
      fetchNextPage: mockFetchNextPage,
      hasNextPage: false,
      isFetchingNextPage: false,
      isPending: true,
      error: null,
    });
    const { container } = render(<DeploymentList onSelect={mockOnSelect} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons).toHaveLength(5);
  });

  it('renders error alert when there is an error', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: undefined,
      fetchNextPage: mockFetchNextPage,
      hasNextPage: false,
      isFetchingNextPage: false,
      isPending: false,
      error: new Error('Test error'),
    });
    const { container } = render(<DeploymentList onSelect={mockOnSelect} />);
    expect(container.querySelector('[data-slot="error-alert"]')).not.toBeNull();
  });

  it('renders deployment items when loaded', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [
        { id: 'dep1', title: 'Deployment 1', stack: 'main', status: 'done' },
        { id: 'dep2', title: 'Deployment 2', stack: 'main', status: 'done' },
      ],
      fetchNextPage: mockFetchNextPage,
      hasNextPage: false,
      isFetchingNextPage: false,
      isPending: false,
      error: null,
    });
    const { container } = render(<DeploymentList onSelect={mockOnSelect} />);
    const items = container.querySelectorAll('[data-slot="deployment-item"]');
    expect(items).toHaveLength(2);
    expect(items[0]?.getAttribute('data-id')).toBe('dep1');
    expect(items[1]?.getAttribute('data-id')).toBe('dep2');
  });

  it('renders empty list when no deployments', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [],
      fetchNextPage: mockFetchNextPage,
      hasNextPage: false,
      isFetchingNextPage: false,
      isPending: false,
      error: null,
    });
    const { container } = render(<DeploymentList onSelect={mockOnSelect} />);
    expect(container.querySelectorAll('[data-slot="deployment-item"]')).toHaveLength(0);
  });

  it('renders spinner when fetching next page', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ id: 'dep1', title: 'D1', stack: 'main', status: 'done' }],
      fetchNextPage: mockFetchNextPage,
      hasNextPage: true,
      isFetchingNextPage: true,
      isPending: false,
      error: null,
    });
    const { container } = render(<DeploymentList onSelect={mockOnSelect} />);
    expect(container.querySelector('[data-slot="spinner"]')).not.toBeNull();
  });
});
