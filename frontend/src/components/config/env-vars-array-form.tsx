import { Equal, Trash2 } from 'lucide-react';
import { Fragment, useEffect } from 'react';
import { Controller, useFieldArray, useWatch, type Control } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Field, FieldError } from '../ui/field';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupTextarea,
} from '../ui/input-group';
import { type FormValues } from './config-form-schema';

export function EnvVarArrayForm({
  control,
  name,
  disabled,
}: {
  control: Control<FormValues, unknown, FormValues>;
  name: 'globalEnvVars' | `services.${number}.envVars`;
  disabled?: boolean;
}) {
  const { t } = useTranslation();
  const {
    fields: varArray,
    append,
    remove,
  } = useFieldArray({
    control,
    name,
  });

  const formData = useWatch({ control, name });

  // add default empty line
  useEffect(() => {
    if (!formData) return;

    const last = formData.at(-1);

    if (last?.key && last.key.trim() !== '') {
      append({ key: '', value: '' });
    }
  }, [formData, append]);

  return (
    <div className="grid grid-cols-2 gap-2 max-w-150">
      {varArray.map((varField, index) => (
        <Fragment key={name + varField.id}>
          <Controller
            name={`${name}.${index}.key`}
            control={control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <InputGroup>
                  <InputGroupInput
                    {...field}
                    aria-invalid={fieldState.invalid}
                    autoComplete="off"
                    placeholder={t('CONFIGURATION.FORM.KEY')}
                  />
                  {!disabled && (
                    <InputGroupAddon align="inline-end">
                      {(index !== varArray.length - 1 || varField.key.trim() !== '') && (
                        <InputGroupButton
                          aria-label={t('CONFIGURATION.FORM.REMOVE_VAR')}
                          size="icon-xs"
                          disabled={disabled}
                          onClick={() => remove(index)}
                        >
                          <Trash2 />
                        </InputGroupButton>
                      )}
                    </InputGroupAddon>
                  )}
                </InputGroup>
                {fieldState.invalid && (
                  <FieldError
                    errors={[{ ...fieldState.error, message: t(fieldState.error?.message ?? '') }]}
                  />
                )}
              </Field>
            )}
          />
          <Controller
            name={`${name}.${index}.value`}
            control={control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <InputGroup className="min-h-9">
                  <InputGroupTextarea
                    {...field}
                    aria-invalid={fieldState.invalid}
                    autoComplete="off"
                    placeholder={t('CONFIGURATION.FORM.VALUE')}
                    rows={1}
                    className="min-h-none py-1"
                  />
                  <InputGroupAddon>
                    <Equal />
                  </InputGroupAddon>
                </InputGroup>
                {fieldState.invalid && (
                  <FieldError
                    errors={[{ ...fieldState.error, message: t(fieldState.error?.message ?? '') }]}
                  />
                )}
              </Field>
            )}
          />
        </Fragment>
      ))}
    </div>
  );
}
