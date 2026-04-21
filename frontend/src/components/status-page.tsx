import { getStatusQueryOptions, useFilteredQuery } from '@/hooks';
import { useTranslation } from 'react-i18next';
import { ServiceStatus, ServiceStatusSkeleton } from './status';
import { ScrollArea } from './ui/scroll-area';
import { ErrorAlert, HeaderLayout, InfoEmpty } from './view';

export function StatusPage() {
  const { t } = useTranslation();
  const { data, isPending, error } = useFilteredQuery(getStatusQueryOptions());

  return (
    <HeaderLayout header={<h2 className="text-2xl font-bold">{t('STATUS.STATUS')}</h2>}>
      <ScrollArea className="space-y-2 min-h-0 flex-1">
        <ErrorAlert title={error && 'ALERT.LOAD_STATUS_ERROR'} details={error?.message} />

        {isPending ? (
          Array(3)
            .fill({})
            .map((_, index) => <ServiceStatusSkeleton key={'status-skeleton-' + index} />)
        ) : data?.length ? (
          data.map((stackStatus) => (
            <ServiceStatus
              key={stackStatus.stackId}
              serviceName={stackStatus.name}
              serviceContainers={stackStatus.services}
              className="m-3"
            />
          ))
        ) : (
          <InfoEmpty
            title="STATUS.NO_STACKS_FOUND"
            details="STATUS.NO_STACKS_FOUND_DESCRIPTION"
          ></InfoEmpty>
        )}
      </ScrollArea>
    </HeaderLayout>
  );
}
