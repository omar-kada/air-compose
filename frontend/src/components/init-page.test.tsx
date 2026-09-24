import { render, screen, waitFor } from '@testing-library/react';
import { InitPage } from './init-page';

const { mockUseQuery, mockUpdateSettings, mockNavigate } = vi.hoisted(() => ({
  mockUseQuery: vi.fn(),
  mockUpdateSettings: vi.fn(),
  mockNavigate: vi.fn(),
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getSettingsQueryOptions: () => ({ queryKey: ['settings'] }),
  getStateQueryOptions: () => ({ queryKey: ['state'] }),
  useUpdateSettings: () => ({ updateSettings: mockUpdateSettings }),
}));


vi.mock('@tanstack/react-query', () => ({
  useQuery: (...args: unknown[]) => mockUseQuery(...args),
}));

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock('./onboarding', () => ({
  OnboardingForm: () => <div data-slot="onboarding-form" />,
}));


describe('InitPage', () => {
  beforeEach(() => {
    mockUpdateSettings.mockClear();
    mockNavigate.mockClear();
    mockUseQuery.mockReset();
    let callCount = 0;
    mockUseQuery.mockImplementation(() => {
      callCount++;
      if (callCount === 1) {
        return { data: undefined, isPending: false, error: null };
      }
      return { data: { initialized: false }, isPending: false, error: null };
    });
  });

  it('renders skeleton while settings are loading', () => {
    mockUseQuery.mockReset();
    let callCount = 0;
    mockUseQuery.mockImplementation(() => {
      callCount++;
      if (callCount === 1) {
        return { data: undefined, isPending: true, error: null };
      }
      return { data: { initialized: false }, isPending: false, error: null };
    });
    const { container } = render(<InitPage />);
    expect(container.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(5);
  });

  it('renders onboarding form when settings are loaded', () => {
    mockUseQuery.mockReset();
    let callCount = 0;
    mockUseQuery.mockImplementation(() => {
      callCount++;
      if (callCount === 1) {
        return {
          data: { repo: 'owner/repo', branch: 'main' },
          isPending: false,
          error: null,
        };
      }
      return { data: { initialized: false }, isPending: false, error: null };
    });
    render(<InitPage />);
    expect(document.querySelector('[data-slot="onboarding-form"]')).not.toBeNull();
    expect(screen.getByText('translated:ONBOARDING.FORM.TITLE')).toBeTruthy();
  });

  it('renders settings error alert when settings query fails', () => {
    mockUseQuery.mockReset();
    let callCount = 0;
    mockUseQuery.mockImplementation(() => {
      callCount++;
      if (callCount === 1) {
        return { data: undefined, isPending: false, error: new Error('Settings error') };
      }
      return { data: { initialized: false }, isPending: false, error: null };
    });
    render(<InitPage />);
    expect(document.querySelector('[data-slot="alert"]')).not.toBeNull();
  });

  it('renders state error alert when state query fails', () => {
    mockUseQuery.mockReset();
    let callCount = 0;
    mockUseQuery.mockImplementation(() => {
      callCount++;
      if (callCount === 1) {
        return { data: { repo: 'test' }, isPending: false, error: null };
      }
      return { data: undefined, isPending: false, error: new Error('State error') };
    });
    render(<InitPage />);
    expect(document.querySelector('[data-slot="alert"]')).not.toBeNull();
  });

  it('navigates to root when state is initialized', async () => {
    mockUseQuery.mockReset();
    mockUseQuery.mockReturnValue({
      data: { initialized: true, services: [], cron: '', retriesOnUnhealthy: 3, retryDelay: 5 },
      isPending: false,
      error: null,
    });
    render(<InitPage />);
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });
  });
});
