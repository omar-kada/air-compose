import { render, screen } from '@testing-library/react';
import { StatusPage } from './status-page';

const { mockUseFilteredQuery, mockGetStatusQueryOptions } = vi.hoisted(() => ({
  mockUseFilteredQuery: vi.fn(),
  mockGetStatusQueryOptions: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
    i18n: { language: 'en' },
  }),
}));

vi.mock('@/hooks', () => ({
  getStatusQueryOptions: mockGetStatusQueryOptions,
  useFilteredQuery: mockUseFilteredQuery,
}));



vi.mock('./status', () => ({
  ServiceStatus: () => <div data-slot="service-status" />,
  ServiceStatusSkeleton: () => <div data-slot="service-status-skeleton" />,
}));



describe('StatusPage', () => {
  beforeEach(() => {
    mockGetStatusQueryOptions.mockReturnValue({ queryKey: ['status'] });
  });

  it('renders the header title', () => {
    mockUseFilteredQuery.mockReturnValue({ data: {}, isPending: false, error: null });
    render(<StatusPage />);
    expect(screen.getByText('t:STATUS.STATUS')).toBeTruthy();
  });

  it('renders skeletons while loading', () => {
    mockUseFilteredQuery.mockReturnValue({ data: undefined, isPending: true, error: null });
    render(<StatusPage />);
    expect(document.querySelectorAll('[data-slot="service-status-skeleton"]')).toHaveLength(3);
  });

  it('renders service status items when data is loaded', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: { web: { id: '1', name: 'web', containers: [] } },
      isPending: false,
      error: null,
    });
    render(<StatusPage />);
    expect(document.querySelectorAll('[data-slot="service-status"]')).toHaveLength(1);
  });

  it('renders info empty when no data', () => {
    mockUseFilteredQuery.mockReturnValue({ data: {}, isPending: false, error: null });
    render(<StatusPage />);
    expect(document.querySelector('[data-slot="empty"]')).not.toBeNull();
  });

  it('renders error alert when there is an error', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: undefined,
      isPending: false,
      error: new Error('Failed to load'),
    });
    render(<StatusPage />);
    expect(document.querySelector('[data-slot="alert"]')).not.toBeNull();
  });
});
