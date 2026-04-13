import { getSettingsAPITestGitConnectionMutationOptions, type GitCredentails } from '@/api/api';
import { useMutation } from '@tanstack/react-query';
import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

export const useTestConnection = () => {
  const { t } = useTranslation();

  const testConnectionMutation = useMutation(getSettingsAPITestGitConnectionMutationOptions());

  const handleTestConnection = useCallback(
    (credentials: GitCredentails) => {
      toast.promise(() => testConnectionMutation.mutateAsync({ data: credentials }), {
        loading: t('ALERT.TESTING_CONNECTION'),
        success: t('ALERT.TESTING_CONNECTION_SUCCESS'),
        error: (error) => {
          console.log(error);
          return {
            message: t(
              error === false
                ? 'ALERT.TESTING_CONNECTION_FAILED'
                : 'ALERT.TESTING_CONNECTION_ERROR',
            ),
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
