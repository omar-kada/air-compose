import { DeploymentStatus } from '@/api/api';
import { getDeploymentOptions, getDeploymentsQueryOptions, useFilteredQuery } from '@/hooks';
import { cn, ROUTES } from '@/lib';
import { useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { DeploymentDiff, DeploymentEventLog, DeploymentStatusBadge } from '.';
import { ScrollArea } from '../ui/scroll-area';
import { Skeleton } from '../ui/skeleton';
import { Spinner } from '../ui/spinner';
import { ErrorAlert, HumanTime } from '../view';

export function DeploymentDetail({ id, className }: { id: string; className?: string }) {
  const { t } = useTranslation();
  const {
    data: deployment,
    error,
    isPending,
    isFetching,
    refetch,
  } = useFilteredQuery(getDeploymentOptions(id));
  const queryClient = useQueryClient();

  useEffect(() => {
    if (deployment?.status === DeploymentStatus.running) {
      setTimeout(() => {
        refetch();
        queryClient.refetchQueries(getDeploymentsQueryOptions());
      }, 1000);
    }
  }, [deployment]);

  if (isPending) {
    return <DeploymentDetailSkeleton />;
  }
  return (
    <div className={cn('flex flex-col', className)}>
      <ErrorAlert
        className="mx-4 mt-4"
        title={error && 'ALERT.LOAD_DEPLOYMENT_ERROR'}
        details={error?.message}
      />
      {deployment && (
        <>
          <div className="flex justify-between items-center-safe m-4">
            <div className="text-2xl font-semibold ">
              <Link to={ROUTES.DEPLOYMENT(id)}>#{id} - </Link>
              {deployment.title}
              <DeploymentStatusBadge
                status={deployment.status}
                className="mx-3"
              ></DeploymentStatusBadge>
            </div>
            {isFetching && <Spinner className="size-6" />}
          </div>
          <ScrollArea className="gap-4 h-1 flex-1">
            <div className="flex flex-col gap-4 mx-4">
              <div className="flex flex-wrap gap-2 text-muted-foreground text-sm">
                <InfoItem
                  label={t('DEPLOYMENTS.AUTHOR')}
                  value={deployment.author !== '' ? deployment.author : t('DEPLOYMENTS.AUTOMATIC')}
                />
                -<HumanTime time={deployment.time}></HumanTime>
              </div>
              <DeploymentDiff fileDiffs={deployment.files ?? []}>
                <InfoItem
                  label={t('DEPLOYMENTS.REPO')}
                  value={`${deployment.repo} ${deployment.branch ? `(${deployment.branch})` : ''}`}
                />
              </DeploymentDiff>
              <DeploymentEventLog events={deployment.events ?? []} />
            </div>
          </ScrollArea>
        </>
      )}
    </div>
  );
}

function InfoItem({
  label,
  value,
  className,
}: {
  label?: string;
  value: string;
  className?: string;
}) {
  return (
    <span className={cn('text-sm font-light', className)}>
      {label} {value}
    </span>
  );
}

export function DeploymentDetailSkeleton() {
  return (
    <div className="flex flex-col space-y-3 m-4">
      <Skeleton className="h-6 mt-2 mb-4 w-2/3" />
      <div className="flex gap-2 mt-4">
        <Skeleton className="h-11 w-35" />
        <Skeleton className="h-11 w-35" />
      </div>

      <Skeleton className="h-30 w-full rounded-lg" />
      <Skeleton className="h-30 w-full rounded-lg" />
    </div>
  );
}
