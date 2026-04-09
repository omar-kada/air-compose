import {
  getSettingsAPIGetQueryKey,
  getSettingsAPISetMutationOptions,
  type Settings,
} from '@/api/api';
import { useMutation } from '@tanstack/react-query';
import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { getStateQueryOptions } from '../stacks';

export const getUpdateSettingsOptions = () => {
  return getSettingsAPISetMutationOptions({
    mutation: {
      onSuccess: (data, _, __, context) => {
        context.client.setQueryData(getSettingsAPIGetQueryKey(), data);
        context.client.refetchQueries(getStateQueryOptions());
      },
    },
  });
};

export const useUpdateSettings = () => {
  const { t } = useTranslation();

  const updateMutation = useMutation(getUpdateSettingsOptions());

  const handleSync = useCallback(
    (Settings: Settings) => {
      return toast.promise(() => updateMutation.mutateAsync({ data: Settings }), {
        success: t('ALERT.SUCCESS'),
        error: t('ALERT.ERROR'),
      });
    },
    [updateMutation.mutateAsync, t],
  );

  return {
    ...updateMutation,
    updateSettings: handleSync,
  };
};
