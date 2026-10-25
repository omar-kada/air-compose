import { Controller, type UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Field, FieldDescription, FieldError, FieldGroup, FieldTitle } from '../ui/field';
import { Input } from '../ui/input';
import { NotificationMultiSelect } from '../view';
import type { NotificationFormValues } from './onboarding-schema';

export function NotificationForm({ form }: { form: UseFormReturn<NotificationFormValues> }) {
  const { t } = useTranslation();
  return (
    <form>
      <p className="my-4 text-center text-primary/70">
        {t('ONBOARDING.FORM.NOTIFICATION_FORM_DESCRIPTION')}
      </p>
      <FieldGroup className="mt-2">
        <NotificationField form={form} name="notificationURL" />
        <NotificationMultiSelect
          name="notificationTypes"
          label="ONBOARDING.FORM.notificationTypes"
          form={form}
          withDescription
          variant="horizontal"
        ></NotificationMultiSelect>
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
  name: keyof NotificationFormValues;
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
          <FieldTitle id={`field-${field.name}`}>{t(`ONBOARDING.FORM.${name}`)}</FieldTitle>
          {withDescription && (
            <FieldDescription>{t(`ONBOARDING.FORM.${name}_DESCRIPTION`)}</FieldDescription>
          )}
          <Input
            {...field}
            aria-labelledby={`field-${field.name}`}
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
