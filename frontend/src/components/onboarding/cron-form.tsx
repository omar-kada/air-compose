import { Ban, Calendar, Clock, Timer } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Field, FieldDescription, FieldError, FieldGroup, FieldTitle } from '../ui/field';
import { Input } from '../ui/input';
import { Toggle } from '../ui/toggle';
import type { CronFormValues } from './onboarding-schema';

const predefinedCrons = [
  { label: 'ONBOARDING.FORM.CRON_DISABLED', value: '', icon: <Ban className="size-7" /> },
  {
    label: 'ONBOARDING.FORM.CRON_EVERY_10_MIN',
    value: '*/10 * * * *',
    icon: <Timer className="size-7" />,
  },
  {
    label: 'ONBOARDING.FORM.CRON_HOURLY',
    value: '0 * * * *',
    icon: <Clock className="size-7" />,
  },
  {
    label: 'ONBOARDING.FORM.CRON_DAILY',
    value: '0 0 * * *',
    icon: <Calendar className="size-7" />,
  },
];

export function CronForm({ form }: { form: UseFormReturn<CronFormValues> }) {
  const { t } = useTranslation();
  const cronValue = form.watch('cron');
  const [activeIdx, setActiveIdx] = useState<number | null>(
    predefinedCrons.findIndex((c) => c.value === cronValue),
  );

  // Keep toggle active if cron matches, deactivate if user changes
  useEffect(() => {
    const idx = predefinedCrons.findIndex((c) => c.value === cronValue);
    setActiveIdx(idx >= 0 ? idx : null);
  }, [cronValue]);

  return (
    <form>
      <p className="my-4 text-center text-primary/70">
        {t('ONBOARDING.FORM.CRON_FORM_DESCRIPTION')}
      </p>
      <FieldGroup className="mt-2">
        <div className="grid grid-cols-4 gap-2 mb-2">
          {predefinedCrons.map((cron, idx) => (
            <Toggle
              variant="outline"
              key={cron.value}
              pressed={activeIdx === idx}
              onPressedChange={(pressed) => {
                if (pressed) {
                  form.setValue('cron', cron.value);
                }
              }}
              className="flex flex-col h-20"
            >
              {cron.icon}
              {t(cron.label)}
            </Toggle>
          ))}
        </div>
        <CronField form={form} name="cron" withDescription withPlaceholder />
      </FieldGroup>
    </form>
  );
}

function CronField({
  form,
  name,
  withDescription = false,
  withPlaceholder = true,
}: {
  form: UseFormReturn<CronFormValues>;
  name: keyof Omit<CronFormValues, 'enableAutosync'>;
  withDescription?: boolean;
  withPlaceholder?: boolean;
}) {
  const { t } = useTranslation();

  return (
    <Controller
      name={name}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid}>
          <FieldTitle>{t(`ONBOARDING.FORM.${name}`)}</FieldTitle>
          {withDescription && (
            <FieldDescription>{t(`ONBOARDING.FORM.${name}_DESCRIPTION`)}</FieldDescription>
          )}

          <Input
            {...field}
            aria-invalid={fieldState.invalid}
            autoComplete="off"
            placeholder={withPlaceholder ? t(`ONBOARDING.FORM.${name}_PLACEHOLDER`) : ''}
          />
          {fieldState.invalid && (
            <FieldError
              errors={[{ ...fieldState.error, message: t(fieldState.error?.message ?? '') }]}
            />
          )}
        </Field>
      )}
    />
  );
}
