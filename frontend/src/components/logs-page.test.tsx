import { render, screen } from '@testing-library/react';
import { LogsPage } from './logs-page';

const { mockUseLogs } = vi.hoisted(() => ({
  mockUseLogs: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

vi.mock('@/hooks', () => ({
  useLogs: mockUseLogs,
}));

vi.mock('./logs/use-filter-logs', () => ({
  useLogFilter: () => ({
    text: '',
    setText: vi.fn(),
    activeLevels: [],
    setActiveLevels: vi.fn(),
    filtered: [],
  }),
}));

vi.mock('./logs', () => ({
  LogEntries: () => <div data-slot="log-entries" />,
  LogFilterBar: () => <div data-slot="log-filter-bar" />,
}));

vi.mock('./view', () => ({
  ErrorAlert: ({ error }: { error?: unknown }) => (error ? <div data-slot="error-alert" /> : null),
  HeaderLayout: ({ header, children }: { header: React.ReactNode; children: React.ReactNode }) => (
    <div data-slot="header-layout">
      <div data-slot="header">{header}</div>
      <div data-slot="content">{children}</div>
    </div>
  ),
  InfoEmpty: ({ title }: { title: string }) => <div data-slot="info-empty">{title}</div>,
}));

vi.mock('./ui/skeleton', () => ({
  Skeleton: () => <div data-slot="skeleton" />,
}));

vi.mock('./ui/scroll-area', () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="scroll-area">{children}</div>
  ),
}));

describe('LogsPage', () => {
  beforeEach(() => {
    mockUseLogs.mockReturnValue({ data: undefined, isPending: false, error: null });
  });

  it('renders the header title', () => {
    mockUseLogs.mockReturnValue({ data: undefined, isPending: false, error: null });
    render(<LogsPage />);
    expect(screen.getByText('translated:LOGS.LOGS')).toBeTruthy();
  });

  it('renders skeletons while loading', () => {
    mockUseLogs.mockReturnValue({ data: undefined, isPending: true, error: null });
    render(<LogsPage />);
    expect(document.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(10);
  });

  it('renders log entries when loaded with data', () => {
    mockUseLogs.mockReturnValue({
      data: { log1: { id: 'log1', time: '2025-01-01', msg: 'test' } },
      isPending: false,
      error: null,
    });
    render(<LogsPage />);
    expect(document.querySelector('[data-slot="log-entries"]')).not.toBeNull();
  });

  it('renders info empty when no logs', () => {
    mockUseLogs.mockReturnValue({ data: {}, isPending: false, error: null });
    render(<LogsPage />);
    expect(document.querySelector('[data-slot="info-empty"]')).not.toBeNull();
  });

  it('renders error alert when there is an error', () => {
    mockUseLogs.mockReturnValue({
      data: undefined,
      isPending: false,
      error: new Error('Failed to load'),
    });
    render(<LogsPage />);
    expect(document.querySelector('[data-slot="error-alert"]')).not.toBeNull();
  });
});
