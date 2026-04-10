import { EventType } from '@/api/api';
import { Controller, type FieldValues, type Path, type UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Field, FieldDescription, FieldError, FieldTitle } from '../ui/field';
import { GroupedCheckboxes, type GroupedCheckboxesVariant } from '../ui/grouped-checkboxes';

const notificationOptions = [
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
      { value: EventType.CONFIGURATION_UPDATED, label: 'EVENT_TYPE.CONFIGURATION_UPDATED' },
      { value: EventType.PASSWORD_UPDATED, label: 'EVENT_TYPE.PASSWORD_UPDATED' },
      { value: EventType.SESSION_REUSED, label: 'EVENT_TYPE.SESSION_REUSED' },
    ],
  },
];

export function NotificationMultiSelect<T extends FieldValues>({
  name,
  form,
  withDescription = false,
  variant,
  label,
}: {
  name: Path<T>;
  form: UseFormReturn<T>;
  withDescription?: boolean;
  label: string;
  variant?: GroupedCheckboxesVariant;
}) {
  const { t } = useTranslation();
  return (
    <Controller
      name={name}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid}>
          <FieldTitle>{t(label)}</FieldTitle>
          {withDescription && <FieldDescription>{t(`${label}_DESCRIPTION`)}</FieldDescription>}
          <GroupedCheckboxes
            groups={notificationOptions}
            value={field.value}
            onChange={field.onChange}
            variant={variant}
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
