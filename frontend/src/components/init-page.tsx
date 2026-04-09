import { getSettingsQueryOptions, getStateQueryOptions, useUpdateSettings } from '@/hooks';
import { ROUTES } from '@/lib';
import { useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { OnboardingForm } from './onboarding';
import { Card, CardContent } from './ui/card';
import { Skeleton } from './ui/skeleton';
import { ErrorAlert } from './view';

export function InitPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const {
    data: settings,
    isPending: isSettingsPending,
    error: settingsError,
  } = useQuery(getSettingsQueryOptions());

  const { data: state, isPending, error: stateError } = useQuery(getStateQueryOptions());
  useEffect(() => {
    if (!isPending && state?.initialized) {
      navigate(ROUTES.ROOT);
    }
  }, [state, isPending]);

  const { updateSettings, isPending: updateSettingsPending } = useUpdateSettings();
  useEffect(() => {
    console.log(' update tsettings pending : ', updateSettingsPending);
  }, [updateSettingsPending]);

  return (
    <div className="p-4 space-y-4 h-full flex items-center flex-col justify-center ">
      <h2 className="text-xl">{t('ONBOARDING.FORM.TITLE')}</h2>
      <Card className="min-w-full sm:min-w-lg md:min-w-2xl">
        <CardContent className="space-y-4 w-full">
          {isSettingsPending ? (
            <FormSkeleton />
          ) : (
            <>
              <ErrorAlert
                className="mx-4 mt-4"
                title={settingsError && 'ALERT.LOAD_SETTINGS_ERROR'}
                details={settingsError?.message}
              />
              {!settingsError && (
                <ErrorAlert
                  className="mx-4 mt-4"
                  title={stateError && 'ALERT.LOAD_SETTINGS_ERROR'}
                  details={stateError?.message}
                />
              )}
              {settings && <OnboardingForm settings={settings} onSubmit={updateSettings} />}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
function FormSkeleton() {
  return (
    <div className="flex flex-col space-y-3 m-4">
      <Skeleton className="h-6 mt-2 mb-4 w-2/3" />
      <div className="flex gap-2 mt-4">
        <Skeleton className="h-11 w-35" />
        <Skeleton className="h-11 w-35" />
      </div>
      <Skeleton className="h-30 w-full rounded-lg" />
      <Skeleton className="h-30 w-full rounded-lg" />
    </div>
  );
}
