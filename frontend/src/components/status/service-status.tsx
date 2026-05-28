import type { ContainerStatus } from '@/api/api';
import { ServiceLogo } from '@/lib';
import { useMemo } from 'react';
import { Item, ItemActions, ItemContent, ItemDescription, ItemMedia, ItemTitle } from '../ui/item';
import { Skeleton } from '../ui/skeleton';
import { HumanTime } from '../view';
import { ContainerStatusBadge } from './container-status-badge';

export function ServiceStatus({
  serviceName,
  serviceContainers,
  className,
}: {
  serviceName: string;
  serviceContainers: Record<string, ContainerStatus>;
  className?: string;
}) {
  const time = useMemo(() => {
    return new Date(
      Math.max(
        ...Object.values(serviceContainers).map((container) =>
          new Date(container.startedAt).getTime(),
        ),
      ),
    );
  }, [serviceContainers]);

  return (
    <Item variant="outline" className={className}>
      <ItemMedia>
        <ServiceLogo service={serviceName} />
      </ItemMedia>
      <ItemContent>
        <ItemTitle>{serviceName}</ItemTitle>
        <ItemDescription className="line-clamp-none">
          <HumanTime time={time} />
        </ItemDescription>
      </ItemContent>
      <ItemActions className="flex-wrap">
        {Object.entries(serviceContainers).map(([_, item]) => (
          <ContainerStatusBadge
            health={item.health}
            state={item.state}
            label={item.name}
            className="mx-1"
            key={`${serviceName}-${item.name}`}
          />
        ))}
      </ItemActions>
    </Item>
  );
}

export function ServiceStatusSkeleton() {
  return (
    <div className="flex flex-wrap items-center gap-4 border rounded-lg w-full p-4">
      <Skeleton className="h-12 w-12 rounded-full" />
      <div className="space-y-2">
        <Skeleton className="h-4 w-25" />
        <Skeleton className="h-2 w-20" />
      </div>
      <div className="flex-1"></div>
      <div className="gap-2 flex items-end-safe h-full">
        <Skeleton className="h-4 w-15" />
        <Skeleton className="h-4 w-15" />
        <Skeleton className="h-4 w-15" />
      </div>
    </div>
  );
}
