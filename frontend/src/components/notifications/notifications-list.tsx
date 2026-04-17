import { type Event } from '@/api/api';
import { getNotificationsQueryOptions, useRelativeTime } from '@/hooks';
import { useInfiniteQuery } from '@tanstack/react-query';
import { Fragment, useCallback, useEffect, useState, type MouseEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { useInView } from 'react-intersection-observer';
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemSeparator,
  ItemTitle,
} from '../ui/item';
import { Skeleton } from '../ui/skeleton';
import { Spinner } from '../ui/spinner';
import { ErrorAlert } from '../view';
import { NotificationBadge } from './notification-badge';

export function NotificationList({
  selectedTypes,
  onNotificationClick,
}: {
  selectedTypes: Array<string>;
  onNotificationClick: (notif: Event) => void;
}) {
  const {
    data: notifications,
    isPending,
    error,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useInfiniteQuery(getNotificationsQueryOptions());
  const { ref, inView } = useInView();
  useEffect(() => {
    if (inView && hasNextPage && !isFetchingNextPage && notifications?.length) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage, notifications?.length]);

  const [filteredNotifications, setFilteredNotifications] = useState(notifications);
  useEffect(() => {
    setFilteredNotifications(notifications?.filter((notif) => selectedTypes.includes(notif.type)));
  }, [notifications, selectedTypes, setFilteredNotifications]);

  return (
    <>
      {isPending && <NotificationSkeleton />}
      {error && <ErrorAlert title="ALERT.LOAD_NOTIFICATIONS_ERROR" details={error.message} />}
      {filteredNotifications && (
        <ItemGroup className="grid auto-rows-min px-4 mb-10">
          {filteredNotifications.map((notification) => (
            <Fragment key={notification.ID}>
              <Notification notification={notification} onClick={onNotificationClick} />
              <ItemSeparator></ItemSeparator>
            </Fragment>
          ))}
        </ItemGroup>
      )}
      <div ref={ref} className="flex justify-around w-full min-h-1">
        {(isFetchingNextPage || (hasNextPage && notifications?.length)) && <Spinner />}
      </div>
    </>
  );
}

function Notification({
  notification,
  onClick,
}: {
  notification: Event;
  onClick: (notif: Event) => void;
}) {
  const { t } = useTranslation();
  const relativeTime = useRelativeTime(notification.time);
  const handleClick = useCallback(
    (e: MouseEvent) => {
      e.preventDefault();
      onClick(notification);
    },
    [onClick, notification],
  );
  return (
    <Item asChild>
      <a href="" onClick={handleClick}>
        <ItemMedia variant="default">
          <NotificationBadge type={notification.type} />
        </ItemMedia>
        <ItemContent>
          <ItemTitle>{t(`EVENT_TYPE.${notification.type}`, { ...notification })}</ItemTitle>
          <ItemDescription>
            {getNotificationObjectTitle(notification)}

            {notification.msg !== '' && `: ${notification.msg}`}
          </ItemDescription>
          <div className="text-xs text-muted-foreground self-end-safe">{relativeTime}</div>
        </ItemContent>
      </a>
    </Item>
  );
}

function getNotificationObjectTitle(notif: Event): string {
  if (notif.objectId && notif.objectName) {
    return `#${notif.objectId} "${notif.objectName}"`;
  }
  return '';
}

export function NotificationSkeleton() {
  return (
    <div className="flex flex-col space-y-3 m-4">
      <Skeleton className="h-15 w-full rounded-lg" />
      <Skeleton className="h-15 w-full rounded-lg" />
      <Skeleton className="h-15 w-full rounded-lg" />
    </div>
  );
}
