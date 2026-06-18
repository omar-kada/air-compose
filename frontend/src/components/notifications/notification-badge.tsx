import { EventType } from '@/api/api';
import { getNotificaitonIcon } from './notification-icon';

export function NotificationBadge({ type }: { type: EventType }) {
  const Icon = getNotificaitonIcon(type);
  const color = getNotificationColor(type);

  return (
    <div className={`p-2 rounded-full ${color}`}>
      <Icon className="h-4 w-4 text-secondary" />
    </div>
  );
}

function getNotificationColor(type: EventType): string {
  switch (type) {
    case EventType.ERROR:
    case EventType.DEPLOYMENT_ERROR:
      return 'bg-destructive';
    case EventType.MISC:
    case EventType.DEPLOYMENT_STARTED:
      return 'bg-blue-500';
    case EventType.DEPLOYMENT_SUCCESS:
      return 'bg-green-500';
    case EventType.PASSWORD_UPDATED:
    case EventType.CONFIGURATION_UPDATED:
    case EventType.HEALTH_CHANGE:
      return 'bg-yellow-500';
    case EventType.SESSION_REUSED:
      return 'bg-purple-500';
    default:
      return 'bg-gray-500';
  }
}
