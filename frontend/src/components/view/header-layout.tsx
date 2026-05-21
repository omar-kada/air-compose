import { type ReactNode } from 'react';
import { Separator } from '../ui/separator';

export function HeaderLayout({ header, children }: { header: ReactNode; children: ReactNode }) {
  return (
    <div className="flex flex-col h-full">
      {header && <div className="py-2">{header}</div>}
      {header && <Separator orientation="horizontal" />}
      <div className="flex flex-col flex-1 overflow-hidden">{children}</div>
    </div>
  );
}
