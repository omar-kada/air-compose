import type { User } from '@/api/api';
import { useLogout, useUnreadNotificationCount, useUser } from '@/hooks';
import { useTheme } from '@/hooks/theme-provider';
import { cn } from '@/lib';
import { Bell, LogOutIcon, Moon, Settings } from 'lucide-react';
import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { NotificationSheet } from './notifications';
import { SettingsSheet } from './settings';
import { Avatar, AvatarFallback } from './ui/avatar';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu';
import { Field, FieldLabel } from './ui/field';
import { Switch } from './ui/switch';

const gotoHome = () => (window.location.href = '/');

export function Topbar({ children, className }: { children?: ReactNode; className?: string }) {
  const { data: user } = useUser();
  const unreadCount = useUnreadNotificationCount();

  return (
    <header className="h-14 min-h-14 border-b w-full flex justify-around bg-sidebar sticky top-0 z-50">
      <div className={cn('h-14 min-h-14 flex items-center justify-between flex-1', className)}>
        {/* Logo */}
        <div className="flex items-center gap-2 me-4 cursor-pointer" onClick={gotoHome}>
          <img src="/air-compose.svg" alt="AirCompose Logo" className="h-12 w-12 flex-0" />
          <div className="text-xl font-semibold mr-5 select-none">AirCompose</div>
        </div>
        <div className="flex-1 w-1 max-w-10">{/*gap*/}</div>
        <div className="flex flex-2 justify-between">{children}</div>

        {user && (
          <div className="flex items-center gap-2">
            <NotificationSheet>
              <Button variant="ghost" size="icon" className="relative rounded-full">
                <Bell className="h-4 w-4" />
                {unreadCount > 0 && (
                  <Badge
                    variant="destructive"
                    className="absolute -top-1 -right-1 h-4 w-4 rounded-full p-0 flex items-center justify-center text-xs"
                  >
                    {unreadCount}
                  </Badge>
                )}{' '}
              </Button>
            </NotificationSheet>
            <UserDropDown user={user} />
          </div>
        )}
      </div>
    </header>
  );
}

function UserDropDown({ user }: { user: User }) {
  const { t } = useTranslation();

  const [initial, setInitial] = useState(user?.username ?? '');

  useEffect(() => {
    if (user) {
      setInitial(user.username.charAt(0).toUpperCase());
    }
  }, [user]);

  const { theme, setTheme } = useTheme();

  const toggleTheme = useCallback(() => {
    setTheme(theme === 'dark' ? 'light' : 'dark');
  }, [theme, setTheme]);

  const [openSettings, setOpenSettings] = useState(false);
  const openSettingsSheet = useCallback(() => setOpenSettings(true), [setOpenSettings]);

  const { logout } = useLogout();
  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="rounded-full cursor-pointer">
            <Avatar>
              <AvatarFallback className="select-none">{initial}</AvatarFallback>
            </Avatar>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuGroup>
            <DropdownMenuLabel>
              {t('MENU.LOGGED_AS')} : {user?.username ?? ''}
            </DropdownMenuLabel>

            <DropdownMenuItem onClick={toggleTheme}>
              <Field className="" orientation="horizontal">
                <Moon></Moon>
                <FieldLabel className="font-normal pe-2">{t('MENU.DARK_MODE')}</FieldLabel>
                <Switch checked={theme === 'dark'} onCheckedChange={toggleTheme} />
              </Field>
            </DropdownMenuItem>

            <DropdownMenuItem onSelect={openSettingsSheet}>
              <Settings />
              {t('MENU.SETTINGS')}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem onSelect={logout}>
            <LogOutIcon />
            {t('ACTION.SIGN_OUT')}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <SettingsSheet open={openSettings} setOpen={setOpenSettings}></SettingsSheet>
    </>
  );
}
