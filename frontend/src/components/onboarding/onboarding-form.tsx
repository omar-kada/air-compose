import type { Settings } from '@/api/api';
import { Button } from '@/components/ui/button';
import {
  Stepper,
  StepperHeader,
  StepperIcon,
  StepperItem,
  StepperSeparator,
} from '@/components/ui/stepper';
import { cn } from '@/lib';
import { zodResolver } from '@hookform/resolvers/zod';
import { ArrowLeft, ArrowRight, Bell, Layers, RefreshCw, Save } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { CronForm } from './cron-form';
import { NotificationForm } from './notification-form';
import {
  cronSchema,
  notificationSchema,
  repoSchema,
  toCronFormValues,
  toNotificationFormValues,
  toRepoFormValues,
  toSettings,
  type CronFormValues,
  type NotificationFormValues,
  type RepoFormValues,
} from './onboarding-schema';
import { RepoForm } from './repo-form';

export function OnboardingForm({
  settings,
  onSubmit,
}: {
  settings: Settings;
  onSubmit: (data: Settings) => void;
}) {
  const { t } = useTranslation();
  const repoForm = useForm<RepoFormValues>({
    resolver: zodResolver(repoSchema),
    defaultValues: {
      repo: '',
    },
  });
  const cronForm = useForm<CronFormValues>({
    resolver: zodResolver(cronSchema),
    defaultValues: {},
  });
  const notificationForm = useForm<NotificationFormValues>({
    resolver: zodResolver(notificationSchema),
    defaultValues: {
      enableNotifications: false,
      notificationTypes: [],
    },
  });
  const forms = [repoForm, cronForm, notificationForm];
  useEffect(() => {
    repoForm.reset(toRepoFormValues(settings));
    cronForm.reset(toCronFormValues(settings));
    notificationForm.reset(toNotificationFormValues(settings));
  }, [settings, repoForm, cronForm, notificationForm]);

  const [step, setStep] = useState(0);

  const handleNextAndSumbit = useCallback(() => {
    if (step === 2) {
      onSubmit(
        toSettings({
          ...repoForm.getValues(),
          ...cronForm.getValues(),
          ...notificationForm.getValues(),
        }),
      );
    } else {
      setStep(step + 1);
    }
  }, [step, setStep, forms]);

  return (
    <div className="w-full">
      <Stepper value={step} onStepChange={setStep} className="relative flex items-center">
        <StepperItem value={0} disabled={step < 0} className="flex-1">
          <StepperHeader className="flex w-full items-center">
            <StepperIcon
              className={cn(
                'relative z-10 flex size-10 shrink-0 items-center justify-center rounded-full border-2 transition-colors',
                step === 0
                  ? 'border-primary bg-primary text-primary-foreground'
                  : step > 0
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-neutral-300 bg-neutral-100 text-neutral-400',
              )}
            >
              <Layers className="h-6 w-6" />
            </StepperIcon>
            <StepperSeparator
              className={cn(
                'mx-2 h-0.5 flex-1 transition-colors',
                step > 0 ? 'bg-primary' : 'bg-neutral-200',
              )}
            />
          </StepperHeader>
        </StepperItem>
        <StepperItem value={1} disabled={step < 1} className="flex-1">
          <StepperHeader className="flex w-full items-center">
            <StepperIcon
              className={cn(
                'relative z-10 flex size-10 shrink-0 items-center justify-center rounded-full border-2 transition-colors',
                step === 1
                  ? 'border-primary bg-primary text-primary-foreground'
                  : step > 1
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-neutral-300 bg-neutral-100 text-neutral-400',
              )}
            >
              <RefreshCw className="h-6 w-6" />
            </StepperIcon>
            <StepperSeparator
              className={cn(
                'mx-2 h-0.5 flex-1 transition-colors',
                step > 1 ? 'bg-primary' : 'bg-neutral-200',
              )}
            />
          </StepperHeader>
        </StepperItem>
        <StepperItem value={2} disabled={step < 2} className="w-fit">
          <StepperHeader className="flex w-full items-center">
            <StepperIcon
              className={cn(
                'relative z-10 flex size-10 shrink-0 items-center justify-center rounded-full border-2 transition-colors',
                step === 2
                  ? 'border-primary bg-primary text-primary-foreground'
                  : 'border-neutral-300 bg-neutral-100 text-neutral-400',
              )}
            >
              <Bell className="h-6 w-6" />
            </StepperIcon>
          </StepperHeader>
        </StepperItem>
      </Stepper>
      <div className="mt-4 min-h-110">
        {step === 0 && <RepoForm form={repoForm} />}
        {step === 1 && <CronForm form={cronForm} />}
        {step === 2 && <NotificationForm form={notificationForm} />}
      </div>

      <div className="mt-8 flex w-full justify-between gap-4">
        <Button disabled={step === 0} onClick={() => setStep(step - 1)}>
          <ArrowLeft />
          {t('ONBOARDING.FORM.PREVIOUS')}
        </Button>
        <Button disabled={!forms[step].formState.isValid} onClick={handleNextAndSumbit}>
          {step === 2 ? t('ONBOARDING.FORM.SUBMIT') : t('ONBOARDING.FORM.NEXT')}
          {step === 2 ? <Save /> : <ArrowRight />}
        </Button>
      </div>
    </div>
  );
}
