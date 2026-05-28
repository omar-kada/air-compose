import type { ContainerHealth, ContainerState } from '@/api/api';
import { Badge } from '@/components/ui/badge';
import { borderForStatus, iconForHealth, textColorForStatus } from '@/lib';
import { cn } from '@/lib/utils';
import { useTranslation } from 'react-i18next';

export function ContainerStatusBadge(props: {
  health?: ContainerHealth;
  state?: ContainerState;
  label?: string;
  className?: string;
  iconOnly?: boolean;
}) {
  const { t } = useTranslation();
  const Icon = iconForHealth(props.health);
  return (
    <Badge
      variant="outline"
      className={cn(borderForStatus(props.state ?? props.health), props.className)}
    >
      <Icon className={textColorForStatus(props.state ?? props.health)} />
      {!props.iconOnly && t(props.label ?? props.health ?? 'unknown')}
    </Badge>
  );
}
