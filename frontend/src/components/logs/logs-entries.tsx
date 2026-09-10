import type { LogLine, LogMessages } from '@/api';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Fragment } from 'react/jsx-runtime';
import { Button } from '../ui/button';
import { AutoScrollArea } from './auto-scroll-area';
import { LEVEL_STYLES } from './log-levels';

export function LogEntries({ logs }: { logs: LogMessages }) {
  const { t } = useTranslation();
  const [expandAll, setExpandAll] = useState(false);
  return (
    <AutoScrollArea className="min-h-0 flex-1" watch={logs.length}>
      <ul className="font-mono text-xs">
        {logs.length === 0 ? (
          <li className="py-4 text-center text-muted-foreground text-sm">
            {t('LOGS.NO_LOGS_FILTERED')}
          </li>
        ) : (
          logs.map((line: LogLine) => (
            <LogEntry key={line.time + line.msg} line={line} expand={expandAll} />
          ))
        )}
      </ul>

      <div className="flex justify-between">
        <LiveIndicator />
        <Button
          size="xs"
          variant="ghost"
          className="text-muted-foreground"
          onClick={() => {
            setExpandAll(!expandAll);
          }}
        >
          {t('LOGS.EXPAND_ALL')}
        </Button>
      </div>
    </AutoScrollArea>
  );
}

function LogEntry({ line, expand }: { line: LogLine; expand?: boolean }) {
  const [expanded, setExpanded] = useState(false);
  useEffect(() => {
    setExpanded(expand ?? expanded);
  }, [expand]);

  return (
    <li className="flex flex-col  border-b border-border/50">
      <div
        className="flex items-start gap-2 py-1"
        onClick={(e) => {
          if (e.target === e.currentTarget) {
            setExpanded(!expanded);
          }
        }}
      >
        <span className="text-muted-foreground shrink-0 tabular-nums">
          {new Date(line.time).toLocaleTimeString()}
        </span>
        <Badge
          variant="outline"
          className={`shrink-0 min-w-10 h-4 justify-center rounded px-1.5 text-[0.6rem] font-semibold uppercase ${
            LEVEL_STYLES[line.level] ?? LEVEL_STYLES.DEBUG
          }`}
        >
          {line.level}
        </Badge>
        <span className="break-all">{line.msg}</span>
        <span className="flex-1" onClick={() => setExpanded(!expanded)}></span>
        {line.meta && Object.keys(line.meta).length > 0 && (
          <Button
            className="size-3 "
            size="icon-xs"
            variant="ghost"
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? <ChevronUp /> : <ChevronDown />}
          </Button>
        )}
      </div>
      {line.meta && Object.keys(line.meta).length > 0 && expanded && (
        <div
          className={cn(
            'text-xs text-muted-foreground my-1 flex flex-wrap',
            expanded || expand ? '' : 'overflow-hidden line-clamp-1',
          )}
        >
          {Object.entries(line.meta).map(([key, value], index) => (
            <Fragment key={line.time + key}>
              <span>
                {key}: {value}
              </span>
              {index + 1 < Object.keys(line.meta).length && <span>&nbsp;|&nbsp;</span>}
            </Fragment>
          ))}
        </div>
      )}
    </li>
  );
}

function LiveIndicator() {
  const { t } = useTranslation();
  return (
    <div className="flex items-center gap-1.5 py-2 text-[10px] text-muted-foreground">
      <span className="relative flex h-1.5 w-1.5">
        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-500 opacity-75" />
        <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-emerald-500" />
      </span>
      {t('LOGS.STREAMING')}
    </div>
  );
}
