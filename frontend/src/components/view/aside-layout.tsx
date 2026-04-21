import { type ReactNode } from 'react';
import { Separator } from '../ui/separator';
import { HeaderLayout } from './header-layout';

export function AsideLayout({
  header,
  aside,
  children,
  focusMain,
}: {
  header: ReactNode;
  aside?: ReactNode;
  children: ReactNode;
  focusMain?: boolean;
}) {
  return (
    <HeaderLayout header={header}>
      {/* Sidebar (hidden on mobile if main is focused) */}
      <div className="w-full flex h-full">
        {aside && (
          <aside
            className={`w-full h-full max-h-full flex flex-col sm:w-75 sm:shrink-0 ${focusMain ? 'hidden sm:flex' : ''}`}
          >
            {aside}
          </aside>
        )}
        {aside && <Separator orientation="vertical" />}

        {/* Main content (hidden on mobile until main is focused) */}
        <main className={`${!focusMain ? 'hidden sm:block sm:flex-1' : 'w-full'}`}>{children}</main>
      </div>
    </HeaderLayout>
  );
}
