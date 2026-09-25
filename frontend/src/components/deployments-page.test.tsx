import { render, screen, fireEvent } from '@testing-library/react';
import { DeploymentsPage } from './deployments-page';

const { mockUseInfiniteQuery, mockUseParams, mockSync, mockDeployNavigate } = vi.hoisted(() => ({
  mockUseInfiniteQuery: vi.fn(),
  mockUseParams: vi.fn(),
  mockSync: vi.fn(),
  mockDeployNavigate: vi.fn(),
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getDeploymentsQueryOptions: () => ({ queryKey: ['deployments'] }),
  useIsMobile: () => false,
  useSync: () => ({ sync: mockSync }),
}));

vi.mock('@/lib', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...(actual as Record<string, unknown>),
    useDeploymentNavigate: () => mockDeployNavigate,
  };
});

vi.mock('@tanstack/react-query', () => ({
  useInfiniteQuery: (...args: unknown[]) => mockUseInfiniteQuery(...args),
}));

vi.mock('react-router-dom', () => ({
  useParams: () => mockUseParams(),
}));

vi.mock('./deployment', () => ({
  DeploymentList: ({ onSelect }: { onSelect?: (item: unknown) => void }) => (
    <div data-slot="deployment-list">
      {onSelect && (
        <button
          data-slot="select-deployment"
          onClick={() => onSelect({ id: 'dep1', title: 'D1', stack: 'main', status: 'done' })}
        >
          Select
        </button>
      )}
    </div>
  ),
  DeploymentDetail: ({ id }: { id: string }) => <div data-slot="deployment-detail" data-id={id} />,
  DeploymentDetailSkeleton: () => <div data-slot="deployment-detail-skeleton" />,
  DeploymentToolbar: () => <div data-slot="deployment-toolbar" />,
}));

describe('DeploymentsPage', () => {
  beforeEach(() => {
    mockUseParams.mockReturnValue({});
    mockUseInfiniteQuery.mockReturnValue({ data: undefined, isPending: true, error: null });
    mockSync.mockClear();
    mockDeployNavigate.mockClear();
  });

  it('renders skeleton while loading', () => {
    const { container } = render(<DeploymentsPage />);
    expect(container.querySelector('[data-slot="deployment-detail-skeleton"]')).not.toBeNull();
  });

  it('renders InfoEmpty with sync button when no deployments', () => {
    mockUseInfiniteQuery.mockReturnValue({ data: [], isPending: false, error: null });
    const { container } = render(<DeploymentsPage />);
    expect(container.querySelector('[data-slot="empty"]')).not.toBeNull();
    expect(container.querySelectorAll('[data-slot="button"]')).toHaveLength(1);
    expect(screen.getByText('translated:ACTION.SYNC_NOW')).toBeTruthy();
  });

  it('renders DeploymentList and DeploymentDetail when loaded', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ id: 'dep1', title: 'D1', stack: 'main', status: 'done' }],
      isPending: false,
      error: null,
    });
    const { container } = render(<DeploymentsPage />);
    expect(container.querySelector('[data-slot="deployment-list"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="deployment-detail"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="deployment-toolbar"]')).not.toBeNull();
  });

  it('renders back button with ArrowLeft icon', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ id: 'dep1', title: 'D1', stack: 'main', status: 'done' }],
      isPending: false,
      error: null,
    });
    render(<DeploymentsPage />);
    expect(screen.getByText('translated:ACTION.BACK')).toBeTruthy();
  });

  it('navigates back when back button is clicked', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ id: 'dep1', title: 'D1', stack: 'main', status: 'done' }],
      isPending: false,
      error: null,
    });
    render(<DeploymentsPage />);
    fireEvent.click(screen.getByText('translated:ACTION.BACK').closest('button') as HTMLElement);
    expect(mockDeployNavigate).toHaveBeenCalledWith();
  });

  it('navigates to deployment when selected from list', () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ id: 'dep1', title: 'D1', stack: 'main', status: 'done' }],
      isPending: false,
      error: null,
    });
    render(<DeploymentsPage />);
    fireEvent.click(screen.getByText('Select').closest('button') as HTMLElement);
    expect(mockDeployNavigate).toHaveBeenCalledWith('dep1');
  });
});
