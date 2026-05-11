import { Separator } from '@/components/ui/separator';

export function TextSeparator({ text }: { text: string }) {
  return (
    <div className="flex items-center">
      <Separator className="flex-1" />
      <span className="flex-0 px-2 text-xs text-muted-foreground text-nowrap">{text}</span>
      <Separator className="flex-1" />
    </div>
  );
}
