import { ContainerHealth, EventType, type Event } from '@/api/api';
import { getNotificaitonIcon } from './notification-icon';

export function NotificationBadge({ notification }: { notification: Event }) {
  const Icon = getNotificaitonIcon(notification.type);
  const color = getNotificationColor(notification);

  return (
    <div className={`p-2 rounded-full ${color}`}>
      <Icon className="h-4 w-4 text-secondary" />
    </div>
  );
}

function getNotificationColor(notification: Event): string {
  switch (notification.type) {
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
      return 'bg-yellow-500';
    case EventType.HEALTH_CHANGE:
      if (notification.msg.includes(ContainerHealth.unhealthy)) {
        return 'bg-destructive';
      } else if (notification.msg.includes(ContainerHealth.healthy)) {
        return 'bg-green-500';
      } else {
        return 'bg-yellow-500';
      }
    case EventType.SESSION_REUSED:
      return 'bg-purple-500';
    default:
      return 'bg-gray-500';
  }
}
