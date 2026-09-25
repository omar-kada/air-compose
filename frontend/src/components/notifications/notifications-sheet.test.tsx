import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { NotificationSheet } from './notifications-sheet';
import { vi } from 'vitest';

const { mockUseInfiniteQuery, mockUseResetUnreadCount, mockDeployNavigate, mockUseUnreadCount } =
  vi.hoisted(() => ({
    mockUseInfiniteQuery: vi.fn(),
    mockUseResetUnreadCount: vi.fn(),
    mockDeployNavigate: vi.fn(),
    mockUseUnreadCount: vi.fn(),
  }));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getNotificationsQueryOptions: () => ({ queryKey: ['notifications'] }),
  useUnreadNotificationCount: mockUseUnreadCount,
  useResetUnreadCount: () => mockUseResetUnreadCount,
  useFilteredQuery: () => ({ data: undefined, error: null, isPending: false }),
  getFeaturesQueryOptions: () => ({ queryKey: ['features'] }),
  getSettingsQueryOptions: () => ({ queryKey: ['settings'] }),
  useRelativeTime: () => 'mocked-time',
  useUpdateSettings: () => ({ updateSettings: vi.fn() }),
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
  useQueryClient: () => ({
    invalidateQueries: vi.fn(),
    refetchQueries: vi.fn(),
  }),
}));

vi.mock('react-intersection-observer', () => ({
  useInView: () => ({ ref: vi.fn(), inView: false }),
}));

vi.mock('react-router-dom', () => ({
  useNavigate: () => vi.fn(),
  Link: ({ children }: { children: React.ReactNode }) => children,
}));

const mockNotifications = [
  {
    id: '1',
    objectId: 'dep1',
    objectName: 'D1',
    type: 'deployment',
    msg: 'Test notification',
    time: '2024-01-01T00:00:00Z',
  },
];

const openSheet = () => {
  fireEvent.click(screen.getByRole('button', { name: 'Open' }));
};

const openSheetAndWaitForItems = async () => {
  openSheet();
  await waitFor(() => {
    const items = document.body.querySelectorAll('a[data-slot="item"]');
    expect(items.length).toBeGreaterThan(0);
  });
};

const getItems = () => document.body.querySelectorAll('a[data-slot="item"]');

describe('NotificationSheet', () => {
  beforeEach(() => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [],
      isPending: false,
      error: null,
      isFetchingNextPage: false,
      hasNextPage: false,
      fetchNextPage: vi.fn(),
    });
    mockUseUnreadCount.mockReturnValue(0);
    mockUseResetUnreadCount.mockClear();
    mockDeployNavigate.mockClear();
  });

  it('renders children as trigger', () => {
    render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    expect(screen.getByRole('button', { name: 'Open' })).toBeTruthy();
  });

  it('renders sheet title and description when opened', async () => {
    render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    openSheet();
    await waitFor(() => {
      expect(screen.getByText('translated:NOTIFICATIONS.NOTIFICATIONS')).toBeTruthy();
    });
    expect(screen.getByText('translated:NOTIFICATIONS.DESCRIPTION')).toBeTruthy();
  });

  it('renders notification list inside sheet', async () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ ...mockNotifications[0] }],
      isPending: false,
      error: null,
      isFetchingNextPage: false,
      hasNextPage: false,
      fetchNextPage: vi.fn(),
    });
    render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    await openSheetAndWaitForItems();
    const items = getItems();
    expect(items.length).toBeGreaterThan(0);
    const allText = Array.from(items, (i) => i.textContent ?? '').join(' ');
    expect(allText.toLowerCase()).toContain('dep1');
  });

  it('renders close button inside sheet', async () => {
    render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    openSheet();
    await waitFor(() => {
      expect(screen.getByText('translated:ACTION.CLOSE')).toBeTruthy();
    });
  });

  it('resets unread count when sheet is closed', async () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ ...mockNotifications[0] }],
      isPending: false,
      error: null,
      isFetchingNextPage: false,
      hasNextPage: false,
      fetchNextPage: vi.fn(),
    });
    render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    await openSheetAndWaitForItems();
    const closeButton = await screen.findByText('translated:ACTION.CLOSE');
    fireEvent.click(closeButton);
    await waitFor(() => {
      expect(mockUseResetUnreadCount).toHaveBeenCalled();
    });
  });

  it('navigates to deployment when notification item clicked', async () => {
    mockUseInfiniteQuery.mockReturnValue({
      data: [{ ...mockNotifications[0] }],
      isPending: false,
      error: null,
      isFetchingNextPage: false,
      hasNextPage: false,
      fetchNextPage: vi.fn(),
    });
    render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    await openSheetAndWaitForItems();
    const items = getItems();
    const targetItem = Array.from(items).find((i) => i.textContent?.toLowerCase().includes('dep1'));
    if (!targetItem) throw new Error('Notification item not found');
    fireEvent.click(targetItem);
    expect(mockDeployNavigate).toHaveBeenCalledWith('dep1');
  });
});
