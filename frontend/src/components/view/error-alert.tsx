import type { Error } from '@/api';
import { cn } from '@/lib';
import type { AxiosError } from 'axios';
import { AlertCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Alert, AlertDescription, AlertTitle } from '../ui/alert';

export function ErrorAlert({
  title,
  error,
  className,
}: {
  title: string;
  error: AxiosError<Error> | null;
  className?: string;
}) {
  const { t } = useTranslation();
  return (
    error && (
      <Alert variant="destructive" className={cn('w-auto', className)}>
        <AlertCircleIcon />
        <AlertTitle>{t(title)}</AlertTitle>
        {error.message && (
          <AlertDescription>{t(error.response?.data.message ?? error.message)}</AlertDescription>
        )}
      </Alert>
    )
  );
}
