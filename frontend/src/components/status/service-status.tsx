import type { ContainerStatus } from '@/api/api';
import { ServiceLogo } from '@/lib';
import { Item, ItemActions, ItemContent, ItemDescription, ItemMedia, ItemTitle } from '../ui/item';
import { Skeleton } from '../ui/skeleton';
import { HumanTime } from '../view/human-time';
import { ContainerStatusBadge } from './container-status-badge';

export function ServiceStatus({
  serviceName,
  serviceContainers,
  className,
}: {
  serviceName: string;
  serviceContainers: Array<ContainerStatus>;
  className?: string;
}) {
  const time = serviceContainers[0]?.startedAt;
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
        {serviceContainers.map((item) => (
          <ContainerStatusBadge
            status={item.health}
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
