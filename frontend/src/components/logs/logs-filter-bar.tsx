import type { Level } from '@/api';
import { Input } from '@/components/ui/input';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { cn } from '@/lib';
import { useTranslation } from 'react-i18next';
import { LEVEL_STYLES, LEVELS } from './log-levels';

export function LogFilterBar({
  text,
  onTextChange,
  activeLevels,
  onLevelsChange,
  className,
}: {
  text: string;
  onTextChange: (value: string) => void;
  activeLevels: Set<Level>;
  onLevelsChange: (levels: Set<Level>) => void;
  className?: string;
}) {
  const { t } = useTranslation();

  return (
    <div className={cn('flex items-center gap-3 flex-wrap min-[400px]:flex-nowrap', className)}>
      <Input
        value={text}
        onChange={(e) => onTextChange(e.target.value)}
        aria-label={t('LOGS.FILTER_LOGS')}
        placeholder={t('LOGS.FILTER_LOGS')}
        className="max-w-50 min-w-30 h-8 text-sm"
      />
      <ToggleGroup
        type="multiple"
        value={[...activeLevels]}
        onValueChange={(next) => onLevelsChange(new Set(next as Level[]))}
        className="flex items-center"
        aria-label={t('LOGS.FILTER_LOG_LEVEL')}
      >
        {LEVELS.map((level) => (
          <ToggleGroupItem
            key={level}
            value={level}
            className={cn(
              'h-7 px-2 text-[11px] font-mono uppercase tracking-wide rounded border transition-colors data-[state=on]:bg-transparent',
              activeLevels.has(level) ? LEVEL_STYLES[level] : LEVEL_STYLES.DEBUG,
            )}
          >
            {level}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </div>
  );
}
