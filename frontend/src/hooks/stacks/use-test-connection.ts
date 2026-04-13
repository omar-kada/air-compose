import { getSettingsAPITestGitConnectionMutationOptions, type GitCredentials } from '@/api/api';
import { useMutation } from '@tanstack/react-query';
import type { AxiosError } from 'axios';
import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

export const useTestConnection = () => {
  const { t } = useTranslation();

  const testConnectionMutation = useMutation(getSettingsAPITestGitConnectionMutationOptions());

  const handleTestConnection = useCallback(
    (credentials: GitCredentials) => {
      toast.promise(() => testConnectionMutation.mutateAsync({ data: credentials }), {
        loading: t('ALERT.TESTING_CONNECTION'),
        success: t('ALERT.TESTING_CONNECTION_SUCCESS'),
        error: (error: AxiosError<string>) => {
          return {
            message: t('ALERT.TESTING_CONNECTION_FAILED'),
            description: error?.response?.data ?? '',
          };
        },
      });
    },
    [testConnectionMutation.mutateAsync, t],
  );

  return {
    ...testConnectionMutation,
    testConnection: handleTestConnection,
  };
};
