// unread-notifications.ts
import { QueryClient, useQuery, useQueryClient } from '@tanstack/react-query';

export const UNREAD_NOTIFICATION_COUNT_KEY = ['unreadNotificationsCount'];

export function useUnreadNotificationCount() {
  const { data } = useQuery({
    queryKey: UNREAD_NOTIFICATION_COUNT_KEY,
    queryFn: () => 0, // never actually refetched, just gives an initial value
    initialData: 0,
    staleTime: Infinity,
  });
  return data;
}

export function incrementUnreadCount(queryClient: QueryClient) {
  queryClient.setQueryData<number>(UNREAD_NOTIFICATION_COUNT_KEY, (prev = 0) => prev + 1);
}

export function useResetUnreadCount() {
  const queryClient = useQueryClient();
  return () => {
    queryClient.setQueryData<number>(UNREAD_NOTIFICATION_COUNT_KEY, 0);
  };
}
