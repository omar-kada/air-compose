import { EventType } from '@/api/api';
import {
  Activity,
  Ban,
  BellIcon,
  Check,
  Clock,
  Cog,
  FileTextIcon,
  Info,
  KeyRoundIcon,
  LogInIcon,
  Rocket,
  X,
  type LucideIcon,
} from 'lucide-react';

type EventGroup = 'MISC' | 'ERROR' | 'DEPLOYMENT' | 'SETTINGS';

export function getNotificaitonIcon(type: EventType | EventGroup): LucideIcon {
  switch (type) {
    case EventType.MISC:
    case 'MISC':
      return Info;
    case EventType.ERROR:
    case 'ERROR':
      return Ban;
    case EventType.DEPLOYMENT_STARTED:
      return Clock;
    case EventType.DEPLOYMENT_SUCCESS:
      return Check;
    case EventType.DEPLOYMENT_ERROR:
      return X;
    case EventType.HEALTH_CHANGE:
      return Activity;
    case 'DEPLOYMENT':
      return Rocket;
    case EventType.PASSWORD_UPDATED:
      return KeyRoundIcon;
    case EventType.CONFIGURATION_UPDATED:
      return FileTextIcon;
    case EventType.SESSION_REUSED:
      return LogInIcon;
    case 'SETTINGS':
      return Cog;
    default:
      return BellIcon;
  }
}
