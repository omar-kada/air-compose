import {
  getDiffQueryOptions,
  getStateQueryOptions,
  useFilteredQuery,
  useIsMobile,
  useSync,
} from '@/hooks';
import { cn } from '@/lib';
import { AlertCircleIcon, CloudSync, FileDiff, History, TriangleAlert } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import { Spinner } from '../ui/spinner';
import { HumanTime } from '../view';
import { DeploymentDiffDialog } from './deployment-diff-dialog';

export function DeploymentToolbar({ className }: { className?: string }) {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const { sync, error: syncError, isPending: isSyncLoading } = useSync();

  const { data: state, isPending, error } = useFilteredQuery(getStateQueryOptions());
  const {
    data: diffs,
    isFetching: isDiffsLoading,
    error: diffError,
  } = useFilteredQuery(getDiffQueryOptions());

  return (
    <div className={cn('flex flex-wrap items-center align-bottom gap-4', className)}>
      <div className="flex items-center p-2 gap-2">
        <span className="text-sm font-light mx-1 flex-1 flex gap-1 items-center">
          <History className="size-4"></History>
          {t('TIME.LAST_X_DAYS', { days: 30 })} :
        </span>
        {error && (
          <>
            <AlertCircleIcon className="size-4 text-destructive" />
            <span className="text-sm text-destructive">{t('ALERT.LOAD_STATE_ERROR')}</span>
          </>
        )}
      </div>

      <div className="flex flex-row items-center gap-1 justify-end-safe flex-1">
        <span className="text-sm font-light text-muted-foreground mr-2">
          {syncError
            ? syncError.message
            : !isMobile && (
                <>
                  {t('DEPLOYMENTS.AUTO_SYNC')} :&nbsp;
                  {error ? (
                    <AlertCircleIcon className="size-4 text-destructive inline" />
                  ) : isPending ? (
                    <Spinner className="inline"></Spinner>
                  ) : (
                    <HumanTime time={state?.nextDeploy} defaultValue={t('DISABLED')}></HumanTime>
                  )}
                </>
              )}
        </span>
        <DeploymentDiffDialog>
          <Button variant="outline">
            <FileDiff />
            {!isMobile && t('DIFF.DIFF')}
            {diffError ? (
              <AlertCircleIcon className="size-4 text-destructive inline" />
            ) : isDiffsLoading ? (
              <Spinner></Spinner>
            ) : (
              diffs != null && (
                <Badge
                  className="h-5 min-w-5 rounded-full px-1 font-mono tabular-nums"
                  variant={diffs.length > 0 ? 'default' : 'outline'}
                >
                  {diffs.length}
                </Badge>
              )
            )}
          </Button>
        </DeploymentDiffDialog>
        <Button variant="outline" onClick={sync} disabled={isSyncLoading}>
          {isSyncLoading ? <Spinner /> : <CloudSync />}
          {!isMobile && t('ACTION.SYNC_NOW')}
          {syncError ? <TriangleAlert className="text-destructive" /> : null}
        </Button>
      </div>
    </div>
  );
}
