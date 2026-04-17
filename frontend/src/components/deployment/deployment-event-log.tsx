import type { Event } from '@/api/api';
import { formatTime, logColor } from '@/lib';
import { useTranslation } from 'react-i18next';
import {
  Timeline,
  TimelineBody,
  TimelineHeader,
  TimelineIcon,
  TimelineItem,
  TimelineSeparator,
} from '../ui/timeline';

export function DeploymentEventLog({ events }: { events: Event[] }) {
  const { t, i18n } = useTranslation();
  return (
    <div className="px-2">
      <h3 className="mb-4 font-bold text-xl">{t('DEPLOYMENTS.EVENTS_LOG')}</h3>

      <Timeline color="secondary" orientation="vertical">
        {events.map((event) => (
          <TimelineItem className={`whitespace-pre-wrap pb-4 ${logColor(event.type)}`}>
            <TimelineHeader>
              <TimelineSeparator />
              <TimelineIcon className="h-3 w-3" />
            </TimelineHeader>
            <TimelineBody className="-translate-y-1.5">
              <div className="space-y-1">
                <span className="flex justify-between w-full">
                  <h3 className="text-base leading-none font-semibold">
                    {t(`EVENT_TYPE.${event.type}`)}
                  </h3>

                  <p className="text-muted-foreground text-xs">
                    {formatTime(event.time, i18n.language)}{' '}
                  </p>
                </span>
                {event.msg && <p className="font-light">{event.msg}</p>}
              </div>
            </TimelineBody>
          </TimelineItem>
        ))}
      </Timeline>
    </div>
  );
}
