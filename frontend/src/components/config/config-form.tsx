import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib';
import { Plus } from 'lucide-react';
import { useCallback } from 'react';
import { useFieldArray, type UseFormReturn } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { type FormValues } from './config-form-schema';
import { EnvVarArrayForm } from './env-vars-array-form';
import { ServiceCard } from './service-card';

export function ConfigForm({
  form,
  className,
  disabled,
}: {
  form: UseFormReturn<FormValues>;
  className?: string;
  disabled?: boolean;
}) {
  const { t } = useTranslation();

  const {
    fields: servicesFields,
    append: appendService,
    remove: removeService,
  } = useFieldArray({
    control: form.control,
    name: 'services',
  });

  const addNewService = useCallback(
    () => appendService({ name: '', envVars: [{ key: '', value: '' }] }, { shouldFocus: true }),
    [appendService],
  );

  return (
    <form className={cn('space-y-4', className)}>
      <h2 className="text-xl">{t('CONFIGURATION.FORM.GLOBAL_ENV_VARS')}</h2>
      <Card>
        <CardContent>
          <EnvVarArrayForm
            control={form.control}
            name="globalEnvVars"
            disabled={disabled}
          ></EnvVarArrayForm>
        </CardContent>
      </Card>
      <h2 className="text-xl">{t('CONFIGURATION.FORM.SERVICES')}</h2>

      {servicesFields.map((service, serviceIndex) => (
        <ServiceCard
          key={service.id}
          service={service}
          name={`services.${serviceIndex}`}
          form={form}
          onRemove={() => removeService(serviceIndex)}
          disabled={disabled}
        />
      ))}
      {!disabled && (
        <Button
          variant="default"
          type="button"
          onClick={addNewService}
          aria-label={t('CONFIGURATION.FORM.ADD_SERVICE')}
        >
          <Plus /> {t('CONFIGURATION.FORM.ADD_SERVICE')}
        </Button>
      )}
    </form>
  );
}
