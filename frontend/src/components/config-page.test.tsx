import { render } from '@testing-library/react';
import { ConfigPage } from './config-page';

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

vi.mock('@/lib', () => ({
  cn: (...args: unknown[]) => args.filter(Boolean).join(' '),
}));

vi.mock('@hookform/resolvers/zod', () => ({
  zodResolver: () => () => ({ values: undefined, errors: {} }),
}));

vi.mock('./config', () => ({
  ConfigForm: () => <div data-slot="config-form" />,
  ConfigViewer: () => <div data-slot="config-viewer" />,
  formSchema: undefined,
  fromConfig: () => ({ globalEnvVars: [], services: [] }),
  toConfig: () => ({}),
  toYaml: () => '',
}));

vi.mock('./ui/alert', () => ({
  Alert: ({ children }: { children: React.ReactNode }) => <div data-slot="alert">{children}</div>,
  AlertDescription: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="alert-description">{children}</div>
  ),
  AlertTitle: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="alert-title">{children}</div>
  ),
}));

vi.mock('./ui/button', () => ({
  Button: ({
    onClick,
    children,
    ...props
  }: {
    onClick?: () => void;
    children: React.ReactNode;
    [key: string]: unknown;
  }) => (
    <button onClick={onClick} data-slot="button" {...props}>
      {children}
    </button>
  ),
}));

vi.mock('./ui/skeleton', () => ({
  Skeleton: () => <div data-slot="skeleton" />,
}));

vi.mock('./ui/spinner', () => ({
  Spinner: () => <div data-slot="spinner" />,
}));

vi.mock('./ui/toggle', () => ({
  Toggle: ({ pressed, children }: { pressed?: boolean; children: React.ReactNode }) => (
    <button data-slot="toggle" data-pressed={pressed}>
      {children}
    </button>
  ),
}));

vi.mock('./view', () => ({
  ErrorAlert: ({ error }: { error?: unknown }) => (error ? <div data-slot="error-alert" /> : null),
  HeaderLayout: ({ children, header }: { children: React.ReactNode; header: React.ReactNode }) => (
    <div data-slot="header-layout">
      {header}
      {children}
    </div>
  ),
  InfoEmpty: ({ title, details }: { title: string; details: string }) => (
    <div data-slot="info-empty" data-title={title} data-details={details} />
  ),
}));

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
    expect(container.querySelector('[data-slot="info-empty"]')).not.toBeNull();
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
    const { container } = render(<ConfigPage />);
    expect(container.querySelector('[data-slot="button"]')).not.toBeNull();
  });

  it('shows error alert when config query has error', () => {
    mockUseFilteredQuery.mockReturnValue({
      data: { displayConfig: true, editConfig: true, services: [] },
      isPending: false,
      error: new Error('Failed to load'),
    });
    const { container } = render(<ConfigPage />);
    expect(container.querySelector('[data-slot="error-alert"]')).not.toBeNull();
  });
});
