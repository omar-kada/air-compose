import { ContainerState, EventType, type ContainerHealth, type DeploymentStatus } from '@/api/api';

export function colorForStatus(status: ContainerHealth | DeploymentStatus): string {
  switch (status) {
    case 'healthy':
    case 'success':
      return 'bg-green-400';
    case 'unhealthy':
    case 'error':
      return 'bg-red-400';
    case 'starting':
    case 'planned':
      return 'bg-slate-400';
    case 'running':
      return 'bg-blue-400';
    default:
      return '';
  }
}

export function borderForStatus(status?: ContainerState | ContainerHealth): string {
  switch (status) {
    case 'running':
    case 'healthy':
      return 'border-green-400';
    case 'dead':
    case 'removing':
    case 'unhealthy':
      return 'border-red-400';
    case 'exited':
    case 'paused':
    case 'none':
      return 'border-slate-400';
    case 'created':
    case 'restarting':
    case 'starting':
      return 'border-blue-400';
    default:
      return '';
  }
}
export function textColorForStatus(status?: ContainerHealth | ContainerState): string {
  switch (status) {
    case 'running':
    case 'healthy':
      return 'text-green-400';
    case 'dead':
    case 'removing':
    case 'unhealthy':
      return 'text-red-400';
    case 'exited':
    case 'paused':
    case 'none':
      return 'text-slate-400';
    case 'created':
    case 'restarting':
    case 'starting':
      return 'text-blue-400';
    default:
      return '';
  }
}

export function logColor(level: EventType): string {
  switch (level) {
    case EventType.ERROR:
    case EventType.DEPLOYMENT_ERROR:
      return 'text-red-700 dark:text-red-300 ';
    case EventType.MISC:
      return 'text-gray-700 dark:text-gray-300';
    default:
      return '';
  }
}
