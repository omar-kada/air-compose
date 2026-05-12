import type { FileDiff } from '@/api/api';
import { GitCompare } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { FileDiffView } from '.';
import { Badge } from '../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Skeleton } from '../ui/skeleton';

export function DeploymentDiff({
  fileDiffs,
  autoOpen,
  children,
}: {
  fileDiffs: FileDiff[];
  autoOpen?: boolean;
  children?: React.ReactNode;
}) {
  const { t } = useTranslation();
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex justify-between">
          <span className="flex flex-nowrap items-center-safe">
            <GitCompare className="size-5 mx-1" />
            {t('DIFF.UPDATED_FILES')}
            <Badge className="size-5 mx-1" variant="outline">
              {fileDiffs.length}
            </Badge>
          </span>
        </CardTitle>
        {children && <CardDescription className="font-light">{children}</CardDescription>}
      </CardHeader>
      <CardContent>
        {fileDiffs.map((fileDiff) => (
          <FileDiffView
            fileDiff={fileDiff}
            key={fileDiff.oldFile}
            autoOpen={autoOpen}
            className="mb-2"
          />
        ))}
      </CardContent>
    </Card>
  );
}
export function DeploymentDiffSkeleton() {
  return (
    <div className="flex flex-col gap-4">
      <Skeleton className="h-6 w-2/3 " />
      <Skeleton className="h-4 w-50" />
      <Skeleton className="h-30 w-full " />
    </div>
  );
}
