import type { Deployment } from '@/api/api';
import { ChevronRight } from 'lucide-react';
import { useCallback } from 'react';
import { Item, ItemActions, ItemContent, ItemDescription, ItemTitle } from '../ui/item';
import { Skeleton } from '../ui/skeleton';
import { HumanTime } from '../view';
import { DeploymentStatusBadge } from './deployment-status-badge';

export function DeploymentListItem({
  deployment,
  isSelected,
  onSelect,
}: {
  deployment: Deployment;
  isSelected: boolean;
  onSelect: (deployment: Deployment) => void;
}) {
  const onDeploymentClick = useCallback(
    (deployment: Deployment) => () => onSelect(deployment),
    [onSelect],
  );

  return (
    <Item
      className={`cursor-pointer ${isSelected ? 'bg-sidebar' : ''}`}
      onClick={onDeploymentClick(deployment)}
      variant="outline"
    >
      <ItemContent>
        <ItemTitle>
          <DeploymentStatusBadge status={deployment.status} iconOnly />
          {deployment.title}
        </ItemTitle>
        <ItemDescription className="text-xs">
          #{deployment.id} - <HumanTime time={deployment.time} />
        </ItemDescription>
      </ItemContent>
      <ItemActions className="flex-col justify-between h-full">
        <ChevronRight />
      </ItemActions>
    </Item>
  );
}

export function DeploymentItemSkeleton() {
  return (
    <div className="flex flex-col gap-2 w-full border rounded-lg p-4">
      <Skeleton className="h-4 " />
      <Skeleton className="h-4 w-2/3" />
    </div>
  );
}
