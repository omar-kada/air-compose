import { EventType } from '@/api/api';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ToggleGroup, ToggleGroupItem } from '../ui/toggle-group';
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip';
import { getNotificaitonIcon } from './notification-icon';

type EventFilter = 'ERROR' | 'DEPLOYMENT' | 'SETTINGS';

const filterMap: Map<EventFilter, Array<EventType>> = new Map([
  ['ERROR', [EventType.DEPLOYMENT_ERROR, EventType.ERROR]],
  [
    'DEPLOYMENT',
    [
      EventType.DEPLOYMENT_ERROR,
      EventType.DEPLOYMENT_STARTED,
      EventType.DEPLOYMENT_SUCCESS,
      EventType.STACKS_UNHEALTHY,
      EventType.STACKS_HEALTHY,
    ],
  ],
  [
    'SETTINGS',
    [EventType.CONFIGURATION_UPDATED, EventType.SESSION_REUSED, EventType.PASSWORD_UPDATED],
  ],
]);

export function NotificationFilter({
  onFilterChanged,
  className,
}: {
  onFilterChanged: (types: Array<EventType>) => void;
  className?: string;
}) {
  const { t } = useTranslation();

  const [selectedFilters, setSelectedFilters] = useState<Array<EventFilter>>([]);

  useEffect(() => {
    if (selectedFilters.length === 0) {
      onFilterChanged(Object.values(EventType));
    } else {
      onFilterChanged(selectedFilters.flatMap((filter) => filterMap.get(filter) ?? []));
    }
  }, [selectedFilters, onFilterChanged]);

  return (
    <ToggleGroup
      size="sm"
      type="multiple"
      value={selectedFilters}
      onValueChange={(newFilters) => setSelectedFilters(newFilters as EventFilter[])}
      className={className}
    >
      {Array.from(filterMap.keys()).map((filter) => (
        <Tooltip key={filter}>
          <TooltipTrigger asChild>
            <ToggleGroupItem
              value={filter}
              variant="outline"
              aria-describedby={t(`EVENT_TYPE.${filter}`)}
            >
              <GroupIcon filter={filter} />
            </ToggleGroupItem>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <p>{t(`EVENT_TYPE.${filter}`)}</p>
          </TooltipContent>
        </Tooltip>
      ))}
    </ToggleGroup>
  );
}

function GroupIcon({ filter }: { filter: EventFilter }) {
  const Icon = getNotificaitonIcon(filter);
  return <Icon></Icon>;
}
