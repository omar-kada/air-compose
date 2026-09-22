import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import type { Settings } from '@/api/api';
import { SettingsSheet } from './settings-sheet';

const mockSettings: Settings = {
  repo: 'owner/repo',
  branch: 'main',
  cron: '0 */10 * * *',
  notificationURL: 'https://example.com',
  notificationTypes: [],
  retriesOnUnhealthy: 3,
  retryDelay: 60000,
};

const { mockUseFilteredQuery, mockUpdateSettings, queryState } = vi.hoisted(() => ({
  mockUseFilteredQuery: vi.fn(),
  mockUpdateSettings: vi.fn(),
  queryState: {
    callCount: 0,
    featuresEditSettings: true,
    isPending: false,
    error: null as Error | null,
  },
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getFeaturesQueryOptions: () => ({ queryKey: ['features'] }),
  getSettingsQueryOptions: () => ({ queryKey: ['settings'] }),
  useFilteredQuery: mockUseFilteredQuery,
  useUpdateSettings: () => ({
    updateSettings: mockUpdateSettings,
    isPending: false,
  }),
}));

vi.mock('../view', () => ({
  ErrorAlert: ({ error }: { error?: unknown }) =>
    error ? <div data-slot="error-alert">Error</div> : null,
}));

vi.mock('./settings-form', () => ({
  SettingsForm: () => <div data-slot="settings-form" />,
}));

describe('SettingsSheet', () => {
  beforeEach(() => {
    queryState.callCount = 0;
    queryState.featuresEditSettings = true;
    queryState.isPending = false;
    queryState.error = null;
    mockUseFilteredQuery.mockImplementation(() => {
      queryState.callCount++;
      if (queryState.callCount === 1) {
        return {
          data: { editSettings: queryState.featuresEditSettings },
          error: queryState.error,
          isPending: queryState.isPending,
        };
      }
      return {
        data: mockSettings,
        error: queryState.error,
        isPending: queryState.isPending,
      };
    });
  });

  it('renders settings form when settings are loaded', () => {
    render(<SettingsSheet open setOpen={vi.fn()} />);
    expect(document.querySelector('[data-slot="settings-form"]')).not.toBeNull();
  });

  it('renders skeleton while loading', () => {
    mockUseFilteredQuery.mockImplementation(() => ({
      data: undefined,
      error: null,
      isPending: true,
    }));
    render(<SettingsSheet open setOpen={vi.fn()} />);
    expect(document.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(3);
  });

  it('renders error alert when there is an error', () => {
    const testError = new Error('Failed to load settings');
    mockUseFilteredQuery.mockImplementation(() => ({
      data: undefined,
      error: testError,
      isPending: false,
    }));
    render(<SettingsSheet open setOpen={vi.fn()} />);
    expect(document.querySelector('[data-slot="error-alert"]')).not.toBeNull();
  });

  it('calls updateSettings on form submit', async () => {
    render(<SettingsSheet open setOpen={vi.fn()} />);
    await screen.findByRole('button', { name: 'translated:ACTION.SAVE' });
    fireEvent.click(screen.getByRole('button', { name: 'translated:ACTION.SAVE' }));
    await waitFor(() => {
      expect(mockUpdateSettings).toHaveBeenCalledWith(
        expect.objectContaining({ repo: 'owner/repo', retryDelay: 60000 }),
      );
    });
  });

  it('renders disabled message when editSettings is false', () => {
    queryState.featuresEditSettings = false;
    render(<SettingsSheet open setOpen={vi.fn()} />);
    expect(screen.getByText(/translated:SETTINGS.DISABLED/)).toBeTruthy();
  });
});
