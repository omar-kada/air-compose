import { render, screen } from '@testing-library/react';
import { ConfigPage } from './config-page';
import React from 'react';

const { mockUseFilteredQuery, mockUpdateConfig } = vi.hoisted(() => ({
  mockUseFilteredQuery: vi.fn(),
  mockUpdateConfig: vi.fn(),
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getConfigQueryOptions: () => ({ queryKey: ['config'] }),
  getFeaturesQueryOptions: () => ({ queryKey: ['features'] }),
  useFilteredQuery: (...args: unknown[]) => mockUseFilteredQuery(...args),
  useIsMobile: () => false,
  useUpdateConfig: () => ({ updateConfig: mockUpdateConfig, isPending: false }),
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

// ConfigForm and ConfigViewer are stubbed because they require FormProvider and
// ThemeProvider context wrappers that ConfigPage does not provide. Everything else
// from ./config (formSchema, fromConfig, toConfig, toYaml) uses real implementations.
vi.mock('./config', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...(actual as Record<string, unknown>),
    ConfigForm: () => <div data-slot="config-form" />,
    ConfigViewer: () => <div data-slot="config-viewer" />,
  };
});

describe('ConfigPage', () => {
  beforeEach(() => {
    mockUseFilteredQuery.mockReset();
    mockUpdateConfig.mockClear();
  });

  it('renders skeletons while features are loading', () => {
    mockUseFilteredQuery.mockReturnValue({ data: undefined, isPending: true, error: null });
    const { container } = render(<ConfigPage />);
    expect(container.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(8);
  });

  it('renders info empty when displayConfig is false', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: { displayConfig: false },
      isPending: false,
      error: null,
    });
    const { container } = render(<ConfigPage />);
    expect(container.querySelector('[data-slot="empty"]')).not.toBeNull();
  });

  it('renders config form and viewer when config is loaded', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: { displayConfig: true, editConfig: true, services: [] },
      isPending: false,
      error: null,
    });
    const { container } = render(<ConfigPage />);
    expect(container.querySelector('[data-slot="config-form"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="config-viewer"]')).not.toBeNull();
  });

  it('renders save button when config is editable', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: { displayConfig: true, editConfig: true, services: [] },
      isPending: false,
      error: null,
    });
    render(<ConfigPage />);
    expect(screen.getByRole('button', { name: 'translated:ACTION.SAVE' })).toBeTruthy();
  });

  it('shows error alert when config query has error', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: { displayConfig: true, editConfig: true, services: [] },
      isPending: false,
      error: new Error('Failed to load'),
    });
    const { container } = render(<ConfigPage />);
    expect(container.querySelector('[data-slot="alert"]')).not.toBeNull();
  });
});
