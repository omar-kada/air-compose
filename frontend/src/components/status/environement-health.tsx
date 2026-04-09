import { getStateQueryOptions, useFilteredQuery, useUser } from '@/hooks';
import { cn, ROUTES } from '@/lib';
import { Link } from 'react-router-dom';
import { ContainerStatusBadge } from '.';
import { Skeleton } from '../ui/skeleton';

export function EnvironementHealth({ className }: { className?: string }) {
  const { data: user } = useUser();
  const { data: state, isPending } = useFilteredQuery(getStateQueryOptions({ enabled: !!user }));

  if (isPending) {
    return <Skeleton className="h-4 w-20 my-auto" />;
  }

  return (
    <Link
      to={ROUTES.STATUS}
      className={cn('flex flex-wrap items-center align-bottom gap-4 m-4 ', className)}
    >
      <ContainerStatusBadge status={state?.health}></ContainerStatusBadge>
    </Link>
  );
}
