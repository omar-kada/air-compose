import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { EnvironementHealth } from './environement-health';

const mockUseUser = vi.hoisted(() => vi.fn());
const mockUseFilteredQuery = vi.hoisted(() => vi.fn());
const mockGetStateQueryOptions = vi.hoisted(() => vi.fn());

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  useUser: mockUseUser,
  useFilteredQuery: mockUseFilteredQuery,
  getStateQueryOptions: mockGetStateQueryOptions,
}));

describe('EnvironementHealth', () => {
  beforeEach(() => {
    mockGetStateQueryOptions.mockReturnValue({ queryKey: ['state'], queryFn: vi.fn() });
  });

  it('renders Skeleton when loading', () => {
    mockUseUser.mockReturnValue({ data: { id: 'user-1' } });
    mockUseFilteredQuery.mockReturnValue({ isPending: true, data: undefined });

    render(
      <MemoryRouter>
        <EnvironementHealth />
      </MemoryRouter>,
    );

    expect(document.querySelector('[data-slot="skeleton"]')).not.toBeNull();
    expect(screen.queryByText('translated:healthy')).toBeNull();
  });

  it('renders ContainerStatusBadge with health when data is loaded', () => {
    mockUseUser.mockReturnValue({ data: { id: 'user-1' } });
    mockUseFilteredQuery.mockReturnValue({
      data: { health: 'healthy', state: 'running' },
      isPending: false,
    });

    render(
      <MemoryRouter>
        <EnvironementHealth />
      </MemoryRouter>,
    );

    expect(screen.getByText('translated:healthy')).toBeTruthy();
    expect(screen.getByRole('link').getAttribute('href')).toBeTruthy();
  });

  it('renders unknown badge when health state is absent (disabled query)', () => {
    // When user is null, getStateQueryOptions gets enabled:false, so the query is
    // disabled and useFilteredQuery returns { isPending: false, data: undefined }.
    // The component gates only on isPending; state?.health is undefined → unknown badge.
    mockUseUser.mockReturnValue({ data: null });
    mockUseFilteredQuery.mockReturnValue({ isPending: false, data: undefined });

    render(
      <MemoryRouter>
        <EnvironementHealth />
      </MemoryRouter>,
    );

    expect(screen.getByText('translated:unknown')).toBeTruthy();
    expect(document.querySelector('[data-slot="skeleton"]')).toBeNull();
  });

  it('applies custom className to the Link', () => {
    mockUseUser.mockReturnValue({ data: { id: 'user-1' } });
    mockUseFilteredQuery.mockReturnValue({
      data: { health: 'healthy', state: 'running' },
      isPending: false,
    });

    render(
      <MemoryRouter>
        <EnvironementHealth className="custom-class" />
      </MemoryRouter>,
    );

    const link = screen.getByRole('link');
    expect(link?.classList.contains('custom-class')).toBe(true);
  });
});
