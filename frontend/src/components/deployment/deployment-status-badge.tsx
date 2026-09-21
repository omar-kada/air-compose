import { useTranslation } from 'react-i18next';
import type { DeploymentStatus } from '@/api/api';
import { Badge } from '@/components/ui/badge';
import { colorForStatus, iconForStatus } from '@/lib';
import { cn } from '@/lib/utils';

export function DeploymentStatusBadge(props: {
  status: DeploymentStatus;
  iconOnly?: boolean;
  label?: string;
  className?: string;
}) {
  const { t } = useTranslation();
  const Icon = iconForStatus(props.status);
  const statusLabel = props.status
    ? t(`DEPLOYMENT_STATUS.${props.status.toUpperCase()}`)
    : 'unknown';
  return (
    <Badge className={cn(colorForStatus(props.status), props.className)}>
      <Icon className={`${props.status === 'running' ? 'animate-spin' : ''}`} />
      {!props.iconOnly && (props.label ?? statusLabel)}
    </Badge>
  );
}
