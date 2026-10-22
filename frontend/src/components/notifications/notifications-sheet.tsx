import { type Event } from '@/api/api';
import { Button } from '@/components/ui/button';
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';
import { getNotificationsQueryOptions, useResetUnreadCount } from '@/hooks';
import { useDeploymentNavigate } from '@/lib';
import { useInfiniteQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { ScrollArea } from '../ui/scroll-area';
import { Separator } from '../ui/separator';
import { NotificationList } from './notifications-list';

export function NotificationSheet({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const depNavigate = useDeploymentNavigate();
  const [open, setOpen] = useState(false);
  const handleNotificationClick = useCallback(
    (event: Event) => {
      setOpen(false);
      if (event.objectId) {
        depNavigate(event.objectId);
      }
    },
    [setOpen, depNavigate],
  );

  const { data: notifications } = useInfiniteQuery(getNotificationsQueryOptions());

  const resetNotifCount = useResetUnreadCount();
  useEffect(() => {
    if (!open && notifications?.length) {
      resetNotifCount();
    }
  }, [open]);
  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>{children}</SheetTrigger>
      <SheetContent
        className="w-full md:w-none flex flex-col h-full"
        aria-describedby={t('NOTIFICATIONS.DESCRIPTION')}
      >
        <SheetHeader>
          <SheetTitle>{t('NOTIFICATIONS.NOTIFICATIONS')}</SheetTitle>
          <div className="flex flex-nowrap justify-between items-center-safe">
            <SheetDescription>{t('NOTIFICATIONS.DESCRIPTION')}</SheetDescription>
          </div>
        </SheetHeader>
        <Separator></Separator>
        <ScrollArea className="h-1 flex-1 gap-2">
          <NotificationList onNotificationClick={handleNotificationClick} />
        </ScrollArea>
        <SheetFooter>
          <SheetClose asChild>
            <Button variant="outline">{t('ACTION.CLOSE')}</Button>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
