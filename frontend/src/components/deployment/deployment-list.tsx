import type { Deployment } from '@/api/api';
import { getDeploymentsQueryOptions } from '@/hooks';
import { useInfiniteQuery } from '@tanstack/react-query';
import { useEffect, type ReactNode } from 'react';
import { useInView } from 'react-intersection-observer';
import { DeploymentItemSkeleton, DeploymentListItem } from '.';
import { ScrollArea } from '../ui/scroll-area';
import { Spinner } from '../ui/spinner';
import { ErrorAlert } from '../view';

export function DeploymentList({
  selectedDeployment,
  onSelect,
  className,
}: {
  selectedDeployment?: string;
  onSelect: (item: Deployment) => void;
  className?: string;
}) {
  const { ref, inView } = useInView();
  const {
    data: deployments,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isPending,
    error,
  } = useInfiniteQuery(getDeploymentsQueryOptions());

  useEffect(() => {
    if (inView && hasNextPage && !isFetchingNextPage && deployments?.length) {
      fetchNextPage();
    }
  }, [inView, hasNextPage, isFetchingNextPage, fetchNextPage, deployments?.length]);

  if (isPending) {
    return DeploymentListSkeleton(className);
  }

  return (
    <DeploymentListContainer className={className}>
      <ErrorAlert title={error && 'ALERT.LOAD_DEPLOYMENTS_ERROR'} details={error?.message} />
      {deployments?.map((deployment) => (
        <DeploymentListItem
          key={deployment.id}
          deployment={deployment}
          isSelected={deployment.id === selectedDeployment}
          onSelect={onSelect}
        ></DeploymentListItem>
      ))}
      <div ref={ref} className="flex justify-around w-full min-h-1">
        {(isFetchingNextPage || (hasNextPage && deployments?.length)) && <Spinner />}
      </div>
    </DeploymentListContainer>
  );
}
function DeploymentListSkeleton(className?: string) {
  return (
    <DeploymentListContainer className={className}>
      {Array.from({ length: 5 }, (_, index) => (
        <DeploymentItemSkeleton key={index} />
      ))}
    </DeploymentListContainer>
  );
}

function DeploymentListContainer({
  className,
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return (
    <ScrollArea className={className}>
      <div className="flex flex-col gap-2 p-3">{children}</div>
    </ScrollArea>
  );
}
