import { EventType } from '@/api/api';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Field, FieldDescription, FieldError, FieldGroup, FieldTitle } from '../ui/field';
import { GroupedCheckboxes, type OptionGroup } from '../ui/grouped-checkboxes';
import { Input } from '../ui/input';
import { Switch } from '../ui/switch';
import type { NotificationFormValues } from './onboarding-schema';

const notificaitonOptions = [
  {
    group: 'EVENT_TYPE.DEPLOYMENT',
    items: [
      { value: EventType.DEPLOYMENT_STARTED, label: 'EVENT_TYPE.DEPLOYMENT_STARTED' },
      { value: EventType.DEPLOYMENT_SUCCESS, label: 'EVENT_TYPE.DEPLOYMENT_SUCCESS' },
      { value: EventType.DEPLOYMENT_ERROR, label: 'EVENT_TYPE.DEPLOYMENT_ERROR' },
    ],
  },
  {
    group: 'EVENT_TYPE.GENERAL',
    items: [
      { value: EventType.ERROR, label: 'EVENT_TYPE.ERROR' },
      { value: EventType.PASSWORD_UPDATED, label: 'EVENT_TYPE.PASSWORD_UPDATED' },
      { value: EventType.CONFIGURATION_UPDATED, label: 'EVENT_TYPE.CONFIGURATION_UPDATED' },
      { value: EventType.SESSION_REUSED, label: 'EVENT_TYPE.SESSION_REUSED' },
    ],
  },
];

export function NotificationForm({ form }: { form: UseFormReturn<NotificationFormValues> }) {
  const { t } = useTranslation();
  return (
    <form>
      <p className="my-4 text-center text-primary/70">
        {t('ONBOARDING.FORM.NOTIFICATION_FORM_DESCRIPTION')}
      </p>
      <FieldGroup className="mt-2">
        <Controller
          name="enableNotifications"
          control={form.control}
          render={({ field }) => (
            <Field orientation="horizontal">
              <Switch checked={field.value} onCheckedChange={field.onChange} />
              <FieldTitle>{t('ONBOARDING.FORM.enableNotifications')}</FieldTitle>
            </Field>
          )}
        />
        {form.watch('enableNotifications') && (
          <>
            <NotificationField form={form} name="notificationURL" />
            <NotificationMultiSelect
              form={form}
              withDescription
              options={notificaitonOptions}
            ></NotificationMultiSelect>
          </>
        )}
      </FieldGroup>
    </form>
  );
}

function NotificationField({
  form,
  name,
  withDescription = false,
  withPlaceholder = true,
}: {
  form: UseFormReturn<NotificationFormValues>;
  name: keyof Omit<NotificationFormValues, 'enableNotifications'>;
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

function NotificationMultiSelect({
  form,
  withDescription = false,
  options,
}: {
  form: UseFormReturn<NotificationFormValues>;
  withDescription?: boolean;
  options: OptionGroup[];
}) {
  const { t } = useTranslation();
  const name = 'notificationTypes';
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
          <GroupedCheckboxes
            groups={options}
            value={field.value}
            onChange={field.onChange}
            variant="horizontal"
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
