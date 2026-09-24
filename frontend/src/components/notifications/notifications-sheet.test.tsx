import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { NotificationSheet } from './notifications-sheet';

const { mockUseInfiniteQuery, mockUseResetUnreadCount, mockDeployNavigate } = vi.hoisted(() => ({
  mockUseInfiniteQuery: vi.fn(),
  mockUseResetUnreadCount: vi.fn(),
  mockDeployNavigate: vi.fn(),
}));

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  getNotificationsQueryOptions: () => ({ queryKey: ['notifications'] }),
  useResetUnreadCount: () => mockUseResetUnreadCount,
}));

vi.mock('@/lib', () => ({
  useDeploymentNavigate: () => mockDeployNavigate,
}));

vi.mock('@tanstack/react-query', () => ({
  useInfiniteQuery: (...args: unknown[]) => mockUseInfiniteQuery(...args),
}));

vi.mock('@/components/ui/scroll-area', () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="scroll-area">{children}</div>
  ),
}));

vi.mock('./notifications-list', () => ({
  NotificationList: ({ onNotificationClick }: { onNotificationClick: (e: unknown) => void }) => (
    <div
      data-slot="notification-list"
      data-has-click-handler={!!onNotificationClick}
      onClick={() => onNotificationClick?.({ objectId: 'dep1' })}
    />
  ),
}));

describe('NotificationSheet', () => {
  beforeEach(() => {
    mockUseInfiniteQuery.mockReturnValue({ data: [] });
    mockUseResetUnreadCount.mockClear();
    mockDeployNavigate.mockClear();
  });

  it('renders children as trigger', () => {
    const { container } = render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    expect(container.querySelector('[data-slot="trigger"]')).not.toBeNull();
  });

  it('renders sheet title and description when opened', async () => {
    const { container } = render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    fireEvent.click(container.querySelector('[data-slot="trigger"]') as HTMLElement);
    await waitFor(() => {
      expect(screen.getByText('translated:NOTIFICATIONS.NOTIFICATIONS')).toBeTruthy();
    });
    expect(screen.getByText('translated:NOTIFICATIONS.DESCRIPTION')).toBeTruthy();
  });

  it('renders notification list inside sheet', async () => {
    const { container } = render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    fireEvent.click(container.querySelector('[data-slot="trigger"]') as HTMLElement);
    await waitFor(() => {
      expect(document.querySelector('[data-slot="notification-list"]')).not.toBeNull();
    });
  });

  it('renders close button inside sheet', async () => {
    const { container } = render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    fireEvent.click(container.querySelector('[data-slot="trigger"]') as HTMLElement);
    await waitFor(() => {
      expect(screen.getByText('translated:ACTION.CLOSE')).toBeTruthy();
    });
  });

  it('resets unread count when sheet is closed with notifications', async () => {
    mockUseInfiniteQuery.mockReturnValue({ data: [{ id: '1', objectId: 'dep1' }] });
    const { container } = render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    fireEvent.click(container.querySelector('[data-slot="trigger"]') as HTMLElement);
    await waitFor(() => {
      const closeButton = screen.getByText('translated:ACTION.CLOSE');
      fireEvent.click(closeButton);
    });
    await waitFor(() => {
      expect(mockUseResetUnreadCount).toHaveBeenCalled();
    });
  });

  it('navigates to deployment when notification is clicked', async () => {
    const { container } = render(
      <NotificationSheet>
        <button data-slot="trigger">Open</button>
      </NotificationSheet>,
    );
    fireEvent.click(container.querySelector('[data-slot="trigger"]') as HTMLElement);
    await waitFor(() => {
      expect(document.querySelector('[data-slot="notification-list"]')).not.toBeNull();
    });
    fireEvent.click(document.querySelector('[data-slot="notification-list"]') as HTMLElement);
    expect(mockDeployNavigate).toHaveBeenCalledWith('dep1');
  });
});
