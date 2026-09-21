import { render } from '@testing-library/react';
import { EventType, type Event } from '@/api/api';
import { NotificationList, NotificationSkeleton } from './notifications-list';
import { useInfiniteQuery, type UseInfiniteQueryResult } from '@tanstack/react-query';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

vi.mock('@/hooks', () => ({
  getNotificationsQueryOptions: () => ({ queryKey: ['notifications'] }),
  useUnreadNotificationCount: () => 0,
  useRelativeTime: (time: string) => time,
}));

vi.mock('@tanstack/react-query', () => ({
  useInfiniteQuery: vi.fn(),
}));

vi.mock('react-intersection-observer', () => ({
  useInView: () => ({ ref: vi.fn(), inView: false }),
}));

const mockEvent = (over: Partial<Event> = {}): Event => ({
  ID: 1,
  time: '2025-01-01T00:00:00Z',
  msg: '',
  type: EventType.DEPLOYMENT_STARTED,
  objectId: 123,
  objectName: 'deploy-1',
  ...over,
});

describe('NotificationSkeleton', () => {
  it('renders three skeleton placeholders', () => {
    const { container } = render(<NotificationSkeleton />);
    expect(container.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(3);
  });
});

describe('NotificationList', () => {
  it('renders skeleton when pending', () => {
    vi.mocked(useInfiniteQuery).mockReturnValue({
      isPending: true,
      data: undefined,
      error: null,
      hasNextPage: false,
      isFetchingNextPage: false,
      fetchNextPage: vi.fn(),
    } as unknown as UseInfiniteQueryResult);
    const { container } = render(<NotificationList onNotificationClick={() => {}} />);
    expect(container.querySelectorAll('[data-slot="skeleton"]')).toHaveLength(3);
  });

  it('renders notifications when loaded', () => {
    vi.mocked(useInfiniteQuery).mockReturnValue({
      isPending: false,
      data: [mockEvent()],
      error: null,
      hasNextPage: false,
      isFetchingNextPage: false,
      fetchNextPage: vi.fn(),
    } as unknown as UseInfiniteQueryResult);
    const { container } = render(<NotificationList onNotificationClick={() => {}} />);
    expect(container.querySelectorAll('[data-slot="item"]')).toHaveLength(1);
  });

  it('renders error alert when an error occurs', () => {
    vi.mocked(useInfiniteQuery).mockReturnValue({
      isPending: false,
      data: undefined,
      error: new Error('Failed to load notifications'),
      hasNextPage: false,
      isFetchingNextPage: false,
      fetchNextPage: vi.fn(),
    } as unknown as UseInfiniteQueryResult);
    const { container } = render(<NotificationList onNotificationClick={() => {}} />);
    expect(container.querySelector('[data-slot="alert"]')).not.toBeNull();
  });
});
