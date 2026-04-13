import { useTestConnection } from '@/hooks';
import { Cable } from 'lucide-react';
import { useCallback, useEffect } from 'react';
import { Controller, type UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { Button } from '../ui/button';
import { Field, FieldDescription, FieldError, FieldGroup, FieldTitle } from '../ui/field';
import { Input } from '../ui/input';
import { Spinner } from '../ui/spinner';
import { Switch } from '../ui/switch';
import { toGitCredentaials, type RepoFormValues } from './onboarding-schema';

export function RepoForm({ form }: { form: UseFormReturn<RepoFormValues> }) {
  const { t } = useTranslation();

  const { testConnection, isPending } = useTestConnection();

  const handleTestConnection = useCallback(() => {
    testConnection(toGitCredentaials(form.getValues()));
  }, [form]);

  useEffect(() => {
    if (!form.getValues('privateRepo')) {
      form.setValue('username', '');
      form.setValue('token', '');
    }
  }, [form.watch('privateRepo'), form]);

  return (
    <form>
      <p className="my-4 text-center text-primary/70">
        {t('ONBOARDING.FORM.REPO_FORM_DESCRIPTION')}
      </p>
      <FieldGroup className="mt-2">
        <RepoField form={form} name="repo" />
        <RepoField form={form} name="branch" />
        <div className="grid grid-cols-2">
          <Controller
            name="privateRepo"
            control={form.control}
            render={({ field }) => (
              <Field orientation="horizontal">
                <Switch checked={field.value} onCheckedChange={field.onChange} />
                <FieldTitle>{t('ONBOARDING.FORM.PRIVATE_REPO')}</FieldTitle>
              </Field>
            )}
          />
          <Button
            type="button"
            onClick={handleTestConnection}
            disabled={isPending || !form.watch('repo')}
          >
            {isPending ? <Spinner /> : <Cable />}
            {t('ONBOARDING.FORM.TEST_CONNECTION')}
          </Button>
        </div>
        {form.watch('privateRepo') && (
          <>
            <RepoField form={form} name="username" />
            <RepoField form={form} name="token" />
          </>
        )}
      </FieldGroup>
    </form>
  );
}

function RepoField({
  form,
  name,
  withDescription = false,
  withPlaceholder = true,
}: {
  form: UseFormReturn<RepoFormValues>;
  name: keyof Omit<RepoFormValues, 'privateRepo'>;
  withDescription?: boolean;
  withPlaceholder?: boolean;
}) {
  const { t } = useTranslation();
  return (
    <Controller
      name={name}
      control={form.control}
      render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid} orientation="horizontal">
          <FieldTitle className="w-40">{t(`SETTINGS.FORM.${name}`)}</FieldTitle>
          {withDescription && (
            <FieldDescription>{t(`SETTINGS.FORM.${name}_DESCRIPTION`)}</FieldDescription>
          )}
          <Input
            {...field}
            aria-invalid={fieldState.invalid}
            autoComplete="off"
            placeholder={withPlaceholder ? t(`SETTINGS.FORM.${name}_PLACEHOLDER`) : ''}
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
